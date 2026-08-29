// Package whatsappadapter menjembatani port.NotificationSender ke
// SessionManager WhatsApp yang sudah berjalan (pilih session CONNECTED).
package whatsappadapter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/driver/whatsapp"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/phone"
)

var ErrNoConnectedSession = errors.New("no connected whatsapp session")

type SenderAdapter struct {
	sessions *whatsapp.SessionManager
	list     func(ctx context.Context) ([]bot.WASession, error)
}

var _ port.NotificationSender = (*SenderAdapter)(nil)

// NewSenderAdapter constructs the adapter. list biasanya pgStore.FindAllSessions.
func NewSenderAdapter(sessions *whatsapp.SessionManager, list func(ctx context.Context) ([]bot.WASession, error)) *SenderAdapter {
	return &SenderAdapter{sessions: sessions, list: list}
}

// Send implements port.NotificationSender: kirim pesan via session WhatsApp
// pertama yang berstatus terhubung. Nomor dinormalisasi ke format 62.
func (a *SenderAdapter) Send(ctx context.Context, rawPhone, content string) error {
	to := phone.Normalize(rawPhone)
	if to == "" {
		return fmt.Errorf("invalid destination number: %q", rawPhone)
	}
	sessionID, err := a.connectedSessionID(ctx)
	if err != nil {
		return err
	}
	if err := a.sessions.SendMessageContext(ctx, sessionID, to, content); err != nil {
		return fmt.Errorf("kirim WA ke %s via session %d: %w", to, sessionID, err)
	}
	return nil
}

func (a *SenderAdapter) connectedSessionID(ctx context.Context) (uint, error) {
	sessions, err := a.list(ctx)
	if err != nil {
		return 0, fmt.Errorf("list sessions: %w", err)
	}
	for _, s := range sessions {
		if isConnectedStatus(string(s.Status)) {
			return s.ID, nil
		}
	}
	return 0, ErrNoConnectedSession
}

func isConnectedStatus(status string) bool {
	switch strings.ToLower(status) {
	case string(bot.StatusOnline):
		return true
	default:
		return false
	}
}
