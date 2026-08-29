package whatsapp

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/pkg/logger"
)

func (c *Client) reconcileLIDChat(ctx context.Context, resolved types.JID) {
	if c.chatRepo == nil || c.waClient == nil || c.waClient.Store == nil || c.waClient.Store.LIDs == nil {
		return
	}
	lid, err := c.waClient.Store.LIDs.GetLIDForPN(ctx, resolved)
	if err != nil || lid.IsEmpty() || lid.Server != "lid" {
		return
	}
	if err := c.chatRepo.MergeChatLID(ctx, c.SessionID, lid.String(), resolved.String()); err != nil {
		logger.WithComponent("WhatsAppDriver").WithError(err).WithField("session_id", c.SessionID).Debug("failed to merge LID chat")
	}
}

func (c *Client) persistMirrorMessage(evt *events.Message) {
	if c.chatRepo == nil {
		return
	}

	chatJID := normalizeJIDFromLID(context.Background(), evt.Info.Chat, c.waClient).String()
	if chatJID == "" || isSkippedJID(chatJID) {
		return
	}

	var displayName string
	if !evt.Info.IsGroup {
		displayName = c.resolveChatDisplayName(context.Background(), evt.Info.Chat, evt.Info.PushName)
	}

	content := extractMessageBody(evt.Message)
	mediaType := extractMediaType(evt.Message)

	if evt.Message == nil || (content == "" && (mediaType == "unknown" || mediaType == "reaction" || mediaType == "system")) {
		return
	}

	senderJID := normalizeJIDFromLID(context.Background(), evt.Info.Sender, c.waClient).String()

	msg := &bot.WAMessage{
		SessionID:   c.SessionID,
		ChatJID:     chatJID,
		WAMessageID: evt.Info.ID,
		SenderJID:   senderJID,
		SenderName:  evt.Info.PushName,
		Content:     content,
		MediaType:   mediaType,
		IsFromMe:    evt.Info.IsFromMe,
		Timestamp:   evt.Info.Timestamp,
	}
	if evt.Info.IsFromMe {
		msg.Status = "sent"
	}
	if msg.WAMessageID == "" {
		msg.WAMessageID = fmt.Sprintf("evt-%d", evt.Info.Timestamp.UnixNano())
	}

	inserted, err := c.chatRepo.UpsertMessage(context.Background(), msg)
	if err != nil {
		logger.WithComponent("WhatsAppDriver").WithError(err).WithField("session_id", c.SessionID).Error("failed to mirror message")
	}

	chat := &bot.WAChat{
		SessionID:          c.SessionID,
		ChatJID:            chatJID,
		DisplayName:        displayName,
		IsGroup:            evt.Info.IsGroup,
		LastMessageID:      msg.WAMessageID,
		LastMessagePreview: previewOf(content, mediaType),
		LastMessageTime:    evt.Info.Timestamp,
	}
	if err := c.chatRepo.UpsertChat(context.Background(), chat); err != nil {
		logger.WithComponent("WhatsAppDriver").WithError(err).WithField("session_id", c.SessionID).Error("failed to mirror chat")
	}

	if inserted && !evt.Info.IsFromMe {
		if err := c.chatRepo.IncrementUnread(context.Background(), c.SessionID, chatJID); err != nil {
			logger.WithComponent("WhatsAppDriver").WithError(err).WithField("session_id", c.SessionID).Debug("failed to increment unread")
		}
	}

	c.notifyChatUpdate(chatJID)
}

func (c *Client) recordOutgoingMessage(jid types.JID, waMessageID string, content string, mediaType string) {
	if c.chatRepo == nil {
		return
	}

	chatJID := jid.String()
	if chatJID == "" {
		return
	}
	if waMessageID == "" {
		waMessageID = fmt.Sprintf("out-%d", time.Now().UnixNano())
	}
	now := time.Now()

	msg := &bot.WAMessage{
		SessionID:   c.SessionID,
		ChatJID:     chatJID,
		WAMessageID: waMessageID,
		Content:     content,
		MediaType:   mediaType,
		IsFromMe:    true,
		Status:      "sent",
		Timestamp:   now,
	}
	if _, err := c.chatRepo.UpsertMessage(context.Background(), msg); err != nil {
		logger.WithComponent("WhatsAppDriver").WithError(err).WithField("session_id", c.SessionID).Error("failed to mirror outgoing message")
	}

	chat := &bot.WAChat{
		SessionID:          c.SessionID,
		ChatJID:            chatJID,
		LastMessageID:      waMessageID,
		LastMessagePreview: previewOf(content, mediaType),
		LastMessageTime:    now,
	}
	if err := c.chatRepo.UpsertChat(context.Background(), chat); err != nil {
		logger.WithComponent("WhatsAppDriver").WithError(err).WithField("session_id", c.SessionID).Error("failed to mirror outgoing chat")
	}

	c.notifyChatUpdate(chatJID)
}

func extractMediaType(msg *waE2E.Message) string {
	if msg == nil {
		return "text"
	}
	switch {
	case msg.GetConversation() != "" || msg.ExtendedTextMessage != nil:
		return "text"
	case msg.ImageMessage != nil:
		return "image"
	case msg.VideoMessage != nil:
		return "video"
	case msg.AudioMessage != nil:
		return "audio"
	case msg.DocumentMessage != nil || msg.DocumentWithCaptionMessage != nil:
		return "document"
	case msg.StickerMessage != nil:
		return "sticker"
	case msg.LocationMessage != nil || msg.LiveLocationMessage != nil:
		return "location"
	case msg.ContactMessage != nil || msg.ContactsArrayMessage != nil:
		return "contact"
	case msg.Call != nil || msg.CallLogMesssage != nil || msg.BcallMessage != nil || msg.ScheduledCallCreationMessage != nil || msg.ScheduledCallEditMessage != nil:
		return "call"
	case msg.ReactionMessage != nil || msg.EncReactionMessage != nil:
		return "reaction"
	case msg.PollCreationMessage != nil || msg.PollUpdateMessage != nil || msg.PollResultSnapshotMessage != nil || msg.PollAddOptionMessage != nil:
		return "poll"
	case msg.ProtocolMessage != nil || msg.PinInChatMessage != nil || msg.KeepInChatMessage != nil:
		return "system"
	default:
		return "unknown"
	}
}

func previewOf(content, mediaType string) string {
	if content != "" {
		return content
	}
	switch mediaType {
	case "image":
		return "📷 Foto"
	case "video":
		return "🎬 Video"
	case "audio":
		return "🎵 Audio"
	case "document":
		return "📄 Dokumen"
	case "sticker":
		return "🖼️ Stiker"
	case "location":
		return "📍 Lokasi"
	case "contact":
		return "👤 Kontak"
	case "call":
		return "📞 Panggilan"
	default:
		return "[media]"
	}
}
