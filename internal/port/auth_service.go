package port

import (
	"context"
	"time"
)

// RateLimiter defines the contract for sliding-window request rate limiting.
type RateLimiter interface {
	IncrementRateLimit(ctx context.Context, scope, window string, ttl time.Duration) (int64, error)
	GetRateLimitCount(ctx context.Context, scope, window string) (int64, error)
	ResetRateLimit(ctx context.Context, scope, window string) error
}

// RoleAuthorizer defines the contract for role and permission inspection & sync.
type RoleAuthorizer interface {
	GetRolesForUser(user string) ([]string, error)
	GetImplicitPermissionsForUser(user string) ([]string, error)
	AddRoleForUser(user, role string) (bool, error)
	DeleteRolesForUser(user string) (bool, error)
}

// TokenService defines the contract for generating, validating, and revoking JWT tokens.
type TokenService interface {
	GenerateAccessToken(userID, username, role string) (string, error)
	ValidateAccessToken(tokenString string) (userID string, username string, role string, err error)
}

// RefreshTokenManager defines the contract for issuing, rotating, and revoking opaque refresh tokens.
type RefreshTokenManager interface {
	IssueToken(ctx context.Context, userID uint, roles []string, tenantID string) (string, error)
	RotateToken(ctx context.Context, oldToken string) (newToken string, userID uint, roles []string, tenantID string, err error)
	RevokeToken(ctx context.Context, token string) error
}
