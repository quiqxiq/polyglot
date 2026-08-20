package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

var (
	ErrRefreshTokenRequired = errors.New("refresh token is required")
	ErrInvalidRefreshToken  = errors.New("invalid or expired refresh token")
)

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	UserID       uint
	Username     string
	Role         string
	Roles        []string
	Permissions  []string
}

type RefreshTokenUseCase struct {
	userRepo   port.UserRepository
	tokenSvc   port.TokenService
	refreshMgr port.RefreshTokenManager
	roles      port.RoleAuthorizer
}

func NewRefreshTokenUseCase(
	userRepo port.UserRepository,
	tokenSvc port.TokenService,
	refreshMgr port.RefreshTokenManager,
	roles port.RoleAuthorizer,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepo:   userRepo,
		tokenSvc:   tokenSvc,
		refreshMgr: refreshMgr,
		roles:      roles,
	}
}

func (u *RefreshTokenUseCase) Refresh(ctx context.Context, currentRefreshToken string) (*RefreshResult, error) {
	if u.refreshMgr == nil {
		return nil, errors.New("refresh token service not available")
	}
	if currentRefreshToken == "" {
		return nil, ErrRefreshTokenRequired
	}

	newRefreshToken, userID, _, _, err := u.refreshMgr.RotateToken(ctx, currentRefreshToken)
	if err != nil {
		logger.WithComponent("RefreshTokenUseCase").WithError(err).Warn("refresh token rotation failed")
		return nil, ErrInvalidRefreshToken
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil || !user.IsActive {
		_ = u.refreshMgr.RevokeToken(ctx, newRefreshToken)
		return nil, ErrUserNotFound
	}

	var userRoles []string
	if u.roles != nil {
		userRoles, _ = u.roles.GetRolesForUser(fmt.Sprintf("%d", userID))
	}
	if len(userRoles) == 0 && user.Role != "" {
		userRoles = []string{user.Role}
	}

	var perms []string
	if u.roles != nil {
		perms, _ = u.roles.GetImplicitPermissionsForUser(fmt.Sprintf("%d", userID))
	}

	primaryRole := user.Role
	if len(userRoles) > 0 {
		primaryRole = userRoles[0]
	}

	accessToken, err := u.tokenSvc.GenerateAccessToken(fmt.Sprintf("%d", userID), user.Username, primaryRole)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		UserID:       userID,
		Username:     user.Username,
		Role:         primaryRole,
		Roles:        userRoles,
		Permissions:  perms,
	}, nil
}

func (u *RefreshTokenUseCase) Revoke(ctx context.Context, refreshToken string) error {
	if u.refreshMgr != nil && refreshToken != "" {
		return u.refreshMgr.RevokeToken(ctx, refreshToken)
	}
	return nil
}
