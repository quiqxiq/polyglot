package middleware

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/auth"
)

// PolicyEnforcer is the subset of auth.CasbinEnforcer the RBAC middleware
// needs. Declared here (not in port/) because this is a gin-specific concern;
// auth.CasbinEnforcer satisfies it.
type PolicyEnforcer interface {
	Enforce(sub, obj, act string) (bool, error)
	GetRolesForUser(user string) ([]string, error)
}

// AuthorizeProcedure enforces Casbin RBAC on ConnectRPC procedures. The URL
// path IS the procedure (e.g. /polyglot.v1.KnowledgeService/CreateKnowledge),
// which the registry maps to a "resource:action" object. Roles are resolved
// from the Casbin group assignments for the authenticated user (set by
// AuthenticateJWT); the single-role JWT claim is only a fallback.
//
// On success it also propagates the identity (userID + roles) into the
// request context so downstream ConnectRPC handlers can read it via
// req.HTTP().Context() (see auth.IdentityFromContext).
func AuthorizeProcedure(enforcer PolicyEnforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Fail-closed: tanpa enforcer, authorization tidak bisa dievaluasi —
		// tolak semua daripada membiarkan akses tanpa kontrol.
		if enforcer == nil {
			log.Printf("[RBAC] Enforcer tidak tersedia — menolak %q (fail closed)", c.Request.URL.Path)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization service unavailable"})
			c.Abort()
			return
		}

		obj, ok := auth.PermissionFor(c.Request.URL.Path)
		if !ok {
			log.Printf("[RBAC] Unknown procedure %q — denying (fail closed)", c.Request.URL.Path)
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Access denied. Unknown resource.",
				"object": c.Request.URL.Path,
			})
			c.Abort()
			return
		}

		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User identity missing in request context"})
			c.Abort()
			return
		}
		userID, ok := userIDVal.(uint)
		if !ok || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user identity"})
			c.Abort()
			return
		}

		// Roles = Casbin group assignments (multi-role, source of truth).
		// Fallback ke klaim JWT tunggal kalau user belum di-assign ke grup.
		roles := rolesForUser(enforcer, fmt.Sprintf("%d", userID), c.GetString("user_role"))
		if len(roles) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User has no roles assigned"})
			c.Abort()
			return
		}

		allowed := false
		for _, role := range roles {
			ok, err := enforcer.Enforce(role, obj, "*")
			if err != nil {
				log.Printf("[RBAC] Error evaluating policy for role '%s' object '%s': %v", role, obj, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to evaluate authorization policy"})
				c.Abort()
				return
			}
			if ok {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Access denied. You do not have permission to perform this action.",
				"role":   roles,
				"object": obj,
			})
			c.Abort()
			return
		}

		// Propagate identity ke connect handler lewat request context.
		c.Request = c.Request.WithContext(auth.WithIdentity(c.Request.Context(), userID, roles))
		c.Next()
	}
}

// rolesForUser resolves roles dari Casbin group (multi-role), fallback ke
// klaim JWT tunggal bila belum ada assignment.
func rolesForUser(enforcer PolicyEnforcer, userID string, fallbackRole string) []string {
	if enforcer != nil {
		if roles, err := enforcer.GetRolesForUser(userID); err == nil && len(roles) > 0 {
			return roles
		}
	}
	if fallbackRole != "" {
		return []string{fallbackRole}
	}
	return nil
}
