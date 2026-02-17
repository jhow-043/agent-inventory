package handler

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"inventario/server/internal/middleware"
	"inventario/server/internal/service"
	"inventario/shared/dto"
)

// AuthHandler handles enrollment, login, and logout.
type AuthHandler struct {
	service       *service.AuthService
	enrollmentKey string
	auditLogger   *middleware.AuditLogger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc *service.AuthService, enrollmentKey string, auditLogger *middleware.AuditLogger) *AuthHandler {
	return &AuthHandler{service: svc, enrollmentKey: enrollmentKey, auditLogger: auditLogger}
}

// Enroll registers a new agent or re-enrolls an existing one.
func (h *AuthHandler) Enroll(c *gin.Context) {
	key := c.GetHeader("X-Enrollment-Key")
	if key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(h.enrollmentKey)) != 1 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid enrollment key"})
		return
	}

	var req dto.EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	resp, err := h.service.Enroll(c.Request.Context(), &req)
	if err != nil {
		slog.Error("enrollment failed", "error", err, "hostname", req.Hostname)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "enrollment failed"})
		return
	}

	slog.Info("device enrolled", "device_id", resp.DeviceID, "hostname", req.Hostname)
	c.JSON(http.StatusCreated, resp)
}

// Login authenticates a dashboard user and sets a JWT session cookie.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}

	tokenString, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		slog.Warn("login failed", "username", req.Username, "ip", c.ClientIP())
		h.auditLogger.LogAuth(c, "auth.login", req.Username, false, nil)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid credentials"})
		return
	}

	// httpOnly=true, secure=false (HTTP in Phase 1).
	c.SetCookie("session", tokenString, 86400, "/", "", false, true)
	slog.Info("user logged in", "username", req.Username)
	h.auditLogger.LogAuth(c, "auth.login", req.Username, true, nil)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "login successful"})
}

// Logout clears the session cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	username := ""
	if val, exists := c.Get("username"); exists && val != nil {
		username = val.(string)
	}
	if username != "" {
		h.auditLogger.LogAuth(c, "auth.logout", username, true, nil)
	}
	c.SetCookie("session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "logout successful"})
}

// Me returns the currently authenticated user's information.
func (h *AuthHandler) Me(c *gin.Context) {
	sub, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("user_role")
	c.JSON(http.StatusOK, dto.MeResponse{
		ID:       sub.(string),
		Username: username.(string),
		Role:     role.(string),
	})
}
