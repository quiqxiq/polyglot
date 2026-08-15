package whatsapp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/quixiq/polyglot/pkg/logger"
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
	logWrapper := waLog.Stdout("WhatsApp", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, logWrapper)
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

		qrCtx, qrCancel := context.WithTimeout(context.Background(), 3*time.Minute)

		qrChan, err := c.waClient.GetQRChannel(qrCtx)
		if err != nil {
			qrCancel()
			if errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
				logger.WithField("session_id", c.SessionID).Info("Store already contains ID, attempting direct connect")
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
	logger.WithField("session_id", c.SessionID).Info("Manual reconnect initiated")
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
