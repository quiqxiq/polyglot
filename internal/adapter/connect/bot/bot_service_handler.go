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
	convService     *convUC.ConversationService
	contextProvider ConversationContextProvider
	skillHandler    *SkillConnectHandler
	llmRepo         port.LLMConfigRepository
	encryptionKey   string
}

func NewBotConnectHandler(
	convService *convUC.ConversationService,
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
	convService *convUC.ConversationService,
	contextProvider ConversationContextProvider,
	skillUC *skillUC.ManageSkillUseCase,
	llmRepo port.LLMConfigRepository,
	encryptionKey string,
) (string, http.Handler) {
	handler := NewBotConnectHandler(convService, contextProvider, skillUC, llmRepo, encryptionKey)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.BotService"
	mux.Handle("/"+serviceName+"/ListConversations", connect.NewUnaryHandler(
		"/"+serviceName+"/ListConversations",
		handler.ListConversations,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetConversation", connect.NewUnaryHandler(
		"/"+serviceName+"/GetConversation",
		handler.GetConversation,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetConversationContext", connect.NewUnaryHandler(
		"/"+serviceName+"/GetConversationContext",
		handler.GetConversationContext,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/TakeOverConversation", connect.NewUnaryHandler(
		"/"+serviceName+"/TakeOverConversation",
		handler.TakeOverConversation,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ResetConversationBot", connect.NewUnaryHandler(
		"/"+serviceName+"/ResetConversationBot",
		handler.ResetConversationBot,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/CloseConversation", connect.NewUnaryHandler(
		"/"+serviceName+"/CloseConversation",
		handler.CloseConversation,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ResetRateLimit", connect.NewUnaryHandler(
		"/"+serviceName+"/ResetRateLimit",
		handler.ResetRateLimit,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetRateLimitStatus", connect.NewUnaryHandler(
		"/"+serviceName+"/GetRateLimitStatus",
		handler.GetRateLimitStatus,
		codecOpt,
	))

	// Skill RPCs (LocalAI Standard)
	if handler.skillHandler != nil {
		mux.Handle("/"+serviceName+"/ListSkills", connect.NewUnaryHandler(
			"/"+serviceName+"/ListSkills",
			handler.skillHandler.ListSkills,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/GetSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/GetSkill",
			handler.skillHandler.GetSkill,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/CreateSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/CreateSkill",
			handler.skillHandler.CreateSkill,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/UpdateSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/UpdateSkill",
			handler.skillHandler.UpdateSkill,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/DeleteSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/DeleteSkill",
			handler.skillHandler.DeleteSkill,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/ExportSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/ExportSkill",
			handler.skillHandler.ExportSkill,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/ImportSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/ImportSkill",
			handler.skillHandler.ImportSkill,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/ListResources", connect.NewUnaryHandler(
			"/"+serviceName+"/ListResources",
			handler.skillHandler.ListResources,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/GetResource", connect.NewUnaryHandler(
			"/"+serviceName+"/GetResource",
			handler.skillHandler.GetResource,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/SaveResource", connect.NewUnaryHandler(
			"/"+serviceName+"/SaveResource",
			handler.skillHandler.SaveResource,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/DeleteResource", connect.NewUnaryHandler(
			"/"+serviceName+"/DeleteResource",
			handler.skillHandler.DeleteResource,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/ListGitRepos", connect.NewUnaryHandler(
			"/"+serviceName+"/ListGitRepos",
			handler.skillHandler.ListGitRepos,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/AddGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/AddGitRepo",
			handler.skillHandler.AddGitRepo,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/UpdateGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/UpdateGitRepo",
			handler.skillHandler.UpdateGitRepo,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/DeleteGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/DeleteGitRepo",
			handler.skillHandler.DeleteGitRepo,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/SyncGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/SyncGitRepo",
			handler.skillHandler.SyncGitRepo,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/ToggleGitRepo", connect.NewUnaryHandler(
			"/"+serviceName+"/ToggleGitRepo",
			handler.skillHandler.ToggleGitRepo,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/ToggleSkill", connect.NewUnaryHandler(
			"/"+serviceName+"/ToggleSkill",
			handler.skillHandler.ToggleSkill,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/GetGlobalPrompt", connect.NewUnaryHandler(
			"/"+serviceName+"/GetGlobalPrompt",
			handler.skillHandler.GetGlobalPrompt,
			codecOpt,
		))
		mux.Handle("/"+serviceName+"/SaveGlobalPrompt", connect.NewUnaryHandler(
			"/"+serviceName+"/SaveGlobalPrompt",
			handler.skillHandler.SaveGlobalPrompt,
			codecOpt,
		))
	}

	// LLM Config RPCs
	mux.Handle("/"+serviceName+"/ListLLMConfigs", connect.NewUnaryHandler(
		"/"+serviceName+"/ListLLMConfigs",
		handler.ListLLMConfigs,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/CreateLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateLLMConfig",
		handler.CreateLLMConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/UpdateLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateLLMConfig",
		handler.UpdateLLMConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ActivateLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/ActivateLLMConfig",
		handler.ActivateLLMConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/TestLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/TestLLMConfig",
		handler.TestLLMConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/DeleteLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteLLMConfig",
		handler.DeleteLLMConfig,
		codecOpt,
	))

	// Technician RPCs
	mux.Handle("/"+serviceName+"/ListTechnicians", connect.NewUnaryHandler(
		"/"+serviceName+"/ListTechnicians",
		handler.ListTechnicians,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/CreateTechnician", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateTechnician",
		handler.CreateTechnician,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/UpdateTechnician", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateTechnician",
		handler.UpdateTechnician,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ToggleTechnicianActive", connect.NewUnaryHandler(
		"/"+serviceName+"/ToggleTechnicianActive",
		handler.ToggleTechnicianActive,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/DeleteTechnician", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteTechnician",
		handler.DeleteTechnician,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
