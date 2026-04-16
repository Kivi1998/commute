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

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in model.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	r, err := h.svc.Login(c.Request.Context(), in)
	if errors.Is(err, service.ErrInvalidCredentials) {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "邮箱或密码错误", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, r)
}

func (h *AuthHandler) Me(c *gin.Context) {
	uid := middleware.GetUserID(c)
	if uid == 0 {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录", nil)
		return
	}
	u, err := h.svc.GetUser(c.Request.Context(), uid)
	if errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户不存在", nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, u)
}
