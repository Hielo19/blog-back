package service

import (
	"context"
	"errors"
	"testing"

	"gin-demo/internal/dto"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
)

type fakePostRepository struct {
	created    *model.Post
	listParams repository.ListPostsParams
	listPosts  []model.Post
	listTotal  int64
}

func (r *fakePostRepository) Create(_ context.Context, post *model.Post) error {
	r.created = post
	post.ID = 1
	return nil
}

func (r *fakePostRepository) List(
	_ context.Context,
	params repository.ListPostsParams,
) ([]model.Post, int64, error) {
	r.listParams = params
	return r.listPosts, r.listTotal, nil
}

func (r *fakePostRepository) GetByID(_ context.Context, _ int64) (*model.Post, error) {
	return nil, repository.ErrPostNotFound
}

func (r *fakePostRepository) Update(
	_ context.Context,
	_ int64,
	_ model.PostChanges,
) (*model.Post, error) {
	return nil, repository.ErrPostNotFound
}

func (r *fakePostRepository) Delete(_ context.Context, _ int64) error {
	return repository.ErrPostNotFound
}

func TestCreateAppliesDefaults(t *testing.T) {
	repository := &fakePostRepository{}
	service := NewPostService(repository)

	post, err := service.Create(context.Background(), dto.CreatePostRequest{
		AuthorID: 1,
		Title:    "  第一篇文章  ",
		Slug:     "  first-post  ",
		Content:  "# Hello",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if post.Title != "第一篇文章" || post.Slug != "first-post" {
		t.Fatalf("Create() did not trim title and slug: %#v", post)
	}
	if post.ContentFormat != "markdown" || post.Status != "draft" {
		t.Fatalf("Create() defaults = (%q, %q)", post.ContentFormat, post.Status)
	}
}

func TestCreatePublishedPostSetsPublishedAt(t *testing.T) {
	repository := &fakePostRepository{}
	service := NewPostService(repository)

	post, err := service.Create(context.Background(), dto.CreatePostRequest{
		AuthorID: 1,
		Title:    "已发布文章",
		Slug:     "published-post",
		Content:  "正文",
		Status:   "published",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if post.PublishedAt == nil {
		t.Fatal("Create() PublishedAt is nil for a published post")
	}
}

func TestListCalculatesPagination(t *testing.T) {
	repository := &fakePostRepository{
		listPosts: []model.Post{{ID: 11}},
		listTotal: 21,
	}
	service := NewPostService(repository)

	result, err := service.List(context.Background(), dto.ListPostsQuery{
		Page:     2,
		PageSize: 10,
		Status:   "published",
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repository.listParams.Offset != 10 || repository.listParams.Limit != 10 {
		t.Fatalf("List() params = %#v", repository.listParams)
	}
	if result.TotalPages != 3 || result.Total != 21 {
		t.Fatalf("List() result = %#v", result)
	}
}

func TestUpdateRejectsEmptyRequest(t *testing.T) {
	service := NewPostService(&fakePostRepository{})

	_, err := service.Update(context.Background(), 1, dto.UpdatePostRequest{})
	if !errors.Is(err, ErrNoChanges) {
		t.Fatalf("Update() error = %v, want ErrNoChanges", err)
	}
}
