package bot

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	chatUC "github.com/quixiq/polyglot/internal/usecase/chat"
)

type WhatsAppConnectHandler struct {
	sessionRepo port.WASessionRepository
	waGateway   port.WhatsAppGateway
	chatService *chatUC.ChatUseCase
}

func NewWhatsAppConnectHandler(sessionRepo port.WASessionRepository, waGateway port.WhatsAppGateway, chatService *chatUC.ChatUseCase) *WhatsAppConnectHandler {
	return &WhatsAppConnectHandler{
		sessionRepo: sessionRepo,
		waGateway:   waGateway,
		chatService: chatService,
	}
}

func NewWhatsAppServiceHandler(sessionRepo port.WASessionRepository, waGateway port.WhatsAppGateway, chatService *chatUC.ChatUseCase) (string, http.Handler) {
	handler := NewWhatsAppConnectHandler(sessionRepo, waGateway, chatService)

	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.WhatsAppService"
	mux.Handle("/"+serviceName+"/ListSessions", connect.NewUnaryHandler(
		"/"+serviceName+"/ListSessions",
		handler.ListSessions,
		opts...,
	))
	mux.Handle("/"+serviceName+"/CreateSession", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateSession",
		handler.CreateSession,
		opts...,
	))
	mux.Handle("/"+serviceName+"/GetQRCode", connect.NewUnaryHandler(
		"/"+serviceName+"/GetQRCode",
		handler.GetQRCode,
		opts...,
	))
	mux.Handle("/"+serviceName+"/GetPairingCode", connect.NewUnaryHandler(
		"/"+serviceName+"/GetPairingCode",
		handler.GetPairingCode,
		opts...,
	))
	mux.Handle("/"+serviceName+"/ToggleBot", connect.NewUnaryHandler(
		"/"+serviceName+"/ToggleBot",
		handler.ToggleBot,
		opts...,
	))
	mux.Handle("/"+serviceName+"/ReconnectSession", connect.NewUnaryHandler(
		"/"+serviceName+"/ReconnectSession",
		handler.ReconnectSession,
		opts...,
	))
	mux.Handle("/"+serviceName+"/LogoutSession", connect.NewUnaryHandler(
		"/"+serviceName+"/LogoutSession",
		handler.LogoutSession,
		opts...,
	))
	mux.Handle("/"+serviceName+"/PurgeSession", connect.NewUnaryHandler(
		"/"+serviceName+"/PurgeSession",
		handler.PurgeSession,
		opts...,
	))
	mux.Handle("/"+serviceName+"/SendTextMessage", connect.NewUnaryHandler(
		"/"+serviceName+"/SendTextMessage",
		handler.SendTextMessage,
		opts...,
	))
	mux.Handle("/"+serviceName+"/ListChats", connect.NewUnaryHandler(
		"/"+serviceName+"/ListChats",
		handler.ListChats,
		opts...,
	))
	mux.Handle("/"+serviceName+"/GetChatMessages", connect.NewUnaryHandler(
		"/"+serviceName+"/GetChatMessages",
		handler.GetChatMessages,
		opts...,
	))
	mux.Handle("/"+serviceName+"/MarkChatRead", connect.NewUnaryHandler(
		"/"+serviceName+"/MarkChatRead",
		handler.MarkChatRead,
		opts...,
	))
	mux.Handle("/"+serviceName+"/ToggleChatBot", connect.NewUnaryHandler(
		"/"+serviceName+"/ToggleChatBot",
		handler.ToggleChatBot,
		opts...,
	))

	return "/" + serviceName + "/", mux
}
