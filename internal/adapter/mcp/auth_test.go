package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(clone)
}

func connectMCP(t *testing.T, srv *httptest.Server, token string) *mcp.ClientSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	transport := &mcp.StreamableClientTransport{
		Endpoint:             srv.URL,
		DisableStandaloneSSE: true,
		HTTPClient: &http.Client{
			Transport: bearerTransport{token: token, base: http.DefaultTransport},
		},
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "auth-test"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func TestMCPHTTP_FailsClosedWithoutToken(t *testing.T) {
	s := newTestServer(t, &testDriver{})
	srv := httptest.NewServer(s.HTTPHandler())
	t.Cleanup(srv.Close)

	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(`{}`))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestMCPHTTP_RejectsInvalidToken(t *testing.T) {
	s := newTestServer(t, &testDriver{}).WithBearerAuth("admin-secret", "read-secret")
	srv := httptest.NewServer(s.HTTPHandler())
	t.Cleanup(srv.Close)

	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(`{}`))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMCPHTTP_ReadTokenHidesAndDeniesMutationTools(t *testing.T) {
	s := newTestServer(t, &testDriver{}).WithBearerAuth("admin-secret", "read-secret")
	srv := httptest.NewServer(s.HTTPHandler())
	t.Cleanup(srv.Close)

	session := connectMCP(t, srv, "read-secret")
	tools, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)

	names := make(map[string]bool, len(tools.Tools))
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}
	assert.True(t, names["list_devices"])
	assert.True(t, names["ont_get_status"])
	assert.False(t, names["routeros_run"])
	assert.False(t, names["push_config"])

	_, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "routeros_run",
		Arguments: map[string]any{"device_id": "dev1", "command": "/system/reboot"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient scope")
}

func TestMCPHTTP_AdminTokenSeesAllTools(t *testing.T) {
	s := newTestServer(t, &testDriver{}).WithBearerAuth("admin-secret", "read-secret")
	srv := httptest.NewServer(s.HTTPHandler())
	t.Cleanup(srv.Close)

	session := connectMCP(t, srv, "admin-secret")
	tools, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)

	names := make(map[string]bool, len(tools.Tools))
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}
	assert.True(t, names["list_devices"])
	assert.True(t, names["routeros_run"])
}
