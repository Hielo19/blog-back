package model

import "time"

type Post struct {
	ID            int64      `json:"id"`
	AuthorID      int64      `json:"author_id"`
	CategoryID    *int64     `json:"category_id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Summary       *string    `json:"summary"`
	Content       string     `json:"content"`
	ContentFormat string     `json:"content_format"`
	CoverImageURL *string    `json:"cover_image_url"`
	Status        string     `json:"status"`
	IsFeatured    bool       `json:"is_featured"`
	PublishedAt   *time.Time `json:"published_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type PostChanges struct {
	AuthorID      *int64
	CategoryID    *int64
	Title         *string
	Slug          *string
	Summary       *string
	Content       *string
	ContentFormat *string
	CoverImageURL *string
	Status        *string
	IsFeatured    *bool
	PublishedAt   *time.Time
}
