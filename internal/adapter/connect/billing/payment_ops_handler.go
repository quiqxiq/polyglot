package billing

import (
	"context"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/pkg/response"
)

// ─── Payment Methods ────────────────────────────────────────────────────────

// ListPaymentMethods returns the configured payment methods.
func (h *BillingConnectHandler) ListPaymentMethods(ctx context.Context, req *connect.Request[polyglotv1.ListPaymentMethodsRequest]) (*connect.Response[polyglotv1.ListPaymentMethodsResponse], error) {
	if h.payMethodUC == nil {
		return nil, response.Unavailable("payment method usecase unavailable")
	}
	methods, err := h.payMethodUC.ListPaymentMethods(ctx, domainTenant.DefaultID, req.Msg.GetActiveOnly())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.PaymentMethod, len(methods))
	for i, m := range methods {
		out[i] = &polyglotv1.PaymentMethod{
			Id:            m.ID,
			Name:          m.Name,
			Type:          m.Type,
			IsActive:      m.IsActive,
			BankName:      m.BankName,
			AccountNumber: m.AccountNumber,
			AccountHolder: m.AccountHolder,
			QrisImageUrl:  m.QRISImageURL,
			Instructions:  m.Instructions,
			SortOrder:     int32(m.SortOrder),
		}
	}
	return connect.NewResponse(&polyglotv1.ListPaymentMethodsResponse{
		PaymentMethods: out,
	}), nil
}

// SavePaymentMethod creates or updates a payment method.
func (h *BillingConnectHandler) SavePaymentMethod(ctx context.Context, req *connect.Request[polyglotv1.SavePaymentMethodRequest]) (*connect.Response[polyglotv1.SavePaymentMethodResponse], error) {
	if h.payMethodUC == nil {
		return nil, response.Unavailable("payment method usecase unavailable")
	}
	pm := req.Msg.GetPaymentMethod()
	saved, err := h.payMethodUC.SavePaymentMethod(ctx, domainBilling.PaymentMethod{
		ID:            pm.GetId(),
		TenantID:      domainTenant.DefaultID,
		Name:          pm.GetName(),
		Type:          pm.GetType(),
		IsActive:      pm.GetIsActive(),
		BankName:      pm.GetBankName(),
		AccountNumber: pm.GetAccountNumber(),
		AccountHolder: pm.GetAccountHolder(),
		QRISImageURL:  pm.GetQrisImageUrl(),
		Instructions:  pm.GetInstructions(),
		SortOrder:     int(pm.GetSortOrder()),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SavePaymentMethodResponse{
		PaymentMethod: &polyglotv1.PaymentMethod{
			Id:            saved.ID,
			Name:          saved.Name,
			Type:          saved.Type,
			IsActive:      saved.IsActive,
			BankName:      saved.BankName,
			AccountNumber: saved.AccountNumber,
			AccountHolder: saved.AccountHolder,
			QrisImageUrl:  saved.QRISImageURL,
			Instructions:  saved.Instructions,
			SortOrder:     int32(saved.SortOrder),
		},
	}), nil
}

// DeletePaymentMethod deactivates a payment method.
func (h *BillingConnectHandler) DeletePaymentMethod(ctx context.Context, req *connect.Request[polyglotv1.DeletePaymentMethodRequest]) (*connect.Response[polyglotv1.DeletePaymentMethodResponse], error) {
	if h.payMethodUC == nil {
		return nil, response.Unavailable("payment method usecase unavailable")
	}
	if err := h.payMethodUC.DeletePaymentMethod(ctx, req.Msg.GetId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.DeletePaymentMethodResponse{
		Message: "Metode pembayaran dinonaktifkan",
	}), nil
}

// ─── Mass Payment ───────────────────────────────────────────────────────────

// CreateMassPayment queues a bulk payment batch.
func (h *BillingConnectHandler) CreateMassPayment(ctx context.Context, req *connect.Request[polyglotv1.CreateMassPaymentRequest]) (*connect.Response[polyglotv1.CreateMassPaymentResponse], error) {
	if h.massPayUC == nil {
		return nil, response.Unavailable("mass payment usecase unavailable")
	}
	var createdBy *uint
	if userID, _, exists := iauth.IdentityFromContext(ctx); exists && userID > 0 {
		createdBy = &userID
	}
	var pmID *string
	if req.Msg.GetPaymentMethodId() != "" {
		id := req.Msg.GetPaymentMethodId()
		pmID = &id
	}
	batch, err := h.massPayUC.CreateMassPayment(ctx, domainTenant.DefaultID, createdBy, pmID, req.Msg.GetInvoiceIds(), req.Msg.GetNotes())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CreateMassPaymentResponse{
		BatchId:      batch.ID,
		JobId:        batch.JobID,
		InvoiceCount: int32(batch.InvoiceCount),
	}), nil
}

// ListMassPayments returns recent bulk payment batches.
func (h *BillingConnectHandler) ListMassPayments(ctx context.Context, req *connect.Request[polyglotv1.ListMassPaymentsRequest]) (*connect.Response[polyglotv1.ListMassPaymentsResponse], error) {
	if h.massPayUC == nil {
		return nil, response.Unavailable("mass payment usecase unavailable")
	}
	limit := int(req.Msg.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	batches, err := h.massPayUC.ListMassPayments(ctx, domainTenant.DefaultID, limit)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.BulkPaymentBatch, len(batches))
	for i, b := range batches {
		out[i] = &polyglotv1.BulkPaymentBatch{
			Id:            b.ID,
			InvoiceCount:  int32(b.InvoiceCount),
			Status:        b.Status,
			JobId:         b.JobID,
			Notes:         b.Notes,
			CreatedAtUnix: b.CreatedAt.Unix(),
		}
	}
	return connect.NewResponse(&polyglotv1.ListMassPaymentsResponse{
		Batches: out,
	}), nil
}

// GetMassPayment returns one bulk payment batch with its items.
func (h *BillingConnectHandler) GetMassPayment(ctx context.Context, req *connect.Request[polyglotv1.GetMassPaymentRequest]) (*connect.Response[polyglotv1.GetMassPaymentResponse], error) {
	if h.massPayUC == nil {
		return nil, response.Unavailable("mass payment usecase unavailable")
	}
	batch, items, err := h.massPayUC.GetMassPayment(ctx, req.Msg.GetBatchId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	outItems := make([]*polyglotv1.BulkPaymentItem, len(items))
	for i, it := range items {
		var procAt int64
		if it.ProcessedAt != nil {
			procAt = it.ProcessedAt.Unix()
		}
		var payID string
		if it.PaymentID != nil {
			payID = *it.PaymentID
		}
		outItems[i] = &polyglotv1.BulkPaymentItem{
			Id:              it.ID,
			InvoiceId:       it.InvoiceID,
			Amount:          it.Amount,
			Status:          it.Status,
			Error:           it.Error,
			PaymentId:       payID,
			ProcessedAtUnix: procAt,
		}
	}
	return connect.NewResponse(&polyglotv1.GetMassPaymentResponse{
		Batch: &polyglotv1.BulkPaymentBatch{
			Id:            batch.ID,
			InvoiceCount:  int32(batch.InvoiceCount),
			Status:        batch.Status,
			JobId:         batch.JobID,
			Notes:         batch.Notes,
			CreatedAtUnix: batch.CreatedAt.Unix(),
		},
		Items: outItems,
	}), nil
}
