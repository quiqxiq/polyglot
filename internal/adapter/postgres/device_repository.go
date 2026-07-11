package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// deviceModel maps the `devices` table to a GORM-friendly struct.
// Tags are deliberately omitted from the model because GORM's PostgreSQL
// driver (via pgx) does not natively scan text[] into []string; they can
// be added later with a custom scanner if the UI needs them.
type deviceModel struct {
	ID             string    `gorm:"column:id;primaryKey"`
	Name           string    `gorm:"column:name;not null"`
	Vendor         string    `gorm:"column:vendor;not null"`
	DriverType     string    `gorm:"column:driver_type;not null"`
	Host           string    `gorm:"column:host;not null"`
	Port           int       `gorm:"column:port;not null;default:0"`
	TimeoutMS      int       `gorm:"column:timeout_ms;not null;default:30000"`
	PollIntervalMS int       `gorm:"column:poll_interval_ms;not null;default:30000"`
	Extra          JSONB     `gorm:"column:extra;type:jsonb;not null;default:'{}'"`
	Enabled        bool      `gorm:"column:enabled;not null;default:true"`
	SiteName       string    `gorm:"column:site_name"`
	IsActive       bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

// TableName returns the explicit table name for the device model.
func (deviceModel) TableName() string {
	return "devices"
}

// toDomain maps a deviceModel to the domain device.Device entity.
func (m deviceModel) toDomain() device.Device {
	extra := map[string]string(m.Extra)
	if extra == nil {
		extra = map[string]string{}
	}

	return device.Device{
		ID:             m.ID,
		Name:           m.Name,
		Vendor:         m.Vendor,
		DriverType:     m.DriverType,
		Host:           m.Host,
		Port:           m.Port,
		TimeoutMS:      m.TimeoutMS,
		PollIntervalMS: m.PollIntervalMS,
		Extra:          extra,
		Enabled:        m.Enabled,
		Tags:           []string{}, // tags not loaded by this model
	}
}

// fromDomain maps a domain device.Device to a deviceModel.
func fromDomain(d device.Device) deviceModel {
	extra := JSONB(d.Extra)
	if extra == nil {
		extra = JSONB{}
	}

	return deviceModel{
		ID:             d.ID,
		Name:           d.Name,
		Vendor:         d.Vendor,
		DriverType:     d.DriverType,
		Host:           d.Host,
		Port:           d.Port,
		TimeoutMS:      d.TimeoutMS,
		PollIntervalMS: d.PollIntervalMS,
		Extra:          extra,
		Enabled:        d.Enabled,
		SiteName:       "",
		IsActive:       true,
	}
}

// DeviceRepository implements port.DeviceRepository backed by PostgreSQL.
type DeviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository returns a port.DeviceRepository backed by GORM/Postgres.
func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// FindByID returns the device inventory record for id, or device.ErrNotFound
// if no such device exists.
func (r *DeviceRepository) FindByID(ctx context.Context, id string) (device.Device, error) {
	var m deviceModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return device.Device{}, fmt.Errorf("device %s: %w", id, device.ErrNotFound)
		}
		return device.Device{}, fmt.Errorf("device %s: %w", id, err)
	}
	return m.toDomain(), nil
}

// FindAll returns all devices ordered by name.
func (r *DeviceRepository) FindAll(ctx context.Context) ([]device.Device, error) {
	var models []deviceModel
	if err := r.db.WithContext(ctx).Order("name").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}

	devices := make([]device.Device, len(models))
	for i, m := range models {
		devices[i] = m.toDomain()
	}
	return devices, nil
}

// Create inserts a new device into the inventory.
func (r *DeviceRepository) Create(ctx context.Context, d device.Device) (device.Device, error) {
	m := fromDomain(d)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return device.Device{}, fmt.Errorf("create device: %w", err)
	}
	return m.toDomain(), nil
}

// Update modifies an existing device inventory record.
func (r *DeviceRepository) Update(ctx context.Context, d device.Device) (device.Device, error) {
	m := fromDomain(d)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return device.Device{}, fmt.Errorf("update device: %w", err)
	}
	return m.toDomain(), nil
}

// Delete removes a device from the inventory. The credentials table has
// ON DELETE CASCADE, so the associated credential row is removed as well.
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&deviceModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete device: %w", err)
	}
	return nil
}

// compile-time check that DeviceRepository implements the intended interface
// (the port package is the source of truth for the interface).
// This will fail if port.DeviceRepository gains methods not implemented here.
var _ port.DeviceRepository = (*DeviceRepository)(nil)
