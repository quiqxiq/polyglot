package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"time"
)

// SecretModel is the GORM model for the generic secrets vault
// (password akun PPPoE / hotspot permanent). Same AES-256-GCM scheme as
// CredentialModel — key is a logical key string, convention:
// "subscription:<id>:password".
type SecretModel struct {
	Key        string    `gorm:"column:key;primaryKey"`
	Ciphertext []byte    `gorm:"column:ciphertext;not null"`
	Nonce      []byte    `gorm:"column:nonce;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (SecretModel) TableName() string {
	return "secrets"
}

// ToSecret decrypts the stored blob into a plaintext string.
func (m *SecretModel) ToSecret(key string) (string, error) {
	if m == nil || len(m.Ciphertext) == 0 {
		return "", fmt.Errorf("empty secret")
	}
	aesGCM, err := newGCM(key)
	if err != nil {
		return "", err
	}
	if len(m.Nonce) != aesGCM.NonceSize() {
		return "", fmt.Errorf("invalid nonce size: %d", len(m.Nonce))
	}
	plaintext, err := aesGCM.Open(nil, m.Nonce, m.Ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt secret failed: %w", err)
	}
	return string(plaintext), nil
}

// SecretModelFromDomain encrypts a plaintext secret for storage. The
// encryption key override follows the CredentialVault convention: ""
// falls back to the shared dev/test default.
func SecretModelFromDomainWithKey(key, secret, encKey string) (*SecretModel, error) {
	aesGCM, err := newGCM(encKey)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := aesGCM.Seal(nil, nonce, []byte(secret), nil)
	return &SecretModel{
		Key:        key,
		Ciphertext: ciphertext,
		Nonce:      nonce,
		UpdatedAt:  time.Now(),
	}, nil
}

// sharedGCM builds the AES-GCM cipher used by both credential and secret
// models so the encryption key policy stays in one place.
func newGCM(key string) (cipher.AEAD, error) {
	if key == "" {
		key = defaultTestEncryptionKey
	}
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes")
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
