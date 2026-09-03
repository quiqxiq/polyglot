package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/session"
	"github.com/quixiq/polyglot/internal/port"
)

// ErrInvalidRefreshToken indicates the presented refresh token is unknown,
// expired, or malformed — the caller should treat the session as invalid.
var ErrInvalidRefreshToken = session.ErrInvalidRefreshToken

// RefreshTokenLifetime is the default 7-day TTL.
const RefreshTokenLifetime = 7 * 24 * time.Hour

// RefreshTokenCookieName is the httpOnly cookie carrying the opaque refresh token.
const RefreshTokenCookieName = "polyglot_refresh"

// RefreshClaims is the payload bound to an issued refresh token.
type RefreshClaims struct {
	UserID   uint     `json:"user_id"`
	Roles    []string `json:"roles,omitempty"`
	TenantID string   `json:"tenant_id,omitempty"`
	IssuedAt int64    `json:"issued_at"`
}

// RefreshTokenService issues, validates, rotates and revokes opaque refresh tokens.
type RefreshTokenService struct {
	store port.RefreshTokenStore
	ttl   time.Duration
}

var _ port.RefreshTokenManager = (*RefreshTokenService)(nil)

// NewRefreshTokenService creates the service with the given store and TTL.
func NewRefreshTokenService(store port.RefreshTokenStore, ttl time.Duration) *RefreshTokenService {
	if ttl <= 0 {
		ttl = RefreshTokenLifetime
	}
	return &RefreshTokenService{store: store, ttl: ttl}
}

// IssueToken implements port.RefreshTokenManager.
func (s *RefreshTokenService) IssueToken(ctx context.Context, userID uint, roles []string, tenantID string) (string, error) {
	return s.Issue(ctx, RefreshClaims{
		UserID:   userID,
		Roles:    roles,
		TenantID: tenantID,
		IssuedAt: time.Now().Unix(),
	})
}

// RotateToken implements port.RefreshTokenManager.
func (s *RefreshTokenService) RotateToken(ctx context.Context, oldToken string) (string, uint, []string, string, error) {
	newToken, claims, err := s.Rotate(ctx, oldToken)
	if err != nil {
		return "", 0, nil, "", err
	}
	return newToken, claims.UserID, claims.Roles, claims.TenantID, nil
}

// RevokeToken implements port.RefreshTokenManager.
func (s *RefreshTokenService) RevokeToken(ctx context.Context, token string) error {
	return s.Revoke(ctx, token)
}

// Issue creates a new opaque refresh token bound to the given identity.
func (s *RefreshTokenService) Issue(ctx context.Context, claims RefreshClaims) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal refresh claims: %w", err)
	}
	if err := s.store.Set(ctx, refreshKey(token), string(payload), int(s.ttl.Seconds())); err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}
	return token, nil
}

// Validate checks the token against the store and returns its claims.
func (s *RefreshTokenService) Validate(ctx context.Context, token string) (RefreshClaims, error) {
	if token == "" {
		return RefreshClaims{}, ErrInvalidRefreshToken
	}
	raw, err := s.store.Get(ctx, refreshKey(token))
	if err != nil {
		return RefreshClaims{}, ErrInvalidRefreshToken
	}
	var claims RefreshClaims
	if err := json.Unmarshal([]byte(raw), &claims); err != nil {
		return RefreshClaims{}, ErrInvalidRefreshToken
	}
	if claims.UserID == 0 {
		return RefreshClaims{}, ErrInvalidRefreshToken
	}
	return claims, nil
}

// Rotate validates the old token, revokes it, and issues a fresh one with the same identity.
func (s *RefreshTokenService) Rotate(ctx context.Context, oldToken string) (string, RefreshClaims, error) {
	claims, err := s.Validate(ctx, oldToken)
	if err != nil {
		return "", RefreshClaims{}, err
	}
	if err := s.store.Delete(ctx, refreshKey(oldToken)); err != nil {
		return "", RefreshClaims{}, fmt.Errorf("revoke old refresh token: %w", err)
	}
	newToken, err := s.Issue(ctx, claims)
	if err != nil {
		return "", RefreshClaims{}, err
	}
	return newToken, claims, nil
}

// Revoke invalidates a refresh token (logout).
func (s *RefreshTokenService) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.Delete(ctx, refreshKey(token))
}

// TTL returns the refresh token lifetime.
func (s *RefreshTokenService) TTL() time.Duration {
	return s.ttl
}

func refreshKey(token string) string {
	return "refresh:" + token
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// SetRefreshTokenCookie sets the refresh token cookie in http.Header.
func SetRefreshTokenCookie(h http.Header, token string, ttl time.Duration, secure bool) {
	SetRefreshCookieHeader(h, token, ttl, secure)
}

// ClearRefreshTokenCookie clears the refresh token cookie in http.Header.
func ClearRefreshTokenCookie(h http.Header, secure bool) {
	ClearRefreshCookieHeader(h, secure)
}

// ExtractRefreshTokenCookie extracts the refresh token from http.Header.
func ExtractRefreshTokenCookie(h http.Header) string {
	return ReadRefreshToken(h, "")
}

// SetRefreshCookieWriter writes the httpOnly refresh cookie on a response writer.
func SetRefreshCookieWriter(w http.ResponseWriter, token string, ttl time.Duration, secure bool) {
	http.SetCookie(w, refreshCookie(token, ttl, secure))
}

// SetRefreshCookieHeader adds the httpOnly refresh cookie to a response header map.
func SetRefreshCookieHeader(h http.Header, token string, ttl time.Duration, secure bool) {
	if c := refreshCookie(token, ttl, secure); c != nil {
		h.Add("Set-Cookie", c.String())
	}
}

// ClearRefreshCookieWriter expires the refresh cookie on a response writer.
func ClearRefreshCookieWriter(w http.ResponseWriter) {
	http.SetCookie(w, expiredRefreshCookie())
}

// ClearRefreshCookieHeader adds an expired refresh cookie to a header map.
func ClearRefreshCookieHeader(h http.Header, secure bool) {
	if c := expiredRefreshCookie(); c != nil {
		h.Add("Set-Cookie", c.String())
	}
}

func refreshCookie(token string, ttl time.Duration, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	}
}

func expiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// ReadRefreshToken extracts the refresh token from the httpOnly cookie.
func ReadRefreshToken(h http.Header, bodyValue string) string {
	if cookieHeader := h.Get("Cookie"); cookieHeader != "" {
		if cookies, err := http.ParseCookie(cookieHeader); err == nil {
			for _, c := range cookies {
				if c.Name == RefreshTokenCookieName && c.Value != "" {
					return c.Value
				}
			}
		}
	}
	return strings.TrimSpace(bodyValue)
}

// UserIDToRef converts a numeric user ID to the string form used in Casbin.
func UserIDToRef(userID uint) string {
	return strconv.FormatUint(uint64(userID), 10)
}
