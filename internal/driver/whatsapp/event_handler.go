package whatsapp

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/domain/whatsapp"
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

// MakePollVoteCallback wraps an asynchronous poll-vote handler into a
// PollVoteCallback. The handler runs outside the whatsmeow event loop so a
// harness turn cannot block message processing.
func (eh *EventHandler) MakePollVoteCallback(handler func(ctx context.Context, sessionID uint, chatJID string, pollID string, voterNumber string, optionHashes [][]byte) error) PollVoteCallback {
	return func(sessionID uint, chatJID string, pollID string, voterNumber string, optionHashes [][]byte) {
		if handler == nil {
			return
		}
		hashes := make([][]byte, len(optionHashes))
		for i, hash := range optionHashes {
			hashes[i] = append([]byte(nil), hash...)
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.WithComponent("EventHandler").WithFields(map[string]any{
						"session_id": sessionID,
						"poll_id":    pollID,
						"panic":      r,
					}).Error("panic during poll vote processing")
				}
			}()
			ctx := context.Background()
			if err := handler(ctx, sessionID, chatJID, pollID, voterNumber, hashes); err != nil {
				logger.WithComponent("EventHandler").WithError(err).WithFields(map[string]any{
					"session_id": sessionID,
					"poll_id":    pollID,
				}).Error("failed to process poll vote")
			}
		}()
	}
}

func (eh *EventHandler) handleStatusUpdate(sessionID uint, status string, qrCode string, jid string, phoneNumber string) {
	if eh.sessionRepo == nil {
		return
	}

	ctx := context.Background()
	sess, err := eh.sessionRepo.FindSessionByID(ctx, sessionID)
	if err == nil {
		sess.Status = whatsapp.WASessionStatus(status)
		if status == string(whatsapp.StatusOnline) {
			sess.ConnectedAt = time.Now()
		}
		if status == string(whatsapp.StatusNeedsRescan) {
			// Logout removed the local device, so the previous JID no longer
			// identifies a paired device and must not be matched on the next
			// connect.
			sess.JID = ""
		} else if jid != "" {
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
			"is_logged_in": status == string(whatsapp.StatusOnline),
		})
	}
}

// MakeMessageCallback wraps an asynchronous message handler function into a MessageCallback.
func (eh *EventHandler) MakeMessageCallback(handler func(ctx context.Context, sessionID uint, messageID string, chatJID string, customerNumber string, content string) error) MessageCallback {
	return func(sessionID uint, messageID string, chatJID string, customerNumber string, content string) {
		if handler == nil {
			return
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.WithComponent("EventHandler").WithFields(map[string]any{
						"session_id": sessionID,
						"message_id": messageID,
						"panic":      r,
					}).Error("panic during incoming message processing")
				}
			}()
			ctx := context.Background()
			if err := handler(ctx, sessionID, messageID, chatJID, customerNumber, content); err != nil {
				logger.WithComponent("EventHandler").WithError(err).WithFields(map[string]any{
					"session_id": sessionID,
					"message_id": messageID,
				}).Error("failed to process incoming message")
			}
		}()
	}
}
