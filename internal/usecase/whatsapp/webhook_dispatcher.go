package whatsapp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quixiq/polyglot/internal/domain/whatsapp"
	"github.com/quixiq/polyglot/pkg/logger"
)

const defaultWebhookTimeout = 5 * time.Second

// WebhookDispatcher dispatches enriched inbound WhatsApp events to an external webhook.
type WebhookDispatcher struct {
	client     *http.Client
	webhookURL string
	secret     string
}

// NewWebhookDispatcher creates a new webhook dispatcher.
func NewWebhookDispatcher(webhookURL string, secret string, client *http.Client) *WebhookDispatcher {
	if client == nil {
		client = &http.Client{Timeout: defaultWebhookTimeout}
	}
	return &WebhookDispatcher{
		client:     client,
		webhookURL: webhookURL,
		secret:     secret,
	}
}

// WebhookURL returns the configured webhook target URL.
func (d *WebhookDispatcher) WebhookURL() string {
	return d.webhookURL
}

// SetWebhookURL dynamically updates the webhook destination.
func (d *WebhookDispatcher) SetWebhookURL(url string) {
	d.webhookURL = url
}

// Dispatch sends an event payload synchronously to the webhook endpoint.
func (d *WebhookDispatcher) Dispatch(ctx context.Context, event whatsapp.InboundEvent) error {
	if d.webhookURL == "" {
		return nil
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal inbound event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to construct webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Polyglot-WhatsApp-Gateway/1.0")

	if d.secret != "" {
		mac := hmac.New(sha256.New, []byte(d.secret))
		mac.Write(body)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
		req.Header.Set("X-Webhook-Secret", d.secret)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook delivery failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("webhook returned non-2xx status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DispatchAsync sends an event payload in a background goroutine with panic protection.
func (d *WebhookDispatcher) DispatchAsync(event whatsapp.InboundEvent) {
	if d.webhookURL == "" {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.WithComponent("WebhookDispatcher").WithFields(map[string]any{
					"session_id": event.SessionID,
					"panic":      r,
				}).Error("panic during webhook event dispatch")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), defaultWebhookTimeout)
		defer cancel()

		if err := d.Dispatch(ctx, event); err != nil {
			logger.WithComponent("WebhookDispatcher").WithError(err).WithFields(map[string]any{
				"session_id": event.SessionID,
				"event":      event.Event,
				"url":        d.webhookURL,
			}).Warn("failed to deliver inbound event to webhook")
		}
	}()
}
