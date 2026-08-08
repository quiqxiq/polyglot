package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
)

type Engine struct {
	cfg          config.Config
	cache        port.CacheStore
	waGateway    port.WhatsAppGateway
	convService  *convUC.ConversationService
	retriever    port.KnowledgeRetriever
	llmConfigRepo port.LLMConfigRepository
	rateLimiter  *RateLimiter
	guardrail    *Guardrail
	contextMgr   *ContextManager
	publisher    port.EventPublisher
}

func NewEngine(
	cfg config.Config,
	cache port.CacheStore,
	waGateway port.WhatsAppGateway,
	convService *convUC.ConversationService,
	retriever port.KnowledgeRetriever,
	llmConfigRepo port.LLMConfigRepository,
	publisher port.EventPublisher,
) *Engine {
	return &Engine{
		cfg:           cfg,
		cache:         cache,
		waGateway:     waGateway,
		convService:   convService,
		retriever:     retriever,
		llmConfigRepo: llmConfigRepo,
		rateLimiter:   NewRateLimiter(cache, cfg),
		guardrail:     NewGuardrail(cfg),
		contextMgr:    NewContextManager(cache, cfg),
		publisher:     publisher,
	}
}

func (e *Engine) HandleIncomingMessage(ctx context.Context, sessionID uint, customerNumber string, messageContent string) error {
	log.Printf("[BotEngine] Processing message from %s (Session %d): %s", customerNumber, sessionID, messageContent)

	conv, err := e.convService.GetOrCreateConversation(sessionID, customerNumber)
	if err != nil {
		return fmt.Errorf("failed to get/create conversation: %w", err)
	}

	custMsg, err := e.convService.AddMessageWithConfig(conv.ID, "customer", messageContent, 0, 0, nil)
	if err == nil && e.publisher != nil {
		e.publisher.PublishEvent("new_message", custMsg)
	}

	if conv.Status == bot.StatusEscalation {
		log.Printf("[BotEngine] Conversation %d is in escalation mode. Message recorded without auto-reply.", conv.ID)
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
		botMsg, _ := e.convService.AddMessageWithConfig(conv.ID, "bot", rateResult.Message, 0, 0, nil)
		if e.publisher != nil && botMsg != nil {
			e.publisher.PublishEvent("new_message", botMsg)
		}
		return nil
	}

	if !e.guardrail.IsTopicAllowed(messageContent) {
		offTopicReply := e.guardrail.FormatOffTopicResponse()
		_ = e.waGateway.SendMessage(sessionID, customerNumber, offTopicReply)
		botMsg, _ := e.convService.AddMessageWithConfig(conv.ID, "bot", offTopicReply, 0, 0, nil)
		if e.publisher != nil && botMsg != nil {
			e.publisher.PublishEvent("new_message", botMsg)
		}
		return nil
	}

	relEntries, _ := e.retriever.Retrieve(ctx, messageContent)

	if e.llmConfigRepo == nil {
		_ = e.convService.Escalate(conv.ID)
		fallbackReply := "Maaf, kami kesulitan memproses pesan Anda. Pesan telah kami teruskan ke admin kami."
		_ = e.waGateway.SendMessage(sessionID, customerNumber, fallbackReply)
		botMsg, _ := e.convService.AddMessageWithConfig(conv.ID, "bot", fallbackReply, 0, 0, nil)
		if e.publisher != nil && botMsg != nil {
			e.publisher.PublishEvent("new_message", botMsg)
		}
		return nil
	}

	llmCfg, err := e.llmConfigRepo.FindActive()
	if err != nil {
		log.Printf("[BotEngine] No active LLM config found! Escalating conversation %d", conv.ID)
		_ = e.convService.Escalate(conv.ID)
		fallbackReply := "Maaf, sistem AI kami sedang dalam pemeliharaan. Pesan Anda telah diteruskan ke tim admin."
		_ = e.waGateway.SendMessage(sessionID, customerNumber, fallbackReply)
		botMsg, _ := e.convService.AddMessageWithConfig(conv.ID, "bot", fallbackReply, 0, 0, nil)
		if e.publisher != nil && botMsg != nil {
			e.publisher.PublishEvent("new_message", botMsg)
		}
		return nil
	}

	systemPrompt, history, err := e.contextMgr.BuildPromptContext(ctx, customerNumber, messageContent, relEntries)
	if err != nil {
		return fmt.Errorf("failed to build prompt context: %w", err)
	}
	_ = systemPrompt
	_ = history
	_ = llmCfg

	return nil
}
