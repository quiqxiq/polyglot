package netops

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/auth"
)

const testSecret = "0123456789abcdef0123456789abcdef"

type fakeAuthorizer struct {
	err   error
	calls int
}

func (f *fakeAuthorizer) AuthorizeAdminUser(context.Context, uint) error {
	f.calls++
	return f.err
}

func TestNewUpstreamTransportDisablesEnvironmentProxy(t *testing.T) {
	transport := newUpstreamTransport()
	assert.Nil(t, transport.Proxy, "loopback harness traffic must not use HTTP_PROXY")
	require.NotNil(t, transport.TLSClientConfig)
	assert.Equal(t, uint16(tls.VersionTLS12), transport.TLSClientConfig.MinVersion)
}

func newTestHandler(t *testing.T, upstreamURL string, authorizer AdminAuthorizer, webToken string) (*Handler, *SessionCodec, *TicketStore) {
	t.Helper()
	upstream, err := url.Parse(upstreamURL)
	require.NoError(t, err)
	session, err := NewSessionCodec(testSecret, time.Hour)
	require.NoError(t, err)
	tickets, err := NewTicketStore(newTestCache(), time.Minute)
	require.NoError(t, err)
	signer, err := NewProxySigner(testSecret)
	require.NoError(t, err)

	handler, err := NewHandler(Options{
		Upstream:   upstream,
		WebToken:   webToken,
		Session:    session,
		Tickets:    tickets,
		Signer:     signer,
		Authorizer: authorizer,
	})
	require.NoError(t, err)
	return handler, session, tickets
}

func sessionRequest(method, target string, session *SessionCodec, userID uint) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: session.Issue(userID, time.Now())})
	return req
}

func TestHandlerRequiresSession(t *testing.T) {
	handler, _, _ := newTestHandler(t, "http://127.0.0.1:1", &fakeAuthorizer{}, "launch")

	req := httptest.NewRequest(http.MethodGet, MountPrefix, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandlerRedeemsTicketAndScopesCookie(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()
	handler, session, tickets := newTestHandler(t, upstream.URL, &fakeAuthorizer{}, "launch")

	ticket, err := tickets.Issue(context.Background(), 42)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, MountPrefix+"?ticket="+url.QueryEscape(ticket), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, MountPrefix, rec.Header().Get("Location"))
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, SessionCookieName, cookies[0].Name)
	assert.Equal(t, MountPath, cookies[0].Path)
	assert.True(t, cookies[0].HttpOnly)
	userID, err := session.Verify(cookies[0].Value, time.Now())
	require.NoError(t, err)
	assert.Equal(t, uint(42), userID)

	replay := httptest.NewRequest(http.MethodGet, MountPrefix+"?ticket="+url.QueryEscape(ticket), nil)
	replayRec := httptest.NewRecorder()
	handler.ServeHTTP(replayRec, replay)
	assert.Equal(t, http.StatusForbidden, replayRec.Code)
}

func TestHandlerRejectsTicketForRevokedAdmin(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()
	handler, _, tickets := newTestHandler(t, upstream.URL, &fakeAuthorizer{err: auth.ErrStreamForbidden}, "launch")

	ticket, err := tickets.Issue(context.Background(), 42)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, MountPrefix+"?ticket="+url.QueryEscape(ticket), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Empty(t, rec.Result().Cookies())
}

func TestHandlerForwardsRequestAndRewritesResponse(t *testing.T) {
	var got *http.Request
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		w.Header().Set("Set-Cookie", "dsh-auth-test=v; Path=/; HttpOnly; SameSite=Strict")
		w.Header().Set("Location", "/next")
		w.WriteHeader(http.StatusFound)
	}))
	defer upstream.Close()
	handler, session, _ := newTestHandler(t, upstream.URL, &fakeAuthorizer{}, "launch")
	signer, err := NewProxySigner(testSecret)
	require.NoError(t, err)

	req := sessionRequest(http.MethodGet, MountPrefix+"api/devices?x=1", session, 42)
	req.Host = "polyglot.example"
	req.Header.Set("Authorization", "Bearer must-not-leak")
	req.AddCookie(&http.Cookie{Name: auth.RefreshTokenCookieName, Value: "refresh"})
	req.AddCookie(&http.Cookie{Name: "dsh-auth-abc", Value: "old"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.NotNil(t, got)
	assert.Equal(t, "/api/devices", got.URL.Path)
	assert.Equal(t, "x=1", got.URL.RawQuery)
	assert.Equal(t, "polyglot.example", got.Host)
	assert.Empty(t, got.Header.Get("Authorization"))
	assert.Equal(t, "dsh-auth-abc=old", got.Header.Get("Cookie"))
	assert.Equal(t, MountPath, got.Header.Get("X-Forwarded-Prefix"))
	assert.Equal(t, "polyglot.example", got.Header.Get("X-Forwarded-Host"))
	require.NoError(t, signer.Verify(got.Header.Get(TrustedProxyHeader), 42, time.Now(), "GET", "/api/devices?x=1", time.Minute))

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, MountPrefix+"next", rec.Header().Get("Location"))
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, MountPath, cookies[0].Path)
	assert.True(t, cookies[0].HttpOnly)
	assert.Empty(t, cookies[0].Domain)
}

func TestHandlerInjectsLaunchTokenForIndex(t *testing.T) {
	var sawToken atomic.Bool
	var gotHeader string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.URL.Query().Get(indexAuthQuery) != "launch" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		sawToken.Store(true)
		gotHeader = r.Header.Get(TrustedProxyHeader)
		w.Header().Set("Set-Cookie", "dsh-auth-new=v; Path=/; HttpOnly")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()
	handler, session, _ := newTestHandler(t, upstream.URL, &fakeAuthorizer{}, "launch")
	signer, err := NewProxySigner(testSecret)
	require.NoError(t, err)

	req := sessionRequest(http.MethodGet, MountPrefix, session, 42)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, sawToken.Load())
	require.NoError(t, signer.Verify(gotHeader, 42, time.Now(), http.MethodGet, "/?token=launch", time.Minute))
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, MountPath, cookies[0].Path)
}

func TestHandlerRetriesStaleHarnessCookie(t *testing.T) {
	var attempts atomic.Int32
	var gotHeader string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		if r.URL.Query().Get(indexAuthQuery) == "launch" {
			gotHeader = r.Header.Get(TrustedProxyHeader)
			w.Header().Set("Set-Cookie", "dsh-auth-new=v; Path=/; HttpOnly")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()
	handler, session, _ := newTestHandler(t, upstream.URL, &fakeAuthorizer{}, "launch")
	signer, err := NewProxySigner(testSecret)
	require.NoError(t, err)

	req := sessionRequest(http.MethodGet, MountPrefix, session, 42)
	req.AddCookie(&http.Cookie{Name: "dsh-auth-stale", Value: "stale"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int32(2), attempts.Load())
	require.NoError(t, signer.Verify(gotHeader, 42, time.Now(), http.MethodGet, "/?token=launch", time.Minute))
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, MountPath, cookies[0].Path)
}

func TestHandlerRejectsRevokedAdminSession(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	tests := []struct {
		name       string
		authorizer error
		wantStatus int
	}{
		{name: "demoted", authorizer: auth.ErrStreamForbidden, wantStatus: http.StatusForbidden},
		{name: "inactive", authorizer: auth.ErrStreamUnauthenticated, wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, session, _ := newTestHandler(t, upstream.URL, &fakeAuthorizer{err: tt.authorizer}, "launch")
			req := sessionRequest(http.MethodGet, MountPrefix, session, 42)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			cookies := rec.Result().Cookies()
			require.Len(t, cookies, 1)
			assert.Equal(t, -1, cookies[0].MaxAge)
		})
	}
}

func TestHandlerRejectsInvalidTicketQuery(t *testing.T) {
	handler, _, _ := newTestHandler(t, "http://127.0.0.1:1", &fakeAuthorizer{}, "launch")

	req := httptest.NewRequest(http.MethodGet, MountPrefix+"?ticket=", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	post := httptest.NewRequest(http.MethodPost, MountPrefix+"?ticket=value", nil)
	postRec := httptest.NewRecorder()
	handler.ServeHTTP(postRec, post)
	assert.Equal(t, http.StatusMethodNotAllowed, postRec.Code)
}

func TestTicketHandlerIssuesTicket(t *testing.T) {
	cache := newTestCache()
	tickets, err := NewTicketStore(cache, time.Minute)
	require.NoError(t, err)
	authorizer := &fakeAuthorizer{}
	handler := NewTicketHandler(tickets, authorizer)

	req := httptest.NewRequest(http.MethodPost, TicketPath, nil)
	req = req.WithContext(auth.WithIdentity(req.Context(), 42, []string{"admin"}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body ticketResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body.Ticket)
	assert.Equal(t, MountPrefix+"?ticket="+url.QueryEscape(body.Ticket), body.URL)
	assert.Equal(t, 60, body.ExpiresIn)

	userID, err := tickets.Redeem(context.Background(), body.Ticket)
	require.NoError(t, err)
	assert.Equal(t, uint(42), userID)
	assert.Equal(t, 1, authorizer.calls)
}

func TestTicketHandlerRejectsNonAdminAndUnconfigured(t *testing.T) {
	tickets, err := NewTicketStore(newTestCache(), time.Minute)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, TicketPath, nil)
	req = req.WithContext(auth.WithIdentity(req.Context(), 42, []string{"teknisi"}))
	rec := httptest.NewRecorder()
	NewTicketHandler(tickets, &fakeAuthorizer{err: auth.ErrStreamForbidden}).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	unconfigured := httptest.NewRecorder()
	NewTicketHandler(nil, nil).ServeHTTP(unconfigured, req)
	assert.Equal(t, http.StatusServiceUnavailable, unconfigured.Code)

	noIdentity := httptest.NewRequest(http.MethodPost, TicketPath, nil)
	noIdentityRec := httptest.NewRecorder()
	NewTicketHandler(tickets, &fakeAuthorizer{}).ServeHTTP(noIdentityRec, noIdentity)
	assert.Equal(t, http.StatusUnauthorized, noIdentityRec.Code)

	getRec := httptest.NewRecorder()
	NewTicketHandler(tickets, &fakeAuthorizer{}).ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, TicketPath, nil))
	assert.Equal(t, http.StatusMethodNotAllowed, getRec.Code)
}

func TestHandlerForwardsWebSocketUpgrade(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		typ, data, err := conn.Read(r.Context())
		if err != nil {
			return
		}
		_ = conn.Write(r.Context(), typ, data)
	}))
	defer upstream.Close()
	handler, session, _ := newTestHandler(t, upstream.URL, &fakeAuthorizer{}, "launch")
	front := httptest.NewServer(handler)
	defer front.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wsURL := "ws" + strings.TrimPrefix(front.URL, "http") + MountPrefix + "api/socket"
	conn, resp, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{ //nolint:bodyclose // a successful upgrade owns the connection through conn.
		HTTPHeader: http.Header{
			"Cookie": {SessionCookieName + "=" + session.Issue(42, time.Now())},
		},
	})
	if resp != nil && err != nil {
		_ = resp.Body.Close()
	}
	require.NoError(t, err)
	defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()

	require.NoError(t, conn.Write(ctx, websocket.MessageText, []byte("ping")))
	typ, data, err := conn.Read(ctx)
	require.NoError(t, err)
	assert.Equal(t, websocket.MessageText, typ)
	assert.Equal(t, "ping", string(data))
}

func TestHandlerRejectsMalformedSessionCookie(t *testing.T) {
	handler, _, _ := newTestHandler(t, "http://127.0.0.1:1", &fakeAuthorizer{}, "launch")

	req := httptest.NewRequest(http.MethodGet, MountPrefix, nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "not-a-session"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
