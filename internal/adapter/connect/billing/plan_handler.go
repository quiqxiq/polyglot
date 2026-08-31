package billing

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListPlans lists service plans.
func (h *BillingConnectHandler) ListPlans(ctx context.Context, req *connect.Request[devicepb.ListPlansRequest]) (*connect.Response[devicepb.ListPlansResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	plans, err := h.planUC.List(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListPlansResponse{
		Plans: toProtoPlanList(plans),
	}), nil
}

// GetPlan retrieves a single plan by identifier.
func (h *BillingConnectHandler) GetPlan(ctx context.Context, req *connect.Request[devicepb.GetPlanRequest]) (*connect.Response[devicepb.GetPlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	pl, err := h.planUC.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetPlanResponse{
		Plan: toProtoPlan(&pl),
	}), nil
}

// CreatePlan creates a new plan.
func (h *BillingConnectHandler) CreatePlan(ctx context.Context, req *connect.Request[devicepb.CreatePlanRequest]) (*connect.Response[devicepb.CreatePlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	p := fromProtoPlan(req.Msg.Plan)
	created, err := h.planUC.Create(ctx, p, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreatePlanResponse{
		Plan: toProtoPlan(&created),
	}), nil
}

// UpdatePlan updates an existing plan.
func (h *BillingConnectHandler) UpdatePlan(ctx context.Context, req *connect.Request[devicepb.UpdatePlanRequest]) (*connect.Response[devicepb.UpdatePlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	p := fromProtoPlan(req.Msg.Plan)
	updated, err := h.planUC.Update(ctx, p, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdatePlanResponse{
		Plan: toProtoPlan(&updated),
	}), nil
}

// DeletePlan deletes an existing plan.
func (h *BillingConnectHandler) DeletePlan(ctx context.Context, req *connect.Request[devicepb.DeletePlanRequest]) (*connect.Response[devicepb.DeletePlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	if err := h.planUC.Delete(ctx, req.Msg.Id, req.Msg.DeviceId); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeletePlanResponse{
		Message: "Plan deleted successfully",
	}), nil
}
