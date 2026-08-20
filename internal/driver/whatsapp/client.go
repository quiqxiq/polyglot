package whatsapp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// MessageCallback meneruskan pesan masuk ke handler (Engine bot).
type MessageCallback func(sessionID uint, chatJID string, customerNumber string, content string)
type StatusCallback func(sessionID uint, status string, qrCode string, jid string, phoneNumber string)

// ChatUpdateCallback memberitahukan bahwa mirror chat berubah (pesan masuk/keluar).
type ChatUpdateCallback func(sessionID uint, chatJID string)

// ChatPresenceCallback memberitahukan event typing/recording dari kontak.
type ChatPresenceCallback func(sessionID uint, chatJID, senderJID, state, media string, isGroup bool)

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
	callbackMu  sync.RWMutex
	onStatus    StatusCallback
	onChatUpd   ChatUpdateCallback
	onChatPres  ChatPresenceCallback
	chatRepo    port.ChatRepository

	qrFlowMu     sync.Mutex
	qrFlowActive bool
}

func NewClient(sessionID uint, deviceStore *store.Device, onMsg MessageCallback, onStat StatusCallback, chatRepo port.ChatRepository) *Client {
	waLogger := waLog.Stdout("WhatsApp", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, waLogger)
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true

	c := &Client{
		SessionID:   sessionID,
		deviceStore: deviceStore,
		waClient:    client,
		onMessage:   onMsg,
		onStatus:    onStat,
		chatRepo:    chatRepo,
	}

	client.AddEventHandler(c.handleEvent)
	return c
}

// SetMessageCallback swaps the incoming-message handler at runtime.
func (c *Client) SetMessageCallback(cb MessageCallback) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onMessage = cb
}

func (c *Client) getMessageCallback() MessageCallback {
	c.callbackMu.RLock()
	defer c.callbackMu.RUnlock()
	return c.onMessage
}

// SetChatUpdateCallback swaps the chat-mirror-change handler at runtime.
func (c *Client) SetChatUpdateCallback(cb ChatUpdateCallback) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onChatUpd = cb
}

func (c *Client) getChatUpdateCallback() ChatUpdateCallback {
	c.callbackMu.RLock()
	defer c.callbackMu.RUnlock()
	return c.onChatUpd
}

func (c *Client) notifyChatUpdate(chatJID string) {
	if cb := c.getChatUpdateCallback(); cb != nil {
		cb(c.SessionID, chatJID)
	}
}

// SetChatPresenceCallback swaps the chat-presence handler at runtime.
func (c *Client) SetChatPresenceCallback(cb ChatPresenceCallback) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onChatPres = cb
}

func (c *Client) getChatPresenceCallback() ChatPresenceCallback {
	c.callbackMu.RLock()
	defer c.callbackMu.RUnlock()
	return c.onChatPres
}

func (c *Client) Connect(ctx context.Context) error {
	if c.waClient.IsConnected() {
		if c.waClient.Store.ID == nil {
			c.EnsureQRFlow()
		}
		return nil
	}

	if c.waClient.Store.ID == nil {
		c.startQRFlow()
		return nil
	}

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

	return nil
}

func (c *Client) Disconnect() {
	if c.waClient.IsConnected() {
		c.waClient.Disconnect()
	}
}

func (c *Client) Logout(ctx context.Context) error {
	if c.waClient.IsConnected() {
		if err := c.waClient.Logout(ctx); err != nil {
			logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Warnf("remote logout failed (best-effort): %v", err)
		}
	}
	c.Disconnect()
	if c.deviceStore != nil {
		_ = c.deviceStore.Delete(ctx)
	}
	return nil
}

func (c *Client) Reconnect(ctx context.Context) error {
	logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Info("manual reconnect initiated")
	c.Disconnect()
	time.Sleep(1 * time.Second)
	return c.Connect(ctx)
}
