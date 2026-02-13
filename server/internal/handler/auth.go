package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"inventario/server/internal/service"
	"inventario/shared/dto"
)

// AuthHandler handles enrollment, login, and logout.
type AuthHandler struct {
	service       *service.AuthService
	enrollmentKey string
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc *service.AuthService, enrollmentKey string) *AuthHandler {
	return &AuthHandler{service: svc, enrollmentKey: enrollmentKey}
}

// Enroll registers a new agent or re-enrolls an existing one.
func (h *AuthHandler) Enroll(c *gin.Context) {
	key := c.GetHeader("X-Enrollment-Key")
	if key == "" || key != h.enrollmentKey {
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
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid credentials"})
		return
	}

	// httpOnly=true, secure=false (HTTP in Phase 1).
	c.SetCookie("session", tokenString, 86400, "/", "", false, true)
	slog.Info("user logged in", "username", req.Username)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "login successful"})
}

// Logout clears the session cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "logout successful"})
}
