package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"gin-demo/internal/dto"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrCategoryNotFound     = repository.ErrCategoryNotFound
	ErrNoCategoryChanges    = errors.New("至少需要提供一个待修改的分类字段")
	ErrInvalidCategory      = errors.New("分类名称不能为空，slug 只能包含小写字母、数字和连字符")
	ErrCategoryNameUsed     = errors.New("分类名称已被使用")
	ErrCategorySlugUsed     = errors.New("分类 slug 已被使用")
	ErrCategoryAlreadyExist = errors.New("分类名称或 slug 已被使用")
)

var categorySlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type CategoryService interface {
	Create(ctx context.Context, request dto.CreateCategoryRequest) (*model.Category, error)
	List(ctx context.Context) ([]model.Category, error)
	GetByID(ctx context.Context, id int64) (*model.Category, error)
	Update(ctx context.Context, id int64, request dto.UpdateCategoryRequest) (*model.Category, error)
	Delete(ctx context.Context, id int64) error
}

type categoryService struct {
	repository repository.CategoryRepository
}

func NewCategoryService(repository repository.CategoryRepository) CategoryService {
	return &categoryService{repository: repository}
}

func (s *categoryService) Create(
	ctx context.Context,
	request dto.CreateCategoryRequest,
) (*model.Category, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Slug = strings.ToLower(strings.TrimSpace(request.Slug))
	request.Description = trimOptional(request.Description)
	if request.Name == "" || !categorySlugPattern.MatchString(request.Slug) {
		return nil, ErrInvalidCategory
	}

	category := &model.Category{
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		SortOrder:   request.SortOrder,
	}
	if err := s.repository.Create(ctx, category); err != nil {
		return nil, translateCategoryDatabaseError(err)
	}
	return category, nil
}

func (s *categoryService) List(ctx context.Context) ([]model.Category, error) {
	return s.repository.List(ctx)
}

func (s *categoryService) GetByID(ctx context.Context, id int64) (*model.Category, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *categoryService) Update(
	ctx context.Context,
	id int64,
	request dto.UpdateCategoryRequest,
) (*model.Category, error) {
	if !request.HasChanges() {
		return nil, ErrNoCategoryChanges
	}

	request.Name = trimOptional(request.Name)
	request.Slug = trimOptional(request.Slug)
	request.Description = trimOptional(request.Description)
	if request.Slug != nil {
		slug := strings.ToLower(*request.Slug)
		request.Slug = &slug
	}
	if request.Name != nil && *request.Name == "" ||
		request.Slug != nil && !categorySlugPattern.MatchString(*request.Slug) {
		return nil, ErrInvalidCategory
	}

	category, err := s.repository.Update(ctx, id, model.CategoryChanges{
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		SortOrder:   request.SortOrder,
	})
	if err != nil {
		return nil, translateCategoryDatabaseError(err)
	}
	return category, nil
}

func (s *categoryService) Delete(ctx context.Context, id int64) error {
	return s.repository.Delete(ctx, id)
}

func translateCategoryDatabaseError(err error) error {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23505" {
		return err
	}

	switch postgresError.ConstraintName {
	case "uk_categories_name":
		return ErrCategoryNameUsed
	case "uk_categories_slug":
		return ErrCategorySlugUsed
	default:
		return ErrCategoryAlreadyExist
	}
}
