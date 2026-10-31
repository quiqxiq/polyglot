package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/port"
)

// ActiveDeviceModel is the GORM model for active_devices (migrasi 000041
// PLAN-AI-ROUTER-OPS v2): binding server-side "router mana yang sedang
// dipilih operator" sebagai fallback/audit di samping Redis.
//
// Kolom updated_at dipakai sebagai BATAS KEDALUWARSA binding (now + TTL) agar
// TTL dapat ditegakkan tanpa kolom tambahan; Get memfilter updated_at > now.
type ActiveDeviceModel struct {
	TenantID  string  `gorm:"column:tenant_id;primaryKey"`
	UserID    string  `gorm:"column:user_id;primaryKey"`
	SessionID string  `gorm:"column:session_id;primaryKey"`
	DeviceID  *string `gorm:"column:device_id;type:uuid"`

	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the database table name for active device bindings.
func (ActiveDeviceModel) TableName() string {
	return "active_devices"
}

// ToDomain converts an active device row to the port contract value.
// UpdatedAt berisi waktu kedaluwarsa binding (now + TTL saat Set).
func (m *ActiveDeviceModel) ToDomain() port.ActiveDevice {
	if m == nil {
		return port.ActiveDevice{}
	}
	return port.ActiveDevice{
		TenantID:  m.TenantID,
		UserID:    m.UserID,
		SessionID: m.SessionID,
		DeviceID:  derefOrEmpty(m.DeviceID),
		UpdatedAt: m.UpdatedAt,
	}
}

// ActiveDeviceModelFromDomain converts a binding to a database row.
// expiresAt adalah nilai yang disimpan di kolom updated_at (batas TTL).
func ActiveDeviceModelFromDomain(a port.ActiveDevice, expiresAt time.Time) *ActiveDeviceModel {
	tenantID := a.TenantID
	if tenantID == "" {
		tenantID = aiDefaultTenantID
	}
	return &ActiveDeviceModel{
		TenantID:  tenantID,
		UserID:    a.UserID,
		SessionID: a.SessionID,
		DeviceID:  aiNullableTextID(a.DeviceID),
		UpdatedAt: expiresAt,
	}
}
