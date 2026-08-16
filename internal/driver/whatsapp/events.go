package whatsapp

import (
	"context"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/quixiq/polyglot/pkg/logger"
)

func (c *Client) handleEvent(evt any) {
	switch v := evt.(type) {
	case *events.Message:
		c.persistMirrorMessage(v)
		if !v.Info.IsFromMe {
			c.handleIncomingMessage(v)
		}
	case *events.Receipt:
		c.handleReceiptEvent(v)
	case *events.ChatPresence:
		c.handleChatPresenceEvent(v)
	case *events.Connected:
		logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Info("connected successfully (auto-reconnect active)")
		c.qrMutex.Lock()
		c.qrCode = ""
		c.qrBase64 = ""
		c.qrMutex.Unlock()
		if c.onStatus != nil && c.waClient.Store.ID != nil {
			c.onStatus(c.SessionID, "online", "", c.waClient.Store.ID.String(), c.waClient.Store.ID.User)
		}
		c.sendPresenceAvailable()
	case *events.AppStateSyncComplete:
		c.sendPresenceAvailable()
	case *events.Disconnected:
		logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Info("disconnected (auto-reconnect will retry)")
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "offline", "", "", "")
		}
	case *events.HistorySync:
		c.handleHistorySync(v)
	case *events.LoggedOut:
		logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Warn("remote logged out event received from phone")
		if c.deviceStore != nil {
			_ = c.deviceStore.Delete(context.Background())
		}
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "needs_rescan", "", "", "")
		}
	}
}

func (c *Client) handleIncomingMessage(evt *events.Message) {
	if evt.Info.IsFromMe {
		return
	}

	chatJID := normalizeJIDFromLID(context.Background(), evt.Info.Chat, c.waClient).String()
	if chatJID == "" || isSkippedJID(chatJID) {
		return
	}

	senderJID := evt.Info.Sender.User
	if senderJID == "" {
		senderJID = evt.Info.Chat.User
	}

	body := extractMessageBody(evt.Message)
	if body == "" {
		return
	}

	logger.WithComponent("WhatsAppDriver").WithFields(map[string]any{
		"session_id": c.SessionID,
		"sender":     senderJID,
		"chat_jid":   chatJID,
	}).Debugf("incoming message: %s", body)

	if cb := c.getMessageCallback(); cb != nil {
		cb(c.SessionID, chatJID, senderJID, body)
	}
}

func (c *Client) sendPresenceAvailable() {
	if c.waClient == nil || !c.waClient.IsConnected() || c.waClient.Store == nil || c.waClient.Store.ID == nil {
		return
	}
	if err := c.waClient.SendPresence(context.Background(), types.PresenceAvailable); err != nil {
		logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Debugf("failed to send presence: %v", err)
		return
	}
	logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Debug("presence available sent")
}

func (c *Client) handleChatPresenceEvent(evt *events.ChatPresence) {
	senderJID := normalizeJIDFromLID(context.Background(), evt.Sender, c.waClient)
	chatJID := evt.Chat.ToNonAD().String()
	senderStr := senderJID.ToNonAD().String()

	state := string(evt.State)
	media := string(evt.Media)

	if cb := c.getChatPresenceCallback(); cb != nil {
		cb(c.SessionID, chatJID, senderStr, state, media, evt.IsGroup)
	}
}

func (c *Client) handleReceiptEvent(evt *events.Receipt) {
	if c.chatRepo == nil || evt == nil || len(evt.MessageIDs) == 0 {
		return
	}
	var status string
	switch evt.Type {
	case types.ReceiptTypeDelivered:
		status = "delivered"
	case types.ReceiptTypeRead, types.ReceiptTypeReadSelf:
		status = "read"
	default:
		return
	}
	chatJID := normalizeJIDFromLID(context.Background(), evt.Chat, c.waClient).String()
	strIDs := make([]string, len(evt.MessageIDs))
	for i, id := range evt.MessageIDs {
		strIDs[i] = string(id)
	}
	if err := c.chatRepo.MarkMessagesStatus(context.Background(), c.SessionID, chatJID, strIDs, status); err != nil {
		logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Debugf("failed to mark messages %s in %s: %v", status, chatJID, err)
		return
	}
	c.notifyChatUpdate(chatJID)
}

func extractMessageBody(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}
	if msg.GetConversation() != "" {
		return msg.GetConversation()
	}
	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.GetText() != "" {
		return msg.ExtendedTextMessage.GetText()
	}
	if msg.ImageMessage != nil && msg.ImageMessage.GetCaption() != "" {
		return msg.ImageMessage.GetCaption()
	}
	if msg.VideoMessage != nil && msg.VideoMessage.GetCaption() != "" {
		return msg.VideoMessage.GetCaption()
	}
	return ""
}
