package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/internal/pkg/amap"
	"github.com/haojia/commute/pkg/response"
)

type MapHandler struct {
	amap *amap.Client
}

func NewMapHandler(client *amap.Client) *MapHandler {
	return &MapHandler{amap: client}
}

func (h *MapHandler) mapAmapError(c *gin.Context, err error) {
	var apiErr *amap.APIError
	if errors.As(err, &apiErr) {
		response.Fail(c, http.StatusBadGateway, response.CodeAmapFailure,
			"高德返回错误："+apiErr.Info, gin.H{"amap_code": apiErr.InfoCode})
		return
	}
	if errors.Is(err, amap.ErrKeyNotConfigured) {
		response.Fail(c, http.StatusServiceUnavailable, response.CodeAmapFailure,
			"高德 Key 未配置", nil)
		return
	}
	response.Fail(c, http.StatusBadGateway, response.CodeAmapFailure, err.Error(), nil)
}

func (h *MapHandler) Geocode(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, "address 必填", nil)
		return
	}
	city := c.Query("city")
	items, err := h.amap.Geocode(c.Request.Context(), address, city)
	if err != nil {
		h.mapAmapError(c, err)
		return
	}
	response.OK(c, gin.H{"results": items})
}

func (h *MapHandler) Regeocode(c *gin.Context) {
	lngStr := c.Query("longitude")
	latStr := c.Query("latitude")
	lng, err1 := strconv.ParseFloat(lngStr, 64)
	lat, err2 := strconv.ParseFloat(latStr, 64)
	if err1 != nil || err2 != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, "经纬度非法", nil)
		return
	}
	res, err := h.amap.Regeocode(c.Request.Context(), lng, lat)
	if err != nil {
		h.mapAmapError(c, err)
		return
	}
	response.OK(c, res)
}

func (h *MapHandler) POISearch(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeBadRequest, "keyword 必填", nil)
		return
	}
	region := c.Query("city")
	pageSize := 10
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 25 {
			pageSize = n
		}
	}
	items, err := h.amap.POISearch(c.Request.Context(), keyword, region, pageSize)
	if err != nil {
		h.mapAmapError(c, err)
		return
	}
	response.OK(c, gin.H{"results": items})
}
