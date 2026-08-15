package whatsapp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

	_, err = c.waClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send WA message: %w", err)
	}

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

	_, err = c.waClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send WA document message: %w", err)
	}

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

	_, err = c.waClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send WA image message: %w", err)
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
