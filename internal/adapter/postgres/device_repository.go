package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

type PostgresDeviceRepository struct {
	db *gorm.DB
}

type PostgresCredentialVault struct {
	db *gorm.DB
}

var _ port.DeviceRepository = (*PostgresDeviceRepository)(nil)
var _ port.CredentialVault = (*PostgresCredentialVault)(nil)

// NewDeviceRepository creates a new port.DeviceRepository.
func NewDeviceRepository(db *gorm.DB) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{db: db}
}

// NewCredentialVault creates a new port.CredentialVault.
func NewCredentialVault(db *gorm.DB) *PostgresCredentialVault {
	return &PostgresCredentialVault{db: db}
}

// Save stores or updates a device record in PostgreSQL.
func (r *PostgresDeviceRepository) Save(ctx context.Context, d device.Device) error {
	m := models.DeviceModelFromDomain(d)
	return r.db.WithContext(ctx).Save(m).Error
}

// FindByID retrieves a device record by ID from PostgreSQL.
func (r *PostgresDeviceRepository) FindByID(ctx context.Context, id string) (device.Device, error) {
	var m models.DeviceModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return device.Device{}, device.ErrNotFound
		}
		return device.Device{}, err
	}
	return m.ToDomain(), nil
}

// FindAll retrieves all device records from PostgreSQL.
func (r *PostgresDeviceRepository) FindAll(ctx context.Context) ([]device.Device, error) {
	var list []models.DeviceModel
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
func (r *PostgresDeviceRepository) Update(ctx context.Context, d device.Device) error {
	return r.Save(ctx, d)
}

// Delete removes a device record and its associated credentials from PostgreSQL.
func (r *PostgresDeviceRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.CredentialModel{}, "device_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&models.DeviceModel{}, "id = ?", id).Error
	})
}

// Get implements port.CredentialVault: retrieves decrypted credentials for a device.
func (v *PostgresCredentialVault) Get(ctx context.Context, deviceID string) (device.Credentials, error) {
	var m models.CredentialModel
	if err := v.db.WithContext(ctx).First(&m, "device_id = ?", deviceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return device.Credentials{}, device.ErrNotFound
		}
		return device.Credentials{}, err
	}
	return m.ToDomain(), nil
}

// Save implements port.CredentialVault: stores credentials for a device.
func (v *PostgresCredentialVault) Save(ctx context.Context, deviceID string, creds device.Credentials) error {
	m := models.CredentialModelFromDomain(deviceID, creds)
	return v.db.WithContext(ctx).Save(m).Error
}
