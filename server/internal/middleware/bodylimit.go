package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySize limits the size of the request body to prevent OOM attacks.
// size is the maximum allowed body size in bytes.
func MaxBodySize(size int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, size)
		}
		c.Next()
	}
}

// DefaultMaxBodySize returns a middleware that limits the body to 10MB.
func DefaultMaxBodySize() gin.HandlerFunc {
	return MaxBodySize(10 << 20) // 10 MB
}
