package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

type DeviceRepository struct {
	db *gorm.DB
}

type CredentialVault struct {
	db  *gorm.DB
	key string
}

// Deprecated alias for backward compatibility
type PostgresDeviceRepository = DeviceRepository
type PostgresCredentialVault = CredentialVault

var _ port.DeviceRepository = (*DeviceRepository)(nil)
var _ port.CredentialVault = (*CredentialVault)(nil)

// NewDeviceRepository creates a new port.DeviceRepository.
func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// NewCredentialVault creates a new port.CredentialVault.
func NewCredentialVault(db *gorm.DB, encryptionKey ...string) *CredentialVault {
	key := ""
	if len(encryptionKey) > 0 {
		key = encryptionKey[0]
	}
	return &CredentialVault{db: db, key: key}
}

// Save stores or updates a device record in PostgreSQL.
func (r *DeviceRepository) Save(ctx context.Context, d device.Device) error {
	m := model.DeviceModelFromDomain(d)
	return r.db.WithContext(ctx).Save(m).Error
}

// FindByID retrieves a device record by ID from PostgreSQL.
func (r *DeviceRepository) FindByID(ctx context.Context, id string) (device.Device, error) {
	var m model.DeviceModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return device.Device{}, device.ErrNotFound
		}
		return device.Device{}, err
	}
	return m.ToDomain(), nil
}

// FindAll retrieves all device records from PostgreSQL.
func (r *DeviceRepository) FindAll(ctx context.Context) ([]device.Device, error) {
	var list []model.DeviceModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}

	result := make([]device.Device, len(list))
	for i, m := range list {
		result[i] = m.ToDomain()
	}
	return result, nil
}

// Update updates an existing device record.
func (r *DeviceRepository) Update(ctx context.Context, d device.Device) error {
	return r.Save(ctx, d)
}

// Delete removes a device record and its associated credentials from PostgreSQL.
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.CredentialModel{}, "device_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&model.DeviceModel{}, "id = ?", id).Error
	})
}

// FindByUserScope returns only devices assigned to the specified user.
func (r *DeviceRepository) FindByUserScope(ctx context.Context, userID uint) ([]device.Device, error) {
	var list []model.DeviceModel
	err := r.db.WithContext(ctx).
		Table("devices").
		Joins("INNER JOIN user_devices ON CAST(user_devices.device_id AS text) = CAST(devices.id AS text)").
		Where("user_devices.user_id = ?", userID).
		Order("devices.created_at desc").
		Find(&list).Error
	if err != nil {
		return nil, err
	}

	result := make([]device.Device, len(list))
	for i, m := range list {
		result[i] = m.ToDomain()
	}
	return result, nil
}

// Get implements port.CredentialVault: retrieves decrypted credentials for a device.
func (v *CredentialVault) Get(ctx context.Context, deviceID string) (device.Credentials, error) {
	var m model.CredentialModel
	if err := v.db.WithContext(ctx).First(&m, "device_id = ?", deviceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return device.Credentials{}, device.ErrNotFound
		}
		return device.Credentials{}, err
	}
	return m.ToDomain(v.key)
}

// Save implements port.CredentialVault: stores credentials for a device.
func (v *CredentialVault) Save(ctx context.Context, deviceID string, creds device.Credentials) error {
	m, err := model.CredentialModelFromDomain(deviceID, creds, v.key)
	if err != nil {
		return err
	}
	return v.db.WithContext(ctx).Save(m).Error
}

// EncryptString implements port.CredentialVault: seals an arbitrary sensitive
// string dengan AES-GCM base64 memakai key vault yang sama.
func (v *CredentialVault) EncryptString(_ context.Context, plaintext string) (string, error) {
	if len(v.key) != 32 {
		if v.key == "" {
			return plaintext, nil
		}
		return "", errors.New("encryption key must be exactly 32 bytes")
	}
	return config.Encrypt(plaintext, v.key)
}

// DecryptString implements port.CredentialVault: opens a ciphertext produced
// by EncryptString.
func (v *CredentialVault) DecryptString(_ context.Context, ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	keysToTry := make([]string, 0, 2)
	if v.key != "" && len(v.key) == 32 {
		keysToTry = append(keysToTry, v.key)
	}
	if v.key != "12345678901234567890123456789012" {
		keysToTry = append(keysToTry, "12345678901234567890123456789012")
	}
	for _, k := range keysToTry {
		if plain, err := config.Decrypt(ciphertext, k); err == nil {
			return plain, nil
		}
	}
	return ciphertext, nil
}
