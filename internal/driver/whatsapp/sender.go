package whatsapp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func (c *Client) SendMessage(ctx context.Context, to string, content string) error {
	if !c.waClient.IsConnected() {
		return fmt.Errorf("WA client for session %d is disconnected", c.SessionID)
	}

	jid, err := parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID (%s): %w", to, err)
	}

	msg := &waE2E.Message{
		Conversation: proto.String(content),
	}

	respID, err := c.waClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send WA message: %w", err)
	}

	c.recordOutgoingMessage(jid, string(respID.ID), content, "text")
	return nil
}

func (c *Client) SendDocument(ctx context.Context, to string, fileBytes []byte, fileName string, contentType string, caption string) error {
	if !c.waClient.IsConnected() {
		return fmt.Errorf("WA client for session %d is disconnected", c.SessionID)
	}

	jid, err := parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID (%s): %w", to, err)
	}

	if contentType == "" {
		contentType = http.DetectContentType(fileBytes)
		if contentType == "text/plain; charset=utf-8" && strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
			contentType = "application/pdf"
		}
	}

	uploadResp, err := c.waClient.Upload(ctx, fileBytes, whatsmeow.MediaDocument)
	if err != nil {
		return fmt.Errorf("failed to upload document to WA servers: %w", err)
	}

	docMsg := &waE2E.DocumentMessage{
		URL:           proto.String(uploadResp.URL),
		DirectPath:    proto.String(uploadResp.DirectPath),
		MediaKey:      uploadResp.MediaKey,
		Mimetype:      proto.String(contentType),
		FileSHA256:    uploadResp.FileSHA256,
		FileEncSHA256: uploadResp.FileEncSHA256,
		FileLength:    proto.Uint64(uploadResp.FileLength),
		FileName:      proto.String(fileName),
	}
	if caption != "" {
		docMsg.Caption = proto.String(caption)
	}

	msg := &waE2E.Message{
		DocumentMessage: docMsg,
	}

	respID, err := c.waClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send WA document message: %w", err)
	}

	c.recordOutgoingMessage(jid, string(respID.ID), caption, "document")
	return nil
}

func (c *Client) SendImage(ctx context.Context, to string, imageBytes []byte, contentType string, caption string) error {
	if !c.waClient.IsConnected() {
		return fmt.Errorf("WA client for session %d is disconnected", c.SessionID)
	}

	jid, err := parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID (%s): %w", to, err)
	}

	if contentType == "" {
		contentType = http.DetectContentType(imageBytes)
	}

	uploadResp, err := c.waClient.Upload(ctx, imageBytes, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload image to WA servers: %w", err)
	}

	imgMsg := &waE2E.ImageMessage{
		URL:           proto.String(uploadResp.URL),
		DirectPath:    proto.String(uploadResp.DirectPath),
		MediaKey:      uploadResp.MediaKey,
		Mimetype:      proto.String(contentType),
		FileSHA256:    uploadResp.FileSHA256,
		FileEncSHA256: uploadResp.FileEncSHA256,
		FileLength:    proto.Uint64(uploadResp.FileLength),
	}
	if caption != "" {
		imgMsg.Caption = proto.String(caption)
	}

	msg := &waE2E.Message{
		ImageMessage: imgMsg,
	}

	respID, err := c.waClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send WA image message: %w", err)
	}

	c.recordOutgoingMessage(jid, string(respID.ID), caption, "image")
	return nil
}

// SendChatPresence updates the typing or recording status for a chat.
func (c *Client) SendChatPresence(ctx context.Context, to string, state string, media string) error {
	if !c.waClient.IsConnected() {
		return fmt.Errorf("WA client for session %d is disconnected", c.SessionID)
	}

	jid, err := parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid recipient JID (%s): %w", to, err)
	}

	presenceState := types.ChatPresence(state)
	if presenceState == "" {
		presenceState = types.ChatPresenceComposing
	}
	presenceMedia := types.ChatPresenceMedia(media)

	if err := c.waClient.SendChatPresence(ctx, jid, presenceState, presenceMedia); err != nil {
		return fmt.Errorf("failed to send WA chat presence: %w", err)
	}
	return nil
}

// MarkRead sends read receipts for the given message IDs.
func (c *Client) MarkRead(ctx context.Context, messageIDs []string, chatJID string, senderJID string) error {
	if !c.waClient.IsConnected() {
		return fmt.Errorf("WA client for session %d is disconnected", c.SessionID)
	}
	if len(messageIDs) == 0 {
		return nil
	}

	chat, err := parseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID (%s): %w", chatJID, err)
	}

	var sender types.JID
	if senderJID != "" {
		sJID, err := parseJID(senderJID)
		if err != nil {
			return fmt.Errorf("invalid sender JID (%s): %w", senderJID, err)
		}
		sender = sJID
	}

	typesMsgIDs := make([]types.MessageID, len(messageIDs))
	for i, id := range messageIDs {
		typesMsgIDs[i] = types.MessageID(id)
	}

	if err := c.waClient.MarkRead(ctx, typesMsgIDs, time.Now(), chat, sender); err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	if c.chatRepo != nil {
		_ = c.chatRepo.MarkMessagesStatus(ctx, c.SessionID, chat.String(), messageIDs, "read")
		_ = c.chatRepo.MarkChatRead(ctx, c.SessionID, chat.String())
	}
	return nil
}

func parseJID(target string) (types.JID, error) {
	target = strings.TrimSpace(target)
	if strings.Contains(target, "@") {
		return types.ParseJID(target)
	}

	cleanNum := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, target)

	if strings.HasPrefix(cleanNum, "0") {
		cleanNum = "62" + cleanNum[1:]
	}

	return types.NewJID(cleanNum, types.DefaultUserServer), nil
}
