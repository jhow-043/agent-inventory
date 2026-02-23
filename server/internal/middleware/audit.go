package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"inventario/server/internal/repository"
)

// AuditLogger provides a convenient way to log audit events from handlers.
type AuditLogger struct {
	repo *repository.AuditLogRepository
}

// NewAuditLogger creates a new AuditLogger.
func NewAuditLogger(repo *repository.AuditLogRepository) *AuditLogger {
	return &AuditLogger{repo: repo}
}

// Log creates an audit log entry with information extracted from the Gin context.
func (a *AuditLogger) Log(c *gin.Context, action, resourceType string, resourceID *uuid.UUID, details interface{}) {
	// Extract user information from context (set by JWTAuth middleware)
	var userID *uuid.UUID
	var username string

	if userIDStr, exists := c.Get("user_id"); exists {
		if uid, ok := userIDStr.(string); ok {
			if parsed, err := uuid.Parse(uid); err == nil {
				userID = &parsed
			}
		}
	}

	if un, exists := c.Get("username"); exists {
		if unStr, ok := un.(string); ok {
			username = unStr
		}
	}

	// For device auth (agent actions)
	if username == "" && userID == nil {
		if deviceID, exists := c.Get("device_id"); exists {
			if did, ok := deviceID.(uuid.UUID); ok {
				username = "device:" + did.String()
			}
		}
	}

	// If still no username, it's an unauthenticated action
	if username == "" {
		username = "anonymous"
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Log asynchronously to avoid blocking the response
	go func() {
		_ = a.repo.CreateWithDetails(context.Background(), userID, username, action, resourceType, resourceID, details, ipAddress, userAgent)
	}()
}

// LogAuth is a specialized method for authentication events.
func (a *AuditLogger) LogAuth(c *gin.Context, action string, username string, success bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["success"] = success

	var userID *uuid.UUID
	if success {
		if userIDStr, exists := c.Get("user_id"); exists {
			if uid, ok := userIDStr.(string); ok {
				if parsed, err := uuid.Parse(uid); err == nil {
					userID = &parsed
				}
			}
		}
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	go func() {
		_ = a.repo.CreateWithDetails(context.Background(), userID, username, action, "session", nil, details, ipAddress, userAgent)
	}()
}
