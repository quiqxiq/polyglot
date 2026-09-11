package billing

import (
	"context"
	"fmt"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/phone"
)

// ReminderResult rekap satu siklus pengingat tagihan.
type ReminderResult struct {
	Queued  int
	Skipped int // sudah dikirim hari ini atau di luar jadwal H-
}

// ReminderWorker mengirim pengingat tagihan WhatsApp pada H-7, H-3, H-1, dan
// hari jatuh tempo (BILL_REMINDER), idempoten per invoice per hari (F3-1).
type ReminderWorker struct {
	subs      port.SubscriptionRepository
	invoices  port.InvoiceRepository
	customers port.CustomerRepository
	notif     port.NotificationRepository

	now     func() time.Time
	offsets []int
}

// NewReminderWorker wires dependencies.
func NewReminderWorker(
	subs port.SubscriptionRepository,
	invoices port.InvoiceRepository,
	customers port.CustomerRepository,
	notif port.NotificationRepository,
) *ReminderWorker {
	return &ReminderWorker{
		subs: subs, invoices: invoices, customers: customers, notif: notif,
		now: time.Now, offsets: []int{7, 3, 1, 0},
	}
}

// WithClock mengganti sumber waktu (test deterministik).
func (w *ReminderWorker) WithClock(now func() time.Time) *ReminderWorker {
	w.now = now
	return w
}

// Run menjalankan satu siklus pengingat untuk semua langganan ACTIVE.
func (w *ReminderWorker) Run(ctx context.Context) (ReminderResult, error) {
	res := ReminderResult{}
	if w.notif == nil {
		return res, nil
	}
	active, err := w.subs.ListActive(ctx)
	if err != nil {
		return res, fmt.Errorf("list active subscriptions: %w", err)
	}
	now := w.now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, sub := range active {
		invoices, err := w.invoices.FindByCustomerID(ctx, sub.CustomerID)
		if err != nil {
			return res, fmt.Errorf("list invoices %s: %w", sub.CustomerID, err)
		}
		for _, inv := range invoices {
			if inv.SubscriptionID == nil || *inv.SubscriptionID != sub.ID {
				continue
			}
			if !isOutstandingInvoice(inv.Status) || inv.DueDate.Before(dayStart) {
				continue
			}
			daysLeft := int(inv.DueDate.Sub(dayStart).Hours() / 24)
			if !containsOffset(w.offsets, daysLeft) {
				continue
			}
			exists, err := w.notif.ExistsForInvoiceSince(ctx, inv.ID, "BILL_REMINDER", dayStart)
			if err != nil {
				return res, fmt.Errorf("check reminder %s: %w", inv.ID, err)
			}
			if exists {
				res.Skipped++
				continue
			}
			if w.queueReminder(ctx, sub, inv) {
				res.Queued++
			} else {
				res.Skipped++
			}
		}
	}
	return res, nil
}

func (w *ReminderWorker) queueReminder(ctx context.Context, sub domainSubscription.Subscription, inv domainBilling.Invoice) bool {
	cust, err := w.customers.FindByID(ctx, sub.CustomerID)
	if err != nil {
		logger.WithComponent("ReminderWorker").WithFields(map[string]any{
			"subscription_id": sub.ID,
		}).WithError(err).Warn("customer not found; reminder skipped")
		return false
	}
	content := renderTemplateOrDefault(ctx, w.notif, cust.TenantID, "BILL_REMINDER",
		"Yth {{customer_name}}, tagihan {{period}} sebesar Rp{{total}} jatuh tempo {{due_date}}. Kode bayar: {{payment_code}}.",
		map[string]string{
			"customer_name": cust.Name,
			"period":        inv.Period,
			"total":         fmt.Sprintf("%.0f", inv.Total),
			"due_date":      inv.DueDate.Format("02 Jan 2006"),
			"payment_code":  inv.ManualPaymentCode,
		})
	n := domainNotification.WANotification{
		ID:             idgen.New("wa"),
		TenantID:       cust.TenantID,
		CustomerID:     &cust.ID,
		InvoiceID:      &inv.ID,
		RecipientPhone: phone.Normalize(cust.Phone),
		MessageType:    "BILL_REMINDER",
		MessageContent: content,
		Status:         domainNotification.StatusQueued,
		CreatedAt:      time.Now(),
	}
	if err := w.notif.Queue(ctx, n); err != nil {
		logger.WithComponent("ReminderWorker").WithError(err).Warn("queue bill reminder failed")
		return false
	}
	return true
}

func containsOffset(offsets []int, v int) bool {
	for _, o := range offsets {
		if o == v {
			return true
		}
	}
	return false
}
