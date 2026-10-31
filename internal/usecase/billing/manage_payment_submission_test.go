package billing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

type paymentFixture struct {
	invRepo   *mocktest.FakeInvoiceRepo
	subs      *mocktest.FakePaymentSubmissionRepo
	processor *mocktest.FakePaymentProcessor
	settings  *mocktest.FakeSettingReader
	useCase   *uc.PaymentSubmissionUseCase
}

func newPaymentFixture(t *testing.T) *paymentFixture {
	t.Helper()
	f := &paymentFixture{
		invRepo:   mocktest.NewFakeInvoiceRepo(),
		subs:      mocktest.NewFakePaymentSubmissionRepo(),
		processor: &mocktest.FakePaymentProcessor{Pay: domainBilling.Payment{ID: "pay-1", Amount: 100000}},
		settings: mocktest.NewFakeSettingReader(map[string]string{
			"isp.manual_payment_cash_account_id":    "ca-kas",
			"isp.manual_payment_income_category_id": "cc-tagihan",
			"isp.manual_payment_method_id":          "pm-bank",
		}),
	}
	f.useCase = uc.NewPaymentSubmissionUseCase(f.subs, f.invRepo, f.processor, f.settings).
		WithClock(func() time.Time { return time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC) })
	require.NoError(t, f.invRepo.Save(context.Background(), domainBilling.Invoice{
		ID: "inv-1", TenantID: "tenant-default", CustomerID: "cust-1",
		Period: "2026-09", Subtotal: 100000, Total: 100000, PaidAmount: 0,
		Status: domainBilling.StatusUnpaid, ManualPaymentCode: "PAY-1",
		QRPayload: "polyglot://invoice/inv-1", DueDate: time.Now(),
	}))
	return f
}

func validSubmissionInput() uc.SubmitManualPaymentInput {
	return uc.SubmitManualPaymentInput{
		ManualPaymentCode: "PAY-1",
		Amount:            100000,
		SenderName:        "Budi",
		SenderBank:        "BCA",
		ProofImageURL:     "/uploads/bukti/budi.jpg",
	}
}

func TestSubmitManualPayment_CreatesPending(t *testing.T) {
	f := newPaymentFixture(t)
	sub, err := f.useCase.Submit(context.Background(), validSubmissionInput())
	require.NoError(t, err)
	assert.Equal(t, domainBilling.PaymentSubmissionPending, sub.Status)
	assert.Equal(t, "inv-1", sub.InvoiceID)
	assert.Equal(t, "tenant-default", sub.TenantID)
	assert.InDelta(t, 100000, sub.Amount, 0.01)
	assert.NotEmpty(t, sub.ID)

	stored, err := f.subs.FindByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domainBilling.PaymentSubmissionPending, stored.Status)
}

func TestSubmitManualPayment_Validation(t *testing.T) {
	f := newPaymentFixture(t)

	// Kode bayar tidak dikenal.
	in := validSubmissionInput()
	in.ManualPaymentCode = "PAY-UNKNOWN"
	_, err := f.useCase.Submit(context.Background(), in)
	assert.ErrorIs(t, err, domainBilling.ErrNotFound)

	// Bukti wajib diisi.
	in = validSubmissionInput()
	in.ProofImageURL = " "
	_, err = f.useCase.Submit(context.Background(), in)
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)

	// Nominal melebihi sisa tagihan.
	in = validSubmissionInput()
	in.Amount = 150000
	_, err = f.useCase.Submit(context.Background(), in)
	assert.ErrorIs(t, err, domainBilling.ErrOverpayment)
}

func TestSubmitManualPayment_RejectsAlreadyPaidInvoice(t *testing.T) {
	f := newPaymentFixture(t)
	inv, err := f.invRepo.FindByPaymentCode(context.Background(), "PAY-1")
	require.NoError(t, err)
	inv.Status = domainBilling.StatusPaid
	inv.PaidAmount = inv.Total
	require.NoError(t, f.invRepo.Save(context.Background(), inv))

	_, err = f.useCase.Submit(context.Background(), validSubmissionInput())
	assert.ErrorIs(t, err, domainBilling.ErrInvalidTransition)
}

func TestApproveManualPayment_ProcessesAtomically(t *testing.T) {
	f := newPaymentFixture(t)
	sub, err := f.useCase.Submit(context.Background(), validSubmissionInput())
	require.NoError(t, err)
	reviewer := uint(7)

	approved, err := f.useCase.Approve(context.Background(), sub.ID, &reviewer)
	require.NoError(t, err)
	assert.Equal(t, domainBilling.PaymentSubmissionApproved, approved.Status)
	require.NotNil(t, approved.PaymentID)
	assert.Equal(t, "pay-1", *approved.PaymentID)
	require.NotNil(t, approved.ReviewedBy)
	assert.Equal(t, reviewer, *approved.ReviewedBy)

	require.Len(t, f.processor.Cmds, 1)
	cmd := f.processor.Cmds[0]
	assert.Equal(t, "inv-1", cmd.InvoiceID)
	assert.Equal(t, "ca-kas", cmd.CashAccountID)
	assert.Equal(t, "cc-tagihan", cmd.IncomeCategoryID)
	require.NotNil(t, cmd.PaymentMethodID)
	assert.Equal(t, "pm-bank", *cmd.PaymentMethodID)
	assert.InDelta(t, 100000, cmd.Amount, 0.01)
	assert.Equal(t, domainBilling.ScanManual, cmd.ScanMethod)

	stored, err := f.subs.FindByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domainBilling.PaymentSubmissionApproved, stored.Status)
}

func TestApproveManualPayment_ProcessorErrorKeepsPending(t *testing.T) {
	f := newPaymentFixture(t)
	sub, err := f.useCase.Submit(context.Background(), validSubmissionInput())
	require.NoError(t, err)
	f.processor.Err = errors.New("db down")

	_, err = f.useCase.Approve(context.Background(), sub.ID, nil)
	require.Error(t, err)
	stored, err := f.subs.FindByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domainBilling.PaymentSubmissionPending, stored.Status)
}

func TestRejectManualPayment(t *testing.T) {
	f := newPaymentFixture(t)
	sub, err := f.useCase.Submit(context.Background(), validSubmissionInput())
	require.NoError(t, err)
	reviewer := uint(3)

	rejected, err := f.useCase.Reject(context.Background(), sub.ID, "nominal tidak sesuai", &reviewer)
	require.NoError(t, err)
	assert.Equal(t, domainBilling.PaymentSubmissionRejected, rejected.Status)
	assert.Equal(t, "nominal tidak sesuai", rejected.RejectionReason)
	require.NotNil(t, rejected.ReviewedAt)

	// Reject kedua harus gagal.
	_, err = f.useCase.Reject(context.Background(), sub.ID, "lagi", &reviewer)
	assert.ErrorIs(t, err, domainBilling.ErrPaymentSubmissionNotPending)
}

func TestRejectManualPayment_RequiresReason(t *testing.T) {
	f := newPaymentFixture(t)
	sub, err := f.useCase.Submit(context.Background(), validSubmissionInput())
	require.NoError(t, err)
	_, err = f.useCase.Reject(context.Background(), sub.ID, "  ", nil)
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)
}

func TestListManualPayments(t *testing.T) {
	f := newPaymentFixture(t)
	_, err := f.useCase.Submit(context.Background(), validSubmissionInput())
	require.NoError(t, err)

	all, err := f.useCase.List(context.Background(), "")
	require.NoError(t, err)
	assert.Len(t, all, 1)

	pending, err := f.useCase.List(context.Background(), "pending")
	require.NoError(t, err)
	assert.Len(t, pending, 1)

	approved, err := f.useCase.List(context.Background(), "APPROVED")
	require.NoError(t, err)
	assert.Empty(t, approved)
}
