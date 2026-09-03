package whatsapp

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

type EventHandler struct {
	sessionRepo port.WASessionRepository
	sseHub      *ws.SSEHub
}

func NewEventHandler(sessionRepo port.WASessionRepository, sseHub *ws.SSEHub) *EventHandler {
	return &EventHandler{
		sessionRepo: sessionRepo,
		sseHub:      sseHub,
	}
}

func (eh *EventHandler) MakeStatusCallback() StatusCallback {
	return func(sessionID uint, status string, qrCode string, jid string, phoneNumber string) {
		go eh.handleStatusUpdate(sessionID, status, qrCode, jid, phoneNumber)
	}
}

// MakeChatUpdateCallback returns the callback yang menyiarkan event SSE
// `chat_update` setiap kali mirror chat berubah (pesan masuk ATAU keluar),
// sehingga Inbox frontend bisa refresh instan tanpa polling.
func (eh *EventHandler) MakeChatUpdateCallback() ChatUpdateCallback {
	return func(sessionID uint, chatJID string) {
		if eh.sseHub != nil {
			eh.sseHub.Broadcast("chat_update", map[string]any{
				"session_id": sessionID,
				"chat_jid":   chatJID,
			})
		}
	}
}

// MakeChatPresenceCallback returns the callback yang menyiarkan event SSE
// `chat_presence` setiap kali kontak mengetik atau merekam voice, sehingga
// frontend bisa menampilkan indikator "mengetik…" / "merekam…" secara real-time.
// Payload: session_id, chat_jid, sender_jid, state ("composing"|"paused"),
// media (""|"audio"), is_group.
func (eh *EventHandler) MakeChatPresenceCallback() ChatPresenceCallback {
	return func(sessionID uint, chatJID, senderJID, state, media string, isGroup bool) {
		if eh.sseHub != nil {
			eh.sseHub.Broadcast("chat_presence", map[string]any{
				"session_id": sessionID,
				"chat_jid":   chatJID,
				"sender_jid": senderJID,
				"state":      state,
				"media":      media,
				"is_group":   isGroup,
			})
		}
	}
}

func (eh *EventHandler) handleStatusUpdate(sessionID uint, status string, qrCode string, jid string, phoneNumber string) {
	if eh.sessionRepo == nil {
		return
	}

	ctx := context.Background()
	sess, err := eh.sessionRepo.FindSessionByID(ctx, sessionID)
	if err == nil {
		sess.Status = bot.WASessionStatus(status)
		if status == string(bot.StatusOnline) {
			sess.ConnectedAt = time.Now()
		}
		if jid != "" {
			sess.JID = jid
		}
		if phoneNumber != "" {
			sess.PhoneNumber = phoneNumber
		}
		_ = eh.sessionRepo.UpdateSession(ctx, sess)
	}

	if eh.sseHub != nil {
		eh.sseHub.Broadcast("session_status", map[string]any{
			"session_id":   sessionID,
			"status":       status,
			"qr_code":      qrCode,
			"jid":          jid,
			"phone_number": phoneNumber,
			"is_logged_in": status == string(bot.StatusOnline),
		})
	}
}

func (eh *EventHandler) MakeMessageCallback(engineHandler func(ctx context.Context, sessionID uint, chatJID string, customerNumber string, content string) error) MessageCallback {
	return func(sessionID uint, chatJID string, customerNumber string, content string) {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.WithComponent("EventHandler").WithFields(map[string]any{
						"session_id": sessionID,
						"panic":      r,
					}).Error("panic during incoming message processing")
				}
			}()
			ctx := context.Background()
			if err := engineHandler(ctx, sessionID, chatJID, customerNumber, content); err != nil {
				logger.WithComponent("EventHandler").WithError(err).WithField("session_id", sessionID).Error("failed to process incoming message")
			}
		}()
	}
}
