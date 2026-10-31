package approval

import (
	"context"
	"encoding/json"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	domainApproval "github.com/quixiq/polyglot/internal/domain/approval"
	"github.com/quixiq/polyglot/internal/port"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectHandler implements polyglot.v1.ApprovalService.
type ConnectHandler struct {
	repo       port.ApprovalRepository
	decider    *networkUC.DecisionUseCase
	authorizer port.DeviceAuthorizer
}

// NewConnectHandler builds the approval handler.
func NewConnectHandler(repo port.ApprovalRepository, decider *networkUC.DecisionUseCase, authorizer port.DeviceAuthorizer) *ConnectHandler {
	return &ConnectHandler{repo: repo, decider: decider, authorizer: authorizer}
}

func (h *ConnectHandler) ListApprovals(ctx context.Context, req *connect.Request[devicepb.ListApprovalsRequest]) (*connect.Response[devicepb.ListApprovalsResponse], error) {
	items, total, err := h.repo.List(ctx, port.ApprovalFilter{
		TenantID: "tenant-default", DeviceID: req.Msg.GetDeviceId(),
		Status: req.Msg.GetStatus(), Limit: int(req.Msg.GetLimit()), Offset: int(req.Msg.GetOffset()),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*devicepb.Approval, len(items))
	for i := range items {
		out[i] = toProtoApproval(items[i])
	}
	return connect.NewResponse(&devicepb.ListApprovalsResponse{Approvals: out, Total: int32(total)}), nil
}

func (h *ConnectHandler) GetApproval(ctx context.Context, req *connect.Request[devicepb.GetApprovalRequest]) (*connect.Response[devicepb.GetApprovalResponse], error) {
	item, err := h.repo.FindByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetApprovalResponse{Approval: toProtoApproval(item)}), nil
}

func (h *ConnectHandler) DecideApproval(ctx context.Context, req *connect.Request[devicepb.DecideApprovalRequest]) (*connect.Response[devicepb.DecideApprovalResponse], error) {
	userID, roles, ok := iauth.IdentityFromContext(ctx)
	if !ok || userID == 0 {
		return nil, response.Unauthenticated("user session required")
	}
	existing, err := h.repo.FindByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	if err := h.checkDevice(ctx, userID, roles, existing.DeviceID); err != nil {
		return nil, err
	}
	decision := decisionFromProto(req.Msg.GetDecision())
	if decision == "" {
		return nil, response.InvalidArgument("invalid decision")
	}
	item, err := h.decider.Decide(ctx, networkUC.DecideInput{
		ApprovalID: req.Msg.GetId(), Decision: decision, ActorID: userID,
		EditedCommand: req.Msg.GetEditedCommand(), EditedArgs: req.Msg.GetEditedArgs(),
		Reason: req.Msg.GetReason(), Revision: int(req.Msg.GetRevision()),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DecideApprovalResponse{Approval: toProtoApproval(item)}), nil
}

func (h *ConnectHandler) CancelApproval(ctx context.Context, req *connect.Request[devicepb.CancelApprovalRequest]) (*connect.Response[devicepb.CancelApprovalResponse], error) {
	userID, roles, ok := iauth.IdentityFromContext(ctx)
	if !ok || userID == 0 {
		return nil, response.Unauthenticated("user session required")
	}
	existing, err := h.repo.FindByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	if err := h.checkDevice(ctx, userID, roles, existing.DeviceID); err != nil {
		return nil, err
	}
	item, err := h.decider.Cancel(ctx, req.Msg.GetId(), userID)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CancelApprovalResponse{Approval: toProtoApproval(item)}), nil
}

func (h *ConnectHandler) checkDevice(ctx context.Context, userID uint, roles []string, deviceID string) error {
	if h.authorizer == nil || deviceID == "" {
		return nil
	}
	ok, err := h.authorizer.CanAccessDevice(ctx, userID, roles, deviceID)
	if err != nil {
		return response.MapDomainError(err)
	}
	if !ok {
		return response.PermissionDenied("no access to this device")
	}
	return nil
}

func decisionFromProto(d devicepb.ApprovalDecision) networkUC.Decision {
	switch d {
	case devicepb.ApprovalDecision_APPROVAL_DECISION_APPROVE:
		return networkUC.DecisionApprove
	case devicepb.ApprovalDecision_APPROVAL_DECISION_REJECT:
		return networkUC.DecisionReject
	case devicepb.ApprovalDecision_APPROVAL_DECISION_EDIT:
		return networkUC.DecisionEdit
	case devicepb.ApprovalDecision_APPROVAL_DECISION_EDIT_AND_APPROVE:
		return networkUC.DecisionEditAndApprove
	default:
		return ""
	}
}

func toProtoApproval(a domainApproval.Request) *devicepb.Approval {
	var decided, executed int64
	if a.DecidedAt != nil {
		decided = a.DecidedAt.Unix()
	}
	if a.ExecutedAt != nil {
		executed = a.ExecutedAt.Unix()
	}
	var approvedBy int32
	if a.ApprovedBy != nil {
		approvedBy = int32(*a.ApprovedBy)
	}
	return &devicepb.Approval{
		Id: a.ID, DeviceId: a.DeviceID, ActorType: string(a.ActorType), ActorId: a.ActorID,
		SessionId: a.SessionID, Command: a.Command, Args: a.Args, Revision: int32(a.Revision),
		EditedCommand: valueOr(a.EditedCommand), EditedArgs: a.EditedArgs, EditCount: int32(a.EditCount),
		Status: string(a.Status), Reason: a.Reason, RestorePointId: a.RestorePointID,
		DiffJson: string(a.Diff), ValidationJson: string(a.Validation), ExecutionResultJson: string(a.ExecutionResult),
		CreatedAtUnix: a.CreatedAt.Unix(), ExpiresAtUnix: a.ExpiresAt.Unix(),
		DecidedAtUnix: decided, ExecutedAtUnix: executed, ApprovedBy: approvedBy,
	}
}

func valueOr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

var _ = json.Marshal
var _ = time.Now
