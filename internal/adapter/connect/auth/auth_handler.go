package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
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

func (h *AuthConnectHandler) Login(ctx context.Context, req *connect.Request[devicepb.LoginRequest]) (*connect.Response[devicepb.LoginResponse], error) {
	if req.Msg.Username == "" || req.Msg.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username and password are required"))
	}

	clientIP := clientIPFromRequest(req)
	res, err := h.authUC.Login(ctx, req.Msg.Username, req.Msg.Password, clientIP)
	if err != nil {
		if errors.Is(err, authUC.ErrTooManyAttempts) {
			return nil, connect.NewError(connect.CodeResourceExhausted, err)
		}
		if errors.Is(err, authUC.ErrInvalidCredentials) {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}
		if errors.Is(err, authUC.ErrAccountInactive) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, response.MapDomainError(err)
	}

	profile := DomainUserToProfile(res.User, res.Roles, res.Permissions)
	now := time.Now()

	resp := connect.NewResponse(&devicepb.LoginResponse{
		Token:         res.AccessToken,
		User:          profile,
		ExpiresAtUnix: now.Add(24 * time.Hour).Unix(),
	})

	if res.RefreshToken != "" {
		iauth.SetRefreshTokenCookie(resp.Header(), res.RefreshToken, iauth.RefreshTokenLifetime, h.secure)
	}

	return resp, nil
}

func (h *AuthConnectHandler) GetMe(ctx context.Context, req *connect.Request[devicepb.GetMeRequest]) (*connect.Response[devicepb.GetMeResponse], error) {
	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authorization header missing"))
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format"))
	}

	tokenStr := parts[1]
	userIDStr, _, _, err := h.authUC.TokenService().ValidateAccessToken(tokenStr)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid or expired token"))
	}

	var uid uint
	fmt.Sscanf(userIDStr, "%d", &uid)

	user, err := h.userUC.GetUser(ctx, uid)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}

	roles, _ := h.userUC.GetRoles(ctx, uid)
	perms, _ := h.userUC.GetPermissions(ctx, uid)

	profile := DomainUserToProfile(user, roles, perms)
	return connect.NewResponse(&devicepb.GetMeResponse{User: profile}), nil
}

func (h *AuthConnectHandler) UpdateMe(ctx context.Context, req *connect.Request[devicepb.UpdateMeRequest]) (*connect.Response[devicepb.UpdateMeResponse], error) {
	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authorization header missing"))
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format"))
	}

	userIDStr, _, _, err := h.authUC.TokenService().ValidateAccessToken(parts[1])
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid or expired token"))
	}

	var uid uint
	fmt.Sscanf(userIDStr, "%d", &uid)

	user, err := h.authUC.UpdateProfile(ctx, uid, req.Msg.FullName, req.Msg.PhoneNumber, req.Msg.Email, req.Msg.Specialization)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	roles, _ := h.userUC.GetRoles(ctx, uid)
	perms, _ := h.userUC.GetPermissions(ctx, uid)

	profile := DomainUserToProfile(user, roles, perms)
	return connect.NewResponse(&devicepb.UpdateMeResponse{User: profile}), nil
}

func (h *AuthConnectHandler) ChangePassword(ctx context.Context, req *connect.Request[devicepb.ChangePasswordRequest]) (*connect.Response[devicepb.ChangePasswordResponse], error) {
	if req.Msg.OldPassword == "" || req.Msg.NewPassword == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("old_password and new_password are required"))
	}

	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authorization header missing"))
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format"))
	}

	userIDStr, _, _, err := h.authUC.TokenService().ValidateAccessToken(parts[1])
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid or expired token"))
	}

	var uid uint
	fmt.Sscanf(userIDStr, "%d", &uid)

	if err := h.authUC.ChangePassword(ctx, uid, req.Msg.OldPassword, req.Msg.NewPassword); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ChangePasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}), nil
}

func (h *AuthConnectHandler) RefreshToken(ctx context.Context, req *connect.Request[devicepb.RefreshTokenRequest]) (*connect.Response[devicepb.RefreshTokenResponse], error) {
	rawToken := req.Msg.RefreshToken
	if rawToken == "" {
		rawToken = iauth.ExtractRefreshTokenCookie(req.Header())
	}
	if rawToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("refresh token is required"))
	}

	res, err := h.refreshUC.Refresh(ctx, rawToken)
	if err != nil {
		if errors.Is(err, authUC.ErrInvalidRefreshToken) {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}
		return nil, response.MapDomainError(err)
	}

	now := time.Now()
	resp := connect.NewResponse(&devicepb.RefreshTokenResponse{
		Token:                res.AccessToken,
		ExpiresAtUnix:        now.Add(24 * time.Hour).Unix(),
		RefreshExpiresAtUnix: now.Add(iauth.RefreshTokenLifetime).Unix(),
	})

	if res.RefreshToken != "" {
		iauth.SetRefreshTokenCookie(resp.Header(), res.RefreshToken, iauth.RefreshTokenLifetime, h.secure)
	}

	return resp, nil
}

func (h *AuthConnectHandler) Logout(ctx context.Context, req *connect.Request[devicepb.LogoutRequest]) (*connect.Response[devicepb.LogoutResponse], error) {
	rawToken := iauth.ExtractRefreshTokenCookie(req.Header())
	if rawToken != "" && h.refreshUC != nil {
		// best-effort: logout tetap mengembalikan sukses meski revoke refresh token gagal.
		_ = h.refreshUC.Revoke(ctx, rawToken)
	}

	resp := connect.NewResponse(&devicepb.LogoutResponse{
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
