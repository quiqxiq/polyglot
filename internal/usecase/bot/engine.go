package bot

import (
	"context"
	"fmt"
	"log"

	llmprovider "github.com/quixiq/polyglot/internal/adapter/llm"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/redis"
	"github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/business"
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
	log.Printf("[BotEngine] Processing message from %s (Session %d): %s", customerNumber, sessionID, messageContent)

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
		log.Printf("[BotEngine] Bot disabled for session %d or conversation %d is in escalation mode. Message recorded without auto-reply.", sessionID, conv.ID)
		return nil
	}

	rateResult, err := e.rateLimiter.Check(ctx, customerNumber, messageContent)
	if err != nil {
		log.Printf("[BotEngine] Rate limit check error: %v", err)
	}

	switch rateResult.Status {
	case StatusMuted, StatusBlocked:
		log.Printf("[BotEngine] Number %s is %v. Ignoring message to save tokens.", customerNumber, rateResult.Status)
		return nil
	case StatusWarned:
		log.Printf("[BotEngine] Number %s hit rate limit warning.", customerNumber)
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
		log.Printf("[BotEngine] No active LLM config found! Escalating conversation %d", conv.ID)
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
		log.Printf("[BotEngine] LLM Call failed: %v. Escalating conversation.", err)
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
		log.Printf("[BotEngine] Failed to send WA reply: %v", err)
	}

	botMsg, err := e.convService.AddMessageWithConfig(conv.ID, "bot", botReply, resp.TokenIn, resp.TokenOut, &llmCfg.ID)
	if err == nil && e.sseHub != nil {
		e.sseHub.Broadcast("new_message", botMsg)
	}

	_ = e.contextMgr.SaveMessageToSession(ctx, customerNumber, messageContent, botReply)
	_ = e.contextMgr.SummarizeSessionIfLong(ctx, customerNumber, provider)

	log.Printf("[BotEngine] Successfully replied to %s (Tokens In: %d, Out: %d)", customerNumber, resp.TokenIn, resp.TokenOut)
	return nil
}
