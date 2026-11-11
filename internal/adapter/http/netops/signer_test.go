package netops

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxySignerVerifiesBoundHeader(t *testing.T) {
	signer, err := NewProxySigner("0123456789abcdef")
	require.NoError(t, err)
	now := time.Unix(1_700_000_000, 0)

	header := signer.Header(42, now, "GET", "/api/devices?x=1")
	require.NoError(t, signer.Verify(header, 42, now, "GET", "/api/devices?x=1", time.Minute))

	require.ErrorIs(t, signer.Verify(header, 43, now, "GET", "/api/devices?x=1", time.Minute), errInvalidProxyHeader)
	require.ErrorIs(t, signer.Verify(header, 42, now, "POST", "/api/devices?x=1", time.Minute), errInvalidProxyHeader)
	require.ErrorIs(t, signer.Verify(header, 42, now.Add(2*time.Minute), "GET", "/api/devices?x=1", time.Minute), errInvalidProxyHeader)

	tampered := strings.Replace(header, "sig=", "sig=0", 1)
	require.ErrorIs(t, signer.Verify(tampered, 42, now, "GET", "/api/devices?x=1", time.Minute), errInvalidProxyHeader)
}

func TestProxySignerRejectsInvalidInput(t *testing.T) {
	_, err := NewProxySigner("")
	require.Error(t, err)

	signer, err := NewProxySigner("0123456789abcdef")
	require.NoError(t, err)
	require.ErrorIs(t, signer.Verify("garbage", 1, time.Now(), "GET", "/", time.Minute), errInvalidProxyHeader)
	assert.NotEmpty(t, signer.Header(1, time.Now(), "GET", "/"))
}
