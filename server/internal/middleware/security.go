package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds standard security headers to every response.
// corsOrigins is used to dynamically build the CSP connect-src directive
// so that allowed API origins don't need to be hardcoded.
func SecurityHeaders(corsOrigins []string) gin.HandlerFunc {
	connectSrc := "'self'"
	if len(corsOrigins) > 0 {
		connectSrc += " " + strings.Join(corsOrigins, " ")
	}

	csp := strings.Join([]string{
		"default-src 'self'",
		"script-src 'self'",
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data:",
		"font-src 'self'",
		"connect-src " + connectSrc,
	}, "; ")

	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", csp)
		c.Next()
	}
}
