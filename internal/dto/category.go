package dto

type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Slug        string  `json:"slug" binding:"required,max=120"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	SortOrder   int     `json:"sort_order" binding:"min=0"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Slug        *string `json:"slug" binding:"omitempty,min=1,max=120"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	SortOrder   *int    `json:"sort_order" binding:"omitempty,min=0"`
}

func (r UpdateCategoryRequest) HasChanges() bool {
	return r.Name != nil || r.Slug != nil || r.Description != nil || r.SortOrder != nil
}
