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
	convService     *convUC.ConversationService
	contextProvider ConversationContextProvider
}

func NewBotConnectHandler(convService *convUC.ConversationService, contextProvider ConversationContextProvider) *BotConnectHandler {
	return &BotConnectHandler{
		convService:     convService,
		contextProvider: contextProvider,
	}
}

func NewBotServiceHandler(convService *convUC.ConversationService, contextProvider ConversationContextProvider) (string, http.Handler) {
	handler := NewBotConnectHandler(convService, contextProvider)
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

	return "/" + serviceName + "/", mux
}
