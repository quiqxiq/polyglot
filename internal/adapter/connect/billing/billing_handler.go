package billing

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	billingUsecase "github.com/quixiq/polyglot/internal/usecase/billing"
)

const (
	StatusUnpaid    = domainBilling.StatusUnpaid
	StatusPaid      = domainBilling.StatusPaid
	StatusActive    = domainSub.StatusActive
	StatusCancelled = domainSub.StatusCancelled
)

type BillingConnectHandler struct {
	invoiceUsecase      *billingUsecase.InvoiceUsecase
	subscriptionUsecase *billingUsecase.SubscriptionUsecase
}

func NewBillingConnectHandler(
	invUsecase *billingUsecase.InvoiceUsecase,
	subUsecase *billingUsecase.SubscriptionUsecase,
) *BillingConnectHandler {
	return &BillingConnectHandler{
		invoiceUsecase:      invUsecase,
		subscriptionUsecase: subUsecase,
	}
}

func (h *BillingConnectHandler) ListInvoices(ctx context.Context, req *connect.Request[devicepb.ListInvoicesRequest]) (*connect.Response[devicepb.ListInvoicesResponse], error) {
	if h.invoiceUsecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("invoice usecase unavailable"))
	}

	invoices, err := h.invoiceUsecase.ListInvoices(ctx, req.Msg.CustomerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbInvoices := make([]*devicepb.Invoice, len(invoices))
	for i, inv := range invoices {
		pbInvoices[i] = &devicepb.Invoice{
			Id:            inv.ID,
			CustomerId:    inv.CustomerID,
			Amount:        inv.Amount,
			Status:        inv.Status,
			DueDateUnix:   inv.DueDate.Unix(),
			CreatedAtUnix: inv.CreatedAt.Unix(),
		}
	}

	return connect.NewResponse(&devicepb.ListInvoicesResponse{Invoices: pbInvoices}), nil
}

func (h *BillingConnectHandler) GetInvoice(ctx context.Context, req *connect.Request[devicepb.GetInvoiceRequest]) (*connect.Response[devicepb.GetInvoiceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invoice id is required"))
	}
	if h.invoiceUsecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("invoice usecase unavailable"))
	}

	inv, err := h.invoiceUsecase.GetInvoice(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&devicepb.GetInvoiceResponse{
		Invoice: &devicepb.Invoice{
			Id:            inv.ID,
			CustomerId:    inv.CustomerID,
			Amount:        inv.Amount,
			Status:        inv.Status,
			DueDateUnix:   inv.DueDate.Unix(),
			CreatedAtUnix: inv.CreatedAt.Unix(),
		},
	}), nil
}

func (h *BillingConnectHandler) CreateInvoice(ctx context.Context, req *connect.Request[devicepb.CreateInvoiceRequest]) (*connect.Response[devicepb.CreateInvoiceResponse], error) {
	if req.Msg.CustomerId == "" || req.Msg.Amount <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer_id and valid amount are required"))
	}
	if h.invoiceUsecase == nil {
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

	created, err := h.invoiceUsecase.CreateInvoice(ctx, newInv)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.CreateInvoiceResponse{
		Invoice: &devicepb.Invoice{
			Id:            created.ID,
			CustomerId:    created.CustomerID,
			Amount:        created.Amount,
			Status:        created.Status,
			DueDateUnix:   created.DueDate.Unix(),
			CreatedAtUnix: created.CreatedAt.Unix(),
		},
	}), nil
}

func (h *BillingConnectHandler) PayInvoice(ctx context.Context, req *connect.Request[devicepb.PayInvoiceRequest]) (*connect.Response[devicepb.PayInvoiceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invoice id is required"))
	}
	if h.invoiceUsecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("invoice usecase unavailable"))
	}

	paid, err := h.invoiceUsecase.PayInvoice(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.PayInvoiceResponse{
		Invoice: &devicepb.Invoice{
			Id:            paid.ID,
			CustomerId:    paid.CustomerID,
			Amount:        paid.Amount,
			Status:        paid.Status,
			DueDateUnix:   paid.DueDate.Unix(),
			CreatedAtUnix: paid.CreatedAt.Unix(),
		},
		Message: "Payment successful",
	}), nil
}

func (h *BillingConnectHandler) ListSubscriptions(ctx context.Context, req *connect.Request[devicepb.ListSubscriptionsRequest]) (*connect.Response[devicepb.ListSubscriptionsResponse], error) {
	if h.subscriptionUsecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}

	subs, err := h.subscriptionUsecase.ListSubscriptions(ctx, req.Msg.CustomerId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbSubs := make([]*devicepb.Subscription, len(subs))
	for i, sub := range subs {
		pbSubs[i] = &devicepb.Subscription{
			Id:            sub.ID,
			CustomerId:    sub.CustomerID,
			PlanId:        sub.PlanID,
			Status:        sub.Status,
			StartDateUnix: sub.StartDate.Unix(),
			EndDateUnix:   sub.EndDate.Unix(),
			Price:         sub.Price,
		}
	}

	return connect.NewResponse(&devicepb.ListSubscriptionsResponse{Subscriptions: pbSubs}), nil
}

func (h *BillingConnectHandler) CreateSubscription(ctx context.Context, req *connect.Request[devicepb.CreateSubscriptionRequest]) (*connect.Response[devicepb.CreateSubscriptionResponse], error) {
	if req.Msg.CustomerId == "" || req.Msg.PlanId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer_id and plan_id are required"))
	}
	if h.subscriptionUsecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}

	now := time.Now()
	subID := fmt.Sprintf("sub-%d", now.UnixNano()%100000)
	newSub := domainSub.Subscription{
		ID:         subID,
		CustomerID: req.Msg.CustomerId,
		PlanID:     req.Msg.PlanId,
		Status:     StatusActive,
		StartDate:  now,
		EndDate:    now.AddDate(1, 0, 0),
		Price:      req.Msg.Price,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	created, err := h.subscriptionUsecase.CreateSubscription(ctx, newSub)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.CreateSubscriptionResponse{
		Subscription: &devicepb.Subscription{
			Id:            created.ID,
			CustomerId:    created.CustomerID,
			PlanId:        created.PlanID,
			Status:        created.Status,
			StartDateUnix: created.StartDate.Unix(),
			EndDateUnix:   created.EndDate.Unix(),
			Price:         created.Price,
		},
	}), nil
}

func (h *BillingConnectHandler) CancelSubscription(ctx context.Context, req *connect.Request[devicepb.CancelSubscriptionRequest]) (*connect.Response[devicepb.CancelSubscriptionResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("subscription id is required"))
	}
	if h.subscriptionUsecase == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}

	cancelled, err := h.subscriptionUsecase.CancelSubscription(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.CancelSubscriptionResponse{
		Subscription: &devicepb.Subscription{
			Id:            cancelled.ID,
			CustomerId:    cancelled.CustomerID,
			PlanId:        cancelled.PlanID,
			Status:        cancelled.Status,
			StartDateUnix: cancelled.StartDate.Unix(),
			EndDateUnix:   cancelled.EndDate.Unix(),
			Price:         cancelled.Price,
		},
		Message: "Subscription cancelled successfully",
	}), nil
}

func NewBillingServiceHandler(
	invUsecase *billingUsecase.InvoiceUsecase,
	subUsecase *billingUsecase.SubscriptionUsecase,
) (string, http.Handler) {
	handler := NewBillingConnectHandler(invUsecase, subUsecase)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(connectJSONCodec{})

	serviceName := "polyglot.v1.BillingService"
	mux.Handle("/"+serviceName+"/ListInvoices", connect.NewUnaryHandler("/"+serviceName+"/ListInvoices", handler.ListInvoices, codecOpt))
	mux.Handle("/"+serviceName+"/GetInvoice", connect.NewUnaryHandler("/"+serviceName+"/GetInvoice", handler.GetInvoice, codecOpt))
	mux.Handle("/"+serviceName+"/CreateInvoice", connect.NewUnaryHandler("/"+serviceName+"/CreateInvoice", handler.CreateInvoice, codecOpt))
	mux.Handle("/"+serviceName+"/PayInvoice", connect.NewUnaryHandler("/"+serviceName+"/PayInvoice", handler.PayInvoice, codecOpt))
	mux.Handle("/"+serviceName+"/ListSubscriptions", connect.NewUnaryHandler("/"+serviceName+"/ListSubscriptions", handler.ListSubscriptions, codecOpt))
	mux.Handle("/"+serviceName+"/CreateSubscription", connect.NewUnaryHandler("/"+serviceName+"/CreateSubscription", handler.CreateSubscription, codecOpt))
	mux.Handle("/"+serviceName+"/CancelSubscription", connect.NewUnaryHandler("/"+serviceName+"/CancelSubscription", handler.CancelSubscription, codecOpt))

	return "/" + serviceName + "/", mux
}
