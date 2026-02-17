package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"inventario/server/internal/repository"
	"inventario/shared/dto"
)

// AuditLogHandler handles audit log queries.
type AuditLogHandler struct {
	repo *repository.AuditLogRepository
}

// NewAuditLogHandler creates a new AuditLogHandler.
func NewAuditLogHandler(repo *repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{repo: repo}
}

// ListAuditLogs returns audit logs with optional filtering.
// Query params: user_id, action, resource_type, resource_id, limit, offset
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	filters := make(map[string]interface{})

	// Parse filters
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filters["user_id"] = userID
		}
	}

	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters["resource_type"] = resourceType
	}

	if resourceIDStr := c.Query("resource_id"); resourceIDStr != "" {
		if resourceID, err := uuid.Parse(resourceIDStr); err == nil {
			filters["resource_id"] = resourceID
		}
	}

	// Parse pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit < 1 || limit > 100 {
		limit = 50
	}

	logs, total, err := h.repo.List(c.Request.Context(), filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch audit logs"})
		return
	}

	// Convert to response DTOs
	resp := dto.AuditLogListResponse{
		Logs:  make([]dto.AuditLogResponse, 0, len(logs)),
		Total: total,
	}

	for _, log := range logs {
		resp.Logs = append(resp.Logs, dto.AuditLogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			Username:     log.Username,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			Details:      log.Details,
			IPAddress:    log.IPAddress,
			UserAgent:    log.UserAgent,
			CreatedAt:    log.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, resp)
}

// GetResourceAuditLogs returns audit logs for a specific resource.
func (h *AuditLogHandler) GetResourceAuditLogs(c *gin.Context) {
	resourceType := c.Param("type")
	resourceIDStr := c.Param("id")

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid resource ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, err := h.repo.GetByResourceID(c.Request.Context(), resourceType, resourceID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch audit logs"})
		return
	}

	// Convert to response DTOs
	resp := dto.AuditLogListResponse{
		Logs:  make([]dto.AuditLogResponse, 0, len(logs)),
		Total: len(logs),
	}

	for _, log := range logs {
		resp.Logs = append(resp.Logs, dto.AuditLogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			Username:     log.Username,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			Details:      log.Details,
			IPAddress:    log.IPAddress,
			UserAgent:    log.UserAgent,
			CreatedAt:    log.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, resp)
}
