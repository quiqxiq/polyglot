package whatsapp

import (
	"context"
	"log"

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

func (eh *EventHandler) MakeStatusCallback() StatusCallback {
	return func(sessionID uint, status string, qrCode string, jid string, phoneNumber string) {
		go eh.handleStatusUpdate(sessionID, status, qrCode, jid, phoneNumber)
	}
}

func (eh *EventHandler) handleStatusUpdate(sessionID uint, status string, qrCode string, jid string, phoneNumber string) {
	if eh.pgStore == nil {
		return
	}

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
	}

	if eh.sseHub != nil {
		eh.sseHub.Broadcast("session_status", map[string]any{
			"session_id": sessionID,
			"status":     status,
			"qr_code":    qrCode,
		})
	}
}

func (eh *EventHandler) MakeMessageCallback(engineHandler func(ctx context.Context, sessionID uint, customerNumber string, content string) error) MessageCallback {
	return func(sessionID uint, customerNumber string, content string) {
		go func() {
			ctx := context.Background()
			if err := engineHandler(ctx, sessionID, customerNumber, content); err != nil {
				log.Printf("[EventHandler] Error processing message from %s: %v", customerNumber, err)
			}
		}()
	}
}
