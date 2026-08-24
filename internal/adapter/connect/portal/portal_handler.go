package portal

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	uc "github.com/quixiq/polyglot/internal/usecase/portal"
	"github.com/quixiq/polyglot/pkg/response"
)

type PortalConnectHandler struct {
	usecase *uc.UseCase
}

func NewPortalConnectHandler(u *uc.UseCase) *PortalConnectHandler {
	return &PortalConnectHandler{usecase: u}
}

func (h *PortalConnectHandler) RequestOTP(ctx context.Context, req *connect.Request[devicepb.RequestOTPRequest]) (*connect.Response[devicepb.RequestOTPResponse], error) {
	if h.usecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("portal usecase unavailable"))
	}
	if req.Msg.Identifier == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("identifier is required"))
	}
	masked, err := h.usecase.RequestOTP(ctx, req.Msg.Identifier)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RequestOTPResponse{
		Message:     "OTP dikirim via WhatsApp ke " + masked,
		MaskedPhone: masked,
	}), nil
}

func (h *PortalConnectHandler) Login(ctx context.Context, req *connect.Request[devicepb.PortalLoginRequest]) (*connect.Response[devicepb.PortalLoginResponse], error) {
	if h.usecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("portal usecase unavailable"))
	}
	if req.Msg.Identifier == "" || req.Msg.Otp == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("identifier and otp are required"))
	}
	token, cust, err := h.usecase.Login(ctx, req.Msg.Identifier, req.Msg.Otp)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.PortalLoginResponse{
		Token:         token,
		ExpiresAtUnix: time.Now().Add(12 * time.Hour).Unix(),
		Customer:      toProtoCustomerPortal(&cust),
	}), nil
}

func (h *PortalConnectHandler) Overview(ctx context.Context, req *connect.Request[devicepb.PortalOverviewRequest]) (*connect.Response[devicepb.PortalOverviewResponse], error) {
	if h.usecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("portal usecase unavailable"))
	}
	cust, err := h.usecase.Authenticate(ctx, req.Msg.Token)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("session token invalid or expired"))
	}
	ov, err := h.usecase.Overview(ctx, cust.ID)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	resp := &devicepb.PortalOverviewResponse{
		Customer:   toProtoCustomerPortal(&ov.Customer),
		Status:     ov.Status,
		PaymentUrl: ov.PaymentURL,
	}
	if ov.Subscription != nil {
		var endUnix int64
		if ov.Subscription.EndDate != nil {
			if t, err := time.Parse("2006-01-02", *ov.Subscription.EndDate); err == nil {
				endUnix = t.Unix()
			}
		}
		resp.Subscription = &devicepb.PortalSubscriptionSummary{
			Id:          ov.Subscription.ID,
			PlanId:      ov.Subscription.PlanID,
			ServiceType: ov.Subscription.ServiceType,
			Status:      ov.Subscription.Status,
			RateLimit:   ov.Subscription.RateLimit,
			BillingDay:  int32(ov.Subscription.BillingDay),
			EndDateUnix: endUnix,
		}
	}
	for _, inv := range ov.UnpaidInvoices {
		resp.UnpaidInvoices = append(resp.UnpaidInvoices, &devicepb.UnpaidInvoiceSummary{
			Id:                inv.ID,
			InvoiceNumber:     inv.InvoiceNumber,
			Period:            inv.Period,
			Total:             inv.Total,
			PaidAmount:        inv.PaidAmount,
			Outstanding:       inv.Outstanding,
			DueDate:           inv.DueDate,
			Status:            inv.Status,
			ManualPaymentCode: inv.ManualPaymentCode,
		})
	}
	return connect.NewResponse(resp), nil
}

func (h *PortalConnectHandler) MyInvoices(ctx context.Context, req *connect.Request[devicepb.MyInvoicesRequest]) (*connect.Response[devicepb.MyInvoicesResponse], error) {
	if h.usecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("portal usecase unavailable"))
	}
	cust, err := h.usecase.Authenticate(ctx, req.Msg.Token)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("session token invalid or expired"))
	}
	invoices, err := h.usecase.Invoices(ctx, cust.ID)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.MyInvoicesResponse{
		Invoices: toProtoInvoicePortalList(invoices),
	}), nil
}

func (h *PortalConnectHandler) MyPayments(ctx context.Context, req *connect.Request[devicepb.MyPaymentsRequest]) (*connect.Response[devicepb.MyPaymentsResponse], error) {
	if h.usecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("portal usecase unavailable"))
	}
	cust, err := h.usecase.Authenticate(ctx, req.Msg.Token)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("session token invalid or expired"))
	}
	limit := int(req.Msg.Limit)
	if limit <= 0 {
		limit = 50
	}
	payments, err := h.usecase.Payments(ctx, cust.ID, limit)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	entries := make([]*devicepb.PaymentEntry, len(payments))
	for i, p := range payments {
		entries[i] = &devicepb.PaymentEntry{
			Id:              p.ID,
			PaymentNo:       p.PaymentNo,
			Amount:          p.Amount,
			PaymentDateUnix: p.PaymentDate.Unix(),
			ScanMethod:      p.ScanMethod,
			Reference:       p.Reference,
		}
	}
	return connect.NewResponse(&devicepb.MyPaymentsResponse{Payments: entries}), nil
}

func (h *PortalConnectHandler) Logout(ctx context.Context, req *connect.Request[devicepb.PortalLogoutRequest]) (*connect.Response[devicepb.PortalLogoutResponse], error) {
	if h.usecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("portal usecase unavailable"))
	}
	_ = h.usecase.Logout(ctx, req.Msg.Token)
	return connect.NewResponse(&devicepb.PortalLogoutResponse{Message: "logged out"}), nil
}

func toProtoCustomerPortal(c *domainCustomer.Customer) *devicepb.Customer {
	if c == nil {
		return nil
	}
	return &devicepb.Customer{
		Id: c.ID, TenantId: c.TenantID, CustomerCode: c.CustomerCode,
		Name: c.Name, Phone: c.Phone, Email: c.Email, Address: c.Address,
		Status: c.Status,
	}
}

func toProtoInvoicePortalList(list []domainBilling.Invoice) []*devicepb.Invoice {
	out := make([]*devicepb.Invoice, len(list))
	for i, inv := range list {
		out[i] = &devicepb.Invoice{
			Id: inv.ID, InvoiceNumber: inv.InvoiceNumber, Period: inv.Period,
			Total: inv.Total, PaidAmount: inv.PaidAmount, Status: inv.Status,
			DueDateUnix: inv.DueDate.Unix(), ManualPaymentCode: inv.ManualPaymentCode,
		}
	}
	return out
}
