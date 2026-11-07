package billing

import (
	"context"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListInvoices returns invoices matching the request filters.
func (h *BillingConnectHandler) ListInvoices(ctx context.Context, req *connect.Request[polyglotv1.ListInvoicesRequest]) (*connect.Response[polyglotv1.ListInvoicesResponse], error) {
	if h.invoiceUC == nil {
		return nil, response.Unavailable("invoice usecase unavailable")
	}
	f := port.InvoiceFilter{
		TenantID:          "tenant-default",
		CustomerID:        req.Msg.GetCustomerId(),
		Status:            req.Msg.GetStatus(),
		Period:            req.Msg.GetPeriod(),
		SubscriptionID:    req.Msg.GetSubscriptionId(),
		OverdueOnly:       req.Msg.GetOverdueOnly(),
		UnpaidOnly:        req.Msg.GetUnpaidOnly(),
		UnpaidDaysDisplay: int(req.Msg.GetUnpaidDaysDisplay()),
		Search:            req.Msg.GetSearch(),
		Limit:             int(req.Msg.GetLimit()),
		Offset:            int(req.Msg.GetOffset()),
	}
	if req.Msg.GetDueFromUnix() > 0 {
		t := time.Unix(req.Msg.GetDueFromUnix(), 0).UTC()
		f.DueFrom = &t
	}
	if req.Msg.GetDueToUnix() > 0 {
		t := time.Unix(req.Msg.GetDueToUnix(), 0).UTC()
		f.DueTo = &t
	}

	invoices, total, err := h.invoiceUC.FindFiltered(ctx, f)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	summary, _ := h.invoiceUC.Summary(ctx, f)

	return connect.NewResponse(&polyglotv1.ListInvoicesResponse{
		Invoices: toProtoInvoiceList(invoices),
		Summary: &polyglotv1.InvoiceSummary{
			Count:            int32(summary.Count),
			TotalBilled:      summary.TotalBilled,
			TotalPaid:        summary.TotalPaid,
			TotalOutstanding: summary.TotalOutstanding,
			TotalOverdue:     summary.TotalOverdue,
		},
		Total: int32(total),
	}), nil
}

// CreateInvoice creates a manual invoice.
func (h *BillingConnectHandler) CreateInvoice(ctx context.Context, req *connect.Request[polyglotv1.CreateInvoiceRequest]) (*connect.Response[polyglotv1.CreateInvoiceResponse], error) {
	if h.invoiceUC == nil {
		return nil, response.Unavailable("invoice usecase unavailable")
	}
	items := make([]billingUC.CreateInvoiceItemInput, len(req.Msg.GetItems()))
	for i, it := range req.Msg.GetItems() {
		items[i] = billingUC.CreateInvoiceItemInput{
			Description: it.GetDescription(),
			Quantity:    it.GetQuantity(),
			UnitPrice:   it.GetUnitPrice(),
			ItemType:    it.GetItemType(),
		}
	}
	var subID *string
	if req.Msg.GetSubscriptionId() != "" {
		s := req.Msg.GetSubscriptionId()
		subID = &s
	}
	var dueDate time.Time
	if req.Msg.GetDueDateUnix() > 0 {
		dueDate = time.Unix(req.Msg.GetDueDateUnix(), 0).UTC()
	}
	inv, err := h.invoiceUC.CreateManualInvoice(ctx, billingUC.CreateManualInvoiceInput{
		TenantID:       "tenant-default",
		CustomerID:     req.Msg.GetCustomerId(),
		SubscriptionID: subID,
		Period:         req.Msg.GetPeriod(),
		DueDate:        dueDate,
		TaxPercent:     req.Msg.GetTaxPercent(),
		Items:          items,
		Notes:          req.Msg.GetNotes(),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CreateInvoiceResponse{
		Invoice: toProtoInvoice(&inv),
	}), nil
}

// ExportInvoices exports invoices to CSV or XLSX format.
func (h *BillingConnectHandler) ExportInvoices(ctx context.Context, req *connect.Request[polyglotv1.ExportInvoicesRequest]) (*connect.Response[polyglotv1.ExportInvoicesResponse], error) {
	if h.invoiceUC == nil {
		return nil, response.Unavailable("invoice usecase unavailable")
	}
	var f port.InvoiceFilter
	if req.Msg.GetFilter() != nil {
		flt := req.Msg.GetFilter()
		f = port.InvoiceFilter{
			TenantID:          "tenant-default",
			CustomerID:        flt.GetCustomerId(),
			Status:            flt.GetStatus(),
			Period:            flt.GetPeriod(),
			SubscriptionID:    flt.GetSubscriptionId(),
			OverdueOnly:       flt.GetOverdueOnly(),
			UnpaidOnly:        flt.GetUnpaidOnly(),
			UnpaidDaysDisplay: int(flt.GetUnpaidDaysDisplay()),
			Search:            flt.GetSearch(),
		}
		if flt.GetDueFromUnix() > 0 {
			t := time.Unix(flt.GetDueFromUnix(), 0).UTC()
			f.DueFrom = &t
		}
		if flt.GetDueToUnix() > 0 {
			t := time.Unix(flt.GetDueToUnix(), 0).UTC()
			f.DueTo = &t
		}
	}
	format := req.Msg.GetFormat()
	if format == "" {
		format = "csv"
	}
	payload, contentType, filename, err := h.invoiceUC.ExportInvoices(ctx, f, format)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ExportInvoicesResponse{
		Payload:     payload,
		ContentType: contentType,
		Filename:    filename,
	}), nil
}

// GetInvoice returns one invoice by identifier.
func (h *BillingConnectHandler) GetInvoice(ctx context.Context, req *connect.Request[polyglotv1.GetInvoiceRequest]) (*connect.Response[polyglotv1.GetInvoiceResponse], error) {
	if h.invoiceUC == nil {
		return nil, response.Unavailable("invoice usecase unavailable")
	}
	inv, err := h.invoiceUC.GetInvoice(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.GetInvoiceResponse{
		Invoice: toProtoInvoice(&inv),
	}), nil
}

// CancelInvoice cancels an invoice officially.
func (h *BillingConnectHandler) CancelInvoice(ctx context.Context, req *connect.Request[polyglotv1.CancelInvoiceRequest]) (*connect.Response[polyglotv1.CancelInvoiceResponse], error) {
	if h.invoiceUC == nil {
		return nil, response.Unavailable("invoice usecase unavailable")
	}
	inv, err := h.invoiceUC.CancelInvoice(ctx, req.Msg.Id, req.Msg.Reason)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CancelInvoiceResponse{
		Invoice: toProtoInvoice(&inv),
	}), nil
}

// CashierResolve resolves an invoice for cashier payment.
func (h *BillingConnectHandler) CashierResolve(ctx context.Context, req *connect.Request[polyglotv1.CashierResolveRequest]) (*connect.Response[polyglotv1.CashierResolveResponse], error) {
	if h.checkoutUC == nil {
		return nil, response.Unavailable("checkout usecase unavailable")
	}
	ident := req.Msg.Identifier

	var inv domainBillingInvoiceAlias
	var err error
	switch req.Msg.Method {
	case polyglotv1.ResolveMethod_RESOLVE_QR:
		inv, err = h.checkoutUC.ResolveByQR(ctx, ident)
	case polyglotv1.ResolveMethod_RESOLVE_PORTAL:
		inv, err = h.checkoutUC.ResolveByPortalCode(ctx, ident)
	default: // RESOLVE_CODE
		inv, err = h.checkoutUC.ResolveByPaymentCode(ctx, ident)
	}
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CashierResolveResponse{
		Invoice: toProtoInvoice(&inv),
	}), nil
}

// CashierPay records a cashier payment for an invoice.
func (h *BillingConnectHandler) CashierPay(ctx context.Context, req *connect.Request[polyglotv1.CashierPayRequest]) (*connect.Response[polyglotv1.CashierPayResponse], error) {
	if h.checkoutUC == nil {
		return nil, response.Unavailable("checkout usecase unavailable")
	}
	scanMethod := req.Msg.ScanMethod
	if scanMethod == "" {
		scanMethod = "MANUAL"
	}
	accountID := req.Msg.CashAccountId
	if accountID == "" {
		accountID = "ca-1001-kas-kantor"
	}
	categoryID := req.Msg.IncomeCategoryId
	if categoryID == "" {
		categoryID = "cc-tagihan"
	}

	var receivedBy *uint
	if userID, _, exists := iauth.IdentityFromContext(ctx); exists && userID > 0 {
		receivedBy = &userID
	}

	pay, err := h.checkoutUC.PayCash(ctx, port.CashPaymentCommand{
		TenantID:         "tenant-default",
		InvoiceID:        req.Msg.InvoiceId,
		Amount:           req.Msg.Amount,
		CashAccountID:    accountID,
		IncomeCategoryID: categoryID,
		ReceivedBy:       receivedBy,
		ScanMethod:       scanMethod,
		Reference:        req.Msg.Reference,
		Notes:            req.Msg.Notes,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	inv, _ := h.invoiceUC.GetInvoice(ctx, req.Msg.InvoiceId)
	return connect.NewResponse(&polyglotv1.CashierPayResponse{
		PaymentId: pay.ID,
		PaymentNo: pay.PaymentNo,
		Invoice:   toProtoInvoice(&inv),
	}), nil
}

// GenerateInvoices runs the billing invoice generation for the given period.
func (h *BillingConnectHandler) GenerateInvoices(ctx context.Context, req *connect.Request[polyglotv1.GenerateInvoicesRequest]) (*connect.Response[polyglotv1.GenerateInvoicesResponse], error) {
	if h.runBilling == nil {
		return nil, response.Unavailable("billing runner unavailable")
	}
	period := req.Msg.Period
	if period == "" {
		period = time.Now().Format("2006-01")
	}
	res, err := h.runBilling.Run(ctx, "tenant-default", period)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.GenerateInvoicesResponse{
		Created: int32(res.Created),
		Skipped: int32(res.Skipped),
	}), nil
}

// ApplyInvoiceDiscount menerapkan diskon/promo code manual pada faktur.
func (h *BillingConnectHandler) ApplyInvoiceDiscount(ctx context.Context, req *connect.Request[polyglotv1.ApplyInvoiceDiscountRequest]) (*connect.Response[polyglotv1.ApplyInvoiceDiscountResponse], error) {
	if h.invoiceUC == nil {
		return nil, response.Unavailable("invoice usecase unavailable")
	}
	inv, err := h.invoiceUC.ApplyDiscount(ctx, req.Msg.InvoiceId, req.Msg.DiscountType, req.Msg.DiscountValue, req.Msg.PromoCode)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ApplyInvoiceDiscountResponse{
		Invoice: toProtoInvoice(&inv),
	}), nil
}

// ─── Payment Methods ────────────────────────────────────────────────────────

func (h *BillingConnectHandler) ListPaymentMethods(ctx context.Context, req *connect.Request[polyglotv1.ListPaymentMethodsRequest]) (*connect.Response[polyglotv1.ListPaymentMethodsResponse], error) {
	if h.payMethodUC == nil {
		return nil, response.Unavailable("payment method usecase unavailable")
	}
	methods, err := h.payMethodUC.ListPaymentMethods(ctx, "tenant-default", req.Msg.GetActiveOnly())
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

func (h *BillingConnectHandler) SavePaymentMethod(ctx context.Context, req *connect.Request[polyglotv1.SavePaymentMethodRequest]) (*connect.Response[polyglotv1.SavePaymentMethodResponse], error) {
	if h.payMethodUC == nil {
		return nil, response.Unavailable("payment method usecase unavailable")
	}
	pm := req.Msg.GetPaymentMethod()
	if pm == nil {
		return nil, response.InvalidArgument("payment_method is required")
	}
	saved, err := h.payMethodUC.SavePaymentMethod(ctx, domainBilling.PaymentMethod{
		ID:            pm.GetId(),
		TenantID:      "tenant-default",
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
	batch, err := h.massPayUC.CreateMassPayment(ctx, "tenant-default", createdBy, pmID, req.Msg.GetInvoiceIds(), req.Msg.GetNotes())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CreateMassPaymentResponse{
		BatchId:      batch.ID,
		JobId:        batch.JobID,
		InvoiceCount: int32(batch.InvoiceCount),
	}), nil
}

func (h *BillingConnectHandler) ListMassPayments(ctx context.Context, req *connect.Request[polyglotv1.ListMassPaymentsRequest]) (*connect.Response[polyglotv1.ListMassPaymentsResponse], error) {
	if h.massPayUC == nil {
		return nil, response.Unavailable("mass payment usecase unavailable")
	}
	limit := int(req.Msg.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	batches, err := h.massPayUC.ListMassPayments(ctx, "tenant-default", limit)
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

// ─── Gateway Operations ─────────────────────────────────────────────────────

func (h *BillingConnectHandler) ListGatewayTransactions(ctx context.Context, req *connect.Request[polyglotv1.ListGatewayTransactionsRequest]) (*connect.Response[polyglotv1.ListGatewayTransactionsResponse], error) {
	if h.gwTxRepo == nil {
		return nil, response.Unavailable("gateway transaction repository unavailable")
	}
	fRepo, ok := h.gwTxRepo.(port.GatewayTransactionFilteredRepository)
	if !ok {
		return nil, response.Unavailable("gateway transaction filtering unavailable")
	}
	txs, total, err := fRepo.ListFiltered(ctx, port.GatewayTransactionFilter{
		TenantID:  "tenant-default",
		Gateway:   req.Msg.GetGateway(),
		Status:    req.Msg.GetStatus(),
		InvoiceID: req.Msg.GetInvoiceId(),
		Limit:     int(req.Msg.GetLimit()),
		Offset:    int(req.Msg.GetOffset()),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.GatewayTransaction, len(txs))
	for i, t := range txs {
		var paidAt int64
		if t.PaidAt != nil {
			paidAt = t.PaidAt.Unix()
		}
		var invID, payID string
		if t.InvoiceID != nil {
			invID = *t.InvoiceID
		}
		if t.PaymentID != nil {
			payID = *t.PaymentID
		}
		out[i] = &polyglotv1.GatewayTransaction{
			Id:             t.ID,
			Gateway:        t.Gateway,
			ExternalId:     t.ExternalID,
			InvoiceId:      invID,
			PaymentId:      payID,
			Amount:         t.Amount,
			FeeAmount:      t.FeeAmount,
			Status:         t.Status,
			PaymentChannel: t.PaymentChannel,
			PaymentUrl:     t.PaymentURL,
			CreatedAtUnix:  t.CreatedAt.Unix(),
			PaidAtUnix:     paidAt,
		}
	}
	return connect.NewResponse(&polyglotv1.ListGatewayTransactionsResponse{
		Transactions: out,
		Total:        int32(total),
	}), nil
}

func (h *BillingConnectHandler) SyncGatewayTransactions(ctx context.Context, _ *connect.Request[polyglotv1.SyncGatewayTransactionsRequest]) (*connect.Response[polyglotv1.SyncGatewayTransactionsResponse], error) {
	if h.chargeUC == nil {
		return nil, response.Unavailable("gateway usecase unavailable")
	}
	synced, expired, err := h.chargeUC.SyncPendingTransactions(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SyncGatewayTransactionsResponse{
		Synced:  int32(synced),
		Expired: int32(expired),
	}), nil
}

func (h *BillingConnectHandler) ListGatewaySettlements(ctx context.Context, req *connect.Request[polyglotv1.ListGatewaySettlementsRequest]) (*connect.Response[polyglotv1.ListGatewaySettlementsResponse], error) {
	if h.settlementRepo == nil {
		return nil, response.Unavailable("gateway settlement repository unavailable")
	}
	settlements, err := h.settlementRepo.List(ctx, port.GatewaySettlementFilter{
		TenantID: "tenant-default",
		Gateway:  req.Msg.GetGateway(),
		Status:   req.Msg.GetStatus(),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.GatewaySettlement, len(settlements))
	for i, s := range settlements {
		var fromUnix, toUnix int64
		if s.PeriodFrom != nil {
			fromUnix = s.PeriodFrom.Unix()
		}
		if s.PeriodTo != nil {
			toUnix = s.PeriodTo.Unix()
		}
		out[i] = &polyglotv1.GatewaySettlement{
			Id:             s.ID,
			Gateway:        s.Gateway,
			SettlementRef:  s.SettlementRef,
			PeriodFromUnix: fromUnix,
			PeriodToUnix:   toUnix,
			GrossAmount:    s.GrossAmount,
			FeeAmount:      s.FeeAmount,
			NetAmount:      s.NetAmount,
			Status:         s.Status,
			Notes:          s.Notes,
		}
	}
	return connect.NewResponse(&polyglotv1.ListGatewaySettlementsResponse{
		Settlements: out,
	}), nil
}

func (h *BillingConnectHandler) SaveGatewaySettlement(ctx context.Context, req *connect.Request[polyglotv1.SaveGatewaySettlementRequest]) (*connect.Response[polyglotv1.SaveGatewaySettlementResponse], error) {
	if h.settlementRepo == nil {
		return nil, response.Unavailable("gateway settlement repository unavailable")
	}
	s := req.Msg.GetSettlement()
	if s == nil {
		return nil, response.InvalidArgument("settlement is required")
	}
	var from, to *time.Time
	if s.GetPeriodFromUnix() > 0 {
		t := time.Unix(s.GetPeriodFromUnix(), 0).UTC()
		from = &t
	}
	if s.GetPeriodToUnix() > 0 {
		t := time.Unix(s.GetPeriodToUnix(), 0).UTC()
		to = &t
	}
	var createdBy *uint
	if userID, _, exists := iauth.IdentityFromContext(ctx); exists && userID > 0 {
		createdBy = &userID
	}
	domainSettlement := domainBilling.GatewaySettlement{
		ID:            s.GetId(),
		TenantID:      "tenant-default",
		Gateway:       s.GetGateway(),
		SettlementRef: s.GetSettlementRef(),
		PeriodFrom:    from,
		PeriodTo:      to,
		GrossAmount:   s.GetGrossAmount(),
		FeeAmount:     s.GetFeeAmount(),
		NetAmount:     s.GetNetAmount(),
		Status:        s.GetStatus(),
		Notes:         s.GetNotes(),
		CreatedBy:     createdBy,
	}
	if err := h.settlementRepo.Save(ctx, domainSettlement); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SaveGatewaySettlementResponse{
		Settlement: s,
	}), nil
}
