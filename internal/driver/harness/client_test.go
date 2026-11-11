package harness

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientAskReturnsFinalText(t *testing.T) {
	t.Setenv("DSH_HOME", t.TempDir())
	patch := writePatch(t)
	bin := writeScript(t, `
case "$*" in
  *"--profile headless"*"--approval-policy never"*"--json -"*) ;;
  *) echo "unexpected args: $*" >&2; exit 2 ;;
esac
input=$(cat)
printf '%s\n' '{"type":"session","sessionId":"s"}'
printf '{"type":"final","sessionId":"s","reason":{"kind":"completed"},"text":"balasan: %s"}\n' "$input"
`)
	client, err := New(Config{
		Bin:           bin,
		PatchPath:     patch,
		Timeout:       5 * time.Second,
		MaxConcurrent: 1,
	})
	require.NoError(t, err)

	got, err := client.Ask(context.Background(), "halo")
	require.NoError(t, err)
	assert.Equal(t, "balasan: halo", got)
}

func TestClientAskExportsMCPEnvironment(t *testing.T) {
	t.Setenv("DSH_HOME", t.TempDir())
	t.Setenv("POLYGLOT_MCP_URL", "http://stale.invalid/mcp")
	t.Setenv("POLYGLOT_MCP_READ_TOKEN", "stale-token")
	for _, key := range harnessScrubbedEnv {
		t.Setenv(key, "server-secret")
	}
	patch := writePatch(t)
	bin := writeScript(t, `
for name in JWT_SECRET ENCRYPTION_KEY DATABASE_URL REDIS_URL POLYGLOT_MCP_TOKEN POLYGLOT_WS_TOKEN HARNESS_WEB_TOKEN HARNESS_PROXY_SECRET; do
  if env | grep -q "^${name}="; then
    echo "server credential leaked into the harness environment: $name" >&2
    exit 3
  fi
done
cat >/dev/null
printf '{"type":"final","sessionId":"s","reason":{"kind":"completed"},"text":"%s|%s"}\n' "$POLYGLOT_MCP_URL" "$POLYGLOT_MCP_READ_TOKEN"
`)
	client, err := New(Config{
		Bin:           bin,
		PatchPath:     patch,
		MCPURL:        "http://127.0.0.1:9999/mcp",
		MCPReadToken:  "read-token",
		Timeout:       5 * time.Second,
		MaxConcurrent: 1,
	})
	require.NoError(t, err)

	got, err := client.Ask(context.Background(), "halo")
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:9999/mcp|read-token", got)
}

func TestClientUsesHarnessHomeWhenWorkDirUnset(t *testing.T) {
	home := t.TempDir()
	t.Setenv("DSH_HOME", home)
	patch := writePatch(t)
	bin := writeScript(t, `
cat >/dev/null
printf '{"type":"final","sessionId":"s","reason":{"kind":"completed"},"text":"%s|%s"}\n' "$(pwd -P)" "$(cd "$DSH_HOME" && pwd -P)"
`)
	client, err := New(Config{
		Bin:           bin,
		PatchPath:     patch,
		Timeout:       5 * time.Second,
		MaxConcurrent: 1,
	})
	require.NoError(t, err)

	got, err := client.Ask(context.Background(), "halo")
	require.NoError(t, err)
	resolvedHome, err := filepath.EvalSymlinks(home)
	require.NoError(t, err)
	assert.Equal(t, resolvedHome+"|"+resolvedHome, got)
}

func TestClientAskSurfacesErrorEvent(t *testing.T) {
	client := newTestClient(t, `
cat >/dev/null
printf '%s\n' '{"type":"error","message":"missing key"}'
exit 1
`)

	_, err := client.Ask(context.Background(), "halo")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnavailable)
	assert.Contains(t, err.Error(), "missing key")
}

func TestClientAskRejectsNonCompletedFinal(t *testing.T) {
	client := newTestClient(t, `
cat >/dev/null
printf '%s\n' '{"type":"final","sessionId":"s","reason":{"kind":"aborted"},"text":"partial"}'
`)

	_, err := client.Ask(context.Background(), "halo")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnavailable)
	assert.Contains(t, err.Error(), "aborted")
}

func TestClientAskRejectsMissingFinal(t *testing.T) {
	client := newTestClient(t, `
cat >/dev/null
printf '%s\n' '{"type":"status","status":"working"}'
`)

	_, err := client.Ask(context.Background(), "halo")
	require.ErrorIs(t, err, ErrNoReply)
}

func TestClientAskTimesOut(t *testing.T) {
	t.Setenv("DSH_HOME", t.TempDir())
	patch := writePatch(t)
	bin := writeScript(t, `
cat >/dev/null
while :; do :; done
`)
	client, err := New(Config{
		Bin:           bin,
		PatchPath:     patch,
		Timeout:       100 * time.Millisecond,
		MaxConcurrent: 1,
	})
	require.NoError(t, err)

	_, err = client.Ask(context.Background(), "halo")
	require.ErrorIs(t, err, ErrTimeout)
}

func TestNewRejectsMissingPatch(t *testing.T) {
	t.Setenv("DSH_HOME", t.TempDir())
	bin := writeScript(t, `exit 0`)
	_, err := New(Config{
		Bin:           bin,
		PatchPath:     filepath.Join(t.TempDir(), "missing.yml"),
		Timeout:       time.Second,
		MaxConcurrent: 1,
	})
	require.ErrorIs(t, err, ErrUnavailable)
}

func TestNewRejectsMissingBinary(t *testing.T) {
	_, err := New(Config{
		Bin:           filepath.Join(t.TempDir(), "missing-dsh"),
		PatchPath:     writePatch(t),
		Timeout:       time.Second,
		MaxConcurrent: 1,
	})
	require.ErrorIs(t, err, ErrUnavailable)
}

func newTestClient(t *testing.T, script string) *Client {
	t.Helper()
	t.Setenv("DSH_HOME", t.TempDir())
	client, err := New(Config{
		Bin:           writeScript(t, script),
		PatchPath:     writePatch(t),
		Timeout:       5 * time.Second,
		MaxConcurrent: 1,
	})
	require.NoError(t, err)
	return client
}

func writePatch(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "netops-cs.patch.yml")
	require.NoError(t, os.WriteFile(path, []byte("[]\n"), 0o600))
	return path
}

func writeScript(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script harness fixture is not available on Windows")
	}
	path := filepath.Join(t.TempDir(), "fake-dsh")
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o700))
	return path
}
