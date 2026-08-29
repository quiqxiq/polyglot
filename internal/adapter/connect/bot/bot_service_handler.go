package bot

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
	botUC "github.com/quixiq/polyglot/internal/usecase/bot"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
)

// ConversationContextProvider abstracts the engine's ability to aggregate the
// LLM-facing state of a conversation (history + summary + token usage) and manage rate limits.
type ConversationContextProvider interface {
	GetConversationContext(ctx context.Context, convID uint) (*bot.ConversationContext, error)
	ResetRateLimit(ctx context.Context, customerNumber string) error
	GetRateLimitStatus(ctx context.Context, customerNumber string) (*botUC.RateLimitStatusInfo, error)
}

type BotConnectHandler struct {
	convService     *convUC.ConversationUseCase
	contextProvider ConversationContextProvider
	skillHandler    *SkillConnectHandler
	llmRepo         port.LLMConfigRepository
	encryptionKey   string
}

func NewBotConnectHandler(
	convService *convUC.ConversationUseCase,
	contextProvider ConversationContextProvider,
	skillUC *skillUC.ManageSkillUseCase,
	llmRepo port.LLMConfigRepository,
	encryptionKey string,
) *BotConnectHandler {
	return &BotConnectHandler{
		convService:     convService,
		contextProvider: contextProvider,
		skillHandler:    NewSkillConnectHandler(skillUC),
		llmRepo:         llmRepo,
		encryptionKey:   encryptionKey,
	}
}

func NewBotServiceHandler(
	convService *convUC.ConversationUseCase,
	contextProvider ConversationContextProvider,
	skillUC *skillUC.ManageSkillUseCase,
	llmRepo port.LLMConfigRepository,
	encryptionKey string,
) (string, http.Handler) {
	handler := NewBotConnectHandler(convService, contextProvider, skillUC, llmRepo, encryptionKey)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.BotService"
	mux.Handle("/"+serviceName+"/ListConversations", connect.NewUnaryHandler(
		"/"+serviceName+"/ListConversations",
		handler.ListConversations,
		opts...,
	))
	mux.Handle("/"+serviceName+"/GetConversation", connect.NewUnaryHandler(
		"/"+serviceName+"/GetConversation",
		handler.GetConversation,
		opts...,
	))
	mux.Handle("/"+serviceName+"/GetConversationContext", connect.NewUnaryHandler(
		"/"+serviceName+"/GetConversationContext",
		handler.GetConversationContext,
		opts...,
	))
	mux.Handle("/"+serviceName+"/TakeOverConversation", connect.NewUnaryHandler(
		"/"+serviceName+"/TakeOverConversation",
		handler.TakeOverConversation,
		opts...,
	))
	mux.Handle("/"+serviceName+"/ResetConversationBot", connect.NewUnaryHandler(
		"/"+serviceName+"/ResetConversationBot",
		handler.ResetConversationBot,
		opts...,
	))
	mux.Handle("/"+serviceName+"/CloseConversation", connect.NewUnaryHandler(
		"/"+serviceName+"/CloseConversation",
		handler.CloseConversation,
		opts...,
	))
	mux.Handle("/"+serviceName+"/ResetRateLimit", connect.NewUnaryHandler(
		"/"+serviceName+"/ResetRateLimit",
		handler.ResetRateLimit,
		opts...,
	))
	mux.Handle("/"+serviceName+"/GetRateLimitStatus", connect.NewUnaryHandler(
		"/"+serviceName+"/GetRateLimitStatus",
		handler.GetRateLimitStatus,
		opts...,
	))

	// Skill RPCs (LocalAI Standard)
	if handler.skillHandler != nil {
		mux.Handle("/"+serviceName+"/ListSkills", connect.NewUnaryHandler(
			"/"+serviceName+"/ListSkills",
			handler.skillHandler.ListSkills,
			opts...,
		))
		mux.Handle("/"+serviceName+"/GetSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/GetSkill",
			handler.skillHandler.GetSkill,
			opts...,
		))
		mux.Handle("/"+serviceName+"/CreateSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/CreateSkill",
			handler.skillHandler.CreateSkill,
			opts...,
		))
		mux.Handle("/"+serviceName+"/UpdateSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/UpdateSkill",
			handler.skillHandler.UpdateSkill,
			opts...,
		))
		mux.Handle("/"+serviceName+"/DeleteSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/DeleteSkill",
			handler.skillHandler.DeleteSkill,
			opts...,
		))
		mux.Handle("/"+serviceName+"/ExportSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/ExportSkill",
			handler.skillHandler.ExportSkill,
			opts...,
		))
		mux.Handle("/"+serviceName+"/ImportSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/ImportSkill",
			handler.skillHandler.ImportSkill,
			opts...,
		))
		mux.Handle("/"+serviceName+"/ListResources", connect.NewUnaryHandler(
			"/"+serviceName+"/ListResources",
			handler.skillHandler.ListResources,
			opts...,
		))
		mux.Handle("/"+serviceName+"/GetResource", connect.NewUnaryHandler(
			"/"+serviceName+"/GetResource",
			handler.skillHandler.GetResource,
			opts...,
		))
		mux.Handle("/"+serviceName+"/SaveResource", connect.NewUnaryHandler(
			"/"+serviceName+"/SaveResource",
			handler.skillHandler.SaveResource,
			opts...,
		))
		mux.Handle("/"+serviceName+"/DeleteResource", connect.NewUnaryHandler(
			"/"+serviceName+"/DeleteResource",
			handler.skillHandler.DeleteResource,
			opts...,
		))
		mux.Handle("/"+serviceName+"/ListGitRepos", connect.NewUnaryHandler(
			"/"+serviceName+"/ListGitRepos",
			handler.skillHandler.ListGitRepos,
			opts...,
		))
		mux.Handle("/"+serviceName+"/AddGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/AddGitRepo",
			handler.skillHandler.AddGitRepo,
			opts...,
		))
		mux.Handle("/"+serviceName+"/UpdateGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/UpdateGitRepo",
			handler.skillHandler.UpdateGitRepo,
			opts...,
		))
		mux.Handle("/"+serviceName+"/DeleteGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/DeleteGitRepo",
			handler.skillHandler.DeleteGitRepo,
			opts...,
		))
		mux.Handle("/"+serviceName+"/SyncGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/SyncGitRepo",
			handler.skillHandler.SyncGitRepo,
			opts...,
		))
		mux.Handle("/"+serviceName+"/ToggleGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/ToggleGitRepo",
			handler.skillHandler.ToggleGitRepo,
			opts...,
		))
		mux.Handle("/"+serviceName+"/ToggleSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/ToggleSkill",
			handler.skillHandler.ToggleSkill,
			opts...,
		))
		mux.Handle("/"+serviceName+"/GetGlobalPrompt", connect.NewUnaryHandler(
			"/"+serviceName+"/GetGlobalPrompt",
			handler.skillHandler.GetGlobalPrompt,
			opts...,
		))
		mux.Handle("/"+serviceName+"/SaveGlobalPrompt", connect.NewUnaryHandler(
			"/"+serviceName+"/SaveGlobalPrompt",
			handler.skillHandler.SaveGlobalPrompt,
			opts...,
		))
	}

	// LLM Config RPCs
	mux.Handle("/"+serviceName+"/ListLLMConfigs", connect.NewUnaryHandler(
		"/"+serviceName+"/ListLLMConfigs",
		handler.ListLLMConfigs,
		opts...,
	))
	mux.Handle("/"+serviceName+"/CreateLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateLLMConfig",
		handler.CreateLLMConfig,
		opts...,
	))
	mux.Handle("/"+serviceName+"/UpdateLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateLLMConfig",
		handler.UpdateLLMConfig,
		opts...,
	))
	mux.Handle("/"+serviceName+"/ActivateLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/ActivateLLMConfig",
		handler.ActivateLLMConfig,
		opts...,
	))
	mux.Handle("/"+serviceName+"/TestLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/TestLLMConfig",
		handler.TestLLMConfig,
		opts...,
	))
	mux.Handle("/"+serviceName+"/DeleteLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteLLMConfig",
		handler.DeleteLLMConfig,
		opts...,
	))

	// Technician RPCs
	mux.Handle("/"+serviceName+"/ListTechnicians", connect.NewUnaryHandler(
		"/"+serviceName+"/ListTechnicians",
		handler.ListTechnicians,
		opts...,
	))
	mux.Handle("/"+serviceName+"/CreateTechnician", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateTechnician",
		handler.CreateTechnician,
		opts...,
	))
	mux.Handle("/"+serviceName+"/UpdateTechnician", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateTechnician",
		handler.UpdateTechnician,
		opts...,
	))
	mux.Handle("/"+serviceName+"/ToggleTechnicianActive", connect.NewUnaryHandler(
		"/"+serviceName+"/ToggleTechnicianActive",
		handler.ToggleTechnicianActive,
		opts...,
	))
	mux.Handle("/"+serviceName+"/DeleteTechnician", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteTechnician",
		handler.DeleteTechnician,
		opts...,
	))

	return "/" + serviceName + "/", mux
}
