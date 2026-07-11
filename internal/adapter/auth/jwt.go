package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/quixiq/polyglot/internal/port"
)

var (
	// ErrInvalidToken is returned when a JWT is invalid, expired, or malformed.
	ErrInvalidToken = errors.New("invalid or expired token")
)

// Claims represents the JWT wire format for Polyglot auth. It is an
// implementation detail of this adapter; callers receive a
// library-agnostic port.Identity from Validate instead.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type jwtHandler struct {
	secret []byte
	expiry time.Duration
}

var _ port.Authenticator = (*jwtHandler)(nil)

// NewJWT builds a port.Authenticator backed by JWT with HMAC SHA-256.
func NewJWT(secret []byte, expiry time.Duration) port.Authenticator {
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

// Validate parses and verifies a JWT token, returning the Identity it
// encodes.
func (j *jwtHandler) Validate(ctx context.Context, tokenString string) (*port.Identity, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &port.Identity{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}, nil
	}

	return nil, ErrInvalidToken
}
