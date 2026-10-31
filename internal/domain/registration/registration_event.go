package registration

import (
	"encoding/json"
	"time"
)

// EventType PSB timeline.
const (
	EventCreated               = "REGISTRATION_CREATED"
	EventSurveyScheduled       = "SURVEY_SCHEDULED"
	EventSurveyCompleted       = "SURVEY_COMPLETED"
	EventApproved              = "APPROVED"
	EventInstallationScheduled = "INSTALLATION_SCHEDULED"
	EventInstallationStarted   = "INSTALLATION_STARTED"
	EventInstallationCompleted = "INSTALLATION_COMPLETED"
	EventConverted             = "CONVERTED"
	EventRejected              = "REJECTED"
	EventCancelled             = "CANCELLED"
)

// RegistrationEvent represents one entry in the PSB timeline
// (DATABASE-SCHEMA-ISP.md §2.3 — registration_events).
// Metadata JSONB mengikuti konvensi domain lain (json.RawMessage,
// mis. billing.GatewayTransaction.RawCallback).
type RegistrationEvent struct {
	ID             string          `json:"id"`
	RegistrationID string          `json:"registration_id"`
	EventType      string          `json:"event_type"`
	ActorID        *uint           `json:"actor_id,omitempty"`
	Title          string          `json:"title"`
	Description    string          `json:"description,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}
