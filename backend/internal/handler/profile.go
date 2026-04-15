package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/repository"
	"github.com/haojia/commute/internal/service"
	"github.com/haojia/commute/pkg/response"
)

type ProfileHandler struct {
	svc *service.ProfileService
}

func NewProfileHandler(svc *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) Get(c *gin.Context) {
	p, err := h.svc.Get(c.Request.Context(), model.DefaultUserID)
	if errors.Is(err, repository.ErrNotFound) {
		response.OK(c, nil)
		return
	}
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, p)
}

func (h *ProfileHandler) Upsert(c *gin.Context) {
	var in model.ProfileUpsertInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, err.Error(), nil)
		return
	}
	p, err := h.svc.Upsert(c.Request.Context(), model.DefaultUserID, in)
	if err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.OK(c, p)
}
