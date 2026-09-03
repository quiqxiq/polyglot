package plan

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
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
func (h *PlanConnectHandler) ListPlans(ctx context.Context, req *connect.Request[devicepb.ListPlansRequest]) (*connect.Response[devicepb.ListPlansResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	plans, err := h.planUC.List(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListPlansResponse{
		Plans: ToProtoPlanList(plans),
	}), nil
}

// GetPlan retrieves a single plan by identifier.
func (h *PlanConnectHandler) GetPlan(ctx context.Context, req *connect.Request[devicepb.GetPlanRequest]) (*connect.Response[devicepb.GetPlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	pl, err := h.planUC.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetPlanResponse{
		Plan: ToProtoPlan(&pl),
	}), nil
}

// CreatePlan creates a new plan.
func (h *PlanConnectHandler) CreatePlan(ctx context.Context, req *connect.Request[devicepb.CreatePlanRequest]) (*connect.Response[devicepb.CreatePlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	p := FromProtoPlan(req.Msg.Plan)
	created, err := h.planUC.Create(ctx, p, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreatePlanResponse{
		Plan: ToProtoPlan(&created),
	}), nil
}

// UpdatePlan updates an existing plan.
func (h *PlanConnectHandler) UpdatePlan(ctx context.Context, req *connect.Request[devicepb.UpdatePlanRequest]) (*connect.Response[devicepb.UpdatePlanResponse], error) {
	if h.planUC == nil {
		return nil, response.Unavailable("plan usecase unavailable")
	}
	p := FromProtoPlan(req.Msg.Plan)
	updated, err := h.planUC.Update(ctx, p, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdatePlanResponse{
		Plan: ToProtoPlan(&updated),
	}), nil
}

// DeletePlan deletes an existing plan.
func (h *PlanConnectHandler) DeletePlan(ctx context.Context, req *connect.Request[devicepb.DeletePlanRequest]) (*connect.Response[devicepb.DeletePlanResponse], error) {
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
