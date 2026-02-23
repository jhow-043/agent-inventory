package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"inventario/shared/dto"
)

// RequireRole ensures the authenticated user has one of the specified roles.
// Must be used after JWTAuth middleware.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "role information not found"})
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "invalid role format"})
			return
		}

		// Check if user's role is in allowed roles
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "insufficient permissions"})
	}
}
