package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"inventario/server/internal/middleware"
	"inventario/server/internal/service"
	"inventario/shared/dto"
)

// UserHandler handles user management endpoints.
type UserHandler struct {
	authService *service.AuthService
	auditLogger *middleware.AuditLogger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(authSvc *service.AuthService, auditLogger *middleware.AuditLogger) *UserHandler {
	return &UserHandler{authService: authSvc, auditLogger: auditLogger}
}

// ListUsers returns all dashboard users.
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers(c.Request.Context())
	if err != nil {
		slog.Error("failed to list users", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list users"})
		return
	}

	resp := dto.UserListResponse{
		Users: make([]dto.UserResponse, 0, len(users)),
		Total: len(users),
	}
	for _, u := range users {
		resp.Users = append(resp.Users, dto.UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Role:      u.Role,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, resp)
}

// CreateUser creates a new dashboard user.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Default to viewer if role not specified
	role := req.Role
	if role == "" {
		role = "viewer"
	}

	if err := h.authService.CreateUser(c.Request.Context(), req.Username, req.Password, role); err != nil {
		slog.Error("failed to create user", "error", err)
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "username already exists or invalid role"})
		return
	}

	h.auditLogger.Log(c, "user.create", "user", nil, map[string]interface{}{"username": req.Username, "role": role})
	c.JSON(http.StatusCreated, dto.MessageResponse{Message: "user created successfully"})
}

// DeleteUser deletes a dashboard user by ID.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user ID"})
		return
	}

	// Extract the requesting user's ID from JWT claims.
	sub, _ := c.Get("user_id")
	requestingUserID, err := uuid.Parse(sub.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "invalid session"})
		return
	}

	if err := h.authService.DeleteUser(c.Request.Context(), requestingUserID, targetID); err != nil {
		slog.Error("failed to delete user", "error", err, "target_id", targetID)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	h.auditLogger.Log(c, "user.delete", "user", &targetID, map[string]interface{}{"target_user_id": targetID})
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "user deleted successfully"})
}
