package notification

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	notificationUC "github.com/quixiq/polyglot/internal/usecase/notification"
)

// NewNotificationServiceHandler creates the notification ConnectRPC service handler.
func NewNotificationServiceHandler(uc *notificationUC.ManageNotificationUseCase) (string, http.Handler) {
	handler := NewNotificationConnectHandler(uc)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.NotificationService"
	mux.Handle("/"+serviceName+"/ListTemplates", connect.NewUnaryHandler("/"+serviceName+"/ListTemplates", handler.ListTemplates, opts...))
	mux.Handle("/"+serviceName+"/GetTemplate", connect.NewUnaryHandler("/"+serviceName+"/GetTemplate", handler.GetTemplate, opts...))
	mux.Handle("/"+serviceName+"/SaveTemplate", connect.NewUnaryHandler("/"+serviceName+"/SaveTemplate", handler.SaveTemplate, opts...))

	mux.Handle("/"+serviceName+"/ListNotifications", connect.NewUnaryHandler("/"+serviceName+"/ListNotifications", handler.ListNotifications, opts...))
	mux.Handle("/"+serviceName+"/PendingCount", connect.NewUnaryHandler("/"+serviceName+"/PendingCount", handler.PendingCount, opts...))
	mux.Handle("/"+serviceName+"/MarkNotificationSent", connect.NewUnaryHandler("/"+serviceName+"/MarkNotificationSent", handler.MarkNotificationSent, opts...))
	mux.Handle("/"+serviceName+"/MarkNotificationFailed", connect.NewUnaryHandler("/"+serviceName+"/MarkNotificationFailed", handler.MarkNotificationFailed, opts...))
	mux.Handle("/"+serviceName+"/TestSend", connect.NewUnaryHandler("/"+serviceName+"/TestSend", handler.TestSend, opts...))

	return "/" + serviceName + "/", mux
}
