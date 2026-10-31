package model

import (
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/registration"
)

// RegistrationEventModel is the GORM model for the PSB timeline/audit events
// (integrasi-router-billing-psb-global-smartapp.md §3.7 — registration_events).
type RegistrationEventModel struct {
	ID string `gorm:"column:id;primaryKey"`

	RegistrationID string `gorm:"column:registration_id;type:text;not null;index"`
	EventType      string `gorm:"column:event_type;type:varchar(50);not null"`
	ActorID        *uint  `gorm:"column:actor_id"`
	Title          string `gorm:"column:title;type:varchar(100);not null"`
	Description    string `gorm:"column:description;type:text"`

	// Metadata JSONB — representasi json.RawMessage mengikuti konvensi model
	// JSONB lain (mis. GatewayTransactionModel.RawCallback).
	Metadata json.RawMessage `gorm:"column:metadata;type:jsonb"`

	CreatedAt time.Time `gorm:"column:created_at;index"`
}

// TableName returns the database table name for registration events.
func (RegistrationEventModel) TableName() string {
	return "registration_events"
}

// ToDomain converts a registration event database model to its domain representation.
func (m *RegistrationEventModel) ToDomain() registration.RegistrationEvent {
	if m == nil {
		return registration.RegistrationEvent{}
	}
	return registration.RegistrationEvent{
		ID:             m.ID,
		RegistrationID: m.RegistrationID,
		EventType:      m.EventType,
		ActorID:        m.ActorID,
		Title:          m.Title,
		Description:    m.Description,
		Metadata:       m.Metadata,
		CreatedAt:      m.CreatedAt,
	}
}

// RegistrationEventModelFromDomain converts a registration event domain entity to a database model.
func RegistrationEventModelFromDomain(e registration.RegistrationEvent) *RegistrationEventModel {
	return &RegistrationEventModel{
		ID:             e.ID,
		RegistrationID: e.RegistrationID,
		EventType:      e.EventType,
		ActorID:        e.ActorID,
		Title:          e.Title,
		Description:    e.Description,
		Metadata:       e.Metadata,
		CreatedAt:      e.CreatedAt,
	}
}
