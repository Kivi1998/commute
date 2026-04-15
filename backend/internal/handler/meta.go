package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/pkg/response"
)

type MetaHandler struct{}

func NewMetaHandler() *MetaHandler { return &MetaHandler{} }

func (h *MetaHandler) Enums(c *gin.Context) {
	response.OK(c, model.AllEnums)
}
