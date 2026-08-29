package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// DeviceModel is the GORM database model for network devices inventory.
type DeviceModel struct {
	ID             string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID       string `gorm:"not null;index;default:tenant-default"`
	Name           string `gorm:"not null"`
	Vendor         string `gorm:"not null;default:mikrotik"`
	DriverType     string `gorm:"not null;default:mikrotik"`
	Host           string `gorm:"not null"`
	Port           int    `gorm:"not null;default:8728"`
	SSHPort        int    `gorm:"column:ssh_port;not null;default:22"`
	TimeoutMS      int    `gorm:"not null;default:10000"`
	PollIntervalMS int    `gorm:"not null;default:30000"`
	ExtraJSON      string `gorm:"type:text"`
	TagsJSON       string `gorm:"type:text"`
	Enabled        bool   `gorm:"not null;default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName returns the database table name for devices.
func (DeviceModel) TableName() string {
	return "devices"
}

// CredentialModel is the GORM database model for encrypted device credentials (AES-256-GCM).
type CredentialModel struct {
	DeviceID   string    `gorm:"column:device_id;primaryKey"`
	Ciphertext []byte    `gorm:"column:ciphertext;not null"`
	Nonce      []byte    `gorm:"column:nonce;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

// TableName returns the database table name for credentials.
func (CredentialModel) TableName() string {
	return "credentials"
}

const defaultTestEncryptionKey = "12345678901234567890123456789012"

// ToDomain converts a device database model to its domain representation.
func (m *DeviceModel) ToDomain() device.Device {
	if m == nil {
		return device.Device{}
	}

	var extra map[string]string
	if m.ExtraJSON != "" {
		_ = json.Unmarshal([]byte(m.ExtraJSON), &extra)
	}

	var tags []string
	if m.TagsJSON != "" {
		_ = json.Unmarshal([]byte(m.TagsJSON), &tags)
	}

	sshPort := m.SSHPort
	if sshPort <= 0 {
		sshPort = 22
	}

	return device.Device{
		ID:             m.ID,
		TenantID:       m.TenantID,
		Name:           m.Name,
		Vendor:         m.Vendor,
		DriverType:     m.DriverType,
		Host:           m.Host,
		Port:           m.Port,
		SSHPort:        sshPort,
		TimeoutMS:      m.TimeoutMS,
		PollIntervalMS: m.PollIntervalMS,
		Extra:          extra,
		Tags:           tags,
		Enabled:        m.Enabled,
	}
}

// DeviceModelFromDomain converts a device domain entity to a database model.
func DeviceModelFromDomain(d device.Device) *DeviceModel {
	extraJSON, _ := json.Marshal(d.Extra)
	tagsJSON, _ := json.Marshal(d.Tags)

	tenantID := d.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}

	sshPort := d.SSHPort
	if sshPort <= 0 {
		sshPort = 22
	}

	return &DeviceModel{
		ID:             d.ID,
		TenantID:       tenantID,
		Name:           d.Name,
		Vendor:         d.Vendor,
		DriverType:     d.DriverType,
		Host:           d.Host,
		Port:           d.Port,
		SSHPort:        sshPort,
		TimeoutMS:      d.TimeoutMS,
		PollIntervalMS: d.PollIntervalMS,
		ExtraJSON:      string(extraJSON),
		TagsJSON:       string(tagsJSON),
		Enabled:        d.Enabled,
	}
}

// ToDomain decrypts credentials and converts them to the domain representation.
func (c *CredentialModel) ToDomain(key string) (device.Credentials, error) {
	if c == nil || len(c.Ciphertext) == 0 {
		return device.Credentials{}, nil
	}
	if key == "" {
		key = defaultTestEncryptionKey
	}
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return device.Credentials{}, fmt.Errorf("encryption key must be exactly 32 bytes")
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return device.Credentials{}, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return device.Credentials{}, err
	}
	if len(c.Nonce) != aesGCM.NonceSize() {
		return device.Credentials{}, fmt.Errorf("invalid nonce size: %d", len(c.Nonce))
	}
	plaintext, err := aesGCM.Open(nil, c.Nonce, c.Ciphertext, nil)
	if err != nil {
		return device.Credentials{}, fmt.Errorf("decrypt credentials failed: %w", err)
	}
	var creds device.Credentials
	if err := json.Unmarshal(plaintext, &creds); err != nil {
		return device.Credentials{}, fmt.Errorf("unmarshal decrypted credentials failed: %w", err)
	}
	return creds, nil
}

// CredentialModelFromDomain encrypts credentials into a database model.
func CredentialModelFromDomain(deviceID string, c device.Credentials, key string) (*CredentialModel, error) {
	if key == "" {
		key = defaultTestEncryptionKey
	}
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes")
	}
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := aesGCM.Seal(nil, nonce, data, nil)
	return &CredentialModel{
		DeviceID:   deviceID,
		Ciphertext: ciphertext,
		Nonce:      nonce,
		UpdatedAt:  time.Now(),
	}, nil
}
