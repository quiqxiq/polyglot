package whatsapp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

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
	qrMutex     sync.RWMutex
	onMessage   MessageCallback
	onStatus    StatusCallback
}

func NewClient(sessionID uint, deviceStore *store.Device, onMsg MessageCallback, onStat StatusCallback) *Client {
	logger := waLog.Stdout("WhatsApp", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, logger)

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
		qrChan, err := c.waClient.GetQRChannel(ctx)
		if err != nil {
			return fmt.Errorf("failed to get QR channel: %w", err)
		}

		if err := c.waClient.Connect(); err != nil {
			return fmt.Errorf("failed to connect WA client: %w", err)
		}

		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					c.qrMutex.Lock()
					c.qrCode = evt.Code
					c.qrMutex.Unlock()

					if c.onStatus != nil {
						c.onStatus(c.SessionID, "needs_rescan", evt.Code, "", "")
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
		return c.waClient.Logout(ctx)
	}
	return nil
}

func (c *Client) Reconnect(ctx context.Context) error {
	c.Disconnect()
	time.Sleep(1 * time.Second)
	return c.Connect(ctx)
}

func (c *Client) GetQRCode() string {
	c.qrMutex.RLock()
	defer c.qrMutex.RUnlock()
	return c.qrCode
}

func (c *Client) GetPairingCode(ctx context.Context, phoneNumber string) (string, error) {
	if !c.waClient.IsConnected() {
		if err := c.waClient.Connect(); err != nil {
			return "", fmt.Errorf("failed to connect for pairing: %w", err)
		}
	}
	code, err := c.waClient.PairPhone(ctx, phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Polyglot)")
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

func (c *Client) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		c.handleIncomingMessage(v)
	case *events.Connected:
		log.Printf("[WhatsApp Client %d] Connected successfully", c.SessionID)
		if c.onStatus != nil && c.waClient.Store.ID != nil {
			c.onStatus(c.SessionID, "online", "", c.waClient.Store.ID.String(), c.waClient.Store.ID.User)
		}
	case *events.Disconnected:
		log.Printf("[WhatsApp Client %d] Disconnected", c.SessionID)
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "offline", "", "", "")
		}
	case *events.LoggedOut:
		log.Printf("[WhatsApp Client %d] Logged out", c.SessionID)
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
