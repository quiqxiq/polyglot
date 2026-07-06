package auth

import "context"

// NewJWT builds the JWT issuer/validator.
// TODO: use github.com/golang-jwt/jwt/v5 per TECH-STACK-DAN-PERSIAPAN.md §4
// — never v4 or the unmaintained dgrijalva/jwt-go.
func NewJWT(ctx context.Context) error {
	return nil
}
