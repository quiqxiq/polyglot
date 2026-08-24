package port

import (
	"context"
	"errors"
	"time"

	"github.com/quixiq/polyglot/internal/domain/notification"
)

// NotificationRepository defines persistence operations for WhatsApp
// notification templates and the send queue/log
// (DATABASE-SCHEMA-ISP.md §2.10). Digabung per area.
type NotificationRepository interface {
	// Templates
	SaveTemplate(ctx context.Context, t notification.NotificationTemplate) error
	FindTemplateByKey(ctx context.Context, tenantID, key string) (notification.NotificationTemplate, error)
	ListTemplates(ctx context.Context, activeOnly bool) ([]notification.NotificationTemplate, error)

	// Queue / log
	Queue(ctx context.Context, n notification.WANotification) error
	FindByID(ctx context.Context, id string) (notification.WANotification, error)
	// Pending returns QUEUED notifications oldest-first for the worker.
	Pending(ctx context.Context, limit int) ([]notification.WANotification, error)
	MarkSent(ctx context.Context, id string, sentAt time.Time) error
	MarkFailed(ctx context.Context, id string, errMsg string) error
	ListByCustomer(ctx context.Context, customerID string, limit int) ([]notification.WANotification, error)
}

// ErrNoWASession menandai tidak ada sesi WhatsApp terhubung — worker
// memperlakukan ini sebagai infrastruktur mati (tidak membakar attempts).
var ErrNoWASession = errors.New("no connected whatsapp session")

// Varian sadar-retry untuk worker pengirim.
type NotificationRetryRepository interface {
	// PendingWithRetryLimit mengembalikan QUEUED dengan attempts < maxAttempts.
	PendingWithRetryLimit(ctx context.Context, limit, maxAttempts int) ([]notification.WANotification, error)
	// MarkFailedWithAttempt menyimpan pesan gagal + jumlah attempts terbaru.
	MarkFailedWithAttempt(ctx context.Context, id, errMsg string, attempts int) error
}
