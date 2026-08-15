package whatsapp

import (
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/quixiq/polyglot/pkg/logger"
)

func (c *Client) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		c.handleIncomingMessage(v)
	case *events.Connected:
		logger.WithField("session_id", c.SessionID).Info("Connected successfully (Auto-reconnect active)")
		if c.onStatus != nil && c.waClient.Store.ID != nil {
			c.onStatus(c.SessionID, "online", "", c.waClient.Store.ID.String(), c.waClient.Store.ID.User)
		}
	case *events.Disconnected:
		logger.WithField("session_id", c.SessionID).Warn("Disconnected (Auto-reconnect will retry)")
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "offline", "", "", "")
		}
	case *events.LoggedOut:
		logger.WithField("session_id", c.SessionID).Warn("Remote LoggedOut event received from phone")
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "needs_rescan", "", "", "")
		}
	}
}

func (c *Client) handleIncomingMessage(evt *events.Message) {
	if evt.Info.IsFromMe {
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

	logger.WithFields(logger.Fields{
		"session_id": c.SessionID,
		"from":       senderJID,
		"body":       body,
	}).Debug("Inbound WhatsApp message")

	if c.onMessage != nil {
		c.onMessage(c.SessionID, senderJID, body)
	}
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
