package middleware

import (
	"fmt"
	"net/http"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// PolicyEnforcer is the subset of auth.CasbinEnforcer the RBAC middleware needs.
type PolicyEnforcer interface {
	Enforce(sub, obj, act string) (bool, error)
	GetRolesForUser(user string) ([]string, error)
}

// AuthorizeProcedure enforces Casbin RBAC on ConnectRPC procedures using standard http.Handler.
func AuthorizeProcedure(enforcer PolicyEnforcer) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enforcer == nil {
				logger.WithComponent("RBAC").WithField("path", r.URL.Path).Warn("enforcer unavailable; denying request")
				response.WriteHTTPStatusError(w, http.StatusInternalServerError, "Authorization service unavailable")
				return
			}

			obj, ok := auth.PermissionFor(r.URL.Path)
			if !ok {
				logger.WithComponent("RBAC").WithField("path", r.URL.Path).Warn("unknown procedure; denying request")
				response.WriteHTTPStatusError(w, http.StatusForbidden, "Access denied. Unknown resource.")
				return
			}

			userID, roles, exists := auth.IdentityFromContext(r.Context())
			if !exists || userID == 0 {
				response.WriteHTTPStatusError(w, http.StatusUnauthorized, "User identity missing in request context")
				return
			}

			effectiveRoles := rolesForUser(enforcer, fmt.Sprintf("%d", userID), roles)
			if len(effectiveRoles) == 0 {
				response.WriteHTTPStatusError(w, http.StatusUnauthorized, "User has no roles assigned")
				return
			}

			allowed := false
			for _, role := range effectiveRoles {
				ok, err := enforcer.Enforce(role, obj, "*")
				if err != nil {
					logger.WithComponent("RBAC").WithError(err).WithFields(map[string]any{"role": role, "object": obj}).Error("policy evaluation failed")
					response.WriteHTTPStatusError(w, http.StatusInternalServerError, "Failed to evaluate authorization policy")
					return
				}
				if ok {
					allowed = true
					break
				}
			}

			if !allowed {
				response.WriteHTTPStatusError(w, http.StatusForbidden, "Access denied. You do not have permission to perform this action.")
				return
			}

			ctx := auth.WithIdentity(r.Context(), userID, effectiveRoles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func rolesForUser(enforcer PolicyEnforcer, userIDStr string, fallbackRoles []string) []string {
	if enforcer == nil {
		return fallbackRoles
	}
	roles, err := enforcer.GetRolesForUser(userIDStr)
	if err != nil || len(roles) == 0 {
		return fallbackRoles
	}
	return roles
}
