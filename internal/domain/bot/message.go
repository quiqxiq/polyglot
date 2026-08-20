package bot

import "time"

// SenderType defines the sender of a message.
type SenderType string

const (
	SenderCustomer SenderType = "customer"
	SenderBot      SenderType = "bot"
	SenderAgent    SenderType = "agent"
)

// Message represents a single message within a conversation.
type Message struct {
	ID             uint       `json:"id"`
	ConversationID uint       `json:"conversation_id"`
	SenderType     SenderType `json:"sender_type"`
	Content        string     `json:"content"`
	TokenIn        int        `json:"token_in"`
	TokenOut       int        `json:"token_out"`
	LLMConfigID    *uint      `json:"llm_config_id"`
	CreatedAt      time.Time  `json:"created_at"`
}
