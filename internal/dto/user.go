package dto

type CreateUserRequest struct {
	Username    string  `json:"username" binding:"required,min=3,max=50"`
	Email       string  `json:"email" binding:"required,email,max=255"`
	Password    string  `json:"password" binding:"required,min=8,max=72"`
	DisplayName *string `json:"display_name" binding:"omitempty,max=100"`
	AvatarURL   *string `json:"avatar_url" binding:"omitempty,max=500"`
	Bio         *string `json:"bio" binding:"omitempty,max=500"`
}
