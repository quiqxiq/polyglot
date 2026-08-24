package notification

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
)

func NewNotificationServiceHandler(repo port.NotificationRepository, sender port.NotificationSender) (string, http.Handler) {
	handler := NewNotificationConnectHandler(repo, sender)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.NotificationService"
	mux.Handle("/"+serviceName+"/ListTemplates", connect.NewUnaryHandler("/"+serviceName+"/ListTemplates", handler.ListTemplates, codecOpt))
	mux.Handle("/"+serviceName+"/GetTemplate", connect.NewUnaryHandler("/"+serviceName+"/GetTemplate", handler.GetTemplate, codecOpt))
	mux.Handle("/"+serviceName+"/SaveTemplate", connect.NewUnaryHandler("/"+serviceName+"/SaveTemplate", handler.SaveTemplate, codecOpt))

	mux.Handle("/"+serviceName+"/ListNotifications", connect.NewUnaryHandler("/"+serviceName+"/ListNotifications", handler.ListNotifications, codecOpt))
	mux.Handle("/"+serviceName+"/PendingCount", connect.NewUnaryHandler("/"+serviceName+"/PendingCount", handler.PendingCount, codecOpt))
	mux.Handle("/"+serviceName+"/MarkNotificationSent", connect.NewUnaryHandler("/"+serviceName+"/MarkNotificationSent", handler.MarkNotificationSent, codecOpt))
	mux.Handle("/"+serviceName+"/MarkNotificationFailed", connect.NewUnaryHandler("/"+serviceName+"/MarkNotificationFailed", handler.MarkNotificationFailed, codecOpt))
	mux.Handle("/"+serviceName+"/TestSend", connect.NewUnaryHandler("/"+serviceName+"/TestSend", handler.TestSend, codecOpt))

	return "/" + serviceName + "/", mux
}
