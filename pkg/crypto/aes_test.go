package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt_Success(t *testing.T) {
	key := "01234567890123456789012345678901" // 32 bytes
	plaintext := "sk-ant-api03-test-super-secret-key"

	ciphertext, err := Encrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_InvalidKeyLength(t *testing.T) {
	shortKey := "short-key"
	plaintext := "secret"

	_, err := Encrypt(plaintext, shortKey)
	assert.ErrorIs(t, err, ErrInvalidKeyLength)

	_, err = Decrypt("someciphertext", shortKey)
	assert.ErrorIs(t, err, ErrInvalidKeyLength)
}

func TestDecrypt_CorruptedOrShortData(t *testing.T) {
	key := "01234567890123456789012345678901"

	_, err := Decrypt("aGVsbG8=", key) // "hello" in base64 (5 bytes, less than 12-byte GCM nonce)
	assert.Error(t, err)
}
