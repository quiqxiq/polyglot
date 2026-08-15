package bot

import (
	"context"
	"fmt"

	llmprovider "github.com/quixiq/polyglot/internal/adapter/llm"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/redis"
	"github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/business"
	"github.com/quixiq/polyglot/pkg/logger"
)

type Engine struct {
	cfg         config.Config
	pgStore     *postgres.Store
	redisStore  *redis.Store
	waGateway   port.WhatsAppGateway
	convService *business.ConversationService
	retriever   port.KnowledgeRetriever
	rateLimiter *RateLimiter
	guardrail   *Guardrail
	contextMgr  *ContextManager
	sseHub      *ws.SSEHub
}

func NewEngine(
	cfg config.Config,
	pgStore *postgres.Store,
	redisStore *redis.Store,
	waGateway port.WhatsAppGateway,
	convService *business.ConversationService,
	retriever port.KnowledgeRetriever,
	sseHub *ws.SSEHub,
) *Engine {
	return &Engine{
		cfg:         cfg,
		pgStore:     pgStore,
		redisStore:  redisStore,
		waGateway:   waGateway,
		convService: convService,
		retriever:   retriever,
		rateLimiter: NewRateLimiter(redisStore, cfg),
		guardrail:   NewGuardrail(cfg),
		contextMgr:  NewContextManager(redisStore, cfg),
		sseHub:      sseHub,
	}
}

func (e *Engine) HandleIncomingMessage(ctx context.Context, sessionID uint, customerNumber string, messageContent string) error {
	logger.WithFields(logger.Fields{
		"customer_number": customerNumber,
		"session_id":      sessionID,
		"content":         messageContent,
	}).Info("[BotEngine] Processing incoming message")

	session, err := e.pgStore.FindSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	conv, err := e.convService.GetOrCreateConversation(sessionID, customerNumber)
	if err != nil {
		return fmt.Errorf("failed to get/create conversation: %w", err)
	}

	custMsg, err := e.convService.AddMessage(conv.ID, "customer", messageContent, 0, 0)
	if err == nil && e.sseHub != nil {
		e.sseHub.Broadcast("new_message", custMsg)
	}

	if !session.IsBotEnabled || conv.Status == bot.StatusEscalation {
		logger.WithFields(logger.Fields{
			"session_id":      sessionID,
			"conversation_id": conv.ID,
		}).Info("[BotEngine] Bot disabled or in escalation mode. Message recorded.")
		return nil
	}

	rateResult, err := e.rateLimiter.Check(ctx, customerNumber, messageContent)
	if err != nil {
		logger.WithError(err).Warn("[BotEngine] Rate limit check error")
	}

	switch rateResult.Status {
	case StatusMuted, StatusBlocked:
		logger.WithFields(logger.Fields{
			"customer_number": customerNumber,
			"status":          rateResult.Status,
		}).Info("[BotEngine] Number blocked/muted. Skipping response.")
		return nil
	case StatusWarned:
		logger.WithField("customer_number", customerNumber).Warn("[BotEngine] Rate limit warning triggered")
		_ = e.waGateway.SendMessage(sessionID, customerNumber, rateResult.Message)
		botMsg, _ := e.convService.AddMessage(conv.ID, "bot", rateResult.Message, 0, 0)
		if e.sseHub != nil && botMsg != nil {
			e.sseHub.Broadcast("new_message", botMsg)
		}
		return nil
	}

	if !e.guardrail.IsTopicAllowed(messageContent) {
		offTopicReply := e.guardrail.FormatOffTopicResponse()
		_ = e.waGateway.SendMessage(sessionID, customerNumber, offTopicReply)
		botMsg, _ := e.convService.AddMessage(conv.ID, "bot", offTopicReply, 0, 0)
		if e.sseHub != nil && botMsg != nil {
			e.sseHub.Broadcast("new_message", botMsg)
		}
		return nil
	}

	relEntries, _ := e.retriever.Retrieve(messageContent)

	llmCfg, err := e.pgStore.FindActiveLLMConfig()
	if err != nil {
		logger.WithField("conversation_id", conv.ID).Warn("[BotEngine] No active LLM config found! Escalating conversation.")
		_ = e.convService.Escalate(conv.ID)
		fallbackReply := "Maaf, sistem AI kami sedang dalam pemeliharaan. Pesan Anda telah diteruskan ke tim admin."
		_ = e.waGateway.SendMessage(sessionID, customerNumber, fallbackReply)
		botMsg, _ := e.convService.AddMessage(conv.ID, "bot", fallbackReply, 0, 0)
		if e.sseHub != nil && botMsg != nil {
			e.sseHub.Broadcast("new_message", botMsg)
		}
		return nil
	}

	provider, err := llmprovider.NewProvider(llmCfg, e.cfg.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to instantiate LLM provider: %w", err)
	}

	systemPrompt, history, err := e.contextMgr.BuildPromptContext(ctx, customerNumber, messageContent, relEntries)
	if err != nil {
		return fmt.Errorf("failed to build prompt context: %w", err)
	}

	maxTokens := llmCfg.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = e.cfg.LLMMaxOutputTokens
	}

	resp, err := provider.Chat(ctx, systemPrompt, history, maxTokens)
	if err != nil {
		logger.WithError(err).Error("[BotEngine] LLM Call failed, escalating conversation")
		_ = e.convService.Escalate(conv.ID)
		fallbackReply := "Maaf, kami kesulitan memproses pesan Anda. Pesan telah kami teruskan ke admin kami."
		_ = e.waGateway.SendMessage(sessionID, customerNumber, fallbackReply)
		botMsg, _ := e.convService.AddMessage(conv.ID, "bot", fallbackReply, 0, 0)
		if e.sseHub != nil && botMsg != nil {
			e.sseHub.Broadcast("new_message", botMsg)
		}
		return nil
	}

	botReply := e.guardrail.SanitizeResponse(resp.Content)

	if err := e.waGateway.SendMessage(sessionID, customerNumber, botReply); err != nil {
		logger.WithError(err).Error("[BotEngine] Failed to send WA reply")
	}

	botMsg, err := e.convService.AddMessageWithConfig(conv.ID, "bot", botReply, resp.TokenIn, resp.TokenOut, &llmCfg.ID)
	if err == nil && e.sseHub != nil {
		e.sseHub.Broadcast("new_message", botMsg)
	}

	_ = e.contextMgr.SaveMessageToSession(ctx, customerNumber, messageContent, botReply)
	_ = e.contextMgr.SummarizeSessionIfLong(ctx, customerNumber, provider)

	logger.WithFields(logger.Fields{
		"customer_number": customerNumber,
		"tokens_in":       resp.TokenIn,
		"tokens_out":      resp.TokenOut,
	}).Info("[BotEngine] Successfully processed and replied to message")

	return nil
}
