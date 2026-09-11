package dto

import "time"

type CreatePostRequest struct {
	AuthorID      int64      `json:"author_id" binding:"required,gt=0"`
	CategoryID    *int64     `json:"category_id" binding:"omitempty,gt=0"`
	Title         string     `json:"title" binding:"required,max=200"`
	Slug          string     `json:"slug" binding:"required,max=220"`
	Summary       *string    `json:"summary" binding:"omitempty,max=500"`
	Content       string     `json:"content" binding:"required"`
	ContentFormat string     `json:"content_format" binding:"omitempty,oneof=markdown"`
	CoverImageURL *string    `json:"cover_image_url" binding:"omitempty,max=500"`
	Status        string     `json:"status" binding:"omitempty,oneof=draft published archived"`
	IsFeatured    bool       `json:"is_featured"`
	PublishedAt   *time.Time `json:"published_at"`
}

type UpdatePostRequest struct {
	AuthorID      *int64     `json:"author_id" binding:"omitempty,gt=0"`
	CategoryID    *int64     `json:"category_id" binding:"omitempty,gt=0"`
	Title         *string    `json:"title" binding:"omitempty,min=1,max=200"`
	Slug          *string    `json:"slug" binding:"omitempty,min=1,max=220"`
	Summary       *string    `json:"summary" binding:"omitempty,max=500"`
	Content       *string    `json:"content" binding:"omitempty,min=1"`
	ContentFormat *string    `json:"content_format" binding:"omitempty,oneof=markdown"`
	CoverImageURL *string    `json:"cover_image_url" binding:"omitempty,max=500"`
	Status        *string    `json:"status" binding:"omitempty,oneof=draft published archived"`
	IsFeatured    *bool      `json:"is_featured"`
	PublishedAt   *time.Time `json:"published_at"`
}

func (r UpdatePostRequest) HasChanges() bool {
	return r.AuthorID != nil || r.CategoryID != nil || r.Title != nil ||
		r.Slug != nil || r.Summary != nil || r.Content != nil ||
		r.ContentFormat != nil || r.CoverImageURL != nil || r.Status != nil ||
		r.IsFeatured != nil || r.PublishedAt != nil
}

type ListPostsQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=draft published archived"`
}
