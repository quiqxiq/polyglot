package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quixiq/polyglot/internal/domain/session"
	"github.com/quixiq/polyglot/internal/port"
)

var (
	// ErrInvalidToken indicates a missing, expired, or malformed JWT.
	ErrInvalidToken = session.ErrInvalidToken
)

// Claims contains the authenticated user claims stored in a JWT.
type Claims struct {
	UserID   uint     `json:"user_id"`
	Email    string   `json:"email"`
	Role     string   `json:"role"`
	Roles    []string `json:"roles,omitempty"`
	TenantID string   `json:"tenant_id"`
	jwt.RegisteredClaims
}

// JWTService signs and validates access tokens.
type JWTService struct {
	secretKey []byte
	expiry    time.Duration
}

var _ port.TokenService = (*JWTService)(nil)

// NewJWTService creates a JWT service with the configured expiration period.
func NewJWTService(secret string, expiryHours int) *JWTService {
	if expiryHours <= 0 {
		expiryHours = 24
	}
	return &JWTService{
		secretKey: []byte(secret),
		expiry:    time.Duration(expiryHours) * time.Hour,
	}
}

// GenerateToken creates a signed JWT for a user and tenant.
func (s *JWTService) GenerateToken(userID uint, email string, roles []string, tenantID string) (string, error) {
	if len(roles) == 0 {
		roles = []string{""}
	}
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Email:    email,
		Role:     roles[0],
		Roles:    roles,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenStr, nil
}

// GenerateAccessToken creates a signed access token from string identifiers.
func (s *JWTService) GenerateAccessToken(userID, username, role string) (string, error) {
	uid, _ := strconv.ParseUint(userID, 10, 64)
	return s.GenerateToken(uint(uid), username, []string{role}, "")
}

// ValidateAccessToken validates an access token and returns its principal fields.
func (s *JWTService) ValidateAccessToken(tokenString string) (string, string, string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", "", "", err
	}
	return fmt.Sprintf("%d", claims.UserID), claims.Email, claims.Role, nil
}

// ExpiryDuration returns the configured token lifetime or the default lifetime.
func (s *JWTService) ExpiryDuration() time.Duration {
	if s == nil || s.expiry <= 0 {
		return 24 * time.Hour
	}
	return s.expiry
}

// ValidateToken parses and validates a signed JWT.
func (s *JWTService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, t.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
