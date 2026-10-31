package audit

import (
	"encoding/json"
	"time"
)

// CommandLog is the detailed, append-only trace of every device command
// attempt (READ, MUTATION pending/approved/executed, DENIED, EXPIRED).
// Distinct from AuditLog (business workflow) and from the legacy AuditEvent.
type CommandLog struct {
	ID             int64
	TenantID       string
	DeviceID       string
	ActorType      string
	ActorID        string
	SessionID      string
	Mode           string // READ | MUTATION
	Command        string
	CommandArgs    json.RawMessage
	ApprovalID     string
	RestorePointID string
	Status         string // SUCCESS | FAILED | DENIED | PENDING_APPROVAL | REJECTED | EXPIRED
	OutputSummary  string
	DurationMS     int64
	CreatedAt      time.Time
}
