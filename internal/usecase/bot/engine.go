package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	skillDomain "github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/internal/port"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
	"github.com/quixiq/polyglot/pkg/logger"
)

const (
	SenderCustomer = "customer"
	SenderBot      = "bot"

	// historyLimit is the default baseline number of recent messages fed to the LLM.
	historyLimit = 10
)

// ProviderFactory creates an LLM provider from a stored LLM configuration.
type ProviderFactory func(cfg *llm.Config) (port.LLMProvider, error)

// SkillProvider loads available skills for LLM agent runtime.
type SkillProvider interface {
	ListSkills(ctx context.Context) ([]skillDomain.SkillInfo, error)
	GetSkillContent(ctx context.Context, name string) (string, error)
}

// GlobalPromptProvider loads global system prompt from database or disk.
type GlobalPromptProvider interface {
	GetGlobalSystemPrompt(ctx context.Context) (string, error)
}

// Engine orchestrates the WhatsApp customer-service bot.
type Engine struct {
	cache         port.CacheStore
	waGateway     port.WhatsAppGateway
	convService   *convUC.ConversationUseCase
	skillProvider SkillProvider
	globalPromptP GlobalPromptProvider
	llmConfigRepo port.LLMConfigRepository
	chatRepo      port.ChatRepository
	userRepo      port.UserRepository
	settingRepo   port.SettingRepository
	rateLimiter   *RateLimiter
	guardrail     *Guardrail
	contextMgr    *ContextManager
	publisher     port.EventPublisher
	providerF     ProviderFactory
}

// NewEngine creates a new bot engine instance with port dependencies.
func NewEngine(
	cache port.CacheStore,
	waGateway port.WhatsAppGateway,
	convService *convUC.ConversationUseCase,
	skillProvider SkillProvider,
	globalPromptP GlobalPromptProvider,
	llmConfigRepo port.LLMConfigRepository,
	chatRepo port.ChatRepository,
	userRepo port.UserRepository,
	settingRepo port.SettingRepository,
	publisher port.EventPublisher,
	providerFactory ProviderFactory,
) *Engine {
	rl := NewRateLimiter(cache, settingRepo, userRepo)

	return &Engine{
		cache:         cache,
		waGateway:     waGateway,
		convService:   convService,
		skillProvider: skillProvider,
		globalPromptP: globalPromptP,
		llmConfigRepo: llmConfigRepo,
		chatRepo:      chatRepo,
		userRepo:      userRepo,
		settingRepo:   settingRepo,
		rateLimiter:   rl,
		guardrail:     NewGuardrail(),
		contextMgr:    NewContextManager(cache),
		publisher:     publisher,
		providerF:     providerFactory,
	}
}

// HandleIncomingMessage processes an incoming message from a WhatsApp session.
func (e *Engine) HandleIncomingMessage(ctx context.Context, sessionID uint, chatJID string, customerNumber string, messageContent string) error {
	if strings.TrimSpace(messageContent) == "" {
		return nil
	}

	// Filter ketat: Bot HANYA memproses dan membalas chat direct 1:1 dari nomor WhatsApp pribadi.
	// Dilarang keras membalas Group (@g.us), Status/Story (@broadcast), Channel/News (@newsletter), atau Akun Sistem (0@s.whatsapp.net).
	if strings.HasSuffix(chatJID, "@g.us") ||
		strings.HasSuffix(chatJID, "@broadcast") ||
		strings.HasSuffix(chatJID, "@newsletter") ||
		chatJID == "status@broadcast" ||
		chatJID == "0@s.whatsapp.net" {
		logger.WithComponent("BotEngine").Debugf("Ignoring non-direct message from %s (group/channel/broadcast)", chatJID)
		return nil
	}

	if e.waGateway == nil {
		return errors.New("bot engine: whatsapp gateway not initialized")
	}

	if e.convService == nil {
		return errors.New("bot engine: conversation service not initialized")
	}

	sessionTimeout := 30
	if e.settingRepo != nil {
		if botSettings, err := e.settingRepo.GetBotSettings(ctx); err == nil && botSettings != nil && botSettings.SessionTimeoutMinutes > 0 {
			sessionTimeout = botSettings.SessionTimeoutMinutes
		}
	}

	conv, err := e.convService.GetOrCreateConversationWithTimeout(ctx, sessionID, customerNumber, sessionTimeout)
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
	case StatusAllowed:
		// Continue with the normal bot flow.
	case StatusWarned, StatusDailyQuotaExceeded, StatusMuted, StatusBlocked:
		logger.WithComponent("BotEngine").Warnf("Number %s rate limit notification (%v).", customerNumber, rateResult.Status)
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, rateResult.Message, 0, 0, nil)
	}

	limit := historyLimit
	if e.settingRepo != nil {
		if botSettings, err := e.settingRepo.GetBotSettings(ctx); err == nil && botSettings != nil && botSettings.SlidingWindowSize > 0 {
			limit = botSettings.SlidingWindowSize
		}
	}

	history, err := e.convService.GetRecentHistory(ctx, conv.ID, limit)
	if err != nil || len(history) == 0 {
		if err != nil {
			logger.WithComponent("BotEngine").Warnf("Failed to fetch conversation history: %v", err)
		}
		history = []bot.Message{{
			ConversationID: conv.ID,
			SenderType:     bot.SenderCustomer,
			Content:        messageContent,
			CreatedAt:      time.Now(),
		}}
	}

	if e.llmConfigRepo == nil {
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, "Maaf, sistem AI kami saat ini tidak tersedia.", 0, 0, nil)
	}

	llmCfg, err := e.llmConfigRepo.FindActive(ctx)
	if err != nil || llmCfg == nil {
		logger.WithComponent("BotEngine").Warnf("No active LLM config found for conversation %d: %v", conv.ID, err)
		return e.sendBotReply(ctx, conv.ID, sessionID, chatJID, "Maaf, konfigurasi asisten AI belum aktif. Pesan Anda telah kami terima.", 0, 0, nil)
	}

	var compositePrompt strings.Builder
	if e.globalPromptP != nil {
		if gp, err := e.globalPromptP.GetGlobalSystemPrompt(ctx); err == nil && strings.TrimSpace(gp) != "" {
			compositePrompt.WriteString(strings.TrimSpace(gp))
			compositePrompt.WriteString("\n\n")
		}
	}

	var botTools []llm.Tool
	botTools = append(botTools, NewGetCurrentTimeTool(), NewPingHostTool(), NewNotifyTechnicianTool(e.userRepo, e.waGateway, sessionID))

	// Inject skills based on active LLM config (LocalAI standard)
	if llmCfg.EnableSkills && e.skillProvider != nil {
		skillsMode := llmCfg.SkillsMode
		if skillsMode == "" {
			skillsMode = llm.SkillsModePrompt
		}

		allSkills, err := e.skillProvider.ListSkills(ctx)
		if err == nil && len(allSkills) > 0 {
			filtered := skillUC.FilterSkills(allSkills, llmCfg.SelectedSkills)
			if skillsMode == llm.SkillsModePrompt || skillsMode == llm.SkillsModeBoth {
				if skillsText := skillUC.RenderSkillsPrompt(filtered, llmCfg.SkillsPrompt); skillsText != "" {
					compositePrompt.WriteString(skillsText)
					compositePrompt.WriteString("\n\n")
				}
			}
			if skillsMode == llm.SkillsModeTools || skillsMode == llm.SkillsModeBoth {
				compositePrompt.WriteString(skillUC.SkillsToolsHint)
				compositePrompt.WriteString("\n\n")
				botTools = append(botTools, NewRequestSkillTool(e.skillProvider))
			}
		}
	}

	if strings.TrimSpace(llmCfg.SystemPrompt) != "" {
		compositePrompt.WriteString(strings.TrimSpace(llmCfg.SystemPrompt))
	}

	systemPrompt, chatMessages, err := e.contextMgr.BuildPromptContext(ctx, conv.ID, history, compositePrompt.String())
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
	if e.settingRepo != nil {
		if botSettings, err := e.settingRepo.GetBotSettings(ctx); err == nil && botSettings != nil && botSettings.LLMMaxOutputTokens > 0 {
			maxTokens = botSettings.LLMMaxOutputTokens
		}
	}
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	resp, err := provider.ChatWithTools(ctx, systemPrompt, chatMessages, botTools, maxTokens)
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
	info, err := e.rateLimiter.GetRateLimitStatus(ctx, customerNumber)
	if err != nil {
		return nil, err
	}

	// Fallback to conversation DB history ONLY if rate limiter has no active cache backend (e.g. redis nil)
	if !e.rateLimiter.HasCache() && e.convService != nil {
		cleaned := cleanPhoneNumber(customerNumber)
		conv, err := e.convService.GetActiveConversationByCustomer(ctx, 0, cleaned)
		if err == nil && conv != nil {
			msgs, err := e.convService.GetHistory(ctx, conv.ID, 100)
			if err == nil {
				now := time.Now()
				today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				count := 0
				for _, m := range msgs {
					if m.SenderType == bot.SenderCustomer && (m.CreatedAt.Equal(today) || m.CreatedAt.After(today)) {
						count++
					}
				}
				if count > 0 {
					info.DailyChatCount = count
				}
			}
		}
	}

	return info, nil
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
