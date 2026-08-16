package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
	"github.com/quixiq/polyglot/pkg/logger"
)

const (
	SenderCustomer = "customer"
	SenderBot      = "bot"

	// historyLimit is the number of recent messages fed to the LLM as context.
	historyLimit = 20
)

// Config holds configuration parameters needed by the bot usecase.
type Config struct {
	SystemPrompt       string
	AllowedTopics      []string
	LLMMaxOutputTokens int
}

// ProviderFactory creates an LLM provider from a stored LLM configuration.
type ProviderFactory func(cfg *llm.Config) (port.LLMProvider, error)

// Engine orchestrates the WhatsApp customer-service bot.
type Engine struct {
	cfg           Config
	cache         port.CacheStore
	waGateway     port.WhatsAppGateway
	convService   *convUC.ConversationService
	retriever     port.KnowledgeRetriever
	chat          port.KnowledgeChat
	llmConfigRepo port.LLMConfigRepository
	chatRepo      port.ChatRepository
	rateLimiter   *RateLimiter
	guardrail     *Guardrail
	contextMgr    *ContextManager
	publisher     port.EventPublisher
	providerF     ProviderFactory
}

func NewEngine(
	cfg Config,
	cache port.CacheStore,
	waGateway port.WhatsAppGateway,
	convService *convUC.ConversationService,
	retriever port.KnowledgeRetriever,
	chat port.KnowledgeChat,
	llmConfigRepo port.LLMConfigRepository,
	chatRepo port.ChatRepository,
	publisher port.EventPublisher,
	providerFactory ProviderFactory,
) *Engine {
	return &Engine{
		cfg:           cfg,
		cache:         cache,
		waGateway:     waGateway,
		convService:   convService,
		retriever:     retriever,
		chat:          chat,
		llmConfigRepo: llmConfigRepo,
		chatRepo:      chatRepo,
		rateLimiter:   NewRateLimiter(cache),
		guardrail:     NewGuardrail(cfg.AllowedTopics),
		contextMgr:    NewContextManager(cache, cfg.SystemPrompt),
		publisher:     publisher,
		providerF:     providerFactory,
	}
}

// HandleIncomingMessage processes a single incoming WhatsApp text message.
func (e *Engine) HandleIncomingMessage(ctx context.Context, sessionID uint, chatJID string, customerNumber string, messageContent string) error {
	if e.waGateway == nil {
		return errors.New("bot engine: whatsapp gateway not initialized")
	}

	logger.WithComponent("BotEngine").Infof("Processing message from %s (Session %d): %s", customerNumber, sessionID, messageContent)

	conv, err := e.convService.GetOrCreateConversation(ctx, sessionID, customerNumber)
	if err != nil {
		return fmt.Errorf("failed to get/create conversation: %w", err)
	}

	custMsg, err := e.convService.AddMessageWithConfig(ctx, conv.ID, SenderCustomer, messageContent, 0, 0, nil)
	if err != nil {
		logger.WithComponent("BotEngine").Warnf("Failed to persist customer message: %v", err)
	}
	if err == nil && e.publisher != nil {
		e.publisher.PublishEvent("new_message", custMsg)
	}

	if conv.Status == bot.StatusEscalation {
		logger.WithComponent("BotEngine").Infof("Conversation %d is in escalation mode. Message recorded without auto-reply.", conv.ID)
		return nil
	}

	if e.chatRepo != nil {
		enabled, err := e.chatRepo.IsChatBotEnabled(ctx, sessionID, chatJID)
		if err != nil {
			logger.WithComponent("BotEngine").Warnf("Failed to check bot_enabled for %s: %v", chatJID, err)
		}
		if !enabled {
			logger.WithComponent("BotEngine").Infof("Bot nonaktif untuk chat %s. Pesan dicatat tanpa auto-reply.", chatJID)
			return nil
		}
	}

	rateResult, err := e.rateLimiter.Check(ctx, customerNumber, messageContent)
	if err != nil {
		logger.WithComponent("BotEngine").Warnf("Rate limit check error: %v", err)
	}

	switch rateResult.Status {
	case StatusMuted, StatusBlocked:
		logger.WithComponent("BotEngine").Infof("Number %s is %v. Ignoring message to save tokens.", customerNumber, rateResult.Status)
		return nil
	case StatusWarned:
		logger.WithComponent("BotEngine").Warnf("Number %s hit rate limit warning.", customerNumber)
		return e.sendBotReply(ctx, conv.ID, sessionID, customerNumber, rateResult.Message, 0, 0, nil)
	}

	if !e.guardrail.IsTopicAllowed(messageContent) {
		offTopicReply := e.guardrail.FormatOffTopicResponse()
		return e.sendBotReply(ctx, conv.ID, sessionID, customerNumber, offTopicReply, 0, 0, nil)
	}

	if e.chat != nil {
		chatResult, chatErr := e.chat.Chat(ctx, messageContent, fmt.Sprintf("conv-%d", conv.ID))
		if chatErr == nil && strings.TrimSpace(chatResult.Content) != "" {
			reply := e.guardrail.SanitizeResponse(chatResult.Content)
			if reply != "" {
				if err := e.sendBotReply(ctx, conv.ID, sessionID, customerNumber, reply, chatResult.TokenIn, chatResult.TokenOut, nil); err != nil {
					return err
				}
				_ = e.contextMgr.SaveMessageToSession(ctx, conv.ID, messageContent, reply)
				return nil
			}
		}
		logger.WithComponent("BotEngine").Warnf("AnythingLLM chat failed/unavailable for conversation %d: %v — falling back to local LLM", conv.ID, chatErr)
	}

	relEntries, _ := e.retriever.Retrieve(ctx, messageContent)

	if e.llmConfigRepo == nil {
		fallbackReply := "Maaf, kami kesulitan memproses pesan Anda. Pesan telah kami teruskan ke admin kami."
		return e.escalateWithFallback(ctx, conv.ID, sessionID, customerNumber, fallbackReply)
	}

	llmCfg, err := e.llmConfigRepo.FindActive(ctx)
	if err != nil {
		logger.WithComponent("BotEngine").Warnf("No active LLM config found! Escalating conversation %d", conv.ID)
		fallbackReply := "Maaf, sistem AI kami sedang dalam pemeliharaan. Pesan Anda telah diteruskan ke tim admin."
		return e.escalateWithFallback(ctx, conv.ID, sessionID, customerNumber, fallbackReply)
	}

	history, err := e.convService.GetRecentHistory(ctx, conv.ID, historyLimit)
	if err != nil || len(history) == 0 {
		if err != nil {
			logger.WithComponent("BotEngine").Warnf("Failed to load recent history for conversation %d: %v", conv.ID, err)
		}
		history = []bot.Message{
			{ConversationID: conv.ID, SenderType: bot.SenderCustomer, Content: messageContent},
		}
	}

	systemPrompt, chatMessages, err := e.contextMgr.BuildPromptContext(ctx, conv.ID, history, relEntries)
	if err != nil {
		return fmt.Errorf("failed to build prompt context: %w", err)
	}

	if e.providerF == nil {
		return e.escalateWithFallback(ctx, conv.ID, sessionID, customerNumber, "Maaf, sistem AI kami sedang tidak tersedia. Pesan Anda telah diteruskan ke tim admin.")
	}

	provider, err := e.providerF(llmCfg)
	if err != nil {
		logger.WithComponent("BotEngine").Errorf("Failed to build LLM provider for conversation %d: %v", conv.ID, err)
		return e.escalateWithFallback(ctx, conv.ID, sessionID, customerNumber, "Maaf, sistem AI kami sedang tidak tersedia. Pesan Anda telah diteruskan ke tim admin.")
	}

	maxTokens := llmCfg.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = e.cfg.LLMMaxOutputTokens
	}

	resp, err := provider.Chat(ctx, systemPrompt, chatMessages, maxTokens)
	if err != nil {
		logger.WithComponent("BotEngine").Errorf("LLM call failed for conversation %d: %v", conv.ID, err)
		return e.escalateWithFallback(ctx, conv.ID, sessionID, customerNumber, "Maaf, sistem AI kami sedang bermasalah. Pesan Anda telah diteruskan ke tim admin kami.")
	}

	reply := e.guardrail.SanitizeResponse(resp.Content)
	if reply == "" {
		logger.WithComponent("BotEngine").Warnf("Empty LLM reply for conversation %d", conv.ID)
		reply = "Maaf, saya tidak menemukan jawaban yang tepat. Tim kami akan segera menghubungi Anda."
	}

	if err := e.sendBotReply(ctx, conv.ID, sessionID, customerNumber, reply, resp.TokenIn, resp.TokenOut, &llmCfg.ID); err != nil {
		return err
	}

	_ = e.contextMgr.SaveMessageToSession(ctx, conv.ID, messageContent, reply)
	_ = e.contextMgr.SummarizeSessionIfLong(ctx, conv.ID, provider)

	return nil
}

// GetConversationContext aggregates the LLM-facing state of a conversation.
func (e *Engine) GetConversationContext(ctx context.Context, convID uint) (*bot.ConversationContext, error) {
	conv, err := e.convService.GetConversationWithMessages(ctx, convID)
	if err != nil {
		return nil, err
	}

	info := &bot.ConversationContext{
		ConversationID: conv.ID,
		Status:         conv.Status,
		ClientPhone:    conv.CustomerWANumber,
		Summary:        e.contextMgr.GetSummary(ctx, conv.ID),
		UpdatedAt:      conv.UpdatedAt,
	}

	const maxRecent = 20
	msgs := conv.Messages
	if len(msgs) > maxRecent {
		msgs = msgs[len(msgs)-maxRecent:]
	}
	info.RecentMessages = msgs

	for _, m := range conv.Messages {
		info.TotalTokenIn += int64(m.TokenIn)
		info.TotalTokenOut += int64(m.TokenOut)
		if m.LLMConfigID != nil {
			info.TotalLLMCalls++
		}
	}
	return info, nil
}

// sendBotReply sends a bot message via the WhatsApp gateway and persists it.
func (e *Engine) sendBotReply(ctx context.Context, convID uint, sessionID uint, customerNumber string, content string, tokenIn, tokenOut int, llmConfigID *uint) error {
	if err := e.waGateway.SendMessage(sessionID, customerNumber, content); err != nil {
		logger.WithComponent("BotEngine").Errorf("Failed to send reply to %s: %v", customerNumber, err)
		return fmt.Errorf("failed to send bot reply: %w", err)
	}

	botMsg, err := e.convService.AddMessageWithConfig(ctx, convID, SenderBot, content, tokenIn, tokenOut, llmConfigID)
	if err != nil {
		logger.WithComponent("BotEngine").Warnf("Failed to persist bot message: %v", err)
	}
	if e.publisher != nil && botMsg != nil {
		e.publisher.PublishEvent("new_message", botMsg)
	}
	return nil
}

// escalateWithFallback marks the conversation as needing a human agent, sends
// a fallback reply and persists it.
func (e *Engine) escalateWithFallback(ctx context.Context, convID uint, sessionID uint, customerNumber string, reply string) error {
	if err := e.convService.Escalate(ctx, convID); err != nil {
		logger.WithComponent("BotEngine").Errorf("Failed to escalate conversation %d: %v", convID, err)
	}
	return e.sendBotReply(ctx, convID, sessionID, customerNumber, reply, 0, 0, nil)
}
