package bot

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/port"
	chatUC "github.com/quixiq/polyglot/internal/usecase/chat"
)

type WhatsAppConnectHandler struct {
	pgStore     *postgres.Store
	waGateway   port.WhatsAppGateway
	chatService *chatUC.ChatService
}

func NewWhatsAppConnectHandler(pgStore *postgres.Store, waGateway port.WhatsAppGateway, chatService *chatUC.ChatService) *WhatsAppConnectHandler {
	return &WhatsAppConnectHandler{
		pgStore:     pgStore,
		waGateway:   waGateway,
		chatService: chatService,
	}
}

func NewWhatsAppServiceHandler(pgStore *postgres.Store, waGateway port.WhatsAppGateway, chatService *chatUC.ChatService) (string, http.Handler) {
	handler := NewWhatsAppConnectHandler(pgStore, waGateway, chatService)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.WhatsAppService"
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
	mux.Handle("/"+serviceName+"/ReconnectSession", connect.NewUnaryHandler(
		"/"+serviceName+"/ReconnectSession",
		handler.ReconnectSession,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/LogoutSession", connect.NewUnaryHandler(
		"/"+serviceName+"/LogoutSession",
		handler.LogoutSession,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/PurgeSession", connect.NewUnaryHandler(
		"/"+serviceName+"/PurgeSession",
		handler.PurgeSession,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/SendTextMessage", connect.NewUnaryHandler(
		"/"+serviceName+"/SendTextMessage",
		handler.SendTextMessage,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListChats", connect.NewUnaryHandler(
		"/"+serviceName+"/ListChats",
		handler.ListChats,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetChatMessages", connect.NewUnaryHandler(
		"/"+serviceName+"/GetChatMessages",
		handler.GetChatMessages,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/MarkChatRead", connect.NewUnaryHandler(
		"/"+serviceName+"/MarkChatRead",
		handler.MarkChatRead,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ToggleChatBot", connect.NewUnaryHandler(
		"/"+serviceName+"/ToggleChatBot",
		handler.ToggleChatBot,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
