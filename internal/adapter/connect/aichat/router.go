package aichat

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
)

// NewServiceHandler mounts AIChatService.
func NewServiceHandler(h *Handler) (string, http.Handler) {
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()
	sn := "polyglot.v1.AIChatService"
	mux.Handle("/"+sn+"/Chat", connect.NewServerStreamHandler("/"+sn+"/Chat", h.Chat, opts...))
	return "/" + sn + "/", mux
}
