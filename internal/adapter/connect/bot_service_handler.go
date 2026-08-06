package connectadapter

import (
	"net/http"

	"connectrpc.com/connect"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/port"
)

type BotConnectHandler struct {
	pgStore   *postgres.Store
	waGateway port.WhatsAppGateway
}

func NewBotConnectHandler(pgStore *postgres.Store, waGateway port.WhatsAppGateway) *BotConnectHandler {
	return &BotConnectHandler{
		pgStore:   pgStore,
		waGateway: waGateway,
	}
}

func NewBotServiceHandler(pgStore *postgres.Store, waGateway port.WhatsAppGateway) (string, http.Handler) {
	handler := NewBotConnectHandler(pgStore, waGateway)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(connectJSONCodec{})

	serviceName := "polyglot.v1.BotService"
	mux.Handle("/"+serviceName+"/ListSessions", connect.NewUnaryHandler(
		"/"+serviceName+"/ListSessions",
		handler.ListSessions,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/CreateSession", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateSession",
		handler.CreateSession,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetQRCode", connect.NewUnaryHandler(
		"/"+serviceName+"/GetQRCode",
		handler.GetQRCode,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetPairingCode", connect.NewUnaryHandler(
		"/"+serviceName+"/GetPairingCode",
		handler.GetPairingCode,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ToggleBot", connect.NewUnaryHandler(
		"/"+serviceName+"/ToggleBot",
		handler.ToggleBot,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/LogoutSession", connect.NewUnaryHandler(
		"/"+serviceName+"/LogoutSession",
		handler.LogoutSession,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/SendTextMessage", connect.NewUnaryHandler(
		"/"+serviceName+"/SendTextMessage",
		handler.SendTextMessage,
		codecOpt,
	))
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
