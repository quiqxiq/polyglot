package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
)

type ContextManager struct {
	cache port.CacheStore
	cfg   config.Config
}

func NewContextManager(cache port.CacheStore, cfg config.Config) *ContextManager {
	return &ContextManager{
		cache: cache,
		cfg:   cfg,
	}
}

func (cm *ContextManager) BuildPromptContext(
	ctx context.Context,
	customerNumber string,
	userMessage string,
	knowledgeEntries []knowledge.KnowledgeEntry,
) (systemPrompt string, history []llm.ChatMessage, err error) {
	var sb strings.Builder
	sb.WriteString(cm.cfg.SystemPrompt)
	sb.WriteString("\n\n")

	if len(knowledgeEntries) > 0 {
		sb.WriteString("### BASIS PENGETAHUAN LOKAL (GNET):\n")
		sb.WriteString("Gunakan informasi berikut sebagai acuan utama untuk menjawab:\n")
		for _, entry := range knowledgeEntries {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", entry.Title, entry.Content))
		}
		sb.WriteString("\n")
	}

	if cm.cache != nil {
		summary, _ := cm.cache.Get(ctx, "summary:"+customerNumber)
		if summary != "" {
			sb.WriteString("### RINGKASAN PERCAKAPAN SEBELUMNYA:\n")
			sb.WriteString(summary)
			sb.WriteString("\n\n")
		}
	}

	systemPrompt = sb.String()

	history = append(history, llm.ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	return systemPrompt, history, nil
}

func (cm *ContextManager) SaveMessageToSession(
	ctx context.Context,
	customerNumber string,
	userMsg string,
	botMsg string,
) error {
	if cm.cache != nil {
		_ = cm.cache.Set(ctx, "history:"+customerNumber, userMsg+" | "+botMsg, 86400)
	}
	return nil
}

func (cm *ContextManager) SummarizeSessionIfLong(
	ctx context.Context,
	customerNumber string,
	provider port.LLMProvider,
) error {
	return nil
}
