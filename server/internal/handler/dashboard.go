package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"inventario/server/internal/service"
	"inventario/shared/dto"
)

// DashboardHandler handles dashboard statistics endpoints.
type DashboardHandler struct {
	service *service.DashboardService
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: svc}
}

// GetStats returns aggregated device statistics (total, online, offline).
func (h *DashboardHandler) GetStats(c *gin.Context) {
	resp, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		slog.Error("failed to get dashboard stats", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
