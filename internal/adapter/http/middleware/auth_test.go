package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/auth"
)

const serviceToken = "service-read-token"

func serviceTokenHandler() http.Handler {
	jwtService := auth.NewJWTService("jwt-secret", 1)
	authenticate := AuthenticateJWTOrServiceToken(jwtService, serviceToken, []string{"/api/devices"})
	return authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, roles, ok := auth.IdentityFromContext(r.Context())
		if !ok {
			http.Error(w, "identity missing", http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-User-ID", auth.IdentityStringFromContext(r.Context()))
		w.Header().Set("X-Role", roles[0])
		w.WriteHeader(http.StatusOK)
		_ = userID
	}))
}

func request(method, target, remoteAddr, bearer string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	req.RemoteAddr = remoteAddr
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	return req
}

func TestAuthenticateJWTOrServiceToken_AllowsLoopbackServiceToken(t *testing.T) {
	rec := httptest.NewRecorder()
	serviceTokenHandler().ServeHTTP(rec, request(http.MethodGet, "/api/devices", "127.0.0.1:51234", serviceToken))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, auth.RoleService, rec.Header().Get("X-Role"))
}

func TestAuthenticateJWTOrServiceToken_AllowsIPv6Loopback(t *testing.T) {
	rec := httptest.NewRecorder()
	serviceTokenHandler().ServeHTTP(rec, request(http.MethodGet, "/api/devices", "[::1]:51234", serviceToken))

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthenticateJWTOrServiceToken_RejectsRemotePeer(t *testing.T) {
	rec := httptest.NewRecorder()
	serviceTokenHandler().ServeHTTP(rec, request(http.MethodGet, "/api/devices", "203.0.113.7:51234", serviceToken))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticateJWTOrServiceToken_RejectsTokenOutsideServicePaths(t *testing.T) {
	rec := httptest.NewRecorder()
	serviceTokenHandler().ServeHTTP(rec, request(http.MethodGet, "/api/cashier/gateways", "127.0.0.1:51234", serviceToken))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticateJWTOrServiceToken_RejectsWrongToken(t *testing.T) {
	rec := httptest.NewRecorder()
	serviceTokenHandler().ServeHTTP(rec, request(http.MethodGet, "/api/devices", "127.0.0.1:51234", "wrong-token"))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticateJWTOrServiceToken_FallsBackToJWT(t *testing.T) {
	jwtService := auth.NewJWTService("jwt-secret", 1)
	token, err := jwtService.GenerateToken(7, "admin@example.test", []string{"admin"}, "default")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	serviceTokenHandler().ServeHTTP(rec, request(http.MethodGet, "/api/devices", "203.0.113.7:51234", token))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "admin", rec.Header().Get("X-Role"))
	assert.Equal(t, "7", rec.Header().Get("X-User-ID"))
}

func TestAuthenticateJWTOrServiceToken_MissingHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	serviceTokenHandler().ServeHTTP(rec, request(http.MethodGet, "/api/devices", "127.0.0.1:51234", ""))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticateJWTOrServiceToken_EmptyTokenDisablesServicePath(t *testing.T) {
	jwtService := auth.NewJWTService("jwt-secret", 1)
	authenticate := AuthenticateJWTOrServiceToken(jwtService, "  ", []string{"/api/devices"})
	handler := authenticate(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, request(http.MethodGet, "/api/devices", "127.0.0.1:51234", ""))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
