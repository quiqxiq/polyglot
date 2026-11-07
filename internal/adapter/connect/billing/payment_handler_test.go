// Handler ConnectRPC pembayaran manual & reversal (planning v2 Fase 3).
package billing_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	billingConnect "github.com/quixiq/polyglot/internal/adapter/connect/billing"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

func newPaymentHandlerFixture(t *testing.T) (*billingConnect.PaymentConnectHandler, *mocktest.FakePaymentSubmissionRepo) {
	t.Helper()
	invRepo := mocktest.NewFakeInvoiceRepo()
	require.NoError(t, invRepo.Save(context.Background(), domainBilling.Invoice{
		ID: "inv-1", TenantID: "tenant-default", CustomerID: "cust-1",
		Period: "2026-09", Subtotal: 100000, Total: 100000,
		Status: domainBilling.StatusUnpaid, ManualPaymentCode: "PAY-1",
		QRPayload: "polyglot://invoice/inv-1",
	}))
	subs := mocktest.NewFakePaymentSubmissionRepo()
	processor := &mocktest.FakePaymentProcessor{Pay: domainBilling.Payment{ID: "pay-1", Amount: 100000}}
	settings := mocktest.NewFakeSettingReader(nil)
	submissions := billingUC.NewPaymentSubmissionUseCase(subs, invRepo, processor, settings)
	reverser := &mocktest.FakePaymentReverser{Pay: domainBilling.Payment{ID: "pay-1", InvoiceID: "inv-1", Amount: 100000}}
	payments := billingUC.NewPaymentUseCase(reverser)
	return billingConnect.NewPaymentConnectHandler(submissions, payments), subs
}

func TestPaymentHandler_SubmitApproveReject(t *testing.T) {
	h, _ := newPaymentHandlerFixture(t)
	ctx := context.Background()

	// Submit publik
	subResp, err := h.SubmitManualPayment(ctx, connect.NewRequest(&polyglotv1.SubmitManualPaymentRequest{
		ManualPaymentCode: "PAY-1",
		Amount:            100000,
		SenderName:        "Budi",
		SenderBank:        "BCA",
		ProofImageUrl:     "/uploads/bukti/budi.jpg",
	}))
	require.NoError(t, err)
	sub := subResp.Msg.Submission
	assert.Equal(t, domainBilling.PaymentSubmissionPending, sub.Status)
	assert.Equal(t, "inv-1", sub.InvoiceId)

	// List pending
	listResp, err := h.ListPaymentSubmissions(ctx, connect.NewRequest(&polyglotv1.ListPaymentSubmissionsRequest{
		Status: "PENDING",
	}))
	require.NoError(t, err)
	require.Len(t, listResp.Msg.Submissions, 1)

	// Reject
	rejResp, err := h.RejectPaymentSubmission(ctx, connect.NewRequest(&polyglotv1.RejectPaymentSubmissionRequest{
		Id: sub.Id, Reason: "nominal tidak sesuai",
	}))
	require.NoError(t, err)
	assert.Equal(t, domainBilling.PaymentSubmissionRejected, rejResp.Msg.Submission.Status)
	assert.Equal(t, "nominal tidak sesuai", rejResp.Msg.Submission.RejectionReason)
}

func TestPaymentHandler_Approve(t *testing.T) {
	h, _ := newPaymentHandlerFixture(t)
	ctx := context.Background()
	subResp, err := h.SubmitManualPayment(ctx, connect.NewRequest(&polyglotv1.SubmitManualPaymentRequest{
		ManualPaymentCode: "PAY-1", Amount: 100000, SenderName: "Budi", ProofImageUrl: "/x.jpg",
	}))
	require.NoError(t, err)

	appResp, err := h.ApprovePaymentSubmission(ctx, connect.NewRequest(&polyglotv1.ApprovePaymentSubmissionRequest{
		Id: subResp.Msg.Submission.Id,
	}))
	require.NoError(t, err)
	assert.Equal(t, domainBilling.PaymentSubmissionApproved, appResp.Msg.Submission.Status)
	assert.Equal(t, "pay-1", appResp.Msg.Submission.PaymentId)
}

func TestPaymentHandler_Reverse(t *testing.T) {
	h, _ := newPaymentHandlerFixture(t)
	resp, err := h.ReversePayment(context.Background(), connect.NewRequest(&polyglotv1.ReversePaymentRequest{
		PaymentId: "pay-1",
		Reason:    "salah input",
	}))
	require.NoError(t, err)
	assert.Equal(t, "pay-1", resp.Msg.PaymentId)
	assert.Equal(t, "inv-1", resp.Msg.InvoiceId)
	assert.InDelta(t, 100000, resp.Msg.Amount, 0.01)
}

func TestPaymentHandler_SubmitUnknownCode(t *testing.T) {
	h, _ := newPaymentHandlerFixture(t)
	_, err := h.SubmitManualPayment(context.Background(), connect.NewRequest(&polyglotv1.SubmitManualPaymentRequest{
		ManualPaymentCode: "PAY-UNKNOWN", Amount: 1000, SenderName: "X", ProofImageUrl: "/x.jpg",
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
