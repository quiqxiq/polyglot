package plan

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
	"github.com/quixiq/polyglot/pkg/response"
)

// PlanConnectHandler implements polyglot.v1.PlanService over ConnectRPC.
type PlanConnectHandler struct {
	useCase *planUC.ManagePlansUseCase
}

func NewPlanConnectHandler(uc *planUC.ManagePlansUseCase) *PlanConnectHandler {
	return &PlanConnectHandler{useCase: uc}
}

func (h *PlanConnectHandler) CreatePlan(ctx context.Context, req *connect.Request[devicepb.CreatePlanRequest]) (*connect.Response[devicepb.CreatePlanResponse], error) {
	if req.Msg.Plan == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("plan is required"))
	}
	created, err := h.useCase.Create(ctx, fromProtoPlan(req.Msg.Plan))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreatePlanResponse{
		Plan:    toProtoPlan(&created),
		Message: "plan created",
	}), nil
}

func (h *PlanConnectHandler) GetPlan(ctx context.Context, req *connect.Request[devicepb.GetPlanRequest]) (*connect.Response[devicepb.GetPlanResponse], error) {
	p, err := h.useCase.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetPlanResponse{Plan: toProtoPlan(&p)}), nil
}

func (h *PlanConnectHandler) ListPlans(ctx context.Context, req *connect.Request[devicepb.ListPlansRequest]) (*connect.Response[devicepb.ListPlansResponse], error) {
	plans, err := h.useCase.List(ctx, req.Msg.ServiceType, req.Msg.ActiveOnly, int(req.Msg.Limit))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*devicepb.Plan, len(plans))
	for i := range plans {
		out[i] = toProtoPlan(&plans[i])
	}
	return connect.NewResponse(&devicepb.ListPlansResponse{Plans: out}), nil
}

func (h *PlanConnectHandler) UpdatePlan(ctx context.Context, req *connect.Request[devicepb.UpdatePlanRequest]) (*connect.Response[devicepb.UpdatePlanResponse], error) {
	if req.Msg.Plan == nil || req.Msg.Plan.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("plan.id is required"))
	}
	updated, err := h.useCase.Update(ctx, fromProtoPlan(req.Msg.Plan))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdatePlanResponse{
		Plan:    toProtoPlan(&updated),
		Message: "plan updated",
	}), nil
}

func (h *PlanConnectHandler) DeletePlan(ctx context.Context, req *connect.Request[devicepb.DeletePlanRequest]) (*connect.Response[devicepb.DeletePlanResponse], error) {
	if err := h.useCase.Delete(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeletePlanResponse{Message: "plan deleted"}), nil
}

// ─── mappers ─────────────────────────────────────────────────────────────

func toProtoPlan(p *domainPlan.Plan) *devicepb.Plan {
	if p == nil {
		return nil
	}
	return &devicepb.Plan{
		Id:            p.ID,
		TenantId:      p.TenantID,
		Name:          p.Name,
		ServiceType:   p.ServiceType,
		RateDownKbps:  int32(p.RateDownKbps),
		RateUpKbps:    int32(p.RateUpKbps),
		Price:         p.Price,
		IpPoolName:    p.IPPoolName,
		ParentQueue:   p.ParentQueue,
		AddressList:   p.AddressList,
		SharedUsers:   int32(p.SharedUsers),
		IsActive:      p.IsActive,
		Description:   p.Description,
		CreatedAtUnix: p.CreatedAt.Unix(),
		UpdatedAtUnix: p.UpdatedAt.Unix(),
	}
}

func fromProtoPlan(pb *devicepb.Plan) domainPlan.Plan {
	if pb == nil {
		return domainPlan.Plan{}
	}
	shared := int(pb.SharedUsers)
	if shared <= 0 {
		shared = 1
	}
	return domainPlan.Plan{
		ID:           pb.Id,
		TenantID:     pb.TenantId,
		Name:         pb.Name,
		ServiceType:  pb.ServiceType,
		RateDownKbps: int(pb.RateDownKbps),
		RateUpKbps:   int(pb.RateUpKbps),
		Price:        pb.Price,
		IPPoolName:   pb.IpPoolName,
		ParentQueue:  pb.ParentQueue,
		AddressList:  pb.AddressList,
		SharedUsers:  shared,
		IsActive:     pb.IsActive,
		Description:  pb.Description,
	}
}
