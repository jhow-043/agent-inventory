package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"inventario/server/internal/service"
	"inventario/shared/dto"
)

// DeviceHandler handles device listing and detail endpoints.
type DeviceHandler struct {
	service *service.DeviceService
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{service: svc}
}

// ListDevices returns all devices, with optional hostname and os query filters.
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	hostname := c.Query("hostname")
	osFilter := c.Query("os")

	resp, err := h.service.ListDevices(c.Request.Context(), hostname, osFilter)
	if err != nil {
		slog.Error("failed to list devices", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list devices"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDevice returns full details for a single device, including hardware, disks, NICs, and software.
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid device ID"})
		return
	}

	detail, err := h.service.GetDeviceDetail(c.Request.Context(), id)
	if err != nil {
		slog.Error("failed to get device detail", "error", err, "device_id", id)
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "device not found"})
		return
	}

	c.JSON(http.StatusOK, detail)
}
