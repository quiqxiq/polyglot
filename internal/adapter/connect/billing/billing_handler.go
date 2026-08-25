package billing

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	"github.com/quixiq/polyglot/pkg/response"
)

const (
	StatusUnpaid   = domainBilling.StatusUnpaid
	StatusPaid     = domainBilling.StatusPaid
	StatusActive   = domainSub.StatusActive
	StatusIsolated = domainSub.StatusIsolated
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
		ID:         invID,
		CustomerID: req.Msg.CustomerId,
		Amount:     req.Msg.Amount,
		Status:     StatusUnpaid,
		DueDate:    dueDate,
		CreatedAt:  now,
		UpdatedAt:  now,
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
	subID := fmt.Sprintf("sub-%s", uuid.NewString()[:8])
	newSub := domainSub.Subscription{
		ID:             subID,
		CustomerID:     req.Msg.CustomerId,
		PlanID:         req.Msg.PlanId,
		DeviceID:       req.Msg.DeviceId,
		RemoteUsername: req.Msg.RemoteUsername,
		BillingDay:     int(req.Msg.BillingDay),
		Status:         domainSub.StatusPendingProvision,
		StartDate:      now,
		CreatedAt:      now,
	}

	created, err := h.subscriptionUC.CreateSubscription(ctx, newSub, req.Msg.Password)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CreateSubscriptionResponse{
		Subscription: toProtoSubscription(&created),
	}), nil
}

func (h *BillingConnectHandler) GetSubscription(ctx context.Context, req *connect.Request[devicepb.GetSubscriptionRequest]) (*connect.Response[devicepb.GetSubscriptionResponse], error) {
	if h.subscriptionUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}
	sub, err := h.subscriptionUC.GetSubscription(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
	}), nil
}

func (h *BillingConnectHandler) UpdateSubscription(ctx context.Context, req *connect.Request[devicepb.UpdateSubscriptionRequest]) (*connect.Response[devicepb.UpdateSubscriptionResponse], error) {
	if h.subscriptionUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}
	current, err := h.subscriptionUC.GetSubscription(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	updated, err := h.subscriptionUC.UpdateSubscription(ctx, current, req.Msg.PlanId, req.Msg.RemoteUsername, req.Msg.Notes, int(req.Msg.BillingDay))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdateSubscriptionResponse{
		Subscription: toProtoSubscription(&updated),
		Message:      "subscription updated",
	}), nil
}

func (h *BillingConnectHandler) DeleteSubscription(ctx context.Context, req *connect.Request[devicepb.DeleteSubscriptionRequest]) (*connect.Response[devicepb.DeleteSubscriptionResponse], error) {
	if h.subscriptionUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}
	if err := h.subscriptionUC.DeleteSubscription(ctx, req.Msg.Id, req.Msg.Deprovision); err != nil {
		return nil, response.MapDomainError(err)
	}
	msg := "subscription terminated"
	if req.Msg.Deprovision {
		msg = "subscription terminated and account removed from router"
	}
	return connect.NewResponse(&devicepb.DeleteSubscriptionResponse{Message: msg}), nil
}

func (h *BillingConnectHandler) IsolateSubscription(ctx context.Context, req *connect.Request[devicepb.IsolateSubscriptionRequest]) (*connect.Response[devicepb.IsolateSubscriptionResponse], error) {
	if h.subscriptionUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}
	sub, err := h.subscriptionUC.IsolateSubscription(ctx, req.Msg.Id, req.Msg.Reason)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.IsolateSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
		Message:      "subscriber isolated; web traffic redirected to payment portal",
	}), nil
}

func (h *BillingConnectHandler) UnisolateSubscription(ctx context.Context, req *connect.Request[devicepb.UnisolateSubscriptionRequest]) (*connect.Response[devicepb.UnisolateSubscriptionResponse], error) {
	if h.subscriptionUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}
	sub, err := h.subscriptionUC.UnisolateSubscription(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UnisolateSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
		Message:      "subscriber restored to active service",
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
