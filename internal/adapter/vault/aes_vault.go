package vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/port"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// credentialRecord maps the encrypted credentials table per migration 000001.
type credentialRecord struct {
	DeviceID   string    `gorm:"column:device_id;primaryKey;type:uuid"`
	Ciphertext []byte    `gorm:"column:ciphertext;not null"`
	Nonce      []byte    `gorm:"column:nonce;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

// TableName returns the explicit table name for the credential record.
func (credentialRecord) TableName() string {
	return "credentials"
}

// credentialBlob is the plaintext shape encrypted by the vault.
type credentialBlob struct {
	Username string            `json:"username"`
	Password string            `json:"password"`
	Extra    map[string]string `json:"extra,omitempty"`
}

// AESVault implements port.CredentialVault using AES-256-GCM.
// It stores one encrypted blob per device in the `credentials` table.
type AESVault struct {
	db *gorm.DB
	// gcm is a reusable AES-GCM cipher for this vault instance. It is safe
	// for concurrent use; each Store call generates a fresh random nonce.
	gcm cipher.AEAD
}

// NewAESVault returns a port.CredentialVault backed by AES-256-GCM. The key
// must be exactly 32 bytes.
func NewAESVault(db *gorm.DB, key []byte) (*AESVault, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("vault: key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("vault: create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("vault: create gcm: %w", err)
	}
	return &AESVault{db: db, gcm: gcm}, nil
}

// Get returns the decrypted credentials for deviceID.
func (v *AESVault) Get(ctx context.Context, deviceID string) (device.Credentials, error) {
	var rec credentialRecord
	if err := v.db.WithContext(ctx).First(&rec, "device_id = ?", deviceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return device.Credentials{}, fmt.Errorf("credentials %s: %w", deviceID, device.ErrNotFound)
		}
		return device.Credentials{}, fmt.Errorf("credentials %s: %w", deviceID, err)
	}

	plaintext, err := v.gcm.Open(nil, rec.Nonce, rec.Ciphertext, nil)
	if err != nil {
		return device.Credentials{}, fmt.Errorf("credentials %s: decrypt failed: %w", deviceID, err)
	}

	var blob credentialBlob
	if err := json.Unmarshal(plaintext, &blob); err != nil {
		return device.Credentials{}, fmt.Errorf("credentials %s: unmarshal failed: %w", deviceID, err)
	}

	return device.Credentials{
		Username: blob.Username,
		Password: blob.Password,
		Extra:    blob.Extra,
	}, nil
}

// Store encrypts and persists credentials for deviceID. If a row already
// exists, it is overwritten.
func (v *AESVault) Store(ctx context.Context, deviceID string, creds device.Credentials) error {
	blob := credentialBlob{
		Username: creds.Username,
		Password: creds.Password,
		Extra:    creds.Extra,
	}
	plaintext, err := json.Marshal(blob)
	if err != nil {
		return fmt.Errorf("credentials %s: marshal failed: %w", deviceID, err)
	}

	nonce := make([]byte, v.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("credentials %s: generate nonce: %w", deviceID, err)
	}

	ciphertext := v.gcm.Seal(nil, nonce, plaintext, nil)

	rec := credentialRecord{
		DeviceID:   deviceID,
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}

	if err := v.db.WithContext(ctx).Save(&rec).Error; err != nil {
		return fmt.Errorf("credentials %s: persist: %w", deviceID, err)
	}
	return nil
}

// Delete removes the credential row for deviceID. It is safe to call even
// if the row does not exist.
func (v *AESVault) Delete(ctx context.Context, deviceID string) error {
	if err := v.db.WithContext(ctx).Delete(&credentialRecord{}, "device_id = ?", deviceID).Error; err != nil {
		return fmt.Errorf("credentials %s: delete: %w", deviceID, err)
	}
	return nil
}

// compile-time check that the port interface is satisfied.
var _ port.CredentialVault = (*AESVault)(nil)
