package bot

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
)

const (
	SenderCustomer = "customer"
	SenderBot      = "bot"

	// historyLimit is the number of recent messages fed to the LLM as context.
	historyLimit = 20
)

// ProviderFactory creates an LLM provider from a stored LLM configuration.
// Di-inject dari adapter layer (internal/adapter/llm) lewat app wiring agar
// usecase tetap hanya bergantung ke domain dan port (clean architecture).
type ProviderFactory func(cfg *llm.LLMConfig) (port.LLMProvider, error)

// Engine orchestrates the WhatsApp customer-service bot: it classifies
// incoming messages, applies rate limits and guardrails, retrieves knowledge,
// calls the active LLM provider, sends the reply via the WhatsApp gateway and
// persists both sides of the conversation.
type Engine struct {
	cfg           config.Config
	cache         port.CacheStore
	waGateway     port.WhatsAppGateway
	convService   *convUC.ConversationService
	retriever     port.KnowledgeRetriever
	llmConfigRepo port.LLMConfigRepository
	chatRepo      port.ChatRepository
	rateLimiter   *RateLimiter
	guardrail     *Guardrail
	contextMgr    *ContextManager
	publisher     port.EventPublisher
	providerF     ProviderFactory
}

func NewEngine(
	cfg config.Config,
	cache port.CacheStore,
	waGateway port.WhatsAppGateway,
	convService *convUC.ConversationService,
	retriever port.KnowledgeRetriever,
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
		llmConfigRepo: llmConfigRepo,
		chatRepo:      chatRepo,
		rateLimiter:   NewRateLimiter(cache, cfg),
		guardrail:     NewGuardrail(cfg),
		contextMgr:    NewContextManager(cache, cfg),
		publisher:     publisher,
		providerF:     providerFactory,
	}
}

// HandleIncomingMessage processes a single incoming WhatsApp text message:
// persist -> escalation check -> per-chat bot toggle -> rate limit ->
// guardrail -> LLM reply.
func (e *Engine) HandleIncomingMessage(ctx context.Context, sessionID uint, chatJID string, customerNumber string, messageContent string) error {
	if e.waGateway == nil {
		return errors.New("bot engine: whatsapp gateway not initialized")
	}

	log.Printf("[BotEngine] Processing message from %s (Session %d): %s", customerNumber, sessionID, messageContent)

	conv, err := e.convService.GetOrCreateConversation(sessionID, customerNumber)
	if err != nil {
		return fmt.Errorf("failed to get/create conversation: %w", err)
	}

	custMsg, err := e.convService.AddMessageWithConfig(conv.ID, SenderCustomer, messageContent, 0, 0, nil)
	if err != nil {
		log.Printf("[BotEngine] Failed to persist customer message: %v", err)
	}
	if err == nil && e.publisher != nil {
		e.publisher.PublishEvent("new_message", custMsg)
	}

	if conv.Status == bot.StatusEscalation {
		log.Printf("[BotEngine] Conversation %d is in escalation mode. Message recorded without auto-reply.", conv.ID)
		return nil
	}

	// Kontrol bot per-chat: pesan tetap dicatat, tapi auto-reply dilewati
	// bila agen mematikan bot untuk chat ini (mis. setelah ambil alih).
	// Fail-open saat pengecekan error (bot tetap balas) — dipilih daripada
	// fail-closed agar gangguan DB sesaat tidak menghentikan semua balasan bot.
	if e.chatRepo != nil {
		enabled, err := e.chatRepo.IsChatBotEnabled(sessionID, chatJID)
		if err != nil {
			log.Printf("[BotEngine] Failed to check bot_enabled for %s: %v", chatJID, err)
		}
		if !enabled {
			log.Printf("[BotEngine] Bot nonaktif untuk chat %s. Pesan dicatat tanpa auto-reply.", chatJID)
			return nil
		}
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
		return e.sendBotReply(conv.ID, sessionID, customerNumber, rateResult.Message, 0, 0, nil)
	}

	if !e.guardrail.IsTopicAllowed(messageContent) {
		offTopicReply := e.guardrail.FormatOffTopicResponse()
		return e.sendBotReply(conv.ID, sessionID, customerNumber, offTopicReply, 0, 0, nil)
	}

	relEntries, _ := e.retriever.Retrieve(ctx, messageContent)

	if e.llmConfigRepo == nil {
		fallbackReply := "Maaf, kami kesulitan memproses pesan Anda. Pesan telah kami teruskan ke admin kami."
		return e.escalateWithFallback(conv.ID, sessionID, customerNumber, fallbackReply)
	}

	llmCfg, err := e.llmConfigRepo.FindActive()
	if err != nil {
		log.Printf("[BotEngine] No active LLM config found! Escalating conversation %d", conv.ID)
		fallbackReply := "Maaf, sistem AI kami sedang dalam pemeliharaan. Pesan Anda telah diteruskan ke tim admin."
		return e.escalateWithFallback(conv.ID, sessionID, customerNumber, fallbackReply)
	}

	history, err := e.convService.GetRecentHistory(conv.ID, historyLimit)
	if err != nil || len(history) == 0 {
		if err != nil {
			log.Printf("[BotEngine] Failed to load recent history for conversation %d: %v", conv.ID, err)
		}
		// Jaminan: pertanyaan user SAAT INI selalu masuk ke prompt LLM, walau
		// history gagal dimuat dari DB (tidak boleh bot membalas tanpa konteks).
		history = []bot.Message{
			{ConversationID: conv.ID, SenderType: bot.SenderCustomer, Content: messageContent},
		}
	}

	systemPrompt, chatMessages, err := e.contextMgr.BuildPromptContext(ctx, conv.ID, history, relEntries)
	if err != nil {
		return fmt.Errorf("failed to build prompt context: %w", err)
	}

	if e.providerF == nil {
		return e.escalateWithFallback(conv.ID, sessionID, customerNumber, "Maaf, sistem AI kami sedang tidak tersedia. Pesan Anda telah diteruskan ke tim admin.")
	}

	provider, err := e.providerF(llmCfg)
	if err != nil {
		log.Printf("[BotEngine] Failed to build LLM provider for conversation %d: %v", conv.ID, err)
		return e.escalateWithFallback(conv.ID, sessionID, customerNumber, "Maaf, sistem AI kami sedang tidak tersedia. Pesan Anda telah diteruskan ke tim admin.")
	}

	maxTokens := llmCfg.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = e.cfg.LLMMaxOutputTokens
	}

	resp, err := provider.Chat(ctx, systemPrompt, chatMessages, maxTokens)
	if err != nil {
		log.Printf("[BotEngine] LLM call failed for conversation %d: %v", conv.ID, err)
		return e.escalateWithFallback(conv.ID, sessionID, customerNumber, "Maaf, sistem AI kami sedang bermasalah. Pesan Anda telah diteruskan ke tim admin kami.")
	}

	reply := e.guardrail.SanitizeResponse(resp.Content)
	if reply == "" {
		log.Printf("[BotEngine] Empty LLM reply for conversation %d", conv.ID)
		reply = "Maaf, saya tidak menemukan jawaban yang tepat. Tim kami akan segera menghubungi Anda."
	}

	if err := e.sendBotReply(conv.ID, sessionID, customerNumber, reply, resp.TokenIn, resp.TokenOut, &llmCfg.ID); err != nil {
		return err
	}

	_ = e.contextMgr.SaveMessageToSession(ctx, conv.ID, messageContent, reply)
	_ = e.contextMgr.SummarizeSessionIfLong(ctx, conv.ID, provider)

	return nil
}

// GetConversationContext aggregates the LLM-facing state of a conversation
// (status, running summary, recent history, token usage) untuk dashboard agen.
// Sumber data: DB (conversation + messages) dan cache (summary per-conv).
func (e *Engine) GetConversationContext(ctx context.Context, convID uint) (*bot.ConversationContext, error) {
	conv, err := e.convService.GetConversationWithMessages(convID)
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

// sendBotReply sends a bot message via the WhatsApp gateway and persists it
// with token usage and the LLM config that produced it.
func (e *Engine) sendBotReply(convID uint, sessionID uint, customerNumber string, content string, tokenIn, tokenOut int, llmConfigID *uint) error {
	if err := e.waGateway.SendMessage(sessionID, customerNumber, content); err != nil {
		log.Printf("[BotEngine] Failed to send reply to %s: %v", customerNumber, err)
		return fmt.Errorf("failed to send bot reply: %w", err)
	}

	botMsg, err := e.convService.AddMessageWithConfig(convID, SenderBot, content, tokenIn, tokenOut, llmConfigID)
	if err != nil {
		log.Printf("[BotEngine] Failed to persist bot message: %v", err)
	}
	if e.publisher != nil && botMsg != nil {
		e.publisher.PublishEvent("new_message", botMsg)
	}
	return nil
}

// escalateWithFallback marks the conversation as needing a human agent, sends
// a fallback reply and persists it.
func (e *Engine) escalateWithFallback(convID uint, sessionID uint, customerNumber string, reply string) error {
	if err := e.convService.Escalate(convID); err != nil {
		log.Printf("[BotEngine] Failed to escalate conversation %d: %v", convID, err)
	}
	return e.sendBotReply(convID, sessionID, customerNumber, reply, 0, 0, nil)
}
