package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Error sentinels for this use case live in the domain layer
// (internal/domain/customer, internal/domain/session) per
// DEVELOPMENT-GUIDELINES.md §6.
const (
	MaxLoginAttempts = 5
	RateLimitWindow  = "15m"
	RateLimitTTL     = 15 * time.Minute
)

type LoginResult struct {
	User         *customer.User
	AccessToken  string
	RefreshToken string
	Roles        []string
	Permissions  []string
}

type AuthUseCase struct {
	userRepo   port.UserRepository
	tokenSvc   port.TokenService
	refreshMgr port.RefreshTokenManager
	rateLimit  port.RateLimiter
	roles      port.RoleAuthorizer
}

func NewAuthUseCase(
	userRepo port.UserRepository,
	tokenSvc port.TokenService,
	refreshMgr port.RefreshTokenManager,
	rateLimit port.RateLimiter,
	roles port.RoleAuthorizer,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		tokenSvc:   tokenSvc,
		refreshMgr: refreshMgr,
		rateLimit:  rateLimit,
		roles:      roles,
	}
}

func (u *AuthUseCase) TokenService() port.TokenService {
	return u.tokenSvc
}

func (u *AuthUseCase) Login(ctx context.Context, username, password, clientIP string) (*LoginResult, error) {
	if username == "" || password == "" {
		return nil, customer.ErrInvalidCredentials
	}

	// 1. Check rate limit
	rlScope := "login:" + username + ":" + clientIP
	if u.rateLimit != nil {
		attempts, err := u.rateLimit.GetRateLimitCount(ctx, rlScope, RateLimitWindow)
		if err == nil && attempts >= MaxLoginAttempts {
			logger.WithComponent("AuthUseCase").WithField("username", username).Warn("login blocked: too many failed attempts")
			return nil, customer.ErrTooManyAttempts
		}
	}

	// 2. Fetch user
	user, err := u.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if u.rateLimit != nil {
			_, _ = u.rateLimit.IncrementRateLimit(ctx, rlScope, RateLimitWindow, RateLimitTTL)
		}
		return nil, customer.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, customer.ErrAccountInactive
	}

	// 3. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if u.rateLimit != nil {
			_, _ = u.rateLimit.IncrementRateLimit(ctx, rlScope, RateLimitWindow, RateLimitTTL)
		}
		return nil, customer.ErrInvalidCredentials
	}

	// 4. Reset rate limit on success
	if u.rateLimit != nil {
		_ = u.rateLimit.ResetRateLimit(ctx, rlScope, RateLimitWindow)
	}

	// 5. Query effective roles & permissions
	var userRoles []string
	if u.roles != nil {
		userRoles, _ = u.roles.GetRolesForUser(fmt.Sprintf("%d", user.ID))
	}
	if len(userRoles) == 0 && user.Role != "" {
		userRoles = []string{user.Role}
	}

	var perms []string
	if u.roles != nil {
		perms, _ = u.roles.GetImplicitPermissionsForUser(fmt.Sprintf("%d", user.ID))
	}

	// 6. Generate access token
	primaryRole := user.Role
	if len(userRoles) > 0 {
		primaryRole = userRoles[0]
	}
	token, err := u.tokenSvc.GenerateAccessToken(fmt.Sprintf("%d", user.ID), user.Username, primaryRole)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 7. Optional refresh token
	var refreshToken string
	if u.refreshMgr != nil {
		rt, err := u.refreshMgr.IssueToken(ctx, user.ID, userRoles, "")
		if err == nil {
			refreshToken = rt
		}
	}

	logger.WithComponent("AuthUseCase").WithField("username", username).Info("user logged in successfully")

	return &LoginResult{
		User:         user,
		AccessToken:  token,
		RefreshToken: refreshToken,
		Roles:        userRoles,
		Permissions:  perms,
	}, nil
}

func (u *AuthUseCase) UpdateProfile(ctx context.Context, userID uint, fullName, phone, email, specialization string) (*customer.User, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, customer.ErrUserNotFound
	}

	email = strings.TrimSpace(email)
	if email != "" && email != user.Email {
		if existing, _ := u.userRepo.FindByEmail(ctx, email); existing != nil && existing.ID != userID {
			return nil, customer.ErrEmailAlreadyUsed
		}
		user.Email = email
	}

	user.FullName = strings.TrimSpace(fullName)
	user.PhoneNumber = strings.TrimSpace(phone)
	if specialization != "" {
		user.Specialization = strings.TrimSpace(specialization)
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}

	logger.WithComponent("AuthUseCase").WithField("user_id", userID).Info("user profile updated")
	return user, nil
}

func (u *AuthUseCase) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return customer.ErrPasswordTooShort
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return customer.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return customer.ErrWrongPassword
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := u.userRepo.UpdatePassword(ctx, userID, string(newHash)); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	logger.WithComponent("AuthUseCase").WithField("user_id", userID).Info("user password changed successfully")
	return nil
}
