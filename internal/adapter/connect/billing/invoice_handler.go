package billing

import (
	"context"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
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
		TenantID:          domainTenant.DefaultID,
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
		TenantID:       domainTenant.DefaultID,
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
			TenantID:          domainTenant.DefaultID,
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
		TenantID:         domainTenant.DefaultID,
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
	res, err := h.runBilling.Run(ctx, domainTenant.DefaultID, period)
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
