package bot

import "time"

// ConversationStatus defines status of a customer conversation.
type ConversationStatus string

const (
	StatusBot        ConversationStatus = "bot"
	StatusEscalation ConversationStatus = "escalation"
	StatusDone       ConversationStatus = "done"
)

// Conversation represents a chat conversation between a customer and the bot/agent.
type Conversation struct {
	ID               uint               `json:"id"`
	SessionID        uint               `json:"session_id"`
	CustomerWANumber string             `json:"customer_wa_number"`
	Status           ConversationStatus `json:"status"`
	AssignedAgentID  *uint              `json:"assigned_agent_id"`
	StartedAt        time.Time          `json:"started_at"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`

	Messages []Message `json:"messages,omitempty"`
}
