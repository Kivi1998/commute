package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/internal/config"
	"github.com/haojia/commute/pkg/response"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db        *pgxpool.Pool
	cfg       *config.Config
	startedAt time.Time
}

func NewHealthHandler(db *pgxpool.Pool, cfg *config.Config, startedAt time.Time) *HealthHandler {
	return &HealthHandler{db: db, cfg: cfg, startedAt: startedAt}
}

type healthResponse struct {
	Status        string            `json:"status"`
	Version       string            `json:"version"`
	UptimeSeconds int64             `json:"uptime_seconds"`
	Dependencies  map[string]string `json:"dependencies"`
}

func (h *HealthHandler) Get(c *gin.Context) {
	deps := map[string]string{
		"database": "ok",
		"amap":     h.checkKey(h.cfg.Amap.WebServiceKey),
		"doubao":   h.checkKey(h.cfg.Doubao.APIKey),
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		deps["database"] = "unreachable: " + err.Error()
	}

	status := "ok"
	if deps["database"] != "ok" {
		status = "degraded"
	}

	response.OK(c, healthResponse{
		Status:        status,
		Version:       h.cfg.App.Version,
		UptimeSeconds: int64(time.Since(h.startedAt).Seconds()),
		Dependencies:  deps,
	})
}

func (h *HealthHandler) checkKey(key string) string {
	if key == "" {
		return "not_configured"
	}
	return "configured"
}
