package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/pkg/amap"
	"github.com/haojia/commute/internal/pkg/doubao"
	"github.com/haojia/commute/internal/service"
	"github.com/haojia/commute/pkg/response"
)

type AIHandler struct {
	svc *service.AIService
}

func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

func (h *AIHandler) RecommendCompanies(c *gin.Context) {
	var in model.AIRecommendInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}

	result, err := h.svc.RecommendCompanies(c.Request.Context(), model.DefaultUserID, in)
	if err != nil {
		var doubaoErr *doubao.APIError
		if errors.As(err, &doubaoErr) {
			response.Fail(c, http.StatusBadGateway, response.CodeDoubaoFailure,
				"豆包返回错误："+doubaoErr.Message, gin.H{"code": doubaoErr.Code})
			return
		}
		if errors.Is(err, doubao.ErrNotConfigured) {
			response.Fail(c, http.StatusServiceUnavailable, response.CodeDoubaoFailure,
				"豆包 API Key 或模型未配置", nil)
			return
		}
		var amapErr *amap.APIError
		if errors.As(err, &amapErr) {
			response.Fail(c, http.StatusBadGateway, response.CodeAmapFailure,
				"高德返回错误："+amapErr.Info, gin.H{"amap_code": amapErr.InfoCode})
			return
		}
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, result)
}
