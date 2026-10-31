package port

import (
	"context"
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/aichat"
)

// AIChatRepository persists AI Ops conversations and their event stream.
type AIChatRepository interface {
	CreateConversation(ctx context.Context, c aichat.Conversation) error
	FindConversation(ctx context.Context, id string) (aichat.Conversation, error)
	UpdateConversation(ctx context.Context, c aichat.Conversation) error
	// AppendEvent assigns the next seq for the conversation transactionally.
	AppendEvent(ctx context.Context, conversationID, eventType string, payload json.RawMessage) (aichat.Event, error)
	ListEvents(ctx context.Context, conversationID string, sinceSeq int64, limit int) ([]aichat.Event, error)
	// DeleteEventsBefore removes chat events older than cutoff (retention job).
	DeleteEventsBefore(ctx context.Context, cutoff time.Time) (int, error)
}
