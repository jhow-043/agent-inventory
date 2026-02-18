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
	service      *service.DeviceService
	auditLogger  *middleware.AuditLogger
	activityRepo *repository.DeviceActivityRepository
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(svc *service.DeviceService, auditLogger *middleware.AuditLogger, activityRepo *repository.DeviceActivityRepository) *DeviceHandler {
	return &DeviceHandler{service: svc, auditLogger: auditLogger, activityRepo: activityRepo}
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

// DeleteDevice deletes a device and all related data.
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid device ID"})
		return
	}

	device, err := h.service.DeleteDevice(c.Request.Context(), id)
	if err != nil {
		slog.Error("failed to delete device", "error", err, "device_id", id)
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "device not found"})
		return
	}

	h.auditLogger.Log(c, "device.delete", "device", &id, map[string]interface{}{
		"hostname":      device.Hostname,
		"serial_number": device.SerialNumber,
	})
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "device deleted successfully"})
}

// BulkUpdateStatus changes the status of multiple devices.
func (h *DeviceHandler) BulkUpdateStatus(c *gin.Context) {
	var req dto.BulkDeviceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}

	affected, err := h.service.BulkUpdateStatus(c.Request.Context(), req.DeviceIDs, req.Status)
	if err != nil {
		slog.Error("failed to bulk update status", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update devices"})
		return
	}

	for _, id := range req.DeviceIDs {
		idCopy := id
		h.auditLogger.Log(c, "device.status.update", "device", &idCopy, map[string]interface{}{"new_status": req.Status, "bulk": true})
	}
	c.JSON(http.StatusOK, dto.BulkActionResponse{Affected: int(affected), Message: fmt.Sprintf("%d device(s) updated", affected)})
}

// BulkUpdateDepartment assigns a department to multiple devices.
func (h *DeviceHandler) BulkUpdateDepartment(c *gin.Context) {
	var req dto.BulkDeviceDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}

	affected, err := h.service.BulkUpdateDepartment(c.Request.Context(), req.DeviceIDs, req.DepartmentID)
	if err != nil {
		slog.Error("failed to bulk update department", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update devices"})
		return
	}

	for _, id := range req.DeviceIDs {
		idCopy := id
		h.auditLogger.Log(c, "device.department.update", "device", &idCopy, map[string]interface{}{"department_id": req.DepartmentID, "bulk": true})
	}
	c.JSON(http.StatusOK, dto.BulkActionResponse{Affected: int(affected), Message: fmt.Sprintf("%d device(s) updated", affected)})
}

// BulkDelete deletes multiple devices.
func (h *DeviceHandler) BulkDelete(c *gin.Context) {
	var req dto.BulkDeviceDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}

	affected, err := h.service.BulkDelete(c.Request.Context(), req.DeviceIDs)
	if err != nil {
		slog.Error("failed to bulk delete devices", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to delete devices"})
		return
	}

	for _, id := range req.DeviceIDs {
		idCopy := id
		h.auditLogger.Log(c, "device.delete", "device", &idCopy, map[string]interface{}{"bulk": true})
	}
	c.JSON(http.StatusOK, dto.BulkActionResponse{Affected: int(affected), Message: fmt.Sprintf("%d device(s) deleted", affected)})
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

// GetDeviceActivity returns the activity log for a device.
func (h *DeviceHandler) GetDeviceActivity(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid device ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	logs, total, err := h.activityRepo.ListByDevice(c.Request.Context(), id, limit, offset)
	if err != nil {
		slog.Error("failed to get device activity", "error", err, "device_id", id)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get device activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": logs,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
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
