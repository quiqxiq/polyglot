package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	authAdapter "github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/domain/auth"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// AuthUseCase orchestrates user authentication, token generation, and profile retrieval.
type AuthUseCase struct {
	userRepo   port.UserRepository
	jwtService *authAdapter.JWTService
}

// NewAuthUseCase constructs a new AuthUseCase.
func NewAuthUseCase(userRepo port.UserRepository, jwtService *authAdapter.JWTService) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Login authenticates a user by email and password, returning a signed JWT token.
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*auth.LoginResult, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("%w: email and password are required", response.ErrInvalidInput)
	}

	if uc.userRepo == nil {
		return nil, fmt.Errorf("%w: user repository unavailable", response.ErrUnavailable)
	}

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		logger.FromContext(ctx).WithField("email", email).Warn("Login failed: user not found")
		return nil, fmt.Errorf("%w: invalid email or password", response.ErrUnauthorized)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		logger.FromContext(ctx).WithField("email", email).Warn("Login failed: incorrect password")
		return nil, fmt.Errorf("%w: invalid email or password", response.ErrUnauthorized)
	}

	tokenStr, err := uc.jwtService.GenerateToken(user.ID, user.Email, user.Role, user.TenantID)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Error("Failed to generate JWT token")
		return nil, fmt.Errorf("%w: failed to generate token", response.ErrInternal)
	}

	exp := time.Now().Add(24 * time.Hour).Unix()

	logger.FromContext(ctx).WithFields(logger.Fields{
		"user_id": user.ID,
		"role":    user.Role,
	}).Info("User authenticated successfully")

	return &auth.LoginResult{
		Token:         tokenStr,
		User:          user,
		ExpiresAtUnix: exp,
	}, nil
}

// GetProfile retrieves the profile for a given user ID.
func (uc *AuthUseCase) GetProfile(ctx context.Context, userID uint) (*auth.User, error) {
	if uc.userRepo == nil {
		return nil, response.ErrUnavailable
	}
	return uc.userRepo.FindByID(ctx, userID)
}
