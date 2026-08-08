package billing

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
)

type BillingConnectHandler struct{}

func NewBillingConnectHandler() *BillingConnectHandler {
	return &BillingConnectHandler{}
}

func (h *BillingConnectHandler) ListInvoices(ctx context.Context, req *connect.Request[devicepb.ListInvoicesRequest]) (*connect.Response[devicepb.ListInvoicesResponse], error) {
	now := time.Now()
	mockInvoices := []*devicepb.Invoice{
		{
			Id:            "inv-1001",
			CustomerId:    req.Msg.CustomerId,
			Amount:        150000.0,
			Status:        "UNPAID",
			DueDateUnix:   now.AddDate(0, 0, 7).Unix(),
			CreatedAtUnix: now.Unix(),
		},
	}
	return connect.NewResponse(&devicepb.ListInvoicesResponse{Invoices: mockInvoices}), nil
}

func (h *BillingConnectHandler) GetInvoice(ctx context.Context, req *connect.Request[devicepb.GetInvoiceRequest]) (*connect.Response[devicepb.GetInvoiceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invoice id is required"))
	}

	now := time.Now()
	return connect.NewResponse(&devicepb.GetInvoiceResponse{
		Invoice: &devicepb.Invoice{
			Id:            req.Msg.Id,
			CustomerId:    "cust-1",
			Amount:        150000.0,
			Status:        "UNPAID",
			DueDateUnix:   now.AddDate(0, 0, 7).Unix(),
			CreatedAtUnix: now.Unix(),
		},
	}), nil
}

func (h *BillingConnectHandler) CreateInvoice(ctx context.Context, req *connect.Request[devicepb.CreateInvoiceRequest]) (*connect.Response[devicepb.CreateInvoiceResponse], error) {
	if req.Msg.CustomerId == "" || req.Msg.Amount <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer_id and valid amount are required"))
	}

	now := time.Now()
	inv := &devicepb.Invoice{
		Id:            fmt.Sprintf("inv-%d", now.UnixNano()%10000),
		CustomerId:    req.Msg.CustomerId,
		Amount:        req.Msg.Amount,
		Status:        "UNPAID",
		DueDateUnix:   req.Msg.DueDateUnix,
		CreatedAtUnix: now.Unix(),
	}

	return connect.NewResponse(&devicepb.CreateInvoiceResponse{Invoice: inv}), nil
}

func (h *BillingConnectHandler) PayInvoice(ctx context.Context, req *connect.Request[devicepb.PayInvoiceRequest]) (*connect.Response[devicepb.PayInvoiceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invoice id is required"))
	}

	now := time.Now()
	inv := &devicepb.Invoice{
		Id:            req.Msg.Id,
		CustomerId:    "cust-1",
		Amount:        150000.0,
		Status:        "PAID",
		DueDateUnix:   now.Unix(),
		CreatedAtUnix: now.AddDate(0, 0, -5).Unix(),
	}

	return connect.NewResponse(&devicepb.PayInvoiceResponse{
		Invoice: inv,
		Message: "Payment successful",
	}), nil
}

func (h *BillingConnectHandler) ListSubscriptions(ctx context.Context, req *connect.Request[devicepb.ListSubscriptionsRequest]) (*connect.Response[devicepb.ListSubscriptionsResponse], error) {
	now := time.Now()
	subs := []*devicepb.Subscription{
		{
			Id:            "sub-101",
			CustomerId:    req.Msg.CustomerId,
			PlanId:        "plan-home-50mbps",
			Status:        "ACTIVE",
			StartDateUnix: now.AddDate(0, -1, 0).Unix(),
			EndDateUnix:   now.AddDate(0, 11, 0).Unix(),
			Price:         250000.0,
		},
	}
	return connect.NewResponse(&devicepb.ListSubscriptionsResponse{Subscriptions: subs}), nil
}

func (h *BillingConnectHandler) CreateSubscription(ctx context.Context, req *connect.Request[devicepb.CreateSubscriptionRequest]) (*connect.Response[devicepb.CreateSubscriptionResponse], error) {
	if req.Msg.CustomerId == "" || req.Msg.PlanId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer_id and plan_id are required"))
	}

	now := time.Now()
	sub := &devicepb.Subscription{
		Id:            fmt.Sprintf("sub-%d", now.UnixNano()%10000),
		CustomerId:    req.Msg.CustomerId,
		PlanId:        req.Msg.PlanId,
		Status:        "ACTIVE",
		StartDateUnix: now.Unix(),
		EndDateUnix:   now.AddDate(1, 0, 0).Unix(),
		Price:         req.Msg.Price,
	}

	return connect.NewResponse(&devicepb.CreateSubscriptionResponse{Subscription: sub}), nil
}

func (h *BillingConnectHandler) CancelSubscription(ctx context.Context, req *connect.Request[devicepb.CancelSubscriptionRequest]) (*connect.Response[devicepb.CancelSubscriptionResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("subscription id is required"))
	}

	now := time.Now()
	sub := &devicepb.Subscription{
		Id:            req.Msg.Id,
		CustomerId:    "cust-1",
		PlanId:        "plan-home-50mbps",
		Status:        "CANCELLED",
		StartDateUnix: now.AddDate(0, -1, 0).Unix(),
		EndDateUnix:   now.Unix(),
		Price:         250000.0,
	}

	return connect.NewResponse(&devicepb.CancelSubscriptionResponse{
		Subscription: sub,
		Message:      "Subscription cancelled successfully",
	}), nil
}

func NewBillingServiceHandler() (string, http.Handler) {
	handler := NewBillingConnectHandler()
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
