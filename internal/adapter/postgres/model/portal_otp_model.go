package model

import (
	"time"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
)

// PortalOTPModel is the GORM model for one-time portal login codes
// (migrasi 000019 — portal_otps).
type PortalOTPModel struct {
	ID         string    `gorm:"primaryKey"`
	TenantID   string    `gorm:"type:text;not null;default:tenant-default"`
	Phone      string    `gorm:"type:varchar(20);not null;index"`
	CodeHash   string    `gorm:"column:code_hash;type:varchar(255);not null"`
	Purpose    string    `gorm:"type:varchar(30);not null;default:PORTAL_LOGIN"`
	Attempts   int       `gorm:"not null;default:0"`
	ExpiresAt  time.Time `gorm:"not null"`
	ConsumedAt *time.Time
	CreatedAt  time.Time
}

func (PortalOTPModel) TableName() string { return "portal_otps" }

func (m *PortalOTPModel) ToDomain() domainCustomer.PortalOTP {
	if m == nil {
		return domainCustomer.PortalOTP{}
	}
	return domainCustomer.PortalOTP{
		ID: m.ID, TenantID: m.TenantID, Phone: m.Phone,
		CodeHash: m.CodeHash, Purpose: m.Purpose, Attempts: m.Attempts,
		ExpiresAt: m.ExpiresAt, ConsumedAt: m.ConsumedAt, CreatedAt: m.CreatedAt,
	}
}

func PortalOTPModelFromDomain(o domainCustomer.PortalOTP) *PortalOTPModel {
	return &PortalOTPModel{
		ID: o.ID, TenantID: o.TenantID, Phone: o.Phone,
		CodeHash: o.CodeHash, Purpose: o.Purpose, Attempts: o.Attempts,
		ExpiresAt: o.ExpiresAt, ConsumedAt: o.ConsumedAt, CreatedAt: o.CreatedAt,
	}
}
