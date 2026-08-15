package middleware

import (
	"fmt"
	"net/http"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// AuthorizeCasbin evaluates RBAC permissions against Casbin policies.
func AuthorizeCasbin(enforcer *auth.CasbinEnforcer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enforcer == nil {
				next.ServeHTTP(w, r)
				return
			}

			role := UserRoleFromContext(r.Context())
			if role == "" {
				response.Fail(w, http.StatusUnauthorized, "UNAUTHORIZED", "User role missing in request context", "")
				return
			}

			path := r.URL.Path
			method := r.Method

			allowed, err := enforcer.Enforce(role, path, method)
			if err != nil {
				logger.FromContext(r.Context()).WithFields(logger.Fields{
					"role":   role,
					"path":   path,
					"method": method,
					"error":  err,
				}).Error("Error evaluating RBAC policy")
				response.Fail(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to evaluate authorization policy", "")
				return
			}

			if !allowed {
				response.Fail(w, http.StatusForbidden, "FORBIDDEN",
					fmt.Sprintf("Access denied for role %q on %s %s", role, method, path), "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
