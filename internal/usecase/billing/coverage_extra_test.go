package billing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func TestCheckout_Guards(t *testing.T) {
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	usecase := uc.NewCheckoutUseCase(invoices, customers, &mocktest.FakePaymentProcessor{})
	ctx := context.Background()

	_, err := usecase.ResolveByPaymentCode(ctx, "")
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)

	_, err = usecase.ResolveByQR(ctx, "")
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)

	_, err = usecase.ResolveByPortalCode(ctx, "")
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)
}

func TestIsolateWorker_TemplateRendered(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	isolator := mocktest.NewFakeRouterAccountManager()
	notif := mocktest.NewFakeNotificationRepo()

	notif.SeedTemplate(domainNotification.NotificationTemplate{
		ID: "nt-iso", TemplateKey: "ISOLATION_NOTICE", Name: "Isolir",
		Content:  "Halo {{customer_name}}, layanan diisolir.",
		IsActive: true,
	})

	deviceID := "dev-tpl"
	sub := seedActiveSub(t, subs, "sub-tpl", nil)
	sub.DeviceID = &deviceID
	sub.RemoteUsername = "TPLUSER"
	sub.ServiceType = "PPPOE"
	require.NoError(t, subs.Save(context.Background(), sub))
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-tpl", sub.CustomerID, sub.ID, -30)))
	cust := customerWithPortal(sub.CustomerID, "22222222")
	require.NoError(t, customers.Save(context.Background(), cust))

	worker := uc.NewIsolateWorker(subs, invoices, customers,
		mocktest.NewFakeServicePlanRepo(), isolator, notif, defaultSettings(nil))
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Isolated)

	queued := notif.Queued()
	require.Len(t, queued, 1)
	assert.Contains(t, queued[0].MessageContent, "Halo Pelanggan "+sub.CustomerID)
	assert.Equal(t, domainNotification.StatusQueued, queued[0].Status)
	_ = domainSubscription.StatusActive
}
