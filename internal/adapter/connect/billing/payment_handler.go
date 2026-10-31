package billing

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	"github.com/quixiq/polyglot/pkg/response"
)

// PaymentConnectHandler menangani pembayaran manual (bukti transfer) yang
// di-review finance serta reversal pembayaran (planning v2 Fase 3).
type PaymentConnectHandler struct {
	submissions *billingUC.PaymentSubmissionUseCase
	payments    *billingUC.PaymentUseCase
}

// NewPaymentConnectHandler wires the payment handlers.
func NewPaymentConnectHandler(
	submissions *billingUC.PaymentSubmissionUseCase,
	payments *billingUC.PaymentUseCase,
) *PaymentConnectHandler {
	return &PaymentConnectHandler{submissions: submissions, payments: payments}
}

func toProtoSubmission(s *domainBilling.PaymentSubmission) *devicepb.ManualPaymentSubmission {
	if s == nil {
		return nil
	}
	var reviewedAt int64
	if s.ReviewedAt != nil {
		reviewedAt = s.ReviewedAt.Unix()
	}
	paymentID := ""
	if s.PaymentID != nil {
		paymentID = *s.PaymentID
	}
	return &devicepb.ManualPaymentSubmission{
		Id:              s.ID,
		InvoiceId:       s.InvoiceID,
		Amount:          s.Amount,
		SenderName:      s.SenderName,
		SenderBank:      s.SenderBank,
		ProofImageUrl:   s.ProofImageURL,
		Notes:           s.Notes,
		Status:          s.Status,
		RejectionReason: s.RejectionReason,
		PaymentId:       paymentID,
		CreatedAtUnix:   s.CreatedAt.Unix(),
		ReviewedAtUnix:  reviewedAt,
	}
}

// ResolveManualPayment — PUBLIC: tampilkan ringkasan tagihan untuk kode bayar.
func (h *PaymentConnectHandler) ResolveManualPayment(ctx context.Context, req *connect.Request[devicepb.ResolveManualPaymentRequest]) (*connect.Response[devicepb.ResolveManualPaymentResponse], error) {
	if h.submissions == nil {
		return nil, response.Unavailable("payment submission usecase unavailable")
	}
	inv, err := h.submissions.ResolveInvoice(ctx, req.Msg.ManualPaymentCode)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ResolveManualPaymentResponse{
		InvoiceId:     inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		Period:        inv.Period,
		Total:         inv.Total,
		PaidAmount:    inv.PaidAmount,
		Outstanding:   inv.Total - inv.PaidAmount,
		Status:        inv.Status,
		DueDateUnix:   inv.DueDate.Unix(),
	}), nil
}

// SubmitManualPayment — PUBLIC: pelanggan mengunggah bukti transfer via kode bayar.
func (h *PaymentConnectHandler) SubmitManualPayment(ctx context.Context, req *connect.Request[devicepb.SubmitManualPaymentRequest]) (*connect.Response[devicepb.SubmitManualPaymentResponse], error) {
	if h.submissions == nil {
		return nil, response.Unavailable("payment submission usecase unavailable")
	}
	sub, err := h.submissions.Submit(ctx, billingUC.SubmitManualPaymentInput{
		ManualPaymentCode: req.Msg.ManualPaymentCode,
		Amount:            req.Msg.Amount,
		SenderName:        req.Msg.SenderName,
		SenderBank:        req.Msg.SenderBank,
		ProofImageURL:     req.Msg.ProofImageUrl,
		Notes:             req.Msg.Notes,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SubmitManualPaymentResponse{
		Submission: toProtoSubmission(&sub),
	}), nil
}

// ListPaymentSubmissions mengembalikan antrean bukti transfer untuk finance.
func (h *PaymentConnectHandler) ListPaymentSubmissions(ctx context.Context, req *connect.Request[devicepb.ListPaymentSubmissionsRequest]) (*connect.Response[devicepb.ListPaymentSubmissionsResponse], error) {
	if h.submissions == nil {
		return nil, response.Unavailable("payment submission usecase unavailable")
	}
	list, err := h.submissions.List(ctx, req.Msg.Status)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*devicepb.ManualPaymentSubmission, len(list))
	for i := range list {
		out[i] = toProtoSubmission(&list[i])
	}
	return connect.NewResponse(&devicepb.ListPaymentSubmissionsResponse{Submissions: out}), nil
}

// ApprovePaymentSubmission menyetujui bukti transfer (lunasi + kas + un-isolir).
func (h *PaymentConnectHandler) ApprovePaymentSubmission(ctx context.Context, req *connect.Request[devicepb.ApprovePaymentSubmissionRequest]) (*connect.Response[devicepb.ApprovePaymentSubmissionResponse], error) {
	if h.submissions == nil {
		return nil, response.Unavailable("payment submission usecase unavailable")
	}
	reviewer := uint(1)
	sub, err := h.submissions.Approve(ctx, req.Msg.Id, &reviewer)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ApprovePaymentSubmissionResponse{
		Submission: toProtoSubmission(&sub),
	}), nil
}

// RejectPaymentSubmission menolak bukti transfer tanpa menyentuh buku kas.
func (h *PaymentConnectHandler) RejectPaymentSubmission(ctx context.Context, req *connect.Request[devicepb.RejectPaymentSubmissionRequest]) (*connect.Response[devicepb.RejectPaymentSubmissionResponse], error) {
	if h.submissions == nil {
		return nil, response.Unavailable("payment submission usecase unavailable")
	}
	reviewer := uint(1)
	sub, err := h.submissions.Reject(ctx, req.Msg.Id, req.Msg.Reason, &reviewer)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RejectPaymentSubmissionResponse{
		Submission: toProtoSubmission(&sub),
	}), nil
}

// ReversePayment membatalkan pembayaran dengan jurnal koreksi kas.
func (h *PaymentConnectHandler) ReversePayment(ctx context.Context, req *connect.Request[devicepb.ReversePaymentRequest]) (*connect.Response[devicepb.ReversePaymentResponse], error) {
	if h.payments == nil {
		return nil, response.Unavailable("payment usecase unavailable")
	}
	actor := uint(1)
	pay, err := h.payments.Reverse(ctx, req.Msg.PaymentId, req.Msg.Reason, &actor)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ReversePaymentResponse{
		PaymentId: pay.ID,
		InvoiceId: pay.InvoiceID,
		Amount:    pay.Amount,
	}), nil
}
