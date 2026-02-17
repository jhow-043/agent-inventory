package handler

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"inventario/server/internal/middleware"
	"inventario/server/internal/repository"
	"inventario/server/internal/service"
	"inventario/shared/dto"
)

// DeviceHandler handles device listing and detail endpoints.
type DeviceHandler struct {
	service     *service.DeviceService
	auditLogger *middleware.AuditLogger
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(svc *service.DeviceService, auditLogger *middleware.AuditLogger) *DeviceHandler {
	return &DeviceHandler{service: svc, auditLogger: auditLogger}
}

// ListDevices returns devices with pagination, sorting, and filtering.
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	params := repository.ListParams{
		Hostname:     c.Query("hostname"),
		OS:           c.Query("os"),
		Status:       c.Query("status"),
		DepartmentID: c.Query("department_id"),
		Sort:         c.Query("sort"),
		Order:        c.Query("order"),
		Page:         page,
		Limit:        limit,
	}

	resp, err := h.service.ListDevices(c.Request.Context(), params)
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

// UpdateStatus changes a device's status (active / inactive).
func (h *DeviceHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid device ID"})
		return
	}

	var req dto.UpdateDeviceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		slog.Error("failed to update device status", "error", err, "device_id", id)
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "device not found"})
		return
	}

	h.auditLogger.Log(c, "device.status.update", "device", &id, map[string]interface{}{"new_status": req.Status})
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "status updated"})
}

// UpdateDepartment assigns or removes a department from a device.
func (h *DeviceHandler) UpdateDepartment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid device ID"})
		return
	}

	var req dto.UpdateDeviceDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}

	if err := h.service.UpdateDepartment(c.Request.Context(), id, req.DepartmentID); err != nil {
		slog.Error("failed to update device department", "error", err, "device_id", id)
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "device not found"})
		return
	}

	h.auditLogger.Log(c, "device.department.update", "device", &id, map[string]interface{}{"department_id": req.DepartmentID})
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "department updated"})
}

// GetHardwareHistory returns hardware change snapshots for a device.
func (h *DeviceHandler) GetHardwareHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid device ID"})
		return
	}

	history, err := h.service.GetDeviceDetail(c.Request.Context(), id)
	if err != nil {
		slog.Error("failed to get hardware history", "error", err, "device_id", id)
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "device not found"})
		return
	}

	c.JSON(http.StatusOK, history.HardwareHistory)
}

// ExportCSV streams a CSV file of devices matching the current filters.
func (h *DeviceHandler) ExportCSV(c *gin.Context) {
	params := repository.ListParams{
		Hostname:     c.Query("hostname"),
		OS:           c.Query("os"),
		Status:       c.Query("status"),
		DepartmentID: c.Query("department_id"),
		Sort:         c.DefaultQuery("sort", "hostname"),
		Order:        c.DefaultQuery("order", "asc"),
	}

	devices, err := h.service.ListForExport(c.Request.Context(), params)
	if err != nil {
		slog.Error("failed to export devices", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to export devices"})
		return
	}

	filename := fmt.Sprintf("devices_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	// Header row
	_ = w.Write([]string{
		"Hostname", "Serial Number", "OS", "OS Version", "OS Build", "Architecture",
		"Logged In User", "Agent Version", "License Status", "Status", "Department",
		"Last Seen", "Created At",
	})

	for _, d := range devices {
		deptName := ""
		if d.DepartmentName != nil {
			deptName = *d.DepartmentName
		}
		_ = w.Write([]string{
			d.Hostname,
			d.SerialNumber,
			d.OSName,
			d.OSVersion,
			d.OSBuild,
			d.OSArch,
			d.LoggedInUser,
			d.AgentVersion,
			d.LicenseStatus,
			d.Status,
			deptName,
			d.LastSeen.Format(time.RFC3339),
			d.CreatedAt.Format(time.RFC3339),
		})
	}
}
