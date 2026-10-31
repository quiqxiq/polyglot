package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// ErrInvalidKeyLength indicates that the AES encryption key is not 32 bytes (256-bit).
var ErrInvalidKeyLength = errors.New("crypto: encryption key must be exactly 32 bytes")

// ErrCiphertextTooShort indicates that the ciphertext is shorter than the GCM nonce size.
var ErrCiphertextTooShort = errors.New("crypto: ciphertext too short")

// newGCM membangun cipher AEAD AES-256-GCM dari key raw 32 byte.
func newGCM(key string) (cipher.AEAD, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("aes new cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher new gcm: %w", err)
	}
	return aesGCM, nil
}

// sealGCM menghasilkan nonce||ciphertext dengan nonce acak di header.
func sealGCM(aesGCM cipher.AEAD, plaintext []byte) ([]byte, error) {
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("read nonce: %w", err)
	}
	return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}

// openGCM mendekripsi nonce||ciphertext yang diproduksi sealGCM.
func openGCM(aesGCM cipher.AEAD, ciphertext []byte) ([]byte, error) {
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt gcm ciphertext: %w", err)
	}
	return plaintext, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce and returns base64 encoded string.
func Encrypt(plaintext string, key string) (string, error) {
	aesGCM, err := newGCM(key)
	if err != nil {
		return "", err
	}

	ciphertext, err := sealGCM(aesGCM, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decodes base64 ciphertext and decrypts it using AES-256-GCM.
func Decrypt(ciphertext string, key string) (string, error) {
	aesGCM, err := newGCM(key)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode base64 ciphertext: %w", err)
	}

	plaintext, err := openGCM(aesGCM, data)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptBytes mengenkripsi plaintext biner dengan AES-256-GCM dan
// mengembalikan ciphertext raw nonce||ciphertext — nonce tersimpan di header
// file sehingga tidak perlu kolom/berkas metadata tambahan. Dipakai untuk
// artifact restore point biner (PLAN-AI-ROUTER-OPS v2 M1).
func EncryptBytes(plaintext []byte, key string) ([]byte, error) {
	aesGCM, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	return sealGCM(aesGCM, plaintext)
}

// DecryptBytes mendekripsi ciphertext raw nonce||ciphertext hasil
// EncryptBytes. Key harus 32 byte; ciphertext lebih pendek dari ukuran nonce
// mengembalikan ErrCiphertextTooShort.
func DecryptBytes(ciphertext []byte, key string) ([]byte, error) {
	aesGCM, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	return openGCM(aesGCM, ciphertext)
}
