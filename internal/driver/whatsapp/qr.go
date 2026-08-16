package whatsapp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"

	"github.com/quixiq/polyglot/pkg/logger"
)

// EnsureQRFlow memastikan aliran QR aktif untuk session yang belum ter-pair.
func (c *Client) EnsureQRFlow() {
	if c.waClient.Store.ID != nil {
		return
	}
	c.startQRFlow()
}

// startQRFlow membuka channel QR dan menyalakan goroutine watcher.
func (c *Client) startQRFlow() {
	c.qrFlowMu.Lock()
	if c.qrFlowActive {
		c.qrFlowMu.Unlock()
		return
	}
	c.qrFlowActive = true
	c.qrFlowMu.Unlock()

	if c.onStatus != nil {
		c.onStatus(c.SessionID, "connecting", "", "", "")
	}
	c.waClient.Disconnect()

	qrCtx, qrCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	qrChan, err := c.waClient.GetQRChannel(qrCtx)
	if err != nil {
		qrCancel()
		c.qrFlowMu.Lock()
		c.qrFlowActive = false
		c.qrFlowMu.Unlock()
		if errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
			logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Info("store already contains ID, attempting direct connect")
			if err := c.waClient.Connect(); err != nil {
				logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Errorf("connect with saved session failed: %v", err)
			}
			return
		}
		logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Errorf("failed to get QR channel: %v", err)
		if c.onStatus != nil {
			c.onStatus(c.SessionID, "needs_rescan", "", "", "")
		}
		return
	}

	if !c.waClient.IsConnected() {
		if err := c.waClient.Connect(); err != nil {
			qrCancel()
			c.qrFlowMu.Lock()
			c.qrFlowActive = false
			c.qrFlowMu.Unlock()
			if c.onStatus != nil {
				c.onStatus(c.SessionID, "needs_rescan", "", "", "")
			}
			logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Errorf("failed to connect WA client: %v", err)
			return
		}
	}

	go func() {
		defer qrCancel()
		defer func() {
			c.qrFlowMu.Lock()
			c.qrFlowActive = false
			c.qrFlowMu.Unlock()
		}()
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
			} else if evt.Event == "timeout" {
				logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Info("QR code expired")
				c.qrMutex.Lock()
				c.qrCode = ""
				c.qrBase64 = ""
				c.qrMutex.Unlock()
				if c.onStatus != nil {
					c.onStatus(c.SessionID, "needs_rescan", "", "", "")
				}
			} else {
				logger.WithComponent("WhatsAppDriver").WithField("session_id", c.SessionID).Debugf("QR event: %s", evt.Event)
			}
		}
	}()
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
