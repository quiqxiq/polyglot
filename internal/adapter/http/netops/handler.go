// Package netops exposes the DeepSeek harness NetOps UI to Polyglot
// administrators through a same-origin reverse proxy. The harness itself stays
// bound to loopback; browsers reach it only after a one-time ticket exchange
// that proves a live Polyglot admin session.
package netops

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/response"
)

const (
	// MountPath is the browser-visible prefix owned by the NetOps proxy.
	MountPath = "/netops"
	// MountPrefix is the subtree registered on the root mux.
	MountPrefix = MountPath + "/"
	// TicketPath issues one-time tickets for the NetOps UI.
	TicketPath = "/api/netops/ticket"
	// ticketQuery is the browser query parameter consumed during redemption.
	ticketQuery = "ticket"
)

// AdminAuthorizer re-checks that a user is still active and still holds an
// administrator role. The proxy calls it on every request so a revoked admin
// loses access without waiting for a browser cookie to expire.
type AdminAuthorizer interface {
	AuthorizeAdminUser(ctx context.Context, userID uint) error
}

// Options configures the NetOps proxy.
type Options struct {
	// Upstream is the loopback URL of the harness web app.
	Upstream *url.URL
	// WebToken is the harness launch token printed by `dsh web`.
	WebToken string
	// Session signs and verifies Polyglot browser sessions for the mount.
	Session *SessionCodec
	// Tickets redeems one-time launch tickets.
	Tickets *TicketStore
	// Signer signs the trusted-proxy assertion sent upstream.
	Signer *ProxySigner
	// Authorizer performs the fresh admin check.
	Authorizer AdminAuthorizer
	// SecureCookies marks browser cookies Secure when Polyglot serves HTTPS.
	SecureCookies bool
}

// Handler serves the NetOps UI and its API under MountPrefix.
type Handler struct {
	upstream      *url.URL
	webToken      string
	session       *SessionCodec
	tickets       *TicketStore
	signer        *ProxySigner
	authorizer    AdminAuthorizer
	secureCookies bool
	proxy         http.Handler
	transport     http.RoundTripper
}

// NewHandler builds the NetOps reverse proxy.
func NewHandler(opts Options) (*Handler, error) {
	if opts.Upstream == nil {
		return nil, errors.New("netops upstream URL is required")
	}
	if opts.Upstream.Scheme != "http" && opts.Upstream.Scheme != "https" {
		return nil, errors.New("netops upstream URL must use http or https")
	}
	if opts.Upstream.Host == "" {
		return nil, errors.New("netops upstream URL must include a host")
	}
	if opts.Upstream.User != nil {
		return nil, errors.New("netops upstream URL must not include credentials")
	}
	if opts.Session == nil {
		return nil, errors.New("netops session codec is required")
	}
	if opts.Tickets == nil {
		return nil, errors.New("netops ticket store is required")
	}
	if opts.Signer == nil {
		return nil, errors.New("netops proxy signer is required")
	}
	if opts.Authorizer == nil {
		return nil, errors.New("netops admin authorizer is required")
	}

	h := &Handler{
		upstream:      opts.Upstream,
		webToken:      strings.TrimSpace(opts.WebToken),
		session:       opts.Session,
		tickets:       opts.Tickets,
		signer:        opts.Signer,
		authorizer:    opts.Authorizer,
		secureCookies: opts.SecureCookies,
	}
	h.transport = newUpstreamTransport()
	h.proxy = h.newProxy()
	return h, nil
}

// ServeHTTP authenticates the browser session and forwards the request.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path != MountPath && !strings.HasPrefix(path, MountPrefix) {
		response.WriteHTTPStatusError(w, http.StatusNotFound, "Not found")
		return
	}
	if path == MountPath {
		target := MountPrefix
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusPermanentRedirect)
		return
	}

	if values, ok := r.URL.Query()[ticketQuery]; ok {
		if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
			response.WriteHTTPStatusError(w, http.StatusBadRequest, "Invalid NetOps ticket")
			return
		}
		h.redeemTicket(w, r, values[0])
		return
	}

	userID, err := h.sessionUser(r)
	if err != nil {
		http.SetCookie(w, h.session.ClearCookie(h.secureCookies))
		writeAuthError(w, err)
		return
	}

	ctx := context.WithValue(r.Context(), netopsUserKey{}, userID)
	h.proxy.ServeHTTP(w, r.WithContext(ctx))
}

// redeemTicket consumes a one-time ticket, establishes the browser session,
// and redirects to the clean mount path so the ticket never stays in history.
func (h *Handler) redeemTicket(w http.ResponseWriter, r *http.Request, ticket string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		response.WriteHTTPStatusError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, err := h.tickets.Redeem(r.Context(), ticket)
	if err != nil {
		if errors.Is(err, errInvalidTicket) {
			response.WriteHTTPStatusError(w, http.StatusForbidden, "Invalid or expired NetOps ticket")
			return
		}
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "NetOps ticket service is unavailable")
		return
	}
	if err := h.authorizer.AuthorizeAdminUser(r.Context(), userID); err != nil {
		writeAuthError(w, err)
		return
	}

	http.SetCookie(w, h.session.Cookie(h.session.Issue(userID, time.Now()), h.secureCookies))

	target := *r.URL
	query := target.Query()
	query.Del(ticketQuery)
	target.RawQuery = query.Encode()
	http.Redirect(w, r, target.RequestURI(), http.StatusSeeOther)
}

// sessionUser verifies the browser cookie and performs the fresh admin check.
func (h *Handler) sessionUser(r *http.Request) (uint, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return 0, auth.ErrStreamUnauthenticated
	}
	userID, err := h.session.Verify(cookie.Value, time.Now())
	if err != nil {
		return 0, auth.ErrStreamUnauthenticated
	}
	if err := h.authorizer.AuthorizeAdminUser(r.Context(), userID); err != nil {
		return 0, err
	}
	return userID, nil
}

func writeAuthError(w http.ResponseWriter, err error) {
	if errors.Is(err, auth.ErrStreamForbidden) {
		response.WriteHTTPStatusError(w, http.StatusForbidden, "NetOps access requires an administrator role")
		return
	}
	response.WriteHTTPStatusError(w, http.StatusUnauthorized, "NetOps authentication required")
}

// NotConfiguredHandler fails closed when the harness web app is not configured.
func NotConfiguredHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response.WriteHTTPStatusError(w, http.StatusServiceUnavailable, "NetOps UI is not configured")
	})
}
