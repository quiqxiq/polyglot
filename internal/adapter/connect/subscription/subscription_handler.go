package subscription

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	subUC "github.com/quixiq/polyglot/internal/usecase/subscription"
	"github.com/quixiq/polyglot/pkg/response"
)

// SubscriptionConnectHandler implements the SubscriptionService ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type SubscriptionConnectHandler struct {
	subUC       *subUC.ManageSubscriptionUseCase
	lifecycleUC *subUC.LifecycleUseCase
}

// NewSubscriptionConnectHandler constructs a SubscriptionService ConnectRPC handler.
func NewSubscriptionConnectHandler(
	subUC *subUC.ManageSubscriptionUseCase,
	lifecycleUC *subUC.LifecycleUseCase,
) *SubscriptionConnectHandler {
	return &SubscriptionConnectHandler{
		subUC:       subUC,
		lifecycleUC: lifecycleUC,
	}
}

func (h *SubscriptionConnectHandler) toEnrichedProto(ctx context.Context, sub *domainSub.Subscription) *devicepb.Subscription {
	if sub == nil {
		return nil
	}
	if h.subUC != nil {
		detail := h.subUC.Enrich(ctx, *sub)
		return ToProtoSubscriptionDetail(&detail)
	}
	return ToProtoSubscription(sub)
}

// ListSubscriptions returns subscriptions for a given customer.
func (h *SubscriptionConnectHandler) ListSubscriptions(ctx context.Context, req *connect.Request[devicepb.ListSubscriptionsRequest]) (*connect.Response[devicepb.ListSubscriptionsResponse], error) {
	if h.subUC == nil {
		return nil, response.Unavailable("subscription usecase unavailable")
	}
	details, err := h.subUC.ListSubscriptions(ctx, req.Msg.CustomerId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListSubscriptionsResponse{
		Subscriptions: ToProtoSubscriptionDetailList(details),
	}), nil
}

// GetSubscription returns a single subscription by identifier.
func (h *SubscriptionConnectHandler) GetSubscription(ctx context.Context, req *connect.Request[devicepb.GetSubscriptionRequest]) (*connect.Response[devicepb.GetSubscriptionResponse], error) {
	if h.subUC == nil {
		return nil, response.Unavailable("subscription usecase unavailable")
	}
	detail, err := h.subUC.GetSubscription(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetSubscriptionResponse{
		Subscription: ToProtoSubscriptionDetail(&detail),
	}), nil
}

// ChangePlan changes the plan assigned to a subscription.
func (h *SubscriptionConnectHandler) ChangePlan(ctx context.Context, req *connect.Request[devicepb.ChangePlanRequest]) (*connect.Response[devicepb.ChangePlanResponse], error) {
	if h.lifecycleUC == nil {
		return nil, response.Unavailable("lifecycle usecase unavailable")
	}
	sub, err := h.lifecycleUC.ChangePlan(ctx, req.Msg.SubscriptionId, req.Msg.NewPlanId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ChangePlanResponse{
		Subscription: h.toEnrichedProto(ctx, &sub),
	}), nil
}

// SuspendSubscription suspends a subscription.
func (h *SubscriptionConnectHandler) SuspendSubscription(ctx context.Context, req *connect.Request[devicepb.SuspendSubscriptionRequest]) (*connect.Response[devicepb.SuspendSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, response.Unavailable("lifecycle usecase unavailable")
	}
	sub, err := h.lifecycleUC.Suspend(ctx, req.Msg.SubscriptionId, req.Msg.Reason)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SuspendSubscriptionResponse{
		Subscription: h.toEnrichedProto(ctx, &sub),
	}), nil
}

// ResumeSubscription resumes a suspended subscription.
func (h *SubscriptionConnectHandler) ResumeSubscription(ctx context.Context, req *connect.Request[devicepb.ResumeSubscriptionRequest]) (*connect.Response[devicepb.ResumeSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, response.Unavailable("lifecycle usecase unavailable")
	}
	sub, err := h.lifecycleUC.Resume(ctx, req.Msg.SubscriptionId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ResumeSubscriptionResponse{
		Subscription: h.toEnrichedProto(ctx, &sub),
	}), nil
}

// TerminateSubscription terminates a subscription.
func (h *SubscriptionConnectHandler) TerminateSubscription(ctx context.Context, req *connect.Request[devicepb.TerminateSubscriptionRequest]) (*connect.Response[devicepb.TerminateSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, response.Unavailable("lifecycle usecase unavailable")
	}
	sub, err := h.lifecycleUC.Terminate(ctx, req.Msg.SubscriptionId, req.Msg.Reason)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.TerminateSubscriptionResponse{
		Subscription: h.toEnrichedProto(ctx, &sub),
	}), nil
}

// ActivateSubscription activates a subscription with a target device.
func (h *SubscriptionConnectHandler) ActivateSubscription(ctx context.Context, req *connect.Request[devicepb.ActivateSubscriptionRequest]) (*connect.Response[devicepb.ActivateSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, response.Unavailable("lifecycle usecase unavailable")
	}
	sub, err := h.lifecycleUC.Activate(ctx, req.Msg.SubscriptionId, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ActivateSubscriptionResponse{
		Subscription: h.toEnrichedProto(ctx, &sub),
	}), nil
}

// IsolateSubscription isolates a customer subscription.
func (h *SubscriptionConnectHandler) IsolateSubscription(ctx context.Context, req *connect.Request[devicepb.IsolateSubscriptionRequest]) (*connect.Response[devicepb.IsolateSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, response.Unavailable("lifecycle usecase unavailable")
	}
	sub, err := h.lifecycleUC.Isolate(ctx, req.Msg.SubscriptionId, req.Msg.Reason)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.IsolateSubscriptionResponse{
		Subscription: h.toEnrichedProto(ctx, &sub),
		Message:      "Pelanggan berhasil diisolir",
	}), nil
}

// RestoreSubscription restores an isolated customer subscription.
func (h *SubscriptionConnectHandler) RestoreSubscription(ctx context.Context, req *connect.Request[devicepb.RestoreSubscriptionRequest]) (*connect.Response[devicepb.RestoreSubscriptionResponse], error) {
	if h.lifecycleUC == nil {
		return nil, response.Unavailable("lifecycle usecase unavailable")
	}
	sub, err := h.lifecycleUC.Restore(ctx, req.Msg.SubscriptionId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RestoreSubscriptionResponse{
		Subscription: h.toEnrichedProto(ctx, &sub),
		Message:      "Layanan pelanggan berhasil dipulihkan",
	}), nil
}

// CreateSubscription creates a new subscription.
func (h *SubscriptionConnectHandler) CreateSubscription(ctx context.Context, req *connect.Request[devicepb.CreateSubscriptionRequest]) (*connect.Response[devicepb.CreateSubscriptionResponse], error) {
	if h.subUC == nil {
		return nil, response.Unavailable("subscription manager usecase unavailable")
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
	var localAddr, remoteAddr, rateLimit string
	var pppoeConf *domainSub.PPPoESubscriptionConfig
	if msg.PppoeConfig != nil {
		localAddr = msg.PppoeConfig.LocalAddress
		remoteAddr = msg.PppoeConfig.RemoteAddress
		rateLimit = msg.PppoeConfig.RateLimit
		pppoeConf = &domainSub.PPPoESubscriptionConfig{
			LocalAddress:  msg.PppoeConfig.LocalAddress,
			RemoteAddress: msg.PppoeConfig.RemoteAddress,
			CallerID:      msg.PppoeConfig.CallerId,
			Routes:        msg.PppoeConfig.Routes,
			RateLimit:     msg.PppoeConfig.RateLimit,
			RouterProfile: msg.PppoeConfig.RouterProfile,
		}
	}
	var hotConf *domainSub.HotspotSubscriptionConfig
	if msg.HotspotConfig != nil {
		if remoteAddr == "" {
			remoteAddr = msg.HotspotConfig.IpAddress
		}
		if rateLimit == "" {
			rateLimit = msg.HotspotConfig.RateLimit
		}
		hotConf = &domainSub.HotspotSubscriptionConfig{
			Server:        msg.HotspotConfig.Server,
			MacAddress:    msg.HotspotConfig.MacAddress,
			IPAddress:     msg.HotspotConfig.IpAddress,
			RateLimit:     msg.HotspotConfig.RateLimit,
			RouterProfile: msg.HotspotConfig.RouterProfile,
			LimitUptime:   msg.HotspotConfig.LimitUptime,
			LimitBytes:    msg.HotspotConfig.LimitBytes,
		}
	}
	sub, err := h.subUC.Create(ctx, subUC.CreateInput{
		CustomerID:     msg.CustomerId,
		PlanID:         msg.PlanId,
		DeviceID:       deviceID,
		ServiceType:    msg.ServiceType,
		RemoteUsername: msg.RemoteUsername,
		RemotePassword: msg.RemotePassword,
		LocalAddress:   localAddr,
		RemoteAddress:  remoteAddr,
		RateLimit:      rateLimit,
		CustomPrice:    customPrice,
		BillingCycle:   msg.BillingCycle,
		BillingDay:     int(msg.BillingDay),
		Notes:          msg.Notes,
		PPPoE:          pppoeConf,
		Hotspot:        hotConf,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreateSubscriptionResponse{
		Subscription: h.toEnrichedProto(ctx, &sub),
	}), nil
}

// UpdateSubscription updates an existing subscription.
func (h *SubscriptionConnectHandler) UpdateSubscription(ctx context.Context, req *connect.Request[devicepb.UpdateSubscriptionRequest]) (*connect.Response[devicepb.UpdateSubscriptionResponse], error) {
	if h.subUC == nil {
		return nil, response.Unavailable("subscription manager usecase unavailable")
	}
	msg := req.Msg
	in := subUC.UpdateInput{Notes: &msg.Notes}
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
	if msg.PppoeConfig != nil {
		if msg.PppoeConfig.LocalAddress != "" {
			in.LocalAddress = &msg.PppoeConfig.LocalAddress
		}
		if msg.PppoeConfig.RemoteAddress != "" {
			in.RemoteAddress = &msg.PppoeConfig.RemoteAddress
		}
		if msg.PppoeConfig.RateLimit != "" {
			in.RateLimit = &msg.PppoeConfig.RateLimit
		}
		in.PPPoE = &domainSub.PPPoESubscriptionConfig{
			LocalAddress:  msg.PppoeConfig.LocalAddress,
			RemoteAddress: msg.PppoeConfig.RemoteAddress,
			CallerID:      msg.PppoeConfig.CallerId,
			Routes:        msg.PppoeConfig.Routes,
			RateLimit:     msg.PppoeConfig.RateLimit,
			RouterProfile: msg.PppoeConfig.RouterProfile,
		}
	}
	if msg.HotspotConfig != nil {
		if msg.HotspotConfig.IpAddress != "" {
			in.RemoteAddress = &msg.HotspotConfig.IpAddress
		}
		if msg.HotspotConfig.RateLimit != "" {
			in.RateLimit = &msg.HotspotConfig.RateLimit
		}
		in.Hotspot = &domainSub.HotspotSubscriptionConfig{
			Server:        msg.HotspotConfig.Server,
			MacAddress:    msg.HotspotConfig.MacAddress,
			IPAddress:     msg.HotspotConfig.IpAddress,
			RateLimit:     msg.HotspotConfig.RateLimit,
			RouterProfile: msg.HotspotConfig.RouterProfile,
			LimitUptime:   msg.HotspotConfig.LimitUptime,
			LimitBytes:    msg.HotspotConfig.LimitBytes,
		}
	}
	sub, err := h.subUC.Update(ctx, msg.Id, in)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdateSubscriptionResponse{
		Subscription: h.toEnrichedProto(ctx, &sub),
	}), nil
}

// DeleteSubscription deletes a subscription.
func (h *SubscriptionConnectHandler) DeleteSubscription(ctx context.Context, req *connect.Request[devicepb.DeleteSubscriptionRequest]) (*connect.Response[devicepb.DeleteSubscriptionResponse], error) {
	if h.subUC == nil {
		return nil, response.Unavailable("subscription manager usecase unavailable")
	}
	if err := h.subUC.Delete(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteSubscriptionResponse{
		Message: "langganan dihapus",
	}), nil
}
