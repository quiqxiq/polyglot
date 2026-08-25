package isolir

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/response"
)

// IsolationConnectHandler implements polyglot.v1.IsolationService.
type IsolationConnectHandler struct {
	useCase *networkUC.ManageIsolationUseCase
}

func NewIsolationConnectHandler(uc *networkUC.ManageIsolationUseCase) *IsolationConnectHandler {
	return &IsolationConnectHandler{useCase: uc}
}

func (h *IsolationConnectHandler) SetupIsolation(ctx context.Context, req *connect.Request[devicepb.SetupIsolationRequest]) (*connect.Response[devicepb.SetupIsolationResponse], error) {
	if req.Msg.DeviceId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}
	res, cfg, err := h.useCase.Setup(ctx, req.Msg.DeviceId, fromProtoOverride(req.Msg.Override))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SetupIsolationResponse{
		PoolExisted:     res.PoolExisted,
		ProfileExisted:  res.ProfileExisted,
		CreatedNatIds:   res.CreatedNATIDs,
		NatRuleIds:      res.NATRuleIDs,
		EffectiveConfig: toProtoConfig(cfg),
		Message:         "isolation infrastructure ready; web traffic from the isolation pool is redirected to the payment portal",
	}), nil
}

func (h *IsolationConnectHandler) GetIsolationStatus(ctx context.Context, req *connect.Request[devicepb.GetIsolationStatusRequest]) (*connect.Response[devicepb.GetIsolationStatusResponse], error) {
	if req.Msg.DeviceId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}
	ins, cfg, warnings, err := h.useCase.Status(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetIsolationStatusResponse{
		PoolExists:           ins.PoolExists,
		PoolName:             ins.PoolName,
		PoolRanges:           ins.PoolRanges,
		ProfileExists:        ins.ProfileExists,
		ProfileName:          ins.ProfileName,
		ProfileRateLimit:     ins.ProfileRateLimit,
		ProfileRemoteAddress: ins.ProfileRemoteAddress,
		NatRules:             toProtoNATRules(ins.NATRules),
		EffectiveConfig:      toProtoConfig(cfg),
		Warnings:             warnings,
	}), nil
}

func (h *IsolationConnectHandler) RemoveIsolation(ctx context.Context, req *connect.Request[devicepb.RemoveIsolationRequest]) (*connect.Response[devicepb.RemoveIsolationResponse], error) {
	if req.Msg.DeviceId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}
	if err := h.useCase.Remove(ctx, req.Msg.DeviceId); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RemoveIsolationResponse{
		Message: "isolation infrastructure removed from router",
	}), nil
}

// ─── mappers ─────────────────────────────────────────────────────────────

func toProtoConfig(c port.IsolirConfig) *devicepb.IsolirConfigMsg {
	return &devicepb.IsolirConfigMsg{
		ProfileName:    c.ProfileName,
		PoolName:       c.PoolName,
		PoolRange:      c.PoolRange,
		PortalIp:       c.PortalIP,
		PortalHttpPort: c.PortalHTTPPort,
		RedirectPorts:  c.RedirectPorts,
	}
}

func fromProtoOverride(pb *devicepb.IsolirConfigMsg) networkUC.IsolirConfigOverride {
	if pb == nil {
		return networkUC.IsolirConfigOverride{}
	}
	return networkUC.IsolirConfigOverride{
		ProfileName:    pb.ProfileName,
		PoolName:       pb.PoolName,
		PoolRange:      pb.PoolRange,
		PortalIP:       pb.PortalIp,
		PortalHTTPPort: pb.PortalHttpPort,
		RedirectPorts:  pb.RedirectPorts,
	}
}

func toProtoNATRules(rules []port.IsolirNATRuleStatus) []*devicepb.NATRuleStatusMsg {
	out := make([]*devicepb.NATRuleStatusMsg, len(rules))
	for i := range rules {
		r := rules[i]
		out[i] = &devicepb.NATRuleStatusMsg{
			Port:        r.Port,
			Exists:      r.Exists,
			RosId:       r.RosID,
			Action:      r.Action,
			ToAddresses: r.ToAddresses,
			ToPorts:     r.ToPorts,
			Comment:     r.Comment,
		}
	}
	return out
}
