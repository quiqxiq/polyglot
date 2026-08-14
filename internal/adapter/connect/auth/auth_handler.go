package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/crypto/bcrypt"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/adapter/auth"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
)

const BearerScheme = "Bearer"

// MaxLoginAttemptsPerWindow caps failed login attempts before lockout.
const MaxLoginAttemptsPerWindow = 5

// LoginRateLimitWindow is the sliding window for login attempts.
const LoginRateLimitWindow = "15m"

// LoginRateLimiter is the subset of the Redis store used for login throttling.
type LoginRateLimiter interface {
	IncrementRateLimit(ctx context.Context, scope, window string, ttl time.Duration) (int64, error)
	GetRateLimitCount(ctx context.Context, scope, window string) (int64, error)
	ResetRateLimit(ctx context.Context, scope, window string) error
}

// RoleResolver resolves the roles of a user from Casbin (multi-role), falling
// back to the single role column for users without group assignments.
type RoleResolver interface {
	GetRolesForUser(user string) ([]string, error)
}

type AuthConnectHandler struct {
	pgStore    *postgres.Store
	jwtService *auth.JWTService
	refreshSvc *auth.RefreshTokenService
	rateLimit  LoginRateLimiter
	roles      RoleResolver
	secure     bool
}

func NewAuthConnectHandler(
	pgStore *postgres.Store,
	jwtService *auth.JWTService,
	refreshSvc *auth.RefreshTokenService,
	rateLimit LoginRateLimiter,
	roles RoleResolver,
	secure bool,
) *AuthConnectHandler {
	return &AuthConnectHandler{
		pgStore:    pgStore,
		jwtService: jwtService,
		refreshSvc: refreshSvc,
		rateLimit:  rateLimit,
		roles:      roles,
		secure:     secure,
	}
}

func (h *AuthConnectHandler) Login(ctx context.Context, req *connect.Request[devicepb.LoginRequest]) (*connect.Response[devicepb.LoginResponse], error) {
	if req.Msg.Username == "" || req.Msg.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username and password are required"))
	}

	if h.pgStore == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("database store unavailable"))
	}

	// Rate limit: 5 percobaan / 15 menit per username+IP. Dilakukan SEBELUM
	// bcrypt supaya attacker tidak bisa memakai endpoint ini sebagai oracle
	// timing/cost (doS). Scope memakai username supaya lockout mengikuti akun,
	// bukan sekadar IP yang bisa berganti-ganti.
	clientIP := clientIPFromRequest(req)
	rlScope := "login:" + req.Msg.Username + ":" + clientIP
	if h.rateLimit != nil {
		attempts, err := h.rateLimit.GetRateLimitCount(ctx, rlScope, LoginRateLimitWindow)
		if err == nil && attempts >= MaxLoginAttemptsPerWindow {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("too many login attempts, try again later"))
		}
	}

	user, err := h.pgStore.FindUserByUsername(req.Msg.Username)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			h.recordFailedLogin(ctx, rlScope)
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid username or password"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Msg.Password)); err != nil {
		h.recordFailedLogin(ctx, rlScope)
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid username or password"))
	}

	// Login sukses → reset counter.
	if h.rateLimit != nil {
		_ = h.rateLimit.ResetRateLimit(ctx, rlScope, LoginRateLimitWindow)
	}

	// Roles dari Casbin g (multi-role, source of truth); fallback ke kolom
	// role tunggal bila user belum punya assignment grup.
	roles := h.resolveRoles(user.ID, user.Role)

	accessToken, err := h.jwtService.GenerateToken(user.ID, user.Email, roles, user.TenantID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate auth token"))
	}

	// Issue refresh token + set cookie httpOnly. Cookie dikirim via response
	// header (browser otomatis mengirimnya di request berikutnya).
	var refreshToken string
	if h.refreshSvc != nil {
		refreshToken, err = h.refreshSvc.Issue(ctx, auth.RefreshClaims{
			UserID:   user.ID,
			Roles:    roles,
			TenantID: user.TenantID,
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to issue refresh token"))
		}
	}

	res := connect.NewResponse(&devicepb.LoginResponse{
		Token: accessToken,
		User: &devicepb.UserProfile{
			Id:       fmt.Sprintf("%d", user.ID),
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
			Roles:    roles,
		},
		ExpiresAtUnix: time.Now().Add(h.jwtService.ExpiryDuration()).Unix(),
	})
	if h.refreshSvc != nil && refreshToken != "" {
		auth.SetRefreshCookieHeader(res.Header(), refreshToken, h.refreshSvc.TTL(), h.secure)
	}
	return res, nil
}

func (h *AuthConnectHandler) GetMe(ctx context.Context, req *connect.Request[devicepb.GetMeRequest]) (*connect.Response[devicepb.GetMeResponse], error) {
	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("unauthorized: missing authorization header"))
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("unauthorized: invalid authorization header format"))
	}

	claims, err := h.jwtService.ValidateToken(parts[1])
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("unauthorized: %w", err))
	}

	if h.pgStore != nil {
		user, err := h.pgStore.FindUserByID(claims.UserID)
		if err == nil && user != nil {
			return connect.NewResponse(&devicepb.GetMeResponse{
				User: &devicepb.UserProfile{
					Id:       fmt.Sprintf("%d", user.ID),
					Username: user.Username,
					Email:    user.Email,
					Role:     user.Role,
					Roles:    claims.Roles,
				},
			}), nil
		}
	}

	return connect.NewResponse(&devicepb.GetMeResponse{
		User: &devicepb.UserProfile{
			Id:       fmt.Sprintf("%d", claims.UserID),
			Username: claims.Email,
			Email:    claims.Email,
			Role:     claims.Role,
			Roles:    claims.Roles,
		},
	}), nil
}

func (h *AuthConnectHandler) RefreshToken(ctx context.Context, req *connect.Request[devicepb.RefreshTokenRequest]) (*connect.Response[devicepb.RefreshTokenResponse], error) {
	if h.refreshSvc == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("refresh token service unavailable"))
	}

	// Refresh token diambil dari cookie httpOnly (browser) atau body (non-browser).
	oldToken := auth.ReadRefreshToken(req.Header(), req.Msg.RefreshToken)
	if oldToken == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing refresh token"))
	}

	newToken, claims, err := h.refreshSvc.Rotate(ctx, oldToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid or expired refresh token"))
	}

	// Rotasi: identity diambil dari refresh claims (bukan dari token lama yang
	// bisa di-forge), akses token baru di-generate.
	accessToken, err := h.jwtService.GenerateToken(claims.UserID, "", claims.Roles, claims.TenantID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate auth token"))
	}

	res := connect.NewResponse(&devicepb.RefreshTokenResponse{
		Token:                accessToken,
		ExpiresAtUnix:        time.Now().Add(h.jwtService.ExpiryDuration()).Unix(),
		RefreshExpiresAtUnix: time.Now().Add(h.refreshSvc.TTL()).Unix(),
	})
	auth.SetRefreshCookieHeader(res.Header(), newToken, h.refreshSvc.TTL(), h.secure)
	return res, nil
}

func (h *AuthConnectHandler) Logout(ctx context.Context, req *connect.Request[devicepb.LogoutRequest]) (*connect.Response[devicepb.LogoutResponse], error) {
	if h.refreshSvc != nil {
		token := auth.ReadRefreshToken(req.Header(), "")
		_ = h.refreshSvc.Revoke(ctx, token)
	}
	res := connect.NewResponse(&devicepb.LogoutResponse{Success: true})
	auth.ClearRefreshCookieHeader(res.Header())
	return res, nil
}

func (h *AuthConnectHandler) recordFailedLogin(ctx context.Context, scope string) {
	if h.rateLimit == nil {
		return
	}
	_, _ = h.rateLimit.IncrementRateLimit(ctx, scope, LoginRateLimitWindow, 15*time.Minute)
}

func (h *AuthConnectHandler) resolveRoles(userID uint, fallbackRole string) []string {
	if h.roles != nil {
		if roles, err := h.roles.GetRolesForUser(auth.UserIDToRef(userID)); err == nil && len(roles) > 0 {
			return roles
		}
	}
	if fallbackRole != "" {
		return []string{fallbackRole}
	}
	return nil
}

func clientIPFromRequest(req *connect.Request[devicepb.LoginRequest]) string {
	xff := req.Header().Get("X-Forwarded-For")
	if xff != "" {
		first := strings.Split(xff, ",")[0]
		if ip := strings.TrimSpace(first); ip != "" {
			return ip
		}
	}
	host := req.Peer().Addr
	if host == "" {
		host = "unknown"
	}
	// strip port
	if idx := strings.LastIndex(host, ":"); idx > 0 {
		host = host[:idx]
	}
	return host
}

func NewAuthServiceHandler(
	pgStore *postgres.Store,
	jwtService *auth.JWTService,
	refreshSvc *auth.RefreshTokenService,
	rateLimit LoginRateLimiter,
	roles RoleResolver,
	secure bool,
) (string, http.Handler) {
	handler := NewAuthConnectHandler(pgStore, jwtService, refreshSvc, rateLimit, roles, secure)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.AuthService"
	mux.Handle("/"+serviceName+"/Login", connect.NewUnaryHandler(
		"/"+serviceName+"/Login",
		handler.Login,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetMe", connect.NewUnaryHandler(
		"/"+serviceName+"/GetMe",
		handler.GetMe,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/RefreshToken", connect.NewUnaryHandler(
		"/"+serviceName+"/RefreshToken",
		handler.RefreshToken,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/Logout", connect.NewUnaryHandler(
		"/"+serviceName+"/Logout",
		handler.Logout,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
