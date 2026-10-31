package aichat

import (
	"encoding/json"
	"time"
)

const (
	StatusActive   = "ACTIVE"
	StatusArchived = "ARCHIVED"
)

// Conversation is one AI Ops chat thread tied to one device (immutable).
type Conversation struct {
	ID            string
	TenantID      string
	UserID        uint
	DeviceID      string
	ModelConfigID string
	Title         string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ArchivedAt    *time.Time
}

// Event types persisted for reconnect and UI rendering.
const (
	EventText       = "text"
	EventReasoning  = "reasoning"
	EventToolCall   = "tool_call"
	EventToolResult = "tool_result"
	EventPlanStep   = "plan_step"
	EventTerminal   = "terminal"
	EventApproval   = "approval"
	EventCheckpoint = "checkpoint"
	EventUsage      = "usage"
	EventDone       = "done"
	EventError      = "error"
)

// Event is one persisted stream event.
type Event struct {
	ID             int64
	ConversationID string
	Seq            int64
	Type           string
	Payload        json.RawMessage
	CreatedAt      time.Time
}
