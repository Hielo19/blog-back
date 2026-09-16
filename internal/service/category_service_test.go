package service

import (
	"context"
	"errors"
	"testing"

	"gin-demo/internal/dto"
	"gin-demo/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
)

type fakeCategoryRepository struct {
	created   *model.Category
	createErr error
}

func (r *fakeCategoryRepository) Create(_ context.Context, category *model.Category) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = category
	category.ID = 1
	return nil
}

func (r *fakeCategoryRepository) List(_ context.Context) ([]model.Category, error) {
	return []model.Category{}, nil
}

func (r *fakeCategoryRepository) GetByID(_ context.Context, _ int64) (*model.Category, error) {
	return nil, ErrCategoryNotFound
}

func (r *fakeCategoryRepository) Update(
	_ context.Context,
	_ int64,
	_ model.CategoryChanges,
) (*model.Category, error) {
	return nil, ErrCategoryNotFound
}

func (r *fakeCategoryRepository) Delete(_ context.Context, _ int64) error {
	return ErrCategoryNotFound
}

func TestCreateCategoryNormalizesValues(t *testing.T) {
	repository := &fakeCategoryRepository{}
	service := NewCategoryService(repository)

	category, err := service.Create(context.Background(), dto.CreateCategoryRequest{
		Name:      "  后端开发  ",
		Slug:      "  Backend-Development  ",
		SortOrder: 10,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if category.Name != "后端开发" || category.Slug != "backend-development" {
		t.Fatalf("Create() values = (%q, %q)", category.Name, category.Slug)
	}
	if category.SortOrder != 10 {
		t.Fatalf("Create() SortOrder = %d", category.SortOrder)
	}
}

func TestCreateCategoryRejectsInvalidSlug(t *testing.T) {
	service := NewCategoryService(&fakeCategoryRepository{})

	_, err := service.Create(context.Background(), dto.CreateCategoryRequest{
		Name: "后端开发",
		Slug: "backend_development",
	})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("Create() error = %v, want ErrInvalidCategory", err)
	}
}

func TestCreateCategoryMapsDuplicateSlug(t *testing.T) {
	service := NewCategoryService(&fakeCategoryRepository{
		createErr: &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "uk_categories_slug",
		},
	})

	_, err := service.Create(context.Background(), dto.CreateCategoryRequest{
		Name: "后端开发",
		Slug: "backend-development",
	})
	if !errors.Is(err, ErrCategorySlugUsed) {
		t.Fatalf("Create() error = %v, want ErrCategorySlugUsed", err)
	}
}

func TestUpdateCategoryRejectsEmptyRequest(t *testing.T) {
	service := NewCategoryService(&fakeCategoryRepository{})

	_, err := service.Update(context.Background(), 1, dto.UpdateCategoryRequest{})
	if !errors.Is(err, ErrNoCategoryChanges) {
		t.Fatalf("Update() error = %v, want ErrNoCategoryChanges", err)
	}
}
