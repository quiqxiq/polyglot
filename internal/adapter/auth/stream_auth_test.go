package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

type stubUserRepo struct {
	port.UserRepository
	user *customer.User
	err  error
}

func (s stubUserRepo) FindByID(context.Context, uint) (*customer.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

type memoryRefreshStore struct {
	values map[string]string
}

func (m *memoryRefreshStore) Set(_ context.Context, key, value string, _ int) error {
	m.values[key] = value
	return nil
}

func (m *memoryRefreshStore) Get(_ context.Context, key string) (string, error) {
	return m.values[key], nil
}

func (m *memoryRefreshStore) Delete(_ context.Context, key string) error {
	delete(m.values, key)
	return nil
}

func newStreamRequest(t *testing.T, jwt *JWTService, role string, userID uint) *http.Request {
	t.Helper()
	token, err := jwt.GenerateToken(userID, "admin@example.com", []string{role}, "default")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestStreamAuthenticator_JWTAdmin(t *testing.T) {
	jwt := NewJWTService("0123456789abcdef0123456789abcdef", 1)
	users := stubUserRepo{user: &customer.User{ID: 7, Role: "admin", IsActive: true}}
	authz := NewStreamAuthenticator(jwt, nil, users, "")

	userID, err := authz.AuthenticateAdmin(newStreamRequest(t, jwt, "admin", 7))
	require.NoError(t, err)
	assert.Equal(t, uint(7), userID)
}

func TestStreamAuthenticator_RejectsNonAdminAndInactive(t *testing.T) {
	jwt := NewJWTService("0123456789abcdef0123456789abcdef", 1)

	nonAdmin := NewStreamAuthenticator(jwt, nil, stubUserRepo{
		user: &customer.User{ID: 8, Role: "teknisi", IsActive: true},
	}, "")
	_, err := nonAdmin.AuthenticateAdmin(newStreamRequest(t, jwt, "teknisi", 8))
	require.ErrorIs(t, err, ErrStreamForbidden)

	inactive := NewStreamAuthenticator(jwt, nil, stubUserRepo{
		user: &customer.User{ID: 9, Role: "admin", IsActive: false},
	}, "")
	_, err = inactive.AuthenticateAdmin(newStreamRequest(t, jwt, "admin", 9))
	require.ErrorIs(t, err, ErrStreamUnauthenticated)

	// A fresh DB role change wins over stale JWT claims.
	demoted := NewStreamAuthenticator(jwt, nil, stubUserRepo{
		user: &customer.User{ID: 10, Role: "agent", IsActive: true},
	}, "")
	_, err = demoted.AuthenticateAdmin(newStreamRequest(t, jwt, "admin", 10))
	require.ErrorIs(t, err, ErrStreamForbidden)
}

func TestStreamAuthenticator_StaticWSToken(t *testing.T) {
	authz := NewStreamAuthenticator(nil, nil, stubUserRepo{}, "operator-secret")

	req := httptest.NewRequest(http.MethodGet, "/ws/devices/dev1/terminal?ws_token=operator-secret", nil)
	userID, err := authz.AuthenticateAdmin(req)
	require.NoError(t, err)
	assert.Zero(t, userID)

	reqTokenParam := httptest.NewRequest(http.MethodGet, "/ws/devices/dev1/terminal?token=operator-secret", nil)
	userID, err = authz.AuthenticateAdmin(reqTokenParam)
	require.NoError(t, err)
	assert.Zero(t, userID)

	bad := httptest.NewRequest(http.MethodGet, "/ws/devices/dev1/terminal?ws_token=wrong", nil)
	_, err = authz.AuthenticateAdmin(bad)
	require.ErrorIs(t, err, ErrStreamUnauthenticated)
}

func TestStreamAuthenticator_JWTQueryToken(t *testing.T) {
	jwt := NewJWTService("0123456789abcdef0123456789abcdef", 1)
	users := stubUserRepo{user: &customer.User{ID: 7, Role: "admin", IsActive: true}}
	authz := NewStreamAuthenticator(jwt, nil, users, "")

	token, err := jwt.GenerateToken(7, "admin@example.com", []string{"admin"}, "default")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/ws/devices/dev1/terminal?token="+token, nil)
	userID, err := authz.AuthenticateAdmin(req)
	require.NoError(t, err)
	assert.Equal(t, uint(7), userID)
}

func TestStreamAuthenticator_RefreshCookie(t *testing.T) {
	jwt := NewJWTService("0123456789abcdef0123456789abcdef", 1)
	store := &memoryRefreshStore{values: map[string]string{}}
	refresh := NewRefreshTokenService(store, 0)
	token, err := refresh.Issue(context.Background(), RefreshClaims{UserID: 11, Roles: []string{"admin"}})
	require.NoError(t, err)

	authz := NewStreamAuthenticator(jwt, refresh, stubUserRepo{
		user: &customer.User{ID: 11, Role: "owner", IsActive: true},
	}, "")
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.AddCookie(&http.Cookie{Name: RefreshTokenCookieName, Value: token})

	userID, err := authz.AuthenticateAdmin(req)
	require.NoError(t, err)
	assert.Equal(t, uint(11), userID)
}

func TestStreamAuthenticator_UserLookupFailure(t *testing.T) {
	jwt := NewJWTService("0123456789abcdef0123456789abcdef", 1)
	authz := NewStreamAuthenticator(jwt, nil, stubUserRepo{err: errors.New("db down")}, "")
	_, err := authz.AuthenticateAdmin(newStreamRequest(t, jwt, "admin", 12))
	require.ErrorIs(t, err, ErrStreamUnauthenticated)
}

func TestStreamAuthenticator_AuthorizeAdminUser(t *testing.T) {
	tests := []struct {
		name    string
		repo    stubUserRepo
		userID  uint
		wantErr error
	}{
		{
			name:   "active admin",
			repo:   stubUserRepo{user: &customer.User{ID: 20, Role: "admin", IsActive: true}},
			userID: 20,
		},
		{
			name:   "active owner",
			repo:   stubUserRepo{user: &customer.User{ID: 21, Role: "owner", IsActive: true}},
			userID: 21,
		},
		{
			name:    "non-admin",
			repo:    stubUserRepo{user: &customer.User{ID: 22, Role: "teknisi", IsActive: true}},
			userID:  22,
			wantErr: ErrStreamForbidden,
		},
		{
			name:    "inactive admin",
			repo:    stubUserRepo{user: &customer.User{ID: 23, Role: "admin", IsActive: false}},
			userID:  23,
			wantErr: ErrStreamUnauthenticated,
		},
		{
			name:    "missing user",
			repo:    stubUserRepo{user: nil},
			userID:  24,
			wantErr: ErrStreamUnauthenticated,
		},
		{
			name:    "lookup failure",
			repo:    stubUserRepo{err: errors.New("db down")},
			userID:  25,
			wantErr: ErrStreamUnauthenticated,
		},
		{
			name:    "zero user id",
			repo:    stubUserRepo{user: &customer.User{ID: 26, Role: "admin", IsActive: true}},
			userID:  0,
			wantErr: ErrStreamUnauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz := NewStreamAuthenticator(nil, nil, tt.repo, "")
			err := authz.AuthorizeAdminUser(context.Background(), tt.userID)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
