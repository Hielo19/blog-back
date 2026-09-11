package handler

import (
	"errors"
	"net/http"
	"strconv"

	"gin-demo/internal/dto"
	"gin-demo/internal/response"
	"gin-demo/internal/service"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	service service.PostService
}

func NewPostHandler(service service.PostService) *PostHandler {
	return &PostHandler{service: service}
}

func (h *PostHandler) Create(c *gin.Context) {
	var request dto.CreatePostRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请求参数格式或内容不正确")
		return
	}

	post, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, post)
}

func (h *PostHandler) List(c *gin.Context) {
	var query dto.ListPostsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", "分页参数或文章状态不正确")
		return
	}

	result, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *PostHandler) GetByID(c *gin.Context) {
	id, ok := postID(c)
	if !ok {
		return
	}

	post, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, post)
}

func (h *PostHandler) Update(c *gin.Context) {
	id, ok := postID(c)
	if !ok {
		return
	}

	var request dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请求参数格式或内容不正确")
		return
	}

	post, err := h.service.Update(c.Request.Context(), id, request)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, post)
}

func (h *PostHandler) Delete(c *gin.Context) {
	id, ok := postID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func postID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", "文章 ID 必须是正整数")
		return 0, false
	}
	return id, true
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPostNotFound):
		response.Failure(c, http.StatusNotFound, "POST_NOT_FOUND", "文章不存在")
	case errors.Is(err, service.ErrNoChanges), errors.Is(err, service.ErrInvalidPost):
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
	case errors.Is(err, service.ErrSlugAlreadyUsed):
		response.Failure(c, http.StatusConflict, "SLUG_ALREADY_USED", err.Error())
	case errors.Is(err, service.ErrInvalidReference), errors.Is(err, service.ErrInvalidState):
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
	default:
		_ = c.Error(err)
		response.Failure(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误")
	}
}
