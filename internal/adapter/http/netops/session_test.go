package netops

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCodecRoundTrip(t *testing.T) {
	codec, err := NewSessionCodec("0123456789abcdef", time.Hour)
	require.NoError(t, err)

	now := time.Unix(1_700_000_000, 0)
	value := codec.Issue(42, now)
	userID, err := codec.Verify(value, now.Add(30*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, uint(42), userID)
}

func TestSessionCodecRejectsTamperingAndExpiry(t *testing.T) {
	codec, err := NewSessionCodec("0123456789abcdef", time.Hour)
	require.NoError(t, err)
	now := time.Unix(1_700_000_000, 0)

	value := codec.Issue(42, now)
	_, err = codec.Verify(value+"x", now)
	require.ErrorIs(t, err, errInvalidSession)

	tampered := "v1.43." + value[len("v1.42."):]
	_, err = codec.Verify(tampered, now)
	require.ErrorIs(t, err, errInvalidSession)

	_, err = codec.Verify(value, now.Add(2*time.Hour))
	require.ErrorIs(t, err, errInvalidSession)
}

func TestSessionCodecRejectsInvalidConfiguration(t *testing.T) {
	_, err := NewSessionCodec("", time.Hour)
	require.Error(t, err)
	_, err = NewSessionCodec("0123456789abcdef", 0)
	require.Error(t, err)
}

func TestSessionCookieIsScopedToMount(t *testing.T) {
	codec, err := NewSessionCodec("0123456789abcdef", time.Hour)
	require.NoError(t, err)

	cookie := codec.Cookie("value", true)
	assert.Equal(t, SessionCookieName, cookie.Name)
	assert.Equal(t, MountPath, cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)

	cleared := codec.ClearCookie(false)
	assert.Equal(t, -1, cleared.MaxAge)
	assert.Equal(t, MountPath, cleared.Path)
}
