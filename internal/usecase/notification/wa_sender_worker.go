package notification

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// notificationRepo menggabungkan antarmuka baca/queue dan varian retry.
type notificationRepo interface {
	port.NotificationRepository
	port.NotificationRetryRepository
}

// WASenderWorker mengirim antrean wa_notifications QUEUED via
// port.NotificationSender. Percobaan gagal menaikkan attempts; melewati
// maxRetry dianggap gagal permanen (tidak dipoll lagi).
type WASenderWorker struct {
	notif    notificationRepo
	sender   port.NotificationSender
	settings port.SettingReader

	now func() time.Time
}

// NewWASenderWorker wires dependencies.
func NewWASenderWorker(
	notif notificationRepo,
	sender port.NotificationSender,
	settings port.SettingReader,
) *WASenderWorker {
	return &WASenderWorker{notif: notif, sender: sender, settings: settings, now: time.Now}
}

// SendResult rekap satu siklus pengiriman.
type SendResult struct {
	Sent      int
	Failed    int
	GaveUp    int // melebihi max retry
	NoSession bool
}

// Run executes one send pass.
func (w *WASenderWorker) Run(ctx context.Context) (SendResult, error) {
	res := SendResult{}
	maxRetry := atoiDefault(w.settings.GetValue(ctx, "isp.wa_send_max_retry", "3"), 3)

	pending, err := w.notif.PendingWithRetryLimit(ctx, 50, maxRetry)
	if err != nil {
		return res, fmt.Errorf("load pending notifications: %w", err)
	}
	for _, n := range pending {
		sendErr := w.sender.Send(ctx, n.RecipientPhone, n.MessageContent)
		if sendErr == nil {
			if err := w.notif.MarkSent(ctx, n.ID, w.now()); err != nil {
				return res, fmt.Errorf("mark notification %s sent: %w", n.ID, err)
			}
			res.Sent++
			continue
		}
		logger.WithComponent("WaSender").WithError(sendErr).WithFields(map[string]any{
			"notification_id": n.ID,
		}).Warn("kirim WA gagal")
		if errors.Is(sendErr, domainNotification.ErrNoWASession) {
			res.NoSession = true
			return res, nil // infrastruktur mati: jangan buang attempts
		}
		newAttempts := n.Attempts + 1
		if newAttempts >= maxRetry {
			_ = w.notif.MarkFailedWithAttempt(ctx, n.ID,
				"gave up after max retries: "+sendErr.Error(), newAttempts)
			res.GaveUp++
			continue
		}
		_ = w.notif.MarkFailedWithAttempt(ctx, n.ID,
			"attempt "+itoa(newAttempts)+": "+sendErr.Error(), newAttempts)
		res.Failed++
	}
	return res, nil
}

func atoiDefault(s string, def int) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if n == 0 {
		return def
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
