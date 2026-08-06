package models

import (
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// DeviceModel is the GORM database model for network devices inventory.
type DeviceModel struct {
	ID             string `gorm:"primaryKey"`
	TenantID       string `gorm:"not null;index;default:tenant-default"`
	Name           string `gorm:"not null"`
	Vendor         string `gorm:"not null;default:mikrotik"`
	DriverType     string `gorm:"not null;default:mikrotik"`
	Host           string `gorm:"not null"`
	Port           int    `gorm:"not null;default:8728"`
	TimeoutMS      int    `gorm:"not null;default:10000"`
	PollIntervalMS int    `gorm:"not null;default:30000"`
	ExtraJSON      string `gorm:"type:text"`
	TagsJSON       string `gorm:"type:text"`
	Enabled        bool   `gorm:"not null;default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CredentialModel is the GORM database model for encrypted device credentials.
type CredentialModel struct {
	DeviceID  string `gorm:"primaryKey"`
	Username  string `gorm:"not null"`
	Password  string `gorm:"not null"`
	ExtraJSON string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

	return device.Device{
		ID:             m.ID,
		TenantID:       m.TenantID,
		Name:           m.Name,
		Vendor:         m.Vendor,
		DriverType:     m.DriverType,
		Host:           m.Host,
		Port:           m.Port,
		TimeoutMS:      m.TimeoutMS,
		PollIntervalMS: m.PollIntervalMS,
		Extra:          extra,
		Tags:           tags,
		Enabled:        m.Enabled,
	}
}

func DeviceModelFromDomain(d device.Device) *DeviceModel {
	extraJSON, _ := json.Marshal(d.Extra)
	tagsJSON, _ := json.Marshal(d.Tags)

	tenantID := d.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}

	return &DeviceModel{
		ID:             d.ID,
		TenantID:       tenantID,
		Name:           d.Name,
		Vendor:         d.Vendor,
		DriverType:     d.DriverType,
		Host:           d.Host,
		Port:           d.Port,
		TimeoutMS:      d.TimeoutMS,
		PollIntervalMS: d.PollIntervalMS,
		ExtraJSON:      string(extraJSON),
		TagsJSON:       string(tagsJSON),
		Enabled:        d.Enabled,
	}
}

func (c *CredentialModel) ToDomain() device.Credentials {
	if c == nil {
		return device.Credentials{}
	}

	var extra map[string]string
	if c.ExtraJSON != "" {
		_ = json.Unmarshal([]byte(c.ExtraJSON), &extra)
	}

	return device.Credentials{
		Username: c.Username,
		Password: c.Password,
		Extra:    extra,
	}
}

func CredentialModelFromDomain(deviceID string, c device.Credentials) *CredentialModel {
	extraJSON, _ := json.Marshal(c.Extra)
	return &CredentialModel{
		DeviceID:  deviceID,
		Username:  c.Username,
		Password:  c.Password,
		ExtraJSON: string(extraJSON),
	}
}
