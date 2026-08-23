package billing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func unpaidInvoice(id, customerID, subID string, dueDaysFromNow int) domainBilling.Invoice {
	subRef := subID
	return domainBilling.Invoice{
		ID:                id,
		CustomerID:        customerID,
		SubscriptionID:    &subRef,
		Period:            "2026-07",
		Subtotal:          100000,
		Total:             110000,
		PaidAmount:        0,
		Status:            domainBilling.StatusUnpaid,
		DueDate:           timeAfter(dueDaysFromNow),
		QRPayload:         "polyglot://invoice/" + id,
		ManualPaymentCode: "PAY-" + id,
	}
}

func TestCheckout_ResolvePaths(t *testing.T) {
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-1", "cust-1", "sub-1", -5)))
	require.NoError(t, customers.Save(context.Background(), customerWithPortal("cust-1", "12345678")))

	usecase := uc.NewCheckoutUseCase(invoices, customers, &mocktest.FakePaymentProcessor{})
	ctx := context.Background()

	seeded, err := invoices.FindByID(ctx, "inv-1")
	require.NoError(t, err)

	got, err := invoices.FindByPaymentCode(ctx, seeded.ManualPaymentCode)
	require.NoError(t, err)
	resolved, err := usecase.ResolveByPaymentCode(ctx, got.ManualPaymentCode)
	require.NoError(t, err)
	assert.Equal(t, "inv-1", resolved.ID)

	resolvedQR, err := usecase.ResolveByQR(ctx, got.QRPayload)
	require.NoError(t, err)
	assert.Equal(t, "inv-1", resolvedQR.ID)

	resolvedPortal, err := usecase.ResolveByPortalCode(ctx, "12345678")
	require.NoError(t, err)
	assert.Equal(t, "inv-1", resolvedPortal.ID)

	_, err = usecase.ResolveByPortalCode(ctx, "00000000")
	assert.ErrorContains(t, err, "portal code")
}

func TestPayCash_DelegatesCommandIntact(t *testing.T) {
	invoices := mocktest.NewFakeInvoiceRepo()
	proc := &mocktest.FakePaymentProcessor{
		Pay: domainBilling.Payment{ID: "pay-x", Amount: 50000},
	}
	cashier := uint(9)
	cmd := port.CashPaymentCommand{
		TenantID: "tenant-default", InvoiceID: "inv-1", Amount: 50000,
		CashAccountID: "ca-1001", IncomeCategoryID: "cc-tagihan",
		ReceivedBy: &cashier, ScanMethod: domainBilling.ScanManual,
	}
	usecase := uc.NewCheckoutUseCase(invoices, mocktest.NewFakeCustomerRepo(), proc)

	pay, err := usecase.PayCash(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "pay-x", pay.ID)
	require.Len(t, proc.Cmds, 1)
	assert.Equal(t, cmd.InvoiceID, proc.Cmds[0].InvoiceID)
	assert.InDelta(t, 50000, proc.Cmds[0].Amount, 0.01)
}
