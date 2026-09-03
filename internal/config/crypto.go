package config

import (
	"fmt"

	"github.com/quixiq/polyglot/pkg/crypto"
)

// Encrypt encrypts plaintext using AES-256-GCM via pkg/crypto.
func Encrypt(plaintext string, key string) (string, error) {
	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		return "", fmt.Errorf("config encrypt: %w", err)
	}
	return ciphertext, nil
}

// Decrypt decodes base64 ciphertext and decrypts it using AES-256-GCM via pkg/crypto.
func Decrypt(ciphertext string, key string) (string, error) {
	plaintext, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		return "", fmt.Errorf("config decrypt: %w", err)
	}
	return plaintext, nil
}
