package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/internal/middleware"
	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/repository"
	"github.com/haojia/commute/internal/service"
	"github.com/haojia/commute/pkg/response"
)

type AddressHandler struct {
	svc *service.AddressService
}

func NewAddressHandler(svc *service.AddressService) *AddressHandler {
	return &AddressHandler{svc: svc}
}

func parseID(c *gin.Context) (int64, bool) {
	raw := c.Param("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidValue, "id 非法", gin.H{"id": raw})
		return 0, false
	}
	return id, true
}

func (h *AddressHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list})
}

func (h *AddressHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	a, err := h.svc.Get(c.Request.Context(), middleware.GetUserID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "地址不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, a)
}

func (h *AddressHandler) Create(c *gin.Context) {
	var in model.HomeAddressCreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	a, err := h.svc.Create(c.Request.Context(), middleware.GetUserID(c), in)
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, response.Body{
		Code: response.CodeOK, Message: "ok", Data: a,
		RequestID: c.GetString("request_id"),
	})
}

func (h *AddressHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in model.HomeAddressUpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	a, err := h.svc.Update(c.Request.Context(), middleware.GetUserID(c), id, in)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "地址不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, a)
}

func (h *AddressHandler) SetDefault(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	a, err := h.svc.SetDefault(c.Request.Context(), middleware.GetUserID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "地址不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, a)
}

func (h *AddressHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	err := h.svc.Delete(c.Request.Context(), middleware.GetUserID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "地址不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}
