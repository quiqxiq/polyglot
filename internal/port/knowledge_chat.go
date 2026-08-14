package port

import "context"

// KnowledgeChatResult is the outcome of a knowledge-workspace chat call:
// the LLM text answer plus the document sources it grounded the answer on.
type KnowledgeChatResult struct {
	Content  string
	Sources  []string
	TokenIn  int
	TokenOut int
}

// KnowledgeChat defines the interface for asking a knowledge workspace to
// answer a message directly (RAG + LLM in one call, with per-session chat
// memory). Implemented by AnythingLLM's workspace chat API; dipakai engine
// bot sebagai primary LLM path dengan fallback ke LLM lokal proyek.
type KnowledgeChat interface {
	Chat(ctx context.Context, message string, sessionID string) (KnowledgeChatResult, error)
}
