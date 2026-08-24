package billing_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func gatewayFixture(t *testing.T) (*mocktest.FakeInvoiceRepo, *mocktest.FakeCustomerRepo,
	*mocktest.FakeGatewayTxRepo, *mocktest.FakePaymentGateway, *FakeProcessorRecorder) {
	t.Helper()
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	gwt := mocktest.NewFakeGatewayTxRepo()
	gateway := &mocktest.FakePaymentGateway{
		IsEnabled: true,
		Charge: port.ChargeResult{
			ExternalID: "INV-GW-1", PaymentURL: "https://pay/x",
			QRString: "qrcode", FeeAmount: 700, Status: "PENDING",
			RawResponse: json.RawMessage(`{"success":true}`),
		},
		Event: port.WebhookEvent{
			ExternalID: "INV-GW-1", MerchantRef: "INV-GW-1",
			Status: "SETTLED", PaidAmount: 110000,
			Raw: json.RawMessage(`{"status":"PAID"}`),
		},
	}
	proc := &FakeProcessorRecorder{}
	return invoices, customers, gwt, gateway, proc
}

// FakeProcessorRecorder memakai processor sqlite nyata via helper lain;
// untuk unit ini cukup fake yang merekam command.
type FakeProcessorRecorder struct {
	Cmds []port.CashPaymentCommand
	Pay  domainBilling.Payment
	Err  error
}

func (f *FakeProcessorRecorder) ProcessCashPayment(_ context.Context, cmd port.CashPaymentCommand) (domainBilling.Payment, error) {
	f.Cmds = append(f.Cmds, cmd)
	if f.Err != nil {
		return domainBilling.Payment{}, f.Err
	}
	return domainBilling.Payment{ID: "pay-gw-1"}, nil
}

func TestCreateForInvoice_Happy(t *testing.T) {
	invoices, customers, gwt, gateway, _ := gatewayFixture(t)
	usecase := uc.NewGatewayChargeUseCase(invoices, customers, gwt, gateway,
		&FakeProcessorRecorder{}, mocktest.NewFakeSettingReader(nil))

	subID := "sub-gw"
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-gw", "cust-gw", subID, 5)))
	require.NoError(t, customers.Save(context.Background(), customerWithPortal("cust-gw", "99999999")))

	res, tx, err := usecase.CreateForInvoice(context.Background(), "inv-gw", "", 60)
	require.NoError(t, err)
	assert.Equal(t, "https://pay/x", res.PaymentURL)
	assert.Equal(t, "PENDING", tx.Status)
	assert.NotNil(t, tx.InvoiceID)

	stored := gwt.Get(tx.ID)
	assert.Equal(t, "FAKE", stored.Gateway)
	assert.InDelta(t, 110000, stored.Amount, 0.01)
}

func TestWebhook_Settled_InvokesProcessor(t *testing.T) {
	invoices, customers, gwt, gateway, proc := gatewayFixture(t)
	reader := mocktest.NewFakeSettingReader(map[string]string{
		"gw.tripay.cash_account_id":    "ca-1",
		"gw.tripay.income_category_id": "cc-1",
	})
	usecase := uc.NewGatewayChargeUseCase(invoices, customers, gwt, gateway, proc, reader)

	subID := "sub-gw"
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-gw", "cust-gw", subID, 5)))
	require.NoError(t, customers.Save(context.Background(), customerWithPortal("cust-gw", "99999999")))

	_, _, err := usecase.CreateForInvoice(context.Background(), "inv-gw", "", 60)
	require.NoError(t, err)

	body := []byte(`{"reference":"INV-GW-1","merchant_ref":"INV-GW-1","status":"PAID","total_amount":110000}`)
	invoiceID, settled, err := usecase.HandleWebhook(context.Background(), body, "valid")
	require.NoError(t, err)
	assert.True(t, settled)
	assert.Equal(t, "inv-gw", invoiceID)

	require.Len(t, proc.Cmds, 1)
	assert.Equal(t, "inv-gw", proc.Cmds[0].InvoiceID)
	assert.Equal(t, domainBilling.ScanPaymentGateway, proc.Cmds[0].ScanMethod)
	assert.Equal(t, "ca-1", proc.Cmds[0].CashAccountID)

	stored := gwt.Get(firstGwtID(gwt))
	assert.Equal(t, domainBilling.GatewayStatusSettled, stored.Status)
	require.NotNil(t, stored.PaymentID)
}

func TestWebhook_BadSignature_AndUnknownRef(t *testing.T) {
	invoices, customers, gwt, gateway, proc := gatewayFixture(t)
	reader := mocktest.NewFakeSettingReader(nil)
	usecase := uc.NewGatewayChargeUseCase(invoices, customers, gwt, gateway, proc, reader)

	parseErr := errors.New("parse gagal")
	gateway.ParseErr = parseErr
	_, _, err := usecase.HandleWebhook(context.Background(), []byte("{}"), "sig")
	assert.ErrorIs(t, err, parseErr)

	gateway.ParseErr = nil
	_, _, err = usecase.HandleWebhook(context.Background(),
		[]byte(`{"reference":"UNKNOWN","merchant_ref":"U","status":"PAID","total_amount":100}`), "sig")
	assert.ErrorContains(t, err, "tidak dikenal")
}

func TestCreateForInvoice_AlreadyPaid(t *testing.T) {
	invoices, customers, gwt, gateway, proc := gatewayFixture(t)
	usecase := uc.NewGatewayChargeUseCase(invoices, customers, gwt, gateway, proc,
		mocktest.NewFakeSettingReader(nil))

	paid := unpaidInvoice("inv-paid", "cust-p", "sub-p", 3)
	paid.Status = domainBilling.StatusPaid
	require.NoError(t, invoices.Save(context.Background(), paid))

	_, _, err := usecase.CreateForInvoice(context.Background(), "inv-paid", "", 60)
	assert.ErrorIs(t, err, port.ErrInvoiceAlreadyPaid)
}

func TestIsolateWorker_ProvisionStatusLifecycle(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	isolator := mocktest.NewFakeRouterAccountManager()

	plans.Seed(newPlan("plan-lc", "PLAN-LC"))
	deviceID := "dev-lc"
	sub := seedActiveSub(t, subs, "sub-lcx", nil)
	sub.PlanID = "plan-lc"
	sub.DeviceID = &deviceID
	sub.RemoteUsername = "LCX"
	sub.RemotePassword = "pw"
	sub.ProvisionStatus = domainSubscription.ProvisionFailed // worker harus mencoba lagi
	require.NoError(t, subs.Save(context.Background(), sub))

	settings := defaultSettings(nil)
	worker := uc.NewIsolateWorker(subs, mocktest.NewFakeInvoiceRepo(),
		mocktest.NewFakeCustomerRepo(), plans, isolator,
		mocktest.NewFakeNotificationRepo(), settings)
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Provisioned)

	got, _ := subs.FindByID(context.Background(), sub.ID)
	assert.Equal(t, domainSubscription.ProvisionOK, got.ProvisionStatus)
}
