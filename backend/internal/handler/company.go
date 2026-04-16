package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/internal/middleware"
	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/repository"
	"github.com/haojia/commute/internal/service"
	"github.com/haojia/commute/pkg/response"
)

type CompanyHandler struct {
	svc *service.CompanyService
}

func NewCompanyHandler(svc *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{svc: svc}
}

func (h *CompanyHandler) List(c *gin.Context) {
	var q model.CompanyListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	result, err := h.svc.List(c.Request.Context(), middleware.GetUserID(c), q)
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *CompanyHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.svc.Get(c.Request.Context(), middleware.GetUserID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公司不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *CompanyHandler) Create(c *gin.Context) {
	var in model.CompanyCreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	item, err := h.svc.Create(c.Request.Context(), middleware.GetUserID(c), in)
	if errors.Is(err, repository.ErrDuplicate) {
		response.Fail(c, http.StatusConflict, response.CodeConflict, "同名同地址公司已存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, response.Body{
		Code: response.CodeOK, Message: "ok", Data: item,
		RequestID: c.GetString("request_id"),
	})
}

func (h *CompanyHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in model.CompanyUpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	item, err := h.svc.Update(c.Request.Context(), middleware.GetUserID(c), id, in)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公司不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *CompanyHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in model.CompanyStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	item, err := h.svc.UpdateStatus(c.Request.Context(), middleware.GetUserID(c), id, in.Status)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公司不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *CompanyHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	err := h.svc.Delete(c.Request.Context(), middleware.GetUserID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公司不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *CompanyHandler) Batch(c *gin.Context) {
	var in model.CompanyBatchInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	result, err := h.svc.Batch(c.Request.Context(), middleware.GetUserID(c), in)
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, result)
}
