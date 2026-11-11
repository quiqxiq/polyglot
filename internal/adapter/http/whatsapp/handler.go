package whatsapp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	domainWA "github.com/quixiq/polyglot/internal/domain/whatsapp"
	"github.com/quixiq/polyglot/internal/port"
	ucWA "github.com/quixiq/polyglot/internal/usecase/whatsapp"
	"github.com/quixiq/polyglot/pkg/response"
)

// Handler serves plain-HTTP REST endpoints for WhatsApp device lifecycle,
// messaging, and sender identity resolution.
type Handler struct {
	sessionRepo port.WASessionRepository
	waGateway   port.WhatsAppGateway
	resolver    *ucWA.IdentityResolver
}

// NewHandler constructs a new WhatsApp REST API handler.
func NewHandler(
	sessionRepo port.WASessionRepository,
	waGateway port.WhatsAppGateway,
	resolver *ucWA.IdentityResolver,
) *Handler {
	return &Handler{
		sessionRepo: sessionRepo,
		waGateway:   waGateway,
		resolver:    resolver,
	}
}

// RegisterProtected mounts all WhatsApp REST endpoints on the protected router.
func (h *Handler) RegisterProtected(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/whatsapp/sessions", h.handleListSessions)
	mux.HandleFunc("POST /api/whatsapp/sessions", h.handleCreateSession)
	mux.HandleFunc("GET /api/whatsapp/session/status", h.handleGetSessionStatus)
	mux.HandleFunc("GET /api/whatsapp/session/qr", h.handleGetSessionQR)
	mux.HandleFunc("POST /api/whatsapp/session/pair-code", h.handleGetPairingCode)
	mux.HandleFunc("POST /api/whatsapp/session/reconnect", h.handleReconnectSession)
	mux.HandleFunc("POST /api/whatsapp/session/logout", h.handleLogoutSession)
	mux.HandleFunc("POST /api/whatsapp/session/purge", h.handlePurgeSession)

	mux.HandleFunc("POST /api/whatsapp/send/message", h.handleSendMessage)
	mux.HandleFunc("POST /api/whatsapp/send/image", h.handleSendImage)
	mux.HandleFunc("POST /api/whatsapp/send/document", h.handleSendDocument)
	mux.HandleFunc("POST /api/whatsapp/presence", h.handleChatPresence)
	mux.HandleFunc("POST /api/whatsapp/message/read", h.handleMarkRead)

	mux.HandleFunc("GET /api/whatsapp/identity", h.handleResolveIdentity)
}

func (h *Handler) handleListSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.sessionRepo == nil {
		_ = json.NewEncoder(w).Encode(map[string]any{"sessions": []SessionDTO{}})
		return
	}

	sessions, err := h.sessionRepo.FindAllSessions(r.Context())
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	dtos := make([]SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		status := "offline"
		if h.waGateway != nil {
			if st, err := h.waGateway.GetStatus(s.ID); err == nil {
				status = st
			}
		}
		dtos = append(dtos, toSessionDTO(s, status))
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"sessions": dtos})
}

func (h *Handler) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.sessionRepo == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "Session repository unavailable")
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	name := strings.TrimSpace(req.DeviceName)
	if name == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "device_name is required")
		return
	}

	session := &domainWA.WASession{
		DeviceName:  name,
		PhoneNumber: strings.TrimSpace(req.PhoneNumber),
	}

	if err := h.sessionRepo.CreateSession(r.Context(), session); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	if h.waGateway != nil {
		_ = h.waGateway.Connect(session)
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"session": toSessionDTO(*session, "connecting"),
	})
}

func (h *Handler) handleGetSessionStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID, err := parseSessionIDFromQuery(r)
	if err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Valid session_id query parameter is required")
		return
	}

	status := "offline"
	if h.waGateway != nil {
		if st, err := h.waGateway.GetStatus(sessionID); err == nil {
			status = st
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"session_id": sessionID,
		"status":     status,
	})
}

func (h *Handler) handleGetSessionQR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID, err := parseSessionIDFromQuery(r)
	if err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Valid session_id query parameter is required")
		return
	}

	if h.waGateway == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "WhatsApp gateway unavailable")
		return
	}

	qr, err := h.waGateway.GetQRCode(sessionID)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"session_id": sessionID,
		"qr_code":    qr,
	})
}

func (h *Handler) handleGetPairingCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req PairingCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.SessionID == 0 {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	phoneNum := strings.TrimSpace(req.PhoneNumber)
	if phoneNum == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "phone_number is required")
		return
	}

	if h.waGateway == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "WhatsApp gateway unavailable")
		return
	}

	code, err := h.waGateway.GetPairingCode(req.SessionID, phoneNum)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(PairingCodeResponse{
		SessionID:   req.SessionID,
		PairingCode: code,
	})
}

func (h *Handler) handleReconnectSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req SessionActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.SessionID == 0 {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	if h.sessionRepo == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "Session repository unavailable")
		return
	}
	session, err := h.sessionRepo.FindSessionByID(r.Context(), req.SessionID)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	if h.waGateway != nil {
		if err := h.waGateway.Reconnect(session); err != nil {
			response.WriteHTTPError(w, err)
			return
		}
	}

	_ = json.NewEncoder(w).Encode(GenericSuccessResponse{
		Success: true,
		Message: "reconnect triggered",
	})
}

func (h *Handler) handleLogoutSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req SessionActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.SessionID == 0 {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	if h.waGateway != nil {
		if err := h.waGateway.Logout(req.SessionID); err != nil {
			response.WriteHTTPError(w, err)
			return
		}
	}

	_ = json.NewEncoder(w).Encode(GenericSuccessResponse{
		Success: true,
		Message: "session logged out",
	})
}

func (h *Handler) handlePurgeSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req SessionActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.SessionID == 0 {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	if h.waGateway != nil {
		if err := h.waGateway.Purge(req.SessionID); err != nil {
			response.WriteHTTPError(w, err)
			return
		}
	}
	if h.sessionRepo != nil {
		if err := h.sessionRepo.PurgeSession(r.Context(), req.SessionID); err != nil {
			response.WriteHTTPError(w, err)
			return
		}
	}

	_ = json.NewEncoder(w).Encode(GenericSuccessResponse{
		Success: true,
		Message: "session purged",
	})
}

func (h *Handler) handleResolveIdentity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	phoneParam := strings.TrimSpace(r.URL.Query().Get("phone"))
	if phoneParam == "" {
		phoneParam = strings.TrimSpace(r.URL.Query().Get("jid"))
	}
	if phoneParam == "" {
		response.WriteHTTPStatusError(w, http.StatusBadRequest, "phone or jid query parameter is required")
		return
	}

	if h.resolver == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "Identity resolver unavailable")
		return
	}

	identity, err := h.resolver.Resolve(r.Context(), phoneParam)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(IdentityResponse{Identity: identity})
}

func parseSessionIDFromQuery(r *http.Request) (uint, error) {
	str := r.URL.Query().Get("session_id")
	if str == "" {
		str = r.URL.Query().Get("id")
	}
	val, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, domainWA.ErrSessionNotFound
	}
	return uint(val), nil
}
