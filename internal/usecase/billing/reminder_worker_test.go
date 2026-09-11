// F3-1: reminder tagihan H-7/H-3/H-1/hari-H, idempoten per hari per invoice.
package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func TestReminderWorker_ScheduleAndDedupe(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	notif := mocktest.NewFakeNotificationRepo()
	notif.SeedTemplate(domainNotification.NotificationTemplate{
		ID: "nt-bill", TenantID: "tenant-default", TemplateKey: "BILL_REMINDER",
		Name: "Reminder", Content: "Halo {{customer_name}}, total Rp{{total}} kode {{payment_code}}",
		IsActive: true,
	})

	base := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	clock := base
	worker := uc.NewReminderWorker(subs, invoices, customers, notif).
		WithClock(func() time.Time { return clock })

	sub := seedActiveSub(t, subs, "sub-rem", nil)
	require.NoError(t, customers.Save(context.Background(), customerWithPortal(sub.CustomerID, "87654321")))

	subRef := sub.ID
	inv := domainBilling.Invoice{
		ID: "inv-rem", CustomerID: sub.CustomerID, SubscriptionID: &subRef,
		Period: "2026-09", Total: 110000,
		Status: domainBilling.StatusUnpaid, DueDate: base.AddDate(0, 0, 7),
		QRPayload: "polyglot://invoice/inv-rem", ManualPaymentCode: "PAY-REM-1",
	}
	require.NoError(t, invoices.Save(context.Background(), inv))

	// Hari ini H-7 → reminder terkirim.
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Queued)

	queued := notif.Queued()
	require.Len(t, queued, 1)
	assert.Equal(t, "BILL_REMINDER", queued[0].MessageType)
	assert.Contains(t, queued[0].MessageContent, "Rp110000")
	assert.Contains(t, queued[0].MessageContent, "PAY-REM-1")

	// Run kedua di hari yang sama → idempoten.
	res2, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, res2.Queued)
	assert.Equal(t, 1, res2.Skipped)

	// H-6 bukan jadwal.
	clock = base.AddDate(0, 0, 1)
	res3, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, res3.Queued)

	// H-3 → terkirim lagi (hari berbeda).
	clock = base.AddDate(0, 0, 4)
	res4, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res4.Queued)
	assert.Len(t, notif.Queued(), 2)
}

func TestReminderWorker_SkipsPaidAndPastDue(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	notif := mocktest.NewFakeNotificationRepo()

	base := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	worker := uc.NewReminderWorker(subs, invoices, customers, notif).
		WithClock(func() time.Time { return base })

	sub := seedActiveSub(t, subs, "sub-skip", nil)
	require.NoError(t, customers.Save(context.Background(), customerWithPortal(sub.CustomerID, "87654321")))
	subRef := sub.ID

	paid := domainBilling.Invoice{
		ID: "inv-paid", CustomerID: sub.CustomerID, SubscriptionID: &subRef,
		Total: 110000, Status: domainBilling.StatusPaid, DueDate: base.AddDate(0, 0, 7),
	}
	late := domainBilling.Invoice{
		ID: "inv-late", CustomerID: sub.CustomerID, SubscriptionID: &subRef,
		Total: 110000, Status: domainBilling.StatusOverdue, DueDate: base.AddDate(0, 0, -2),
	}
	require.NoError(t, invoices.Save(context.Background(), paid))
	require.NoError(t, invoices.Save(context.Background(), late))

	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, res.Queued)
	assert.Empty(t, notif.Queued())
}
