package portal

import (
	"context"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
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
		return nil, response.Unavailable("portal usecase unavailable")
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
		return nil, response.Unavailable("portal usecase unavailable")
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
		return nil, response.Unavailable("portal usecase unavailable")
	}
	cust, err := h.usecase.Authenticate(ctx, req.Msg.Token)
	if err != nil {
		return nil, response.Unauthenticated("session token invalid or expired")
	}
	ov, err := h.usecase.Overview(ctx, cust.ID)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	resp := &devicepb.PortalOverviewResponse{
		Customer:     toProtoCustomerPortal(&ov.Customer),
		Status:       ov.Status,
		PaymentUrl:   ov.PaymentURL,
		Subscription: toProtoSubscriptionSummary(ov.Subscription),
	}
	for _, inv := range ov.UnpaidInvoices {
		resp.UnpaidInvoices = append(resp.UnpaidInvoices, toProtoUnpaidInvoiceSummary(inv))
	}
	return connect.NewResponse(resp), nil
}

func (h *PortalConnectHandler) MyInvoices(ctx context.Context, req *connect.Request[devicepb.MyInvoicesRequest]) (*connect.Response[devicepb.MyInvoicesResponse], error) {
	if h.usecase == nil {
		return nil, response.Unavailable("portal usecase unavailable")
	}
	cust, err := h.usecase.Authenticate(ctx, req.Msg.Token)
	if err != nil {
		return nil, response.Unauthenticated("session token invalid or expired")
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
		return nil, response.Unavailable("portal usecase unavailable")
	}
	cust, err := h.usecase.Authenticate(ctx, req.Msg.Token)
	if err != nil {
		return nil, response.Unauthenticated("session token invalid or expired")
	}
	limit := int(req.Msg.Limit)
	if limit <= 0 {
		limit = 50
	}
	payments, err := h.usecase.Payments(ctx, cust.ID, limit)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.MyPaymentsResponse{
		Payments: toProtoPaymentEntries(payments),
	}), nil
}

func (h *PortalConnectHandler) Logout(ctx context.Context, req *connect.Request[devicepb.PortalLogoutRequest]) (*connect.Response[devicepb.PortalLogoutResponse], error) {
	if h.usecase == nil {
		return nil, response.Unavailable("portal usecase unavailable")
	}
	_ = h.usecase.Logout(ctx, req.Msg.Token)
	return connect.NewResponse(&devicepb.PortalLogoutResponse{Message: "logged out"}), nil
}
