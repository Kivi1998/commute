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

type CommuteHandler struct {
	svc *service.CommuteService
}

func NewCommuteHandler(svc *service.CommuteService) *CommuteHandler {
	return &CommuteHandler{svc: svc}
}

func (h *CommuteHandler) Calculate(c *gin.Context) {
	var in model.CommuteCalculateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	// 只允许 depart_at（MVP）
	if in.Morning.Strategy == "arrive_by" || in.Evening.Strategy == "arrive_by" {
		response.Fail(c, http.StatusBadRequest, response.CodeInvalidValue, "arrive_by 策略暂未实现，请用 depart_at", nil)
		return
	}

	warning := ""
	if len(in.CompanyIDs) > 20 {
		warning = "soft_limit_exceeded"
	}

	result, err := h.svc.Calculate(c.Request.Context(), middleware.GetUserID(c), in)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Fail(c, http.StatusNotFound, response.CodeNotFound, err.Error(), nil)
			return
		}
		response.Internal(c, err.Error())
		return
	}

	if warning != "" {
		c.JSON(http.StatusOK, response.Body{
			Code:      response.CodeOK,
			Message:   "已计算" + strconv.Itoa(len(in.CompanyIDs)) + "家公司，建议不超过 20 家",
			Data:      gin.H{"warning": warning, "result": result},
			RequestID: c.GetString("request_id"),
		})
		return
	}
	response.OK(c, result)
}

func (h *CommuteHandler) GetResult(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	r, err := h.svc.GetResultDetail(c.Request.Context(), middleware.GetUserID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "结果不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, r)
}

func (h *CommuteHandler) ListByQuery(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	list, err := h.svc.ListQueryResults(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list})
}

func (h *CommuteHandler) ListQueries(c *gin.Context) {
	limit := 50
	list, err := h.svc.ListQueries(c.Request.Context(), middleware.GetUserID(c), limit)
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list})
}

func (h *CommuteHandler) GetQuery(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	q, err := h.svc.GetQueryDetail(c.Request.Context(), middleware.GetUserID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "查询不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, q)
}

func (h *CommuteHandler) DeleteQuery(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	err := h.svc.DeleteQuery(c.Request.Context(), middleware.GetUserID(c), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "查询不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}
