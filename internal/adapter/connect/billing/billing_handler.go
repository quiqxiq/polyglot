package billing

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	"github.com/quixiq/polyglot/pkg/response"
)

const (
	StatusUnpaid    = domainBilling.StatusUnpaid
	StatusPaid      = domainBilling.StatusPaid
	StatusActive    = domainSub.StatusActive
	StatusCancelled = domainSub.StatusCancelled
)

type BillingConnectHandler struct {
	invoiceUC      *billingUC.InvoiceUseCase
	subscriptionUC *billingUC.SubscriptionUseCase
}

func NewBillingConnectHandler(
	invUC *billingUC.InvoiceUseCase,
	subUC *billingUC.SubscriptionUseCase,
) *BillingConnectHandler {
	return &BillingConnectHandler{
		invoiceUC:      invUC,
		subscriptionUC: subUC,
	}
}

func (h *BillingConnectHandler) ListInvoices(ctx context.Context, req *connect.Request[devicepb.ListInvoicesRequest]) (*connect.Response[devicepb.ListInvoicesResponse], error) {
	if h.invoiceUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("invoice usecase unavailable"))
	}

	invoices, err := h.invoiceUC.ListInvoices(ctx, req.Msg.CustomerId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListInvoicesResponse{
		Invoices: toProtoInvoiceList(invoices),
	}), nil
}

func (h *BillingConnectHandler) GetInvoice(ctx context.Context, req *connect.Request[devicepb.GetInvoiceRequest]) (*connect.Response[devicepb.GetInvoiceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invoice id is required"))
	}
	if h.invoiceUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("invoice usecase unavailable"))
	}

	inv, err := h.invoiceUC.GetInvoice(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetInvoiceResponse{
		Invoice: toProtoInvoice(&inv),
	}), nil
}

func (h *BillingConnectHandler) CreateInvoice(ctx context.Context, req *connect.Request[devicepb.CreateInvoiceRequest]) (*connect.Response[devicepb.CreateInvoiceResponse], error) {
	if req.Msg.CustomerId == "" || req.Msg.Amount <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer_id and valid amount are required"))
	}
	if h.invoiceUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("invoice usecase unavailable"))
	}

	now := time.Now()
	dueDate := time.Unix(req.Msg.DueDateUnix, 0)
	if req.Msg.DueDateUnix <= 0 {
		dueDate = now.AddDate(0, 0, 7)
	}

	invID := fmt.Sprintf("inv-%d", now.UnixNano()%100000)
	newInv := domainBilling.Invoice{
		ID:                invID,
		InvoiceNumber:     fmt.Sprintf("INV-%s-%05d", now.Format("200601"), now.UnixNano()%100000),
		CustomerID:        req.Msg.CustomerId,
		Period:            dueDate.Format("2006-01"),
		Subtotal:          req.Msg.Amount,
		Total:             req.Msg.Amount,
		Status:            StatusUnpaid,
		QRPayload:         fmt.Sprintf("polyglot://invoice/%s", invID),
		ManualPaymentCode: fmt.Sprintf("PAY-%06d", now.UnixNano()%1000000),
		DueDate:           dueDate,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	created, err := h.invoiceUC.CreateInvoice(ctx, newInv)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CreateInvoiceResponse{
		Invoice: toProtoInvoice(&created),
	}), nil
}

func (h *BillingConnectHandler) PayInvoice(ctx context.Context, req *connect.Request[devicepb.PayInvoiceRequest]) (*connect.Response[devicepb.PayInvoiceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invoice id is required"))
	}
	if h.invoiceUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("invoice usecase unavailable"))
	}

	paid, err := h.invoiceUC.PayInvoice(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.PayInvoiceResponse{
		Invoice: toProtoInvoice(&paid),
		Message: "Payment successful",
	}), nil
}

func (h *BillingConnectHandler) ListSubscriptions(ctx context.Context, req *connect.Request[devicepb.ListSubscriptionsRequest]) (*connect.Response[devicepb.ListSubscriptionsResponse], error) {
	if h.subscriptionUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}

	subs, err := h.subscriptionUC.ListSubscriptions(ctx, req.Msg.CustomerId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListSubscriptionsResponse{
		Subscriptions: toProtoSubscriptionList(subs),
	}), nil
}

func (h *BillingConnectHandler) CreateSubscription(ctx context.Context, req *connect.Request[devicepb.CreateSubscriptionRequest]) (*connect.Response[devicepb.CreateSubscriptionResponse], error) {
	if req.Msg.CustomerId == "" || req.Msg.PlanId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer_id and plan_id are required"))
	}
	if h.subscriptionUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}

	now := time.Now()
	subID := fmt.Sprintf("sub-%d", now.UnixNano()%100000)
	endDate := now.AddDate(1, 0, 0)
	newSub := domainSub.Subscription{
		ID:           subID,
		CustomerID:   req.Msg.CustomerId,
		PlanID:       req.Msg.PlanId,
		Status:       StatusActive,
		ServiceType:  domainPlan.TypePPPoE,
		BillingCycle: domainSub.CycleMonthly,
		BillingDay:   int(now.Day()),
		StartDate:    now,
		EndDate:      &endDate,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if req.Msg.Price > 0 {
		newSub.CustomPrice = &req.Msg.Price
	}

	created, err := h.subscriptionUC.CreateSubscription(ctx, newSub)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CreateSubscriptionResponse{
		Subscription: toProtoSubscription(&created),
	}), nil
}

func (h *BillingConnectHandler) CancelSubscription(ctx context.Context, req *connect.Request[devicepb.CancelSubscriptionRequest]) (*connect.Response[devicepb.CancelSubscriptionResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("subscription id is required"))
	}
	if h.subscriptionUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}

	cancelled, err := h.subscriptionUC.CancelSubscription(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CancelSubscriptionResponse{
		Subscription: toProtoSubscription(&cancelled),
		Message:      "Subscription cancelled successfully",
	}), nil
}
