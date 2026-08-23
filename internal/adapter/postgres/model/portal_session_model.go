package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// PortalSessionModel is the GORM model for customer self-service portal
// sessions (DATABASE-SCHEMA-ISP.md §2.4 — customer_portal_sessions).
type PortalSessionModel struct {
	ID           string `gorm:"primaryKey"`
	TenantID     string `gorm:"type:text;not null;default:tenant-default"`
	CustomerID   string `gorm:"type:text;not null;index"`
	SessionToken string `gorm:"type:varchar(128);unique;not null"`
	IPAddress    string `gorm:"type:varchar(45)"`
	UserAgent    string `gorm:"type:text"`
	ExpiresAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time
}

func (PortalSessionModel) TableName() string {
	return "customer_portal_sessions"
}

func (m *PortalSessionModel) ToDomain() customer.PortalSession {
	if m == nil {
		return customer.PortalSession{}
	}
	return customer.PortalSession{
		ID:           m.ID,
		TenantID:     m.TenantID,
		CustomerID:   m.CustomerID,
		SessionToken: m.SessionToken,
		IPAddress:    m.IPAddress,
		UserAgent:    m.UserAgent,
		ExpiresAt:    m.ExpiresAt,
		CreatedAt:    m.CreatedAt,
	}
}

func PortalSessionModelFromDomain(s customer.PortalSession) *PortalSessionModel {
	return &PortalSessionModel{
		ID:           s.ID,
		TenantID:     s.TenantID,
		CustomerID:   s.CustomerID,
		SessionToken: s.SessionToken,
		IPAddress:    s.IPAddress,
		UserAgent:    s.UserAgent,
		ExpiresAt:    s.ExpiresAt,
		CreatedAt:    s.CreatedAt,
	}
}
