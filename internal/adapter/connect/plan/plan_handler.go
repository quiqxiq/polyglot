package plan

import (
	"context"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
	"github.com/quixiq/polyglot/pkg/response"
)

// PlanConnectHandler implements the PlanService ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type PlanConnectHandler struct {
	planUC *planUC.ManagePlanUseCase
}

// NewPlanConnectHandler constructs a PlanService ConnectRPC handler.
func NewPlanConnectHandler(planUC *planUC.ManagePlanUseCase) *PlanConnectHandler {
	return &PlanConnectHandler{
		planUC: planUC,
	}
}

// ListPlans lists service plans.
func (h *PlanConnectHandler) ListPlans(ctx context.Context, req *connect.Request[polyglotv1.ListPlansRequest]) (*connect.Response[polyglotv1.ListPlansResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	plans, err := h.planUC.List(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ListPlansResponse{
		Plans: ToProtoPlanList(plans),
	}), nil
}

// GetPlan retrieves a single plan by identifier.
func (h *PlanConnectHandler) GetPlan(ctx context.Context, req *connect.Request[polyglotv1.GetPlanRequest]) (*connect.Response[polyglotv1.GetPlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	pl, err := h.planUC.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.GetPlanResponse{
		Plan: ToProtoPlan(&pl),
	}), nil
}

// CreatePlan creates a new plan.
func (h *PlanConnectHandler) CreatePlan(ctx context.Context, req *connect.Request[polyglotv1.CreatePlanRequest]) (*connect.Response[polyglotv1.CreatePlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	p := FromProtoPlan(req.Msg.Plan)
	created, err := h.planUC.Create(ctx, p, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CreatePlanResponse{
		Plan: ToProtoPlan(&created),
	}), nil
}

// UpdatePlan updates an existing plan.
func (h *PlanConnectHandler) UpdatePlan(ctx context.Context, req *connect.Request[polyglotv1.UpdatePlanRequest]) (*connect.Response[polyglotv1.UpdatePlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	p := FromProtoPlan(req.Msg.Plan)
	updated, err := h.planUC.Update(ctx, p, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.UpdatePlanResponse{
		Plan: ToProtoPlan(&updated),
	}), nil
}

// DeletePlan deletes an existing plan.
func (h *PlanConnectHandler) DeletePlan(ctx context.Context, req *connect.Request[polyglotv1.DeletePlanRequest]) (*connect.Response[polyglotv1.DeletePlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	if err := h.planUC.Delete(ctx, req.Msg.Id, req.Msg.DeviceId); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.DeletePlanResponse{
		Message: "Plan deleted successfully",
	}), nil
}
