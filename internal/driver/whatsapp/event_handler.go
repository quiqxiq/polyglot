package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

type EventHandler struct {
	pgStore *postgres.Store
	sseHub  *ws.SSEHub
}

func NewEventHandler(pgStore *postgres.Store, sseHub *ws.SSEHub) *EventHandler {
	return &EventHandler{
		pgStore: pgStore,
		sseHub:  sseHub,
	}
}

func (eh *EventHandler) OnStatusChanged(sessionID uint, status string, qrCode string, jid string, phoneNumber string) {
	log.Printf("[EventHandler] Session %d status changed to: %s (JID: %s, Phone: %s)", sessionID, status, jid, phoneNumber)

	sess, err := eh.pgStore.FindSessionByID(sessionID)
	if err == nil {
		sess.Status = bot.WASessionStatus(status)
		if jid != "" {
			sess.JID = jid
		}
		if phoneNumber != "" {
			sess.PhoneNumber = phoneNumber
		}
		_ = eh.pgStore.UpdateSession(sess)

		if sess.WebhookURL != "" {
			dispatchWebhook(sess.WebhookURL, map[string]interface{}{
				"event":        "session_status",
				"session_id":   sessionID,
				"status":       status,
				"phone_number": phoneNumber,
				"timestamp":    time.Now().Unix(),
			})
		}
	}

	if eh.sseHub != nil {
		eh.sseHub.Broadcast("session_status", map[string]interface{}{
			"session_id": sessionID,
			"status":     status,
			"qr_code":    qrCode,
		})
	}
}

func (eh *EventHandler) MakeMessageCallback(engineHandler func(ctx context.Context, sessionID uint, customerNumber string, content string) error) MessageCallback {
	return func(sessionID uint, customerNumber string, content string) {
		go func() {
			sess, _ := eh.pgStore.FindSessionByID(sessionID)
			if sess != nil && sess.WebhookURL != "" {
				dispatchWebhook(sess.WebhookURL, map[string]interface{}{
					"event":           "message",
					"session_id":      sessionID,
					"customer_number": customerNumber,
					"content":         content,
					"timestamp":       time.Now().Unix(),
				})
			}

			ctx := context.Background()
			if err := engineHandler(ctx, sessionID, customerNumber, content); err != nil {
				log.Printf("[EventHandler] Error processing message from %s: %v", customerNumber, err)
			}
		}()
	}
}

func dispatchWebhook(webhookURL string, payload interface{}) {
	if webhookURL == "" {
		return
	}
	go func() {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[Webhook] Failed to marshal payload: %v", err)
			return
		}

		req, err := http.NewRequestWithContext(context.Background(), "POST", webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[Webhook] Failed to create request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Webhook] Failed to send webhook to %s: %v", webhookURL, err)
			return
		}
		defer resp.Body.Close()
		log.Printf("[Webhook] Successfully sent payload to %s (Status: %s)", webhookURL, resp.Status)
	}()
}
