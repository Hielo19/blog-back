package handler

import (
	"errors"
	"net/http"

	"gin-demo/internal/dto"
	"gin-demo/internal/response"
	"gin-demo/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(c *gin.Context) {
	var request dto.CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请求参数格式或内容不正确")
		return
	}

	user, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, user)
}

func handleUserServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidUsername), errors.Is(err, service.ErrInvalidPassword):
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
	case errors.Is(err, service.ErrUsernameAlreadyUsed):
		response.Failure(c, http.StatusConflict, "USERNAME_ALREADY_USED", err.Error())
	case errors.Is(err, service.ErrEmailAlreadyUsed):
		response.Failure(c, http.StatusConflict, "EMAIL_ALREADY_USED", err.Error())
	case errors.Is(err, service.ErrUserAlreadyExists):
		response.Failure(c, http.StatusConflict, "USER_ALREADY_EXISTS", err.Error())
	default:
		_ = c.Error(err)
		response.Failure(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误")
	}
}
