package bot

import (
	"net/http"

	"connectrpc.com/connect"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
)

type BotConnectHandler struct {
	convService *convUC.ConversationService
}

func NewBotConnectHandler(convService *convUC.ConversationService) *BotConnectHandler {
	return &BotConnectHandler{convService: convService}
}

func NewBotServiceHandler(convService *convUC.ConversationService) (string, http.Handler) {
	handler := NewBotConnectHandler(convService)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(connectJSONCodec{})

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

	return "/" + serviceName + "/", mux
}
