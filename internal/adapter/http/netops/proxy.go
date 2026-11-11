package netops

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

const (
	// TrustedProxyHeader is defined in signer.go.
	trustedProxyPrefixHeader = "X-Forwarded-Prefix"
	harnessAuthCookiePrefix  = "dsh-auth-"
	indexAuthQuery           = "token"
)

type netopsUserKey struct{}

func newUpstreamTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// The harness is a loopback service. Never route its traffic (including
	// the launch token) through an environment-configured forward proxy.
	transport.Proxy = nil
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	return transport
}

func (h *Handler) newProxy() *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite:        h.rewrite,
		Transport:      h.transport,
		FlushInterval:  -1,
		ModifyResponse: h.modifyResponse,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, _ error) {
			logger.WithComponent("NetOpsProxy").
				WithField("path", r.URL.Path).
				Error("netops upstream request failed")
			response.WriteHTTPStatusError(w, http.StatusBadGateway, "NetOps upstream is unavailable")
		},
	}
}

// rewrite turns the inbound request into an upstream request while preserving
// the public Host the harness trust fence checks.
func (h *Handler) rewrite(pr *httputil.ProxyRequest) {
	out := pr.Out
	userID, _ := out.Context().Value(netopsUserKey{}).(uint)

	out.URL.Scheme = h.upstream.Scheme
	out.URL.Host = h.upstream.Host
	out.URL.Path = joinUpstreamPath(h.upstream.Path, stripMount(pr.In.URL.Path))
	out.URL.RawPath = ""
	out.URL.RawQuery = pr.In.URL.RawQuery
	out.RequestURI = ""
	out.Host = pr.In.Host

	out.Header.Del("Authorization")
	out.Header.Del("Proxy-Authorization")
	out.Header.Del(TrustedProxyHeader)
	pruneRequestCookies(out.Header)
	pr.SetXForwarded()
	out.Header.Set(trustedProxyPrefixHeader, MountPath)

	if h.shouldInjectIndexToken(out) {
		query := out.URL.Query()
		query.Set(indexAuthQuery, h.webToken)
		out.URL.RawQuery = query.Encode()
	}
	out.Header.Set(TrustedProxyHeader, h.signer.Header(userID, time.Now(), out.Method, out.URL.RequestURI()))
}

// modifyResponse keeps upstream redirects and cookies inside the mount and
// retries the index once when a stale harness cookie blocks a fresh launch.
func (h *Handler) modifyResponse(res *http.Response) error {
	h.rewriteLocation(res)
	h.rewriteSetCookies(res)

	if res.StatusCode != http.StatusUnauthorized || !h.shouldRetryIndex(res.Request) {
		return nil
	}
	retry, err := h.retryIndex(res.Request) //nolint:bodyclose // ownership transfers to res below.
	if err != nil {
		return fmt.Errorf("retry netops index authentication: %w", err)
	}
	_ = res.Body.Close()
	*res = *retry
	h.rewriteLocation(res)
	h.rewriteSetCookies(res)
	return nil
}

func (h *Handler) shouldInjectIndexToken(out *http.Request) bool {
	if h.webToken == "" || out.Method != http.MethodGet || out.URL.Path != "/" {
		return false
	}
	if out.URL.Query().Has(indexAuthQuery) {
		return false
	}
	return !hasHarnessAuthCookie(out.Header)
}

func (h *Handler) shouldRetryIndex(req *http.Request) bool {
	if req == nil || h.webToken == "" || req.Method != http.MethodGet || req.URL.Path != "/" {
		return false
	}
	if req.URL.Query().Has(indexAuthQuery) {
		return false
	}
	return hasHarnessAuthCookie(req.Header)
}

// retryIndex exchanges the launch token for a fresh harness cookie. The stale
// cookie stays on the request: the harness checks a valid token first.
func (h *Handler) retryIndex(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	target := *req.URL
	query := target.Query()
	query.Set(indexAuthQuery, h.webToken)
	target.RawQuery = query.Encode()
	clone.URL = &target
	clone.RequestURI = ""
	clone.Body = nil
	userID, _ := clone.Context().Value(netopsUserKey{}).(uint)
	clone.Header.Set(TrustedProxyHeader, h.signer.Header(userID, time.Now(), clone.Method, clone.URL.RequestURI()))
	return h.transport.RoundTrip(clone)
}

func (h *Handler) rewriteLocation(res *http.Response) {
	location := res.Header.Get("Location")
	if location == "" {
		return
	}
	target, err := url.Parse(location)
	if err != nil {
		return
	}
	if target.IsAbs() {
		publicHost := ""
		if res.Request != nil {
			publicHost = res.Request.Host
		}
		sameUpstream := strings.EqualFold(target.Scheme, h.upstream.Scheme) &&
			strings.EqualFold(target.Host, h.upstream.Host)
		samePublic := publicHost != "" && strings.EqualFold(target.Host, publicHost)
		if !sameUpstream && !samePublic {
			return
		}
		target.Scheme = ""
		target.Host = ""
	}
	if target.Path == "" || !strings.HasPrefix(target.Path, "/") {
		return
	}
	if target.Path == MountPath || strings.HasPrefix(target.Path, MountPrefix) {
		return
	}
	if target.Path == "/" {
		target.Path = MountPrefix
	} else {
		target.Path = MountPath + target.Path
	}
	res.Header.Set("Location", target.String())
}

func (h *Handler) rewriteSetCookies(res *http.Response) {
	values := res.Header.Values("Set-Cookie")
	if len(values) == 0 {
		return
	}
	res.Header.Del("Set-Cookie")
	for _, raw := range values {
		cookie, err := http.ParseSetCookie(raw)
		if err != nil {
			res.Header.Add("Set-Cookie", raw)
			continue
		}
		cookie.Path = mountCookiePath(cookie.Path)
		cookie.Domain = ""
		cookie.Secure = h.secureCookies
		res.Header.Add("Set-Cookie", cookie.String())
	}
}

func mountCookiePath(path string) string {
	if path == "" || path == "/" {
		return MountPath
	}
	if path == MountPath || strings.HasPrefix(path, MountPrefix) {
		return path
	}
	if strings.HasPrefix(path, "/") {
		return MountPath + path
	}
	return path
}

func stripMount(path string) string {
	if path == "" || path == MountPath {
		return "/"
	}
	if strings.HasPrefix(path, MountPrefix) {
		return strings.TrimPrefix(path, MountPath)
	}
	return path
}

func joinUpstreamPath(base string, rest string) string {
	if base == "" || base == "/" {
		if rest == "" {
			return "/"
		}
		return rest
	}
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(rest, "/")
}

func pruneRequestCookies(header http.Header) {
	raw := header.Get("Cookie")
	if raw == "" {
		return
	}
	kept := make([]string, 0, strings.Count(raw, ";")+1)
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, _, _ := strings.Cut(part, "=")
		if name == SessionCookieName || name == auth.RefreshTokenCookieName {
			continue
		}
		kept = append(kept, part)
	}
	if len(kept) == 0 {
		header.Del("Cookie")
		return
	}
	header.Set("Cookie", strings.Join(kept, "; "))
}

func hasHarnessAuthCookie(header http.Header) bool {
	for _, part := range strings.Split(header.Get("Cookie"), ";") {
		name, _, _ := strings.Cut(strings.TrimSpace(part), "=")
		if strings.HasPrefix(name, harnessAuthCookiePrefix) {
			return true
		}
	}
	return false
}
