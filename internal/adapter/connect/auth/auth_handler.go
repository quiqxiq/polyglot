package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	authUC "github.com/quixiq/polyglot/internal/usecase/auth"
	userUC "github.com/quixiq/polyglot/internal/usecase/user"
	"github.com/quixiq/polyglot/pkg/response"
)

const BearerScheme = "Bearer"

type AuthConnectHandler struct {
	authUC    *authUC.AuthUseCase
	refreshUC *authUC.RefreshTokenUseCase
	userUC    *userUC.ManageUserUseCase
	secure    bool
}

func NewAuthConnectHandler(
	authUC *authUC.AuthUseCase,
	refreshUC *authUC.RefreshTokenUseCase,
	userUC *userUC.ManageUserUseCase,
	secure bool,
) *AuthConnectHandler {
	return &AuthConnectHandler{
		authUC:    authUC,
		refreshUC: refreshUC,
		userUC:    userUC,
		secure:    secure,
	}
}

func (h *AuthConnectHandler) Login(ctx context.Context, req *connect.Request[polyglotv1.LoginRequest]) (*connect.Response[polyglotv1.LoginResponse], error) {
	clientIP := clientIPFromRequest(req)
	res, err := h.authUC.Login(ctx, req.Msg.Username, req.Msg.Password, clientIP)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	profile := DomainUserToProfile(res.User, res.Roles, res.Permissions)
	now := time.Now()

	resp := connect.NewResponse(&polyglotv1.LoginResponse{
		Token:         res.AccessToken,
		User:          profile,
		ExpiresAtUnix: now.Add(h.authUC.TokenService().ExpiryDuration()).Unix(),
	})

	if res.RefreshToken != "" {
		iauth.SetRefreshTokenCookie(resp.Header(), res.RefreshToken, iauth.RefreshTokenLifetime, h.secure)
	}

	return resp, nil
}

func (h *AuthConnectHandler) GetMe(ctx context.Context, req *connect.Request[polyglotv1.GetMeRequest]) (*connect.Response[polyglotv1.GetMeResponse], error) {
	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		return nil, response.Unauthenticated("authorization header missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
		return nil, response.Unauthenticated("invalid authorization header format")
	}

	tokenStr := parts[1]
	userIDStr, _, _, err := h.authUC.TokenService().ValidateAccessToken(tokenStr)
	if err != nil {
		return nil, response.Unauthenticated("invalid or expired token")
	}

	var uid uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &uid); err != nil {
		return nil, response.Unauthenticated("invalid user identity")
	}

	user, err := h.userUC.GetUser(ctx, uid)
	if err != nil {
		return nil, response.MapDomainError(domainCustomer.ErrUserNotFound)
	}

	roles, _ := h.userUC.GetRoles(ctx, uid)
	perms, _ := h.userUC.GetPermissions(ctx, uid)

	profile := DomainUserToProfile(user, roles, perms)
	return connect.NewResponse(&polyglotv1.GetMeResponse{User: profile}), nil
}

func (h *AuthConnectHandler) UpdateMe(ctx context.Context, req *connect.Request[polyglotv1.UpdateMeRequest]) (*connect.Response[polyglotv1.UpdateMeResponse], error) {
	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		return nil, response.Unauthenticated("authorization header missing")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
		return nil, response.Unauthenticated("invalid authorization header format")
	}

	userIDStr, _, _, err := h.authUC.TokenService().ValidateAccessToken(parts[1])
	if err != nil {
		return nil, response.Unauthenticated("invalid or expired token")
	}

	var uid uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &uid); err != nil {
		return nil, response.Unauthenticated("invalid user identity")
	}

	user, err := h.authUC.UpdateProfile(ctx, uid, req.Msg.FullName, req.Msg.PhoneNumber, req.Msg.Email, req.Msg.Specialization)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	roles, _ := h.userUC.GetRoles(ctx, uid)
	perms, _ := h.userUC.GetPermissions(ctx, uid)

	profile := DomainUserToProfile(user, roles, perms)
	return connect.NewResponse(&polyglotv1.UpdateMeResponse{User: profile}), nil
}

func (h *AuthConnectHandler) ChangePassword(ctx context.Context, req *connect.Request[polyglotv1.ChangePasswordRequest]) (*connect.Response[polyglotv1.ChangePasswordResponse], error) {
	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		return nil, response.Unauthenticated("authorization header missing")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
		return nil, response.Unauthenticated("invalid authorization header format")
	}

	userIDStr, _, _, err := h.authUC.TokenService().ValidateAccessToken(parts[1])
	if err != nil {
		return nil, response.Unauthenticated("invalid or expired token")
	}

	var uid uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &uid); err != nil {
		return nil, response.Unauthenticated("invalid user identity")
	}

	if err := h.authUC.ChangePassword(ctx, uid, req.Msg.OldPassword, req.Msg.NewPassword); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ChangePasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}), nil
}

func (h *AuthConnectHandler) RefreshToken(ctx context.Context, req *connect.Request[polyglotv1.RefreshTokenRequest]) (*connect.Response[polyglotv1.RefreshTokenResponse], error) {
	rawToken := req.Msg.RefreshToken
	if rawToken == "" {
		rawToken = iauth.ExtractRefreshTokenCookie(req.Header())
	}
	if rawToken == "" {
		return nil, response.InvalidArgument("refresh token is required")
	}

	res, err := h.refreshUC.Refresh(ctx, rawToken)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	now := time.Now()
	resp := connect.NewResponse(&polyglotv1.RefreshTokenResponse{
		Token:                res.AccessToken,
		ExpiresAtUnix:        now.Add(h.authUC.TokenService().ExpiryDuration()).Unix(),
		RefreshExpiresAtUnix: now.Add(iauth.RefreshTokenLifetime).Unix(),
	})

	if res.RefreshToken != "" {
		iauth.SetRefreshTokenCookie(resp.Header(), res.RefreshToken, iauth.RefreshTokenLifetime, h.secure)
	}

	return resp, nil
}

func (h *AuthConnectHandler) Logout(ctx context.Context, req *connect.Request[polyglotv1.LogoutRequest]) (*connect.Response[polyglotv1.LogoutResponse], error) {
	rawToken := iauth.ExtractRefreshTokenCookie(req.Header())
	if rawToken != "" && h.refreshUC != nil {
		// best-effort: logout tetap mengembalikan sukses meski revoke refresh token gagal.
		_ = h.refreshUC.Revoke(ctx, rawToken)
	}

	resp := connect.NewResponse(&polyglotv1.LogoutResponse{
		Success: true,
	})
	iauth.ClearRefreshTokenCookie(resp.Header(), h.secure)

	return resp, nil
}

func clientIPFromRequest[T any](req *connect.Request[T]) string {
	if xff := req.Header().Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	if xri := req.Header().Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	return "unknown"
}
