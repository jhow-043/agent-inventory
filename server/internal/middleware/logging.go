package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging records structured JSON logs for every HTTP request.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start)
		requestID, _ := c.Get("request_id")

		attrs := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"ip", c.ClientIP(),
		}

		if deviceID, exists := c.Get("device_id"); exists {
			attrs = append(attrs, "device_id", deviceID)
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.Log(c.Request.Context(), level, "request completed", attrs...)
	}
}
