package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"inventario/server/internal/service"
	"inventario/shared/dto"
)

// InventoryHandler handles inventory submissions from agents.
type InventoryHandler struct {
	service *service.InventoryService
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: svc}
}

// SubmitInventory processes a full inventory snapshot from an authenticated agent.
func (h *InventoryHandler) SubmitInventory(c *gin.Context) {
	deviceIDRaw, exists := c.Get("device_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "device not authenticated"})
		return
	}
	deviceID, ok := deviceIDRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "invalid device session"})
		return
	}

	var req dto.InventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("invalid inventory payload", "error", err, "device_id", deviceID)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	if err := h.service.ProcessInventory(c.Request.Context(), deviceID, &req); err != nil {
		slog.Error("failed to process inventory", "error", err, "device_id", deviceID)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to process inventory"})
		return
	}

	slog.Info("inventory processed", "device_id", deviceID, "hostname", req.Hostname)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "inventory received"})
}
