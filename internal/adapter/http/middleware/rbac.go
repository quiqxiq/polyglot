package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/port"
)

// RBACRequired creates a Gin middleware that enforces role-based access control.
// It assumes AuthRequired has already successfully authenticated the request and
// placed the user's role in the Gin context.
func RBACRequired(enforcer port.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role information missing from context"})
			return
		}

		role, ok := roleVal.(string)
		if !ok || role == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid role type in context"})
			return
		}

		// Use the request URL Path as the resource, and HTTP Method as the action.
		resource := c.Request.URL.Path
		action := c.Request.Method

		allowed, err := enforcer.Enforce(c.Request.Context(), role, resource, action)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error during authorization check"})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
			return
		}

		c.Next()
	}
}
