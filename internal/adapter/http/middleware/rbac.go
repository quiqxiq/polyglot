package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/auth"
)

func AuthorizeCasbin(enforcer *auth.CasbinEnforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role missing in request context"})
			c.Abort()
			return
		}

		role, ok := roleVal.(string)
		if !ok || role == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user role"})
			c.Abort()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		allowed, err := enforcer.Enforce(role, path, method)
		if err != nil {
			log.Printf("[RBAC Middleware] Error evaluating policy for role '%s' path '%s' method '%s': %v", role, path, method, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to evaluate authorization policy"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied. You do not have permission to perform this action.",
				"role":    role,
				"resource": path,
				"action":  method,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
