package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/redis"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
)

type ContextManager struct {
	redisStore *redis.Store
	cfg        config.Config
}

func NewContextManager(redisStore *redis.Store, cfg config.Config) *ContextManager {
	return &ContextManager{
		redisStore: redisStore,
		cfg:        cfg,
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

	summary, _ := cm.redisStore.GetSessionSummary(ctx, customerNumber)
	if summary != "" {
		sb.WriteString("### RINGKASAN PERCAKAPAN SEBELUMNYA:\n")
		sb.WriteString(summary)
		sb.WriteString("\n\n")
	}

	systemPrompt = sb.String()

	sessionMsgs, err := cm.redisStore.GetSessionMessages(ctx, customerNumber)
	if err != nil {
		sessionMsgs = nil
	}

	windowSize := cm.cfg.SlidingWindowSize
	if len(sessionMsgs) > windowSize {
		sessionMsgs = sessionMsgs[len(sessionMsgs)-windowSize:]
	}

	history = append(sessionMsgs, llm.ChatMessage{
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
	ttl := time.Duration(cm.cfg.SessionTimeoutMinutes) * time.Minute

	msgs, err := cm.redisStore.GetSessionMessages(ctx, customerNumber)
	if err != nil {
		msgs = []llm.ChatMessage{}
	}

	msgs = append(msgs,
		llm.ChatMessage{Role: "user", Content: userMsg},
		llm.ChatMessage{Role: "assistant", Content: botMsg},
	)

	if err := cm.redisStore.SaveSessionMessages(ctx, customerNumber, msgs, ttl); err != nil {
		return err
	}

	return cm.redisStore.RefreshSessionTTL(ctx, customerNumber, ttl)
}

func (cm *ContextManager) SummarizeSessionIfLong(
	ctx context.Context,
	customerNumber string,
	provider port.LLMProvider,
) error {
	msgs, err := cm.redisStore.GetSessionMessages(ctx, customerNumber)
	if err != nil || len(msgs) < 15 {
		return nil
	}

	half := len(msgs) / 2
	oldMsgs := msgs[:half]
	recentMsgs := msgs[half:]

	var sb strings.Builder
	for _, m := range oldMsgs {
		sb.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
	}

	prompt := "Ringkas percakapan berikut dalam 1-2 kalimat padat yang mencatat poin penting dan kebutuhan user:"
	resp, err := provider.Chat(ctx, prompt, []llm.ChatMessage{{Role: "user", Content: sb.String()}}, 150)
	if err != nil {
		return err
	}

	ttl := time.Duration(cm.cfg.SessionTimeoutMinutes) * time.Minute
	_ = cm.redisStore.SetSessionSummary(ctx, customerNumber, resp.Content, ttl)
	_ = cm.redisStore.SaveSessionMessages(ctx, customerNumber, recentMsgs, ttl)

	return nil
}
