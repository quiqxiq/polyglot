package connectadapter

import (
	"net/http"

	"connectrpc.com/connect"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/port"
)

type WhatsAppConnectHandler struct {
	pgStore   *postgres.Store
	waGateway port.WhatsAppGateway
}

func NewWhatsAppConnectHandler(pgStore *postgres.Store, waGateway port.WhatsAppGateway) *WhatsAppConnectHandler {
	return &WhatsAppConnectHandler{
		pgStore:   pgStore,
		waGateway: waGateway,
	}
}

func NewWhatsAppServiceHandler(pgStore *postgres.Store, waGateway port.WhatsAppGateway) (string, http.Handler) {
	handler := NewWhatsAppConnectHandler(pgStore, waGateway)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(connectJSONCodec{})

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

	return "/" + serviceName + "/", mux
}
