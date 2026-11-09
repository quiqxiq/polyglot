package whatsapp_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainWA "github.com/quixiq/polyglot/internal/domain/whatsapp"
	ucWA "github.com/quixiq/polyglot/internal/usecase/whatsapp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookDispatcher_Dispatch(t *testing.T) {
	secret := "test-secret-key-123"
	var receivedBody []byte
	var receivedSig string
	var receivedSecret string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSig = r.Header.Get("X-Hub-Signature-256")
		receivedSecret = r.Header.Get("X-Webhook-Secret")
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	dispatcher := ucWA.NewWebhookDispatcher(server.URL, secret, server.Client())
	assert.Equal(t, server.URL, dispatcher.WebhookURL())

	event := domainWA.InboundEvent{
		Event:     "message.received",
		SessionID: 1,
		MessageID: "msg-12345",
		Timestamp: time.Now().Unix(),
		Sender: domainWA.SenderIdentity{
			PhoneNumber:   "628123456789",
			JID:           "628123456789@s.whatsapp.net",
			Role:          domainWA.RoleTechnician,
			DisplayName:   "Agus Teknisi",
			AllowedScopes: domainWA.DefaultScopesForRole(domainWA.RoleTechnician),
		},
		Message: domainWA.InboundMessage{
			Type: "text",
			Text: "cek redaman odp 5",
		},
	}

	err := dispatcher.Dispatch(context.Background(), event)
	require.NoError(t, err)

	var decoded domainWA.InboundEvent
	err = json.Unmarshal(receivedBody, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "message.received", decoded.Event)
	assert.Equal(t, "msg-12345", decoded.MessageID)
	assert.Equal(t, domainWA.RoleTechnician, decoded.Sender.Role)
	assert.Equal(t, "cek redaman odp 5", decoded.Message.Text)

	// Verify HMAC signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(receivedBody)
	expectedSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expectedSig, receivedSig)
	assert.Equal(t, secret, receivedSecret)
}

func TestWebhookDispatcher_ErrorResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server crashed"))
	}))
	defer server.Close()

	dispatcher := ucWA.NewWebhookDispatcher(server.URL, "", server.Client())
	err := dispatcher.Dispatch(context.Background(), domainWA.InboundEvent{Event: "ping"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-2xx status: 500")
}

func TestWebhookDispatcher_EmptyURL(t *testing.T) {
	dispatcher := ucWA.NewWebhookDispatcher("", "", nil)
	err := dispatcher.Dispatch(context.Background(), domainWA.InboundEvent{Event: "test"})
	assert.NoError(t, err)

	// Async call should not panic or fail
	dispatcher.DispatchAsync(domainWA.InboundEvent{Event: "async_test"})
}
