package bot

import "time"

// ConversationContext aggregates the LLM-facing state of a conversation for
// agent-side dashboards: the running summary, recent history fed to the LLM,
// and accumulated token usage. Dihasilkan oleh bot engine dari ConversationService
// (DB) + ContextManager (cache Redis), bukan dari driver WhatsApp.
type ConversationContext struct {
	ConversationID uint
	Status         ConversationStatus
	ClientPhone    string
	Summary        string
	RecentMessages []Message
	TotalTokenIn   int64
	TotalTokenOut  int64
	TotalLLMCalls  int64
	UpdatedAt      time.Time
}
