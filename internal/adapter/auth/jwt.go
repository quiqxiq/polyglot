package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when a JWT is invalid, expired, or malformed.
	ErrInvalidToken = errors.New("invalid or expired token")
)

// Claims represents the JWT claims for Polyglot auth.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTHandler defines the interface for issuing and validating tokens.
type JWTHandler interface {
	Issue(ctx context.Context, userID, username, role string) (string, error)
	Validate(ctx context.Context, tokenString string) (*Claims, error)
}

type jwtHandler struct {
	secret []byte
	expiry time.Duration
}

// NewJWT builds a JWT issuer/validator using HMAC SHA-256.
func NewJWT(secret []byte, expiry time.Duration) JWTHandler {
	return &jwtHandler{
		secret: secret,
		expiry: expiry,
	}
}

// Issue generates a signed JWT.
func (j *jwtHandler) Issue(ctx context.Context, userID, username, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "polyglot",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// Validate parses and verifies a JWT token.
func (j *jwtHandler) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
