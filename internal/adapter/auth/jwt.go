package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quixiq/polyglot/internal/port"
)

var (
	ErrInvalidToken = errors.New("invalid or expired JWT token")
)

type Claims struct {
	UserID   uint     `json:"user_id"`
	Email    string   `json:"email"`
	Role     string   `json:"role"`
	Roles    []string `json:"roles,omitempty"`
	TenantID string   `json:"tenant_id"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey []byte
	expiry    time.Duration
}

var _ port.TokenService = (*JWTService)(nil)

func NewJWTService(secret string, expiryHours int) *JWTService {
	if expiryHours <= 0 {
		expiryHours = 24
	}
	return &JWTService{
		secretKey: []byte(secret),
		expiry:    time.Duration(expiryHours) * time.Hour,
	}
}

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

func (s *JWTService) GenerateAccessToken(userID, username, role string) (string, error) {
	uid, _ := strconv.ParseUint(userID, 10, 64)
	return s.GenerateToken(uint(uid), username, []string{role}, "")
}

func (s *JWTService) ValidateAccessToken(tokenString string) (string, string, string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", "", "", err
	}
	return fmt.Sprintf("%d", claims.UserID), claims.Email, claims.Role, nil
}

func (s *JWTService) ExpiryDuration() time.Duration {
	if s == nil || s.expiry <= 0 {
		return 24 * time.Hour
	}
	return s.expiry
}

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
