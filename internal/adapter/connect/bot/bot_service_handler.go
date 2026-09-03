package bot

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/domain/bot"
	botUC "github.com/quixiq/polyglot/internal/usecase/bot"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
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
}

func NewBotConnectHandler(
	convService *convUC.ConversationUseCase,
	contextProvider ConversationContextProvider,
) *BotConnectHandler {
	return &BotConnectHandler{
		convService:     convService,
		contextProvider: contextProvider,
	}
}

func NewBotServiceHandler(
	convService *convUC.ConversationUseCase,
	contextProvider ConversationContextProvider,
) (string, http.Handler) {
	handler := NewBotConnectHandler(convService, contextProvider)
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
