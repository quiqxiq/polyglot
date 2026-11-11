package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

// Stream authentication errors. Callers map these to 401/403 responses.
var (
	ErrStreamUnauthenticated = errors.New("stream authentication required")
	ErrStreamForbidden       = errors.New("stream access requires admin role")
)

// StreamAuthenticator authorizes long-lived SSE and WebSocket connections.
// Browsers cannot set Authorization headers on EventSource/WebSocket, so it
// also accepts the httpOnly refresh-token cookie and validates the user's
// current role from the database.
type StreamAuthenticator struct {
	jwt       *JWTService
	refresh   *RefreshTokenService
	users     port.UserRepository
	wsToken   string
	adminOnly bool
}

// NewStreamAuthenticator builds a stream authenticator. All admin-only stream
// endpoints (SSE, terminal WS) use this single policy.
func NewStreamAuthenticator(jwt *JWTService, refresh *RefreshTokenService, users port.UserRepository, wsToken string) *StreamAuthenticator {
	return &StreamAuthenticator{
		jwt:       jwt,
		refresh:   refresh,
		users:     users,
		wsToken:   strings.TrimSpace(wsToken),
		adminOnly: true,
	}
}

// AuthenticateAdmin returns the authenticated admin user ID. A static WS token
// authenticates as user 0 because it is an operator-held secret.
func (a *StreamAuthenticator) AuthenticateAdmin(r *http.Request) (uint, error) {
	if a == nil {
		return 0, ErrStreamUnauthenticated
	}

	if token := bearerToken(r); token != "" {
		return a.authenticateJWT(r, token)
	}

	queryToken := strings.TrimSpace(r.URL.Query().Get("token"))

	if a.wsToken != "" {
		got := strings.TrimSpace(r.URL.Query().Get("ws_token"))
		if got == "" {
			got = queryToken
		}
		if got == "" {
			got = strings.TrimSpace(r.Header.Get("X-Polyglot-WS-Token"))
		}
		if got != "" && subtle.ConstantTimeCompare([]byte(got), []byte(a.wsToken)) == 1 {
			return 0, nil
		}
	}

	if queryToken != "" {
		if uid, err := a.authenticateJWT(r, queryToken); err == nil {
			return uid, nil
		}
	}

	if a.refresh != nil {
		return a.authenticateRefreshCookie(r)
	}
	return 0, ErrStreamUnauthenticated
}

func (a *StreamAuthenticator) authenticateJWT(r *http.Request, token string) (uint, error) {
	if a.jwt == nil {
		return 0, ErrStreamUnauthenticated
	}
	claims, err := a.jwt.ValidateToken(token)
	if err != nil {
		return 0, ErrStreamUnauthenticated
	}
	return a.authorizeUserID(r, claims.UserID)
}

func (a *StreamAuthenticator) authenticateRefreshCookie(r *http.Request) (uint, error) {
	raw := ExtractRefreshTokenCookie(r.Header)
	if raw == "" {
		return 0, ErrStreamUnauthenticated
	}
	claims, err := a.refresh.Validate(r.Context(), raw)
	if err != nil {
		return 0, ErrStreamUnauthenticated
	}
	return a.authorizeUserID(r, claims.UserID)
}

// authorizeUserID performs a fresh active/role check so deactivated or
// re-roled users cannot keep a stream open with stale token claims.
func (a *StreamAuthenticator) authorizeUserID(r *http.Request, userID uint) (uint, error) {
	if a.adminOnly {
		if err := a.AuthorizeAdminUser(r.Context(), userID); err != nil {
			return 0, err
		}
		return userID, nil
	}
	if a.users == nil {
		return 0, ErrStreamUnauthenticated
	}
	user, err := a.users.FindByID(r.Context(), userID)
	if err != nil || user == nil || !user.IsActive {
		return 0, ErrStreamUnauthenticated
	}
	return user.ID, nil
}

// AuthorizeAdminUser re-checks that userID still exists, is active, and holds
// an administrator role. Callers use it for long-lived browser sessions that
// outlive the token that created them.
func (a *StreamAuthenticator) AuthorizeAdminUser(ctx context.Context, userID uint) error {
	if a == nil || a.users == nil || userID == 0 {
		return ErrStreamUnauthenticated
	}
	user, err := a.users.FindByID(ctx, userID)
	if err != nil || user == nil || !user.IsActive {
		return ErrStreamUnauthenticated
	}
	if !customer.IsAdminRole(user.Role) {
		return ErrStreamForbidden
	}
	return nil
}

func bearerToken(r *http.Request) string {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
