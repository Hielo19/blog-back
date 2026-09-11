package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"gin-demo/internal/dto"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrPostNotFound     = repository.ErrPostNotFound
	ErrNoChanges        = errors.New("至少需要提供一个待修改字段")
	ErrInvalidPost      = errors.New("文章标题、slug 和正文不能为空")
	ErrSlugAlreadyUsed  = errors.New("文章 slug 已被使用")
	ErrInvalidReference = errors.New("作者或分类不存在")
	ErrInvalidState     = errors.New("文章状态或发布时间不符合约束")
)

type ListPostsResult struct {
	Items      []model.Post `json:"items"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	Total      int64        `json:"total"`
	TotalPages int          `json:"total_pages"`
}

type PostService interface {
	Create(ctx context.Context, request dto.CreatePostRequest) (*model.Post, error)
	List(ctx context.Context, query dto.ListPostsQuery) (ListPostsResult, error)
	GetByID(ctx context.Context, id int64) (*model.Post, error)
	Update(ctx context.Context, id int64, request dto.UpdatePostRequest) (*model.Post, error)
	Delete(ctx context.Context, id int64) error
}

type postService struct {
	repository repository.PostRepository
}

func NewPostService(repository repository.PostRepository) PostService {
	return &postService{repository: repository}
}

func (s *postService) Create(
	ctx context.Context,
	request dto.CreatePostRequest,
) (*model.Post, error) {
	request.Title = strings.TrimSpace(request.Title)
	request.Slug = strings.TrimSpace(request.Slug)
	if request.Title == "" || request.Slug == "" || strings.TrimSpace(request.Content) == "" {
		return nil, ErrInvalidPost
	}

	if request.ContentFormat == "" {
		request.ContentFormat = "markdown"
	}
	if request.Status == "" {
		request.Status = "draft"
	}
	if request.Status == "published" && request.PublishedAt == nil {
		now := time.Now().UTC()
		request.PublishedAt = &now
	}
	request.Summary = trimOptional(request.Summary)
	request.CoverImageURL = trimOptional(request.CoverImageURL)

	post := &model.Post{
		AuthorID:      request.AuthorID,
		CategoryID:    request.CategoryID,
		Title:         request.Title,
		Slug:          request.Slug,
		Summary:       request.Summary,
		Content:       request.Content,
		ContentFormat: request.ContentFormat,
		CoverImageURL: request.CoverImageURL,
		Status:        request.Status,
		IsFeatured:    request.IsFeatured,
		PublishedAt:   request.PublishedAt,
	}

	if err := s.repository.Create(ctx, post); err != nil {
		return nil, translateDatabaseError(err)
	}
	return post, nil
}

func (s *postService) List(
	ctx context.Context,
	query dto.ListPostsQuery,
) (ListPostsResult, error) {
	offset := (query.Page - 1) * query.PageSize
	posts, total, err := s.repository.List(ctx, repository.ListPostsParams{
		Status: query.Status,
		Limit:  query.PageSize,
		Offset: offset,
	})
	if err != nil {
		return ListPostsResult{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}

	return ListPostsResult{
		Items:      posts,
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *postService) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *postService) Update(
	ctx context.Context,
	id int64,
	request dto.UpdatePostRequest,
) (*model.Post, error) {
	if !request.HasChanges() {
		return nil, ErrNoChanges
	}

	request.Title = trimOptional(request.Title)
	request.Slug = trimOptional(request.Slug)
	request.Summary = trimOptional(request.Summary)
	request.CoverImageURL = trimOptional(request.CoverImageURL)
	if request.Title != nil && *request.Title == "" ||
		request.Slug != nil && *request.Slug == "" ||
		request.Content != nil && strings.TrimSpace(*request.Content) == "" {
		return nil, ErrInvalidPost
	}

	if request.Status != nil && *request.Status == "published" && request.PublishedAt == nil {
		currentPost, err := s.repository.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if currentPost.PublishedAt == nil {
			now := time.Now().UTC()
			request.PublishedAt = &now
		}
	}

	post, err := s.repository.Update(ctx, id, model.PostChanges{
		AuthorID:      request.AuthorID,
		CategoryID:    request.CategoryID,
		Title:         request.Title,
		Slug:          request.Slug,
		Summary:       request.Summary,
		Content:       request.Content,
		ContentFormat: request.ContentFormat,
		CoverImageURL: request.CoverImageURL,
		Status:        request.Status,
		IsFeatured:    request.IsFeatured,
		PublishedAt:   request.PublishedAt,
	})
	if err != nil {
		return nil, translateDatabaseError(err)
	}
	return post, nil
}

func (s *postService) Delete(ctx context.Context, id int64) error {
	return s.repository.Delete(ctx, id)
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func translateDatabaseError(err error) error {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return err
	}

	switch postgresError.Code {
	case "23505":
		return ErrSlugAlreadyUsed
	case "23503":
		return ErrInvalidReference
	case "23514":
		return ErrInvalidState
	default:
		return err
	}
}
