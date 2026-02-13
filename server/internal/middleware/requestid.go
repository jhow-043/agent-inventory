// Package middleware provides HTTP middleware for the Gin router.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-Id"

// RequestID generates a unique request ID for every incoming request
// and sets it both in the Gin context and as a response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.New().String()
		c.Set("request_id", id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}
