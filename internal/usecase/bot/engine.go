package bot

import (
	"context"
	"errors"
	"fmt"

	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
	"github.com/quixiq/polyglot/pkg/logger"
)

const (
	SenderCustomer = "customer"
	SenderBot      = "bot"

	// historyLimit is the number of recent messages fed to the LLM as context (max 6 to preserve tokens).
	historyLimit = 6
)

// Config holds configuration parameters needed by the bot usecase.
type Config struct {
	SystemPrompt       string
	AllowedTopics      []string
	LLMMaxOutputTokens int
	BurstLimit         int
	BurstWindowSeconds int
	Mute1HourSeconds   int
	Ban24HourSeconds   int
	DailyChatLimit     int
	WhitelistPhones    []string
}

// ProviderFactory creates an LLM provider from a stored LLM configuration.
type ProviderFactory func(cfg *llm.Config) (port.LLMProvider, error)

// PromptBuilder constructs composite system prompt from active modular skills and global SOP.
type PromptBuilder interface {
	BuildCompositeSystemPrompt(ctx context.Context) (string, error)
}

// Engine orchestrates the WhatsApp customer-service bot.
type Engine struct {
	cfg           Config
	cache         port.CacheStore
	waGateway     port.WhatsAppGateway
	convService   *convUC.ConversationService
	promptBuilder PromptBuilder
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
	promptBuilder PromptBuilder,
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
		promptBuilder: promptBuilder,
		llmConfigRepo: llmConfigRepo,
		chatRepo:      chatRepo,
		rateLimiter: NewRateLimiter(cache, config.Config{
			BotBurstLimit:      cfg.BurstLimit,
			BotBurstWindowSecs: cfg.BurstWindowSeconds,
			BotMute1HourSecs:   cfg.Mute1HourSeconds,
			BotBan24HourSecs:   cfg.Ban24HourSeconds,
			BotDailyChatLimit:  cfg.DailyChatLimit,
			BotWhitelistPhones: cfg.WhitelistPhones,
		}),
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

	if e.convService == nil {
		return errors.New("bot engine: conversation service not initialized")
	}

	conv, err := e.convService.GetOrCreateConversation(ctx, sessionID, customerNumber)
	if err != nil {
		return fmt.Errorf("failed to get/create conversation: %w", err)
	}

	msg, err := e.convService.AddMessage(ctx, conv.ID, SenderCustomer, messageContent, 0, 0)
	if err != nil {
		logger.WithComponent("BotEngine").Warnf("Failed to persist incoming customer message: %v", err)
	}
	if e.publisher != nil && msg != nil {
		e.publisher.PublishEvent("new_message", msg)
	}

	if conv.Status == bot.StatusEscalation {
		if conv.AssignedAgentID != nil {
			logger.WithComponent("BotEngine").Infof("Conversation %d is assigned to human agent %d. Skipping automated bot reply.", conv.ID, *conv.AssignedAgentID)
			return nil
		}
		// Percakapan dieskalasi otomatis oleh sistem karena error sementara (tanpa admin manusia).
		// Pulihkan status ke 'bot' agar pesan baru dari pelanggan langsung dilayani kembali oleh AI.
		logger.WithComponent("BotEngine").Infof("Conversation %d was unassigned escalation. Auto-recovering to bot mode.", conv.ID)
		_ = e.convService.ResetBot(ctx, conv.ID)
		conv.Status = bot.StatusBot
	}

	if e.chatRepo != nil {
		enabled, err := e.chatRepo.IsChatBotEnabled(ctx, sessionID, chatJID)
		if err != nil {
			logger.WithComponent("BotEngine").Warnf("Failed to query bot_enabled for chat %s: %v — assuming enabled", chatJID, err)
			enabled = true
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
	case StatusWarned, StatusDailyQuotaExceeded:
		logger.WithComponent("BotEngine").Warnf("Number %s rate limit notification (%v).", customerNumber, rateResult.Status)
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, rateResult.Message, 0, 0, nil)
	}

	if !e.guardrail.IsTopicAllowed(messageContent) {
		offTopicReply := e.guardrail.FormatOffTopicResponse()
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, offTopicReply, 0, 0, nil)
	}

	var compositePrompt string
	if e.promptBuilder != nil {
		if cp, err := e.promptBuilder.BuildCompositeSystemPrompt(ctx); err == nil && cp != "" {
			compositePrompt = cp
		} else if err != nil {
			logger.WithComponent("BotEngine").Warnf("Failed to build composite skills prompt: %v", err)
		}
	}

	if e.llmConfigRepo == nil {
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, "Maaf, sistem AI kami saat ini tidak tersedia.", 0, 0, nil)
	}

	llmCfg, err := e.llmConfigRepo.FindActive(ctx)
	if err != nil {
		logger.WithComponent("BotEngine").Warnf("No active LLM config found for conversation %d", conv.ID)
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, "Maaf, konfigurasi asisten AI belum aktif. Pesan Anda telah kami terima.", 0, 0, nil)
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

	systemPrompt, chatMessages, err := e.contextMgr.BuildPromptContext(ctx, conv.ID, history, compositePrompt)
	if err != nil {
		return fmt.Errorf("failed to build prompt context: %w", err)
	}

	if e.providerF == nil {
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, "Maaf, konfigurasi asisten AI belum aktif. Pesan Anda telah kami terima.", 0, 0, nil)
	}

	provider, err := e.providerF(llmCfg)
	if err != nil {
		logger.WithComponent("BotEngine").Errorf("Failed to build LLM provider for conversation %d: %v", conv.ID, err)
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, "Maaf, sistem asisten AI sedang dalam pemeliharaan. Silakan coba kembali beberapa saat lagi.", 0, 0, nil)
	}

	maxTokens := llmCfg.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = e.cfg.LLMMaxOutputTokens
	}

	resp, err := provider.Chat(ctx, systemPrompt, chatMessages, maxTokens)
	if err != nil {
		logger.WithComponent("BotEngine").Errorf("LLM call failed for conversation %d: %v", conv.ID, err)
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, "Mohon maaf, sistem AI kami sedang sibuk atau mengalami kendala koneksi. Silakan ulangi pesan Anda dalam beberapa saat.", 0, 0, nil)
	}

	reply := e.guardrail.SanitizeResponse(resp.Content)
	if reply == "" {
		logger.WithComponent("BotEngine").Warnf("Empty LLM reply for conversation %d", conv.ID)
		reply = "Maaf, saya tidak dapat memahami permintaan Anda. Bisakah Anda menjelaskan lebih detail?"
	}

	if err := e.sendBotReply(ctx, conv.ID, sessionID, chatJID, reply, resp.TokenIn, resp.TokenOut, &llmCfg.ID); err != nil {
		return err
	}

	_ = e.rateLimiter.IncrementDailyQuota(ctx, customerNumber)
	_ = e.contextMgr.SaveMessageToSession(ctx, conv.ID, messageContent, reply)
	_ = e.contextMgr.SummarizeSessionIfLong(ctx, conv.ID, provider)

	return nil
}

// ResetRateLimit resets all rate limit and daily quota counters for a customer phone number.
func (e *Engine) ResetRateLimit(ctx context.Context, customerNumber string) error {
	if e.rateLimiter == nil {
		return nil
	}
	return e.rateLimiter.ResetRateLimit(ctx, customerNumber)
}

// GetRateLimitStatus queries the rate limit and daily quota state for a customer phone number.
func (e *Engine) GetRateLimitStatus(ctx context.Context, customerNumber string) (*RateLimitStatusInfo, error) {
	if e.rateLimiter == nil {
		return &RateLimitStatusInfo{PhoneNumber: customerNumber}, nil
	}
	return e.rateLimiter.GetRateLimitStatus(ctx, customerNumber)
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
func (e *Engine) sendBotReply(ctx context.Context, convID uint, sessionID uint, targetJID string, content string, tokenIn, tokenOut int, llmConfigID *uint) error {
	if err := e.waGateway.SendMessage(sessionID, targetJID, content); err != nil {
		logger.WithComponent("BotEngine").Errorf("Failed to send reply to %s: %v", targetJID, err)
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
func (e *Engine) escalateWithFallback(ctx context.Context, convID uint, sessionID uint, targetJID string, reply string) error {
	if err := e.convService.Escalate(ctx, convID); err != nil {
		logger.WithComponent("BotEngine").Errorf("Failed to escalate conversation %d: %v", convID, err)
	}
	return e.sendBotReply(ctx, convID, sessionID, targetJID, reply, 0, 0, nil)
}
