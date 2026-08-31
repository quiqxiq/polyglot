package billing

import (
	"context"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListInvoices returns invoices matching the request filters.
func (h *BillingConnectHandler) ListInvoices(ctx context.Context, req *connect.Request[devicepb.ListInvoicesRequest]) (*connect.Response[devicepb.ListInvoicesResponse], error) {
	if h.invoiceUC == nil {
		return nil, response.Unavailable("invoice usecase unavailable")
	}
	invoices, err := h.invoiceUC.ListInvoices(ctx, req.Msg.CustomerId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	if req.Msg.Status != "" {
		filtered := invoices[:0]
		for _, inv := range invoices {
			if inv.Status == req.Msg.Status {
				filtered = append(filtered, inv)
			}
		}
		invoices = filtered
	}
	return connect.NewResponse(&devicepb.ListInvoicesResponse{
		Invoices: toProtoInvoiceList(invoices),
	}), nil
}

// GetInvoice returns one invoice by identifier.
func (h *BillingConnectHandler) GetInvoice(ctx context.Context, req *connect.Request[devicepb.GetInvoiceRequest]) (*connect.Response[devicepb.GetInvoiceResponse], error) {
	if h.invoiceUC == nil {
		return nil, response.Unavailable("invoice usecase unavailable")
	}
	inv, err := h.invoiceUC.GetInvoice(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetInvoiceResponse{
		Invoice: toProtoInvoice(&inv),
	}), nil
}

// CashierResolve resolves an invoice for cashier payment.
func (h *BillingConnectHandler) CashierResolve(ctx context.Context, req *connect.Request[devicepb.CashierResolveRequest]) (*connect.Response[devicepb.CashierResolveResponse], error) {
	if h.checkoutUC == nil {
		return nil, response.Unavailable("checkout usecase unavailable")
	}
	ident := req.Msg.Identifier

	var inv domainBillingInvoiceAlias
	var err error
	switch req.Msg.Method {
	case devicepb.ResolveMethod_RESOLVE_QR:
		inv, err = h.checkoutUC.ResolveByQR(ctx, ident)
	case devicepb.ResolveMethod_RESOLVE_PORTAL:
		inv, err = h.checkoutUC.ResolveByPortalCode(ctx, ident)
	default: // RESOLVE_CODE
		inv, err = h.checkoutUC.ResolveByPaymentCode(ctx, ident)
	}
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CashierResolveResponse{
		Invoice: toProtoInvoice(&inv),
	}), nil
}

// CashierPay records a cashier payment for an invoice.
func (h *BillingConnectHandler) CashierPay(ctx context.Context, req *connect.Request[devicepb.CashierPayRequest]) (*connect.Response[devicepb.CashierPayResponse], error) {
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

	pay, err := h.checkoutUC.PayCash(ctx, port.CashPaymentCommand{
		TenantID:         "tenant-default",
		InvoiceID:        req.Msg.InvoiceId,
		Amount:           req.Msg.Amount,
		CashAccountID:    accountID,
		IncomeCategoryID: categoryID,
		ScanMethod:       scanMethod,
		Reference:        req.Msg.Reference,
		Notes:            req.Msg.Notes,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	inv, _ := h.invoiceUC.GetInvoice(ctx, req.Msg.InvoiceId)
	return connect.NewResponse(&devicepb.CashierPayResponse{
		PaymentId: pay.ID,
		PaymentNo: pay.PaymentNo,
		Invoice:   toProtoInvoice(&inv),
	}), nil
}

// GenerateInvoices runs the billing invoice generation for the given period.
func (h *BillingConnectHandler) GenerateInvoices(ctx context.Context, req *connect.Request[devicepb.GenerateInvoicesRequest]) (*connect.Response[devicepb.GenerateInvoicesResponse], error) {
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
	return connect.NewResponse(&devicepb.GenerateInvoicesResponse{
		Created: int32(res.Created),
		Skipped: int32(res.Skipped),
	}), nil
}
