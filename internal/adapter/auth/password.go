package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DefaultBcryptCost is the bcrypt cost factor used for hashing passwords.
const DefaultBcryptCost = 10

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultBcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password with a bcrypt hash.
// Returns nil on match, or an error if they don't match.
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
