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
	ErrPostNotFound = errors.New("文章不存在")
	ErrNoChanges    = errors.New("没有需要更新的字段")
)

const postColumns = `id, author_id, category_id, title, slug, summary, content,
content_format, cover_image_url, status, is_featured, published_at, created_at, updated_at`

type ListPostsParams struct {
	Status string
	Limit  int
	Offset int
}

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	List(ctx context.Context, params ListPostsParams) ([]model.Post, int64, error)
	GetByID(ctx context.Context, id int64) (*model.Post, error)
	Update(ctx context.Context, id int64, changes model.PostChanges) (*model.Post, error)
	Delete(ctx context.Context, id int64) error
}

type PostgresPostRepository struct {
	db *pgxpool.Pool
}

func NewPostgresPostRepository(db *pgxpool.Pool) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

func (r *PostgresPostRepository) Create(ctx context.Context, post *model.Post) error {
	query := `
INSERT INTO posts (
    author_id, category_id, title, slug, summary, content, content_format,
    cover_image_url, status, is_featured, published_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		post.AuthorID,
		post.CategoryID,
		post.Title,
		post.Slug,
		post.Summary,
		post.Content,
		post.ContentFormat,
		post.CoverImageURL,
		post.Status,
		post.IsFeatured,
		post.PublishedAt,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (r *PostgresPostRepository) List(
	ctx context.Context,
	params ListPostsParams,
) ([]model.Post, int64, error) {
	where := "deleted_at IS NULL"
	args := make([]any, 0, 3)
	if params.Status != "" {
		args = append(args, params.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM posts WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, params.Limit)
	limitPosition := len(args)
	args = append(args, params.Offset)
	offsetPosition := len(args)

	query := fmt.Sprintf(`
SELECT %s
FROM posts
WHERE %s
ORDER BY COALESCE(published_at, created_at) DESC, id DESC
LIMIT $%d OFFSET $%d`, postColumns, where, limitPosition, offsetPosition)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	posts := make([]model.Post, 0, params.Limit)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, *post)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *PostgresPostRepository) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	query := fmt.Sprintf("SELECT %s FROM posts WHERE id = $1 AND deleted_at IS NULL", postColumns)
	post, err := scanPost(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPostNotFound
	}
	return post, err
}

func (r *PostgresPostRepository) Update(
	ctx context.Context,
	id int64,
	changes model.PostChanges,
) (*model.Post, error) {
	sets := make([]string, 0, 11)
	args := make([]any, 0, 12)
	add := func(column string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	if changes.AuthorID != nil {
		add("author_id", *changes.AuthorID)
	}
	if changes.CategoryID != nil {
		add("category_id", *changes.CategoryID)
	}
	if changes.Title != nil {
		add("title", *changes.Title)
	}
	if changes.Slug != nil {
		add("slug", *changes.Slug)
	}
	if changes.Summary != nil {
		add("summary", *changes.Summary)
	}
	if changes.Content != nil {
		add("content", *changes.Content)
	}
	if changes.ContentFormat != nil {
		add("content_format", *changes.ContentFormat)
	}
	if changes.CoverImageURL != nil {
		add("cover_image_url", *changes.CoverImageURL)
	}
	if changes.Status != nil {
		add("status", *changes.Status)
	}
	if changes.IsFeatured != nil {
		add("is_featured", *changes.IsFeatured)
	}
	if changes.PublishedAt != nil {
		add("published_at", *changes.PublishedAt)
	}
	if len(sets) == 0 {
		return nil, ErrNoChanges
	}

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE posts SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s",
		strings.Join(sets, ", "),
		len(args),
		postColumns,
	)

	post, err := scanPost(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPostNotFound
	}
	return post, err
}

func (r *PostgresPostRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, `
UPDATE posts
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPostNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPost(row rowScanner) (*model.Post, error) {
	var post model.Post
	err := row.Scan(
		&post.ID,
		&post.AuthorID,
		&post.CategoryID,
		&post.Title,
		&post.Slug,
		&post.Summary,
		&post.Content,
		&post.ContentFormat,
		&post.CoverImageURL,
		&post.Status,
		&post.IsFeatured,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &post, nil
}
