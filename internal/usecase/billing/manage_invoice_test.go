// F3-6: InvoiceUseCase.CancelInvoice memakai canceller atomik + default
// akun/kategori koreksi kas.
package billing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func TestCancelInvoice_UsesAtomicCanceller(t *testing.T) {
	invoices := mocktest.NewFakeInvoiceRepo()
	canceller := mocktest.NewFakeInvoiceCanceller(invoices)
	usecase := uc.NewInvoiceUseCase(invoices).WithCanceller(canceller)
	ctx := context.Background()

	inv := unpaidInvoice("inv-cancel", "cust-1", "sub-1", 5)
	inv.Status = domainBilling.StatusPartial
	inv.PaidAmount = 50000
	require.NoError(t, invoices.Save(ctx, inv))

	cancelled, err := usecase.CancelInvoice(ctx, "inv-cancel", "salah tagih")
	require.NoError(t, err)
	assert.Equal(t, domainBilling.StatusCancelled, cancelled.Status)

	require.Len(t, canceller.Commands, 1)
	assert.Equal(t, "ca-1001-kas-kantor", canceller.Commands[0].CashAccountID)
	assert.Equal(t, "cc-refund", canceller.Commands[0].ExpenseCategoryID)
	assert.Equal(t, "salah tagih", canceller.Commands[0].Reason)

	// Invoice lunas tidak boleh dibatalkan.
	paid := unpaidInvoice("inv-paid-c", "cust-1", "sub-1", 5)
	paid.Status = domainBilling.StatusPaid
	require.NoError(t, invoices.Save(ctx, paid))
	_, err = usecase.CancelInvoice(ctx, "inv-paid-c", "x")
	assert.ErrorIs(t, err, domainBilling.ErrInvoiceAlreadyPaid)
}
