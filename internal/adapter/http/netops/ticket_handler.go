package netops

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

type ticketResponse struct {
	Ticket    string `json:"ticket"`
	URL       string `json:"url"`
	ExpiresIn int    `json:"expires_in"`
}

// TicketHandler issues a one-time ticket to an authenticated admin. The ticket
// is the only credential the browser puts in the NetOps URL; the long-lived
// Polyglot JWT and refresh cookie never leave the API origin.
type TicketHandler struct {
	tickets    *TicketStore
	authorizer AdminAuthorizer
}

// NewTicketHandler returns the ticket endpoint handler.
func NewTicketHandler(tickets *TicketStore, authorizer AdminAuthorizer) *TicketHandler {
	return &TicketHandler{tickets: tickets, authorizer: authorizer}
}

func (h *TicketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		response.WriteHTTPStatusError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.tickets == nil || h.authorizer == nil {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "NetOps UI is not configured")
		return
	}

	userID, _, ok := auth.IdentityFromContext(r.Context())
	if !ok || userID == 0 {
		response.WriteHTTPStatusError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	if err := h.authorizer.AuthorizeAdminUser(r.Context(), userID); err != nil {
		writeAuthError(w, err)
		return
	}

	ticket, err := h.tickets.Issue(r.Context(), userID)
	if err != nil {
		logger.WithComponent("NetOpsTicket").WithField("user_id", userID).Error("failed to issue netops ticket")
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "NetOps ticket service is unavailable")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ticketResponse{
		Ticket:    ticket,
		URL:       MountPrefix + "?" + url.Values{ticketQuery: {ticket}}.Encode(),
		ExpiresIn: int(h.tickets.TTL().Seconds()),
	})
}
