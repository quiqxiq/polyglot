package webhook

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// RouterEventPayload represents the payload sent from RouterOS on-up/on-down or on-login/on-logout scripts.
type RouterEventPayload struct {
	Token     string `json:"token"`
	Event     string `json:"event"`   // "on-up" | "on-down" | "on-login" | "on-logout"
	Service   string `json:"service"` // "pppoe" | "hotspot"
	User      string `json:"user"`
	IP        string `json:"ip,omitempty"`
	MAC       string `json:"mac,omitempty"`
	Interface string `json:"interface,omitempty"`
	Uptime    string `json:"uptime,omitempty"`
	BytesIn   int64  `json:"bytes_in,omitempty"`
	BytesOut  int64  `json:"bytes_out,omitempty"`
}

// Handler handles public RouterOS webhook callbacks.
type Handler struct{}

// NewHandler constructs a new RouterOS webhook handler.
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterPublic registers the webhook endpoint on the public ServeMux.
func (h *Handler) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/webhooks/mikrotik/events", h.handleRouterEvent)
}

func (h *Handler) handleRouterEvent(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var ev RouterEventPayload
	if err := json.Unmarshal(body, &ev); err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	if ev.Token == "" || ev.User == "" || ev.Event == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "missing required fields: token, user, event")
		return
	}

	logger.WithComponent("RouterWebhook").WithFields(map[string]any{
		"event":     ev.Event,
		"service":   ev.Service,
		"user":      ev.User,
		"ip":        ev.IP,
		"mac":       ev.MAC,
		"interface": ev.Interface,
	}).Info("router lifecycle event received")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
