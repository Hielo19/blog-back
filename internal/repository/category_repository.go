package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gin-demo/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCategoryNotFound  = errors.New("分类不存在")
	ErrNoCategoryChanges = errors.New("没有需要更新的分类字段")
)

const categoryColumns = `id, name, slug, description, sort_order, created_at, updated_at`

type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	List(ctx context.Context) ([]model.Category, error)
	GetByID(ctx context.Context, id int64) (*model.Category, error)
	Update(ctx context.Context, id int64, changes model.CategoryChanges) (*model.Category, error)
	Delete(ctx context.Context, id int64) error
}

type PostgresCategoryRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCategoryRepository(db *pgxpool.Pool) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (r *PostgresCategoryRepository) Create(ctx context.Context, category *model.Category) error {
	query := `
INSERT INTO categories (name, slug, description, sort_order)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		category.Name,
		category.Slug,
		category.Description,
		category.SortOrder,
	).Scan(
		&category.ID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
}

func (r *PostgresCategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	query := fmt.Sprintf(`
SELECT %s
FROM categories
WHERE deleted_at IS NULL
ORDER BY sort_order ASC, id ASC`, categoryColumns)

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]model.Category, 0)
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *PostgresCategoryRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.Category, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM categories WHERE id = $1 AND deleted_at IS NULL",
		categoryColumns,
	)
	category, err := scanCategory(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCategoryNotFound
	}
	return category, err
}

func (r *PostgresCategoryRepository) Update(
	ctx context.Context,
	id int64,
	changes model.CategoryChanges,
) (*model.Category, error) {
	sets := make([]string, 0, 4)
	args := make([]any, 0, 5)
	add := func(column string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	if changes.Name != nil {
		add("name", *changes.Name)
	}
	if changes.Slug != nil {
		add("slug", *changes.Slug)
	}
	if changes.Description != nil {
		add("description", *changes.Description)
	}
	if changes.SortOrder != nil {
		add("sort_order", *changes.SortOrder)
	}
	if len(sets) == 0 {
		return nil, ErrNoCategoryChanges
	}

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE categories SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s",
		strings.Join(sets, ", "),
		len(args),
		categoryColumns,
	)

	category, err := scanCategory(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCategoryNotFound
	}
	return category, err
}

func (r *PostgresCategoryRepository) Delete(ctx context.Context, id int64) error {
	transaction, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	result, err := transaction.Exec(ctx, `
UPDATE categories
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	if _, err := transaction.Exec(ctx, `
UPDATE posts
SET category_id = NULL
WHERE category_id = $1 AND deleted_at IS NULL`, id); err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

func scanCategory(row rowScanner) (*model.Category, error) {
	var category model.Category
	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.SortOrder,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &category, nil
}
