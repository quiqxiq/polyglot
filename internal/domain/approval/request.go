package approval

import (
	"encoding/json"
	"time"
)

// Status is the approval lifecycle state (PLAN-AI-ROUTER-OPS v2).
type Status string

const (
	StatusPending   Status = "PENDING"
	StatusApproved  Status = "APPROVED"
	StatusExecuting Status = "EXECUTING"
	StatusExecuted  Status = "EXECUTED"
	StatusFailed    Status = "FAILED"
	StatusRejected  Status = "REJECTED"
	StatusExpired   Status = "EXPIRED"
	StatusCancelled Status = "CANCELLED"
)

// ActorType identifies who created the proposal. AI/MCP proposals are AGENT;
// only a USER session may decide an approval.
type ActorType string

const (
	ActorAgent  ActorType = "AGENT"
	ActorUser   ActorType = "USER"
	ActorSystem ActorType = "SYSTEM"
)

// Request is one human-in-the-loop approval for a router mutation.
//
// Diff/Validation/ExecutionResult are stored as raw JSON so the domain stays
// vendor-agnostic; usecase/network owns their concrete shape.
type Request struct {
	ID          string
	TenantID    string
	DeviceID    string
	SessionID   string
	ActorType   ActorType
	ActorID     string
	RequestedBy string

	Command       string
	Args          map[string]string
	Revision      int
	EditedCommand *string
	EditedArgs    map[string]string
	EditedBy      *uint
	EditCount     int

	Diff           json.RawMessage
	Validation     json.RawMessage
	RestorePointID string

	Status     Status
	Reason     string
	ApprovedBy *uint
	DecidedBy  *uint

	IdempotencyKey  string
	ExecutionNonce  string
	ExecutionResult json.RawMessage

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiresAt  time.Time
	DecidedAt  *time.Time
	ExecutedAt *time.Time
}

// EffectiveCommand returns the edited command when present, else the original.
func (r Request) EffectiveCommand() string {
	if r.EditedCommand != nil && *r.EditedCommand != "" {
		return *r.EditedCommand
	}
	return r.Command
}

// EffectiveArgs returns the edited args when present, else the original.
func (r Request) EffectiveArgs() map[string]string {
	if r.EditedArgs != nil {
		return r.EditedArgs
	}
	return r.Args
}

// IsExpired reports whether the approval TTL has passed.
func (r Request) IsExpired(now time.Time) bool {
	return !r.ExpiresAt.IsZero() && now.After(r.ExpiresAt)
}

// CanDecide reports whether the request is still in a decidable state.
func (r Request) CanDecide(now time.Time) bool {
	return r.Status == StatusPending && !r.IsExpired(now)
}

// CanExecute reports whether the request can be claimed for execution.
func (r Request) CanExecute(now time.Time) bool {
	return r.Status == StatusApproved && !r.IsExpired(now)
}
