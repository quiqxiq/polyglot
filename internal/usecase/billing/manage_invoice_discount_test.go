package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func seedInvoice(t *testing.T, repo *mocktest.FakeInvoiceRepo, status string) {
	t.Helper()
	require.NoError(t, repo.Save(context.Background(), domainBilling.Invoice{
		ID: "inv-d", TenantID: "tenant-default", CustomerID: "cust-1",
		InvoiceNumber: "INV-D", Period: "2026-09",
		Subtotal: 100000, TaxAmount: 11000, Total: 111000,
		Status: status, DueDate: time.Now(),
	}))
}

func TestApplyInvoiceDiscount_Percent(t *testing.T) {
	repo := mocktest.NewFakeInvoiceRepo()
	seedInvoice(t, repo, domainBilling.StatusUnpaid)
	useCase := uc.NewInvoiceUseCase(repo)

	inv, err := useCase.ApplyDiscount(context.Background(), "inv-d", "PERCENT", 10, "PROMO10")
	require.NoError(t, err)
	assert.Equal(t, "PERCENT", inv.DiscountType)
	assert.InDelta(t, 10, inv.DiscountValue, 0.01)
	assert.InDelta(t, 10000, inv.DiscountAmount, 0.01)
	assert.InDelta(t, 10000, inv.Discount, 0.01)
	assert.Equal(t, "PROMO10", inv.PromoCode)
	assert.InDelta(t, 101000, inv.Total, 0.01) // 100000 - 10000 + 11000
}

func TestApplyInvoiceDiscount_AmountAndReset(t *testing.T) {
	repo := mocktest.NewFakeInvoiceRepo()
	seedInvoice(t, repo, domainBilling.StatusPartial)
	useCase := uc.NewInvoiceUseCase(repo)

	inv, err := useCase.ApplyDiscount(context.Background(), "inv-d", "AMOUNT", 25000, "")
	require.NoError(t, err)
	assert.InDelta(t, 25000, inv.DiscountAmount, 0.01)
	assert.InDelta(t, 86000, inv.Total, 0.01) // 100000-25000+11000

	reset, err := useCase.ApplyDiscount(context.Background(), "inv-d", "NONE", 0, "")
	require.NoError(t, err)
	assert.InDelta(t, 0, reset.Discount, 0.01)
	assert.InDelta(t, 111000, reset.Total, 0.01)
}

func TestApplyInvoiceDiscount_Validation(t *testing.T) {
	repo := mocktest.NewFakeInvoiceRepo()
	seedInvoice(t, repo, domainBilling.StatusUnpaid)
	useCase := uc.NewInvoiceUseCase(repo)

	_, err := useCase.ApplyDiscount(context.Background(), "inv-d", "PERCENT", 0, "")
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)

	_, err = useCase.ApplyDiscount(context.Background(), "inv-d", "SALAH", 10, "")
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)

	paidRepo := mocktest.NewFakeInvoiceRepo()
	seedInvoice(t, paidRepo, domainBilling.StatusPaid)
	paidUC := uc.NewInvoiceUseCase(paidRepo)
	_, err = paidUC.ApplyDiscount(context.Background(), "inv-d", "AMOUNT", 1000, "")
	assert.ErrorIs(t, err, domainBilling.ErrInvalidTransition)
}
