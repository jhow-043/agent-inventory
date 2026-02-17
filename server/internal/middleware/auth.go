package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"inventario/server/internal/repository"
	"inventario/shared/dto"
)

// DeviceAuth validates the device token from the Authorization header.
// On success it sets "device_id" (uuid.UUID) in the Gin context.
func DeviceAuth(tokenRepo *repository.TokenRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "missing or invalid authorization header"})
			return
		}

		rawToken := strings.TrimPrefix(header, "Bearer ")
		hash := SHA256Hex(rawToken)

		token, err := tokenRepo.GetByHash(c.Request.Context(), hash)
		if err != nil {
			slog.Warn("device auth failed", "error", err, "ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid device token"})
			return
		}

		c.Set("device_id", token.DeviceID)
		c.Next()
	}
}

// JWTAuth validates the JWT cookie and extracts user claims.
// On success it sets "user_id" and "username" in the Gin context.
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "authentication required"})
			return
		}

		token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{"HS256"}))

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid token claims"})
			return
		}

		exp, err := claims.GetExpirationTime()
		if err != nil || exp.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "token expired"})
			return
		}

		c.Set("user_id", claims["sub"])
		c.Set("username", claims["username"])

		// Extract role from claims (with fallback to viewer for older tokens)
		role, _ := claims["role"].(string)
		if role == "" {
			role = "viewer"
		}
		c.Set("user_role", role)

		c.Next()
	}
}

// SHA256Hex computes the SHA-256 hex digest of the given string.
func SHA256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
