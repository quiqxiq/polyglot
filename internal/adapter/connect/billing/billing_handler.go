package billing

import (
	"context"
	"fmt"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	"github.com/quixiq/polyglot/pkg/response"
)

// BillingConnectHandler implements the billing ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type BillingConnectHandler struct {
	invoiceUC   *billingUC.InvoiceUseCase
	checkoutUC  *billingUC.CheckoutUseCase
	subUC       *billingUC.SubscriptionUseCase
	lifecycleUC *billingUC.SubscriptionLifecycleUseCase
	planUC      *billingUC.PlanUseCase
	runBilling  *billingUC.RunBillingUseCase
	manageSubUC *billingUC.ManageSubscriptionUseCase
}

// NewBillingConnectHandler constructs a billing ConnectRPC handler.
func NewBillingConnectHandler(
	invUC *billingUC.InvoiceUseCase,
	checkoutUC *billingUC.CheckoutUseCase,
	subUC *billingUC.SubscriptionUseCase,
	lifecycleUC *billingUC.SubscriptionLifecycleUseCase,
	planUC *billingUC.PlanUseCase,
	runBilling *billingUC.RunBillingUseCase,
	manageSubUC *billingUC.ManageSubscriptionUseCase,
) *BillingConnectHandler {
	return &BillingConnectHandler{
		invoiceUC:   invUC,
		checkoutUC:  checkoutUC,
		subUC:       subUC,
		lifecycleUC: lifecycleUC,
		planUC:      planUC,
		runBilling:  runBilling,
		manageSubUC: manageSubUC,
	}
}

// ─── Faktur ─────────────────────────────────────────────────────────────

// ListInvoices returns invoices matching the request filters.
func (h *BillingConnectHandler) ListInvoices(ctx context.Context, req *connect.Request[devicepb.ListInvoicesRequest]) (*connect.Response[devicepb.ListInvoicesResponse], error) {
	if h.invoiceUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("invoice usecase unavailable"))
	}
	invoices, err := h.invoiceUC.ListInvoices(ctx, req.Msg.CustomerId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	// Filter status bila diminta
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

// ─── Kasir: resolve & bayar ─────────────────────────────────────────────

// CashierResolve resolves an invoice for cashier payment.
func (h *BillingConnectHandler) CashierResolve(ctx context.Context, req *connect.Request[devicepb.CashierResolveRequest]) (*connect.Response[devicepb.CashierResolveResponse], error) {
	if h.checkoutUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("checkout usecase unavailable"))
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
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("checkout usecase unavailable"))
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

// ─── Langganan & lifecycle ──────────────────────────────────────────────

func (h *BillingConnectHandler) ListSubscriptions(ctx context.Context, req *connect.Request[devicepb.ListSubscriptionsRequest]) (*connect.Response[devicepb.ListSubscriptionsResponse], error) {
	if h.subUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}
	subs, err := h.subUC.ListSubscriptions(ctx, req.Msg.CustomerId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListSubscriptionsResponse{
		Subscriptions: toProtoSubscriptionList(subs),
	}), nil
}

func (h *BillingConnectHandler) GetSubscription(ctx context.Context, req *connect.Request[devicepb.GetSubscriptionRequest]) (*connect.Response[devicepb.GetSubscriptionResponse], error) {
	if h.subUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription usecase unavailable"))
	}
	sub, err := h.subUC.GetSubscription(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
	}), nil
}

func (h *BillingConnectHandler) ChangePlan(ctx context.Context, req *connect.Request[devicepb.ChangePlanRequest]) (*connect.Response[devicepb.ChangePlanResponse], error) {
	if h.lifecycleUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("lifecycle usecase unavailable"))
	}
	sub, err := h.lifecycleUC.ChangePlan(ctx, req.Msg.SubscriptionId, req.Msg.NewPlanId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ChangePlanResponse{
		Subscription: toProtoSubscription(&sub),
	}), nil
}

func (h *BillingConnectHandler) SuspendSubscription(ctx context.Context, req *connect.Request[devicepb.SuspendSubscriptionRequest]) (*connect.Response[devicepb.SuspendSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("lifecycle usecase unavailable"))
	}
	sub, err := h.lifecycleUC.Suspend(ctx, req.Msg.SubscriptionId, req.Msg.Reason)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SuspendSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
	}), nil
}

func (h *BillingConnectHandler) ResumeSubscription(ctx context.Context, req *connect.Request[devicepb.ResumeSubscriptionRequest]) (*connect.Response[devicepb.ResumeSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("lifecycle usecase unavailable"))
	}
	sub, err := h.lifecycleUC.Resume(ctx, req.Msg.SubscriptionId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ResumeSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
	}), nil
}

func (h *BillingConnectHandler) TerminateSubscription(ctx context.Context, req *connect.Request[devicepb.TerminateSubscriptionRequest]) (*connect.Response[devicepb.TerminateSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("lifecycle usecase unavailable"))
	}
	sub, err := h.lifecycleUC.Terminate(ctx, req.Msg.SubscriptionId, req.Msg.Reason)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.TerminateSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
	}), nil
}

func (h *BillingConnectHandler) ActivateSubscription(ctx context.Context, req *connect.Request[devicepb.ActivateSubscriptionRequest]) (*connect.Response[devicepb.ActivateSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("lifecycle usecase unavailable"))
	}
	sub, err := h.lifecycleUC.Activate(ctx, req.Msg.SubscriptionId, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ActivateSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
	}), nil
}

// ─── CRUD langganan ─────────────────────────────────────────────────────

func (h *BillingConnectHandler) CreateSubscription(ctx context.Context, req *connect.Request[devicepb.CreateSubscriptionRequest]) (*connect.Response[devicepb.CreateSubscriptionResponse], error) {
	if h.manageSubUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription manager usecase unavailable"))
	}
	msg := req.Msg
	var deviceID *string
	if msg.DeviceId != "" {
		deviceID = &msg.DeviceId
	}
	var customPrice *float64
	if msg.CustomPrice > 0 {
		customPrice = &msg.CustomPrice
	}
	sub, err := h.manageSubUC.Create(ctx, billingUC.CreateInput{
		CustomerID:     msg.CustomerId,
		PlanID:         msg.PlanId,
		DeviceID:       deviceID,
		ServiceType:    msg.ServiceType,
		RemoteUsername: msg.RemoteUsername,
		RemotePassword: msg.RemotePassword,
		CustomPrice:    customPrice,
		BillingCycle:   msg.BillingCycle,
		BillingDay:     int(msg.BillingDay),
		Notes:          msg.Notes,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreateSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
	}), nil
}

func (h *BillingConnectHandler) UpdateSubscription(ctx context.Context, req *connect.Request[devicepb.UpdateSubscriptionRequest]) (*connect.Response[devicepb.UpdateSubscriptionResponse], error) {
	if h.manageSubUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription manager usecase unavailable"))
	}
	msg := req.Msg
	in := billingUC.UpdateInput{Notes: &msg.Notes}
	if msg.RemoteUsername != "" {
		in.RemoteUsername = &msg.RemoteUsername
	}
	if msg.RemotePassword != "" {
		in.RemotePassword = &msg.RemotePassword
	}
	if msg.CustomPrice > 0 {
		in.CustomPrice = &msg.CustomPrice
	}
	if msg.BillingCycle != "" {
		in.BillingCycle = &msg.BillingCycle
	}
	if msg.BillingDay != 0 {
		day := int(msg.BillingDay)
		in.BillingDay = &day
	}
	if msg.DeviceId != "" {
		in.DeviceID = &msg.DeviceId
	}
	sub, err := h.manageSubUC.Update(ctx, msg.Id, in)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdateSubscriptionResponse{
		Subscription: toProtoSubscription(&sub),
	}), nil
}

func (h *BillingConnectHandler) DeleteSubscription(ctx context.Context, req *connect.Request[devicepb.DeleteSubscriptionRequest]) (*connect.Response[devicepb.DeleteSubscriptionResponse], error) {
	if h.manageSubUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription manager usecase unavailable"))
	}
	if err := h.manageSubUC.Delete(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteSubscriptionResponse{
		Message: "langganan dihapus",
	}), nil
}

// ─── Paket layanan ──────────────────────────────────────────────────────

func (h *BillingConnectHandler) ListPlans(ctx context.Context, req *connect.Request[devicepb.ListPlansRequest]) (*connect.Response[devicepb.ListPlansResponse], error) {
	if h.planUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("plan usecase unavailable"))
	}
	plans, err := h.planUC.List(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListPlansResponse{
		Plans: toProtoPlanList(plans),
	}), nil
}

func (h *BillingConnectHandler) GetPlan(ctx context.Context, req *connect.Request[devicepb.GetPlanRequest]) (*connect.Response[devicepb.GetPlanResponse], error) {
	if h.planUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("plan usecase unavailable"))
	}
	pl, err := h.planUC.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetPlanResponse{
		Plan: toProtoPlan(&pl),
	}), nil
}

func (h *BillingConnectHandler) CreatePlan(ctx context.Context, req *connect.Request[devicepb.CreatePlanRequest]) (*connect.Response[devicepb.CreatePlanResponse], error) {
	if h.planUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("plan usecase unavailable"))
	}
	p := fromProtoPlan(req.Msg.Plan)
	created, err := h.planUC.Create(ctx, p)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreatePlanResponse{
		Plan: toProtoPlan(&created),
	}), nil
}

func (h *BillingConnectHandler) UpdatePlan(ctx context.Context, req *connect.Request[devicepb.UpdatePlanRequest]) (*connect.Response[devicepb.UpdatePlanResponse], error) {
	if h.planUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("plan usecase unavailable"))
	}
	p := fromProtoPlan(req.Msg.Plan)
	updated, err := h.planUC.Update(ctx, p)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdatePlanResponse{
		Plan: toProtoPlan(&updated),
	}), nil
}

func (h *BillingConnectHandler) DeletePlan(ctx context.Context, req *connect.Request[devicepb.DeletePlanRequest]) (*connect.Response[devicepb.DeletePlanResponse], error) {
	if h.planUC == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("plan usecase unavailable"))
	}
	if err := h.planUC.Delete(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeletePlanResponse{
		Message: "Plan deleted successfully",
	}), nil
}

// ─── Generator tagihan ──────────────────────────────────────────────────

func (h *BillingConnectHandler) GenerateInvoices(ctx context.Context, req *connect.Request[devicepb.GenerateInvoicesRequest]) (*connect.Response[devicepb.GenerateInvoicesResponse], error) {
	if h.runBilling == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("billing runner unavailable"))
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

type domainBillingInvoiceAlias = domainBilling.Invoice
