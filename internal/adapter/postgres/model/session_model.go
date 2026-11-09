package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/whatsapp"
)

// WASessionModel is the GORM database model for WA Sessions.
type WASessionModel struct {
	ID          uint   `gorm:"primaryKey"`
	DeviceName  string `gorm:"not null"`
	PhoneNumber string
	JID         string `gorm:"column:jid"`
	Status      string `gorm:"default:offline"`
	ConnectedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName maps WASessionModel ke tabel migrasi `wa_sessions`.
func (WASessionModel) TableName() string { return "wa_sessions" }

// ToDomain converts a WhatsApp session database model to its domain representation.
func (m *WASessionModel) ToDomain() *whatsapp.WASession {
	if m == nil {
		return nil
	}
	return &whatsapp.WASession{
		ID:          m.ID,
		DeviceName:  m.DeviceName,
		PhoneNumber: m.PhoneNumber,
		JID:         m.JID,
		Status:      whatsapp.WASessionStatus(m.Status),
		ConnectedAt: m.ConnectedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// WASessionModelFromDomain converts a WhatsApp session domain entity to a database model.
func WASessionModelFromDomain(s *whatsapp.WASession) *WASessionModel {
	if s == nil {
		return nil
	}
	return &WASessionModel{
		ID:          s.ID,
		DeviceName:  s.DeviceName,
		PhoneNumber: s.PhoneNumber,
		JID:         s.JID,
		Status:      string(s.Status),
		ConnectedAt: s.ConnectedAt,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}
