package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/quixiq/polyglot/internal/port"
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

// Handler handles public RouterOS webhook callbacks. Setiap event wajib
// membawa device id (?device=...) dan token yang cocok dengan token perangkat
// (F5-11). Handler tanpa deviceRepo menolak semua event.
type Handler struct {
	devices port.DeviceRepository
}

// NewHandler constructs a new RouterOS webhook handler.
func NewHandler(devices port.DeviceRepository) *Handler {
	return &Handler{devices: devices}
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

	deviceID := r.URL.Query().Get("device")
	if ev.Token == "" || ev.User == "" || ev.Event == "" || deviceID == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "missing required fields: token, user, event, device")
		return
	}
	if h.devices == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "webhook verification unavailable")
		return
	}

	dev, err := h.devices.FindByID(r.Context(), deviceID)
	if err != nil {
		response.WriteHTTPStatusError(w, http.StatusUnauthorized, "unknown device token")
		return
	}
	expected := ""
	if dev.Extra != nil {
		expected = dev.Extra["webhook_token"]
	}
	if expected == "" {
		// Kompatibilitas skrip lama yang dibuat sebelum token dipersist.
		expected = "rtr_" + dev.ID
	}
	if subtle.ConstantTimeCompare([]byte(expected), []byte(ev.Token)) != 1 {
		response.WriteHTTPStatusError(w, http.StatusUnauthorized, "invalid device token")
		return
	}

	logger.WithComponent("RouterWebhook").WithFields(map[string]any{
		"device_id": dev.ID,
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
