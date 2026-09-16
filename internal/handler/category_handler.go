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

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var request dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请求参数格式或内容不正确")
		return
	}

	category, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		handleCategoryServiceError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, category)
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.service.List(c.Request.Context())
	if err != nil {
		handleCategoryServiceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, categories)
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, ok := categoryID(c)
	if !ok {
		return
	}

	category, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleCategoryServiceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := categoryID(c)
	if !ok {
		return
	}

	var request dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请求参数格式或内容不正确")
		return
	}

	category, err := h.service.Update(c.Request.Context(), id, request)
	if err != nil {
		handleCategoryServiceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, ok := categoryID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleCategoryServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func categoryID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", "分类 ID 必须是正整数")
		return 0, false
	}
	return id, true
}

func handleCategoryServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCategoryNotFound):
		response.Failure(c, http.StatusNotFound, "CATEGORY_NOT_FOUND", err.Error())
	case errors.Is(err, service.ErrNoCategoryChanges), errors.Is(err, service.ErrInvalidCategory):
		response.Failure(c, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
	case errors.Is(err, service.ErrCategoryNameUsed):
		response.Failure(c, http.StatusConflict, "CATEGORY_NAME_ALREADY_USED", err.Error())
	case errors.Is(err, service.ErrCategorySlugUsed):
		response.Failure(c, http.StatusConflict, "CATEGORY_SLUG_ALREADY_USED", err.Error())
	case errors.Is(err, service.ErrCategoryAlreadyExist):
		response.Failure(c, http.StatusConflict, "CATEGORY_ALREADY_EXISTS", err.Error())
	default:
		_ = c.Error(err)
		response.Failure(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误")
	}
}
