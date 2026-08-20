package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/llm"
)

// LLMProvider defines the interface for interacting with LLM services.
type LLMProvider interface {
	Chat(ctx context.Context, systemPrompt string, messages []llm.ChatMessage, maxTokens int) (*llm.ChatResponse, error)
	ChatWithTools(ctx context.Context, systemPrompt string, messages []llm.ChatMessage, tools []llm.Tool, maxTokens int) (*llm.ChatResponse, error)
	TestConnection(ctx context.Context) error
}
