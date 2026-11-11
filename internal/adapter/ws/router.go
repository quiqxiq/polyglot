package ws

import (
	"errors"
	"net/http"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/logger"
)

// AdminAuthenticator authenticates an SSE or WebSocket request and returns the
// admin user ID. Implementations must enforce owner/superadmin/admin only.
type AdminAuthenticator interface {
	AuthenticateAdmin(r *http.Request) (uint, error)
}

// RegisterEventRoutes registers the realtime SSE and terminal WebSocket
// endpoints. Both are admin-only; a nil authenticator fails closed.
func RegisterEventRoutes(
	mux *http.ServeMux,
	sseHub *SSEHub,
	openTermUC *network.OpenTerminalUseCase,
	authz AdminAuthenticator,
	allowedOrigins []string,
) {
	mux.Handle("GET /events", requireAdminStream(authz, http.HandlerFunc(sseHub.ServeHTTP)))

	termHandler := NewTerminalHandler(openTermUC).WithAllowedOrigins(allowedOrigins)
	mux.Handle("GET /ws/devices/{id}/terminal", requireAdminStream(authz, http.HandlerFunc(termHandler.ServeHTTP)))
}

func requireAdminStream(authz AdminAuthenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authz == nil {
			http.Error(w, "stream authentication is not configured", http.StatusServiceUnavailable)
			return
		}
		userID, err := authz.AuthenticateAdmin(r)
		if err != nil {
			if errors.Is(err, auth.ErrStreamForbidden) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if userID != 0 {
			logger.WithComponent("StreamAuth").WithField("user_id", userID).Debug("admin stream authenticated")
		}
		next.ServeHTTP(w, r)
	})
}
