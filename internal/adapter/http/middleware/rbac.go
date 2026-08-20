package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/logger"
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
				logger.WithComponent("RBAC").Warnf("enforcer unavailable — denying %q (fail closed)", r.URL.Path)
				writeJSONError(w, http.StatusInternalServerError, "Authorization service unavailable")
				return
			}

			obj, ok := auth.PermissionFor(r.URL.Path)
			if !ok {
				logger.WithComponent("RBAC").Warnf("unknown procedure %q — denying (fail closed)", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":  "Access denied. Unknown resource.",
					"object": r.URL.Path,
				})
				return
			}

			userID, roles, exists := auth.IdentityFromContext(r.Context())
			if !exists || userID == 0 {
				writeJSONError(w, http.StatusUnauthorized, "User identity missing in request context")
				return
			}

			effectiveRoles := rolesForUser(enforcer, fmt.Sprintf("%d", userID), roles)
			if len(effectiveRoles) == 0 {
				writeJSONError(w, http.StatusUnauthorized, "User has no roles assigned")
				return
			}

			allowed := false
			for _, role := range effectiveRoles {
				ok, err := enforcer.Enforce(role, obj, "*")
				if err != nil {
					logger.WithComponent("RBAC").Errorf("error evaluating policy for role '%s' object '%s': %v", role, obj, err)
					writeJSONError(w, http.StatusInternalServerError, "Failed to evaluate authorization policy")
					return
				}
				if ok {
					allowed = true
					break
				}
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":  "Access denied. You do not have permission to perform this action.",
					"roles":  effectiveRoles,
					"object": obj,
				})
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
