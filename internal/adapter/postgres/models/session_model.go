package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/bot"
)

// WASessionModel is the GORM database model for WA Sessions.
// TableName eksplisit ke `wa_sessions` (migrasi 000002) — tanpa ini GORM
// menyimpan session ke `wa_session_models` (AutoMigrate dev) sehingga divergen
// dari migrasi (prod) dan FK wa_chats/wa_messages yang menunjuk ke wa_sessions
// tidak konsisten.
// Field JID memakai tag `column:jid` eksplisit karena NamingStrategy GORM
// memetakan akronim `JID` menjadi `j_id`, sementara migrasi 000002 dan pemakai
// lain memakai `jid`.
type WASessionModel struct {
	ID           uint   `gorm:"primaryKey"`
	DeviceName   string `gorm:"not null"`
	PhoneNumber  string
	JID          string `gorm:"column:jid"`
	Status       string `gorm:"default:offline"`
	IsBotEnabled bool   `gorm:"default:true"`
	ConnectedAt  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName maps WASessionModel ke tabel migrasi `wa_sessions`.
func (WASessionModel) TableName() string { return "wa_sessions" }

func (m *WASessionModel) ToDomain() *bot.WASession {
	if m == nil {
		return nil
	}
	return &bot.WASession{
		ID:           m.ID,
		DeviceName:   m.DeviceName,
		PhoneNumber:  m.PhoneNumber,
		JID:          m.JID,
		Status:       bot.WASessionStatus(m.Status),
		IsBotEnabled: m.IsBotEnabled,
		ConnectedAt:  m.ConnectedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func WASessionModelFromDomain(s *bot.WASession) *WASessionModel {
	if s == nil {
		return nil
	}
	return &WASessionModel{
		ID:           s.ID,
		DeviceName:   s.DeviceName,
		PhoneNumber:  s.PhoneNumber,
		JID:          s.JID,
		Status:       string(s.Status),
		IsBotEnabled: s.IsBotEnabled,
		ConnectedAt:  s.ConnectedAt,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}
