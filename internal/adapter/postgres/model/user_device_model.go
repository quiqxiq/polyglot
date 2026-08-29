package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// UserDeviceModel is the GORM database model for the user_devices junction table.
type UserDeviceModel struct {
	UserID     uint      `gorm:"primaryKey;not null"`
	DeviceID   string    `gorm:"primaryKey;type:uuid;not null"`
	AssignedBy *uint     `gorm:"type:bigint"`
	CreatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName explicitly maps UserDeviceModel to the user_devices migration table.
func (UserDeviceModel) TableName() string { return "user_devices" }

// ToDomain converts a user-device database model to its domain representation.
func (m *UserDeviceModel) ToDomain() *customer.UserDeviceAssignment {
	if m == nil {
		return nil
	}
	return &customer.UserDeviceAssignment{
		UserID:     m.UserID,
		DeviceID:   m.DeviceID,
		AssignedBy: m.AssignedBy,
		CreatedAt:  m.CreatedAt,
	}
}
