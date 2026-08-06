package whatsapp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type MessageCallback func(sessionID uint, customerNumber string, content string)
type StatusCallback func(sessionID uint, status string, qrCode string, jid string, phoneNumber string)

func init() {
	store.SetOSInfo("Polyglot NetOps WA Bot", [3]uint32{2, 3000, 1015901307})
}

type Client struct {
	SessionID   uint
	deviceStore *store.Device
	waClient    *whatsmeow.Client
	qrCode      string
	qrBase64    string
	qrMutex     sync.RWMutex
	onMessage   MessageCallback
	onStatus    StatusCallback
}

func NewClient(sessionID uint, deviceStore *store.Device, onMsg MessageCallback, onStat StatusCallback) *Client {
	logger := waLog.Stdout("WhatsApp", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, logger)
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true

	c := &Client{
		SessionID:   sessionID,
		deviceStore: deviceStore,
		waClient:    client,
		onMessage:   onMsg,
		onStatus:    onStat,
	}

	client.AddEventHandler(c.handleEvent)
	return c
}

func (c *Client) Connect(ctx context.Context) error {
	if c.waClient.IsConnected() {
		return nil
	}

	if c.waClient.Store.ID == nil {
		c.waClient.Disconnect()

		// Detached 3-minute context for QR watcher so HTTP request timeouts don't cancel QR polling
		qrCtx, qrCancel := context.WithTimeout(context.Background(), 3*time.Minute)

		qrChan, err := c.waClient.GetQRChannel(qrCtx)
		if err != nil {
			qrCancel()
			if errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
				log.Printf("[WhatsApp Client %d] Store already contains ID, attempting direct connect...", c.SessionID)
				if err := c.waClient.Connect(); err != nil {
					return fmt.Errorf("connect with saved session failed: %w", err)
				}
				return nil
			}
			return fmt.Errorf("failed to get QR channel: %w", err)
		}

		if err := c.waClient.Connect(); err != nil {
			qrCancel()
			return fmt.Errorf("failed to connect WA client: %w", err)
		}

		go func() {
			defer qrCancel()
			for evt := range qrChan {
				if evt.Event == "code" {
					c.qrMutex.Lock()
					c.qrCode = evt.Code
					if pngBytes, err := qrcode.Encode(evt.Code, qrcode.Medium, 256); err == nil {
						c.qrBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
					}
					c.qrMutex.Unlock()

					if c.onStatus != nil {
						c.onStatus(c.SessionID, "needs_rescan", c.qrBase64, "", "")
					}
				} else {
					log.Printf("[WhatsApp Client %d] QR event: %s", c.SessionID, evt.Event)
				}
			}
		}()
	} else {
		if err := c.waClient.Connect(); err != nil {
			return fmt.Errorf("failed to connect WA client: %w", err)
		}
		if c.onStatus != nil {
			userJID := ""
			phoneNum := ""
			if c.waClient.Store.ID != nil {
				userJID = c.waClient.Store.ID.String()
				phoneNum = c.waClient.Store.ID.User
			}
			c.onStatus(c.SessionID, "online", "", userJID, phoneNum)
		}
	}

	return nil
}

func (c *Client) Disconnect() {
	if c.waClient.IsConnected() {
		c.waClient.Disconnect()
	}
}

func (c *Client) Logout(ctx context.Context) error {
	if c.waClient.IsConnected() {
		err := c.waClient.Logout(ctx)
		c.Disconnect()
		return err
	}
	c.Disconnect()
	return nil
}

func (c *Client) Reconnect(ctx context.Context) error {
	log.Printf("[WhatsApp Client %d] Manual reconnect initiated...", c.SessionID)
	c.Disconnect()
	time.Sleep(1 * time.Second)
	return c.Connect(ctx)
}

func (c *Client) GetQRCode() string {
	c.qrMutex.RLock()
	defer c.qrMutex.RUnlock()
	if c.qrBase64 != "" {
		return c.qrBase64
	}
	return c.qrCode
}

func (c *Client) GetPairingCode(ctx context.Context, phoneNumber string) (string, error) {
	if !c.waClient.IsConnected() {
		if err := c.waClient.Connect(); err != nil {
			return "", fmt.Errorf("failed to connect for pairing: %w", err)
		}
	}
	code, err := c.waClient.PairPhone(ctx, phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Polyglot NetOps)")
	if err != nil {
		return "", fmt.Errorf("failed to request pairing code: %w", err)
	}
	return code, nil
}

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

func (c *Client) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		c.handleIncomingMessage(v)
	case *events.Connected:
		log.Printf("[WhatsApp Client %d] Connected successfully (Auto-reconnect active)", c.SessionID)
		if c.onStatus != nil && c.waClient.Store.ID != nil {
			c.onStatus(c.SessionID, "online", "", c.waClient.Store.ID.String(), c.waClient.Store.ID.User)
		}
	case *events.Disconnected:
		log.Printf("[WhatsApp Client %d] Disconnected (Auto-reconnect will retry)", c.SessionID)
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "offline", "", "", "")
		}
	case *events.LoggedOut:
		log.Printf("[WhatsApp Client %d] Remote LoggedOut event received from phone", c.SessionID)
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

	log.Printf("[WhatsApp Client %d] Message from %s: %s", c.SessionID, senderJID, body)

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
