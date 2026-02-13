// Package handler implements the HTTP handlers (controllers) for the API.
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"inventario/shared/dto"
)

// HealthHandler provides liveness and readiness probes.
type HealthHandler struct {
	db *sqlx.DB
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Healthz is the liveness probe — returns 200 if the process is running.
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "ok"})
}

// Readyz is the readiness probe — returns 200 only if the database is reachable.
func (h *HealthHandler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, dto.ReadyResponse{
			Status:   "not ready",
			Database: "error",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ReadyResponse{
		Status:   "ready",
		Database: "ok",
	})
}
