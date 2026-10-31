package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/quixiq/polyglot/internal/domain/approval"
	"github.com/quixiq/polyglot/internal/port"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

// ProposalService is the subset of usecase/network used by MCP tools.
type ProposalService interface {
	ProposeCommand(ctx context.Context, in networkUC.ProposeCommandInput) (networkUC.Proposal, error)
	ProposeConfig(ctx context.Context, in networkUC.ProposeCommandInput) (networkUC.Proposal, error)
}

// ApprovalCanceler cancels a PENDING approval through the server service.
type ApprovalCanceler interface {
	Cancel(ctx context.Context, approvalID string, actorID uint) (approval.Request, error)
}

// ActiveDeviceResolver resolves the operator's active router binding.
type ActiveDeviceResolver interface {
	Resolve(ctx context.Context, tenantID, userID, sessionID string) (string, error)
}

// resolveDevice returns the explicit device_id or the active binding.
func (s *Server) resolveDevice(ctx context.Context, req *mcp.CallToolRequest, actorID uint, explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return explicit, nil
	}
	if s.activeDevices == nil {
		return "", fmt.Errorf("no active router selected: pass device_id or select a router in the Polyglot UI")
	}
	sessionID := ""
	if req != nil && req.Extra != nil && req.Extra.Header != nil {
		sessionID = req.Extra.Header.Get("Mcp-Session-Id")
		if sessionID == "" {
			sessionID = req.Extra.Header.Get("X-Polyglot-Session-ID")
		}
	}
	return s.activeDevices.Resolve(ctx, "tenant-default", fmt.Sprintf("%d", actorID), sessionID)
}

type proposeCommandArgs struct {
	DeviceID string            `json:"device_id,omitempty" jsonschema:"target device id (defaults to the operator's active router)"`
	Command  string            `json:"command" jsonschema:"vendor-native mutation command (e.g. /ip/firewall/filter/add)"`
	Args     map[string]string `json:"args,omitempty" jsonschema:"optional command attributes"`
}

type proposeOutput struct {
	Status          string `json:"status"`
	Executed        bool   `json:"executed"`
	ApprovalID      string `json:"approval_id"`
	DeviceID        string `json:"device_id"`
	Command         string `json:"command"`
	Revision        int32  `json:"revision"`
	ExpiresAtUnix   int64  `json:"expires_at_unix,omitempty"`
	RestorePointID  string `json:"restore_point_id"`
	RestoreStatus   string `json:"restore_point_status"`
	RestoreSizeByte int64  `json:"restore_point_size_bytes"`
	DiffJSON        string `json:"diff_json,omitempty"`
	ValidationJSON  string `json:"validation_json,omitempty"`
	Message         string `json:"message"`
}

func (s *Server) proposeCommand(ctx context.Context, req *mcp.CallToolRequest, args proposeCommandArgs) (*mcp.CallToolResult, proposeOutput, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(proposeOutput{Status: "error", Message: "unauthenticated: a valid user session is required"})
	}
	deviceID, err := s.resolveDevice(ctx, req, actorID, args.DeviceID)
	if err != nil {
		return toolError(proposeOutput{Status: "error", Message: err.Error()})
	}
	if s.proposals == nil {
		return toolError(proposeOutput{Status: "error", DeviceID: deviceID, Message: "approval pipeline not configured"})
	}
	res, err := s.proposals.ProposeCommand(ctx, networkUC.ProposeCommandInput{
		TenantID: "tenant-default", DeviceID: deviceID,
		ActorID: fmt.Sprintf("%d", actorID), ActorType: approval.ActorAgent,
		Command: args.Command, Args: args.Args,
	})
	if err != nil {
		return toolError(proposeOutput{Status: "error", DeviceID: deviceID, Command: args.Command, Message: err.Error()})
	}
	return toolOK(proposeOutput{
		Status: "pending_approval", Executed: false,
		ApprovalID: res.Approval.ID, DeviceID: deviceID, Command: res.Approval.Command,
		Revision: int32(res.Approval.Revision), ExpiresAtUnix: res.Approval.ExpiresAt.Unix(),
		RestorePointID: res.RestorePoint.ID, RestoreStatus: string(res.RestorePoint.Status),
		RestoreSizeByte: res.RestorePoint.TotalSizeBytes,
		DiffJSON:        string(res.Approval.Diff), ValidationJSON: string(res.Approval.Validation),
		Message: "Mutation was NOT executed. Ask the operator to approve it in the AI Ops approval inbox.",
	})
}

type proposeConfigArgs struct {
	DeviceID string            `json:"device_id,omitempty" jsonschema:"target device id (defaults to the operator's active router)"`
	Script   string            `json:"script" jsonschema:"multi-line RouterOS configuration script (max configured lines)"`
	Args     map[string]string `json:"args,omitempty" jsonschema:"optional command attributes applied to each line"`
}

func (s *Server) proposeConfig(ctx context.Context, req *mcp.CallToolRequest, args proposeConfigArgs) (*mcp.CallToolResult, proposeOutput, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(proposeOutput{Status: "error", Message: "unauthenticated: a valid user session is required"})
	}
	deviceID, err := s.resolveDevice(ctx, req, actorID, args.DeviceID)
	if err != nil {
		return toolError(proposeOutput{Status: "error", Message: err.Error()})
	}
	if s.proposals == nil {
		return toolError(proposeOutput{Status: "error", DeviceID: deviceID, Message: "approval pipeline not configured"})
	}
	res, err := s.proposals.ProposeConfig(ctx, networkUC.ProposeCommandInput{
		TenantID: "tenant-default", DeviceID: deviceID,
		ActorID: fmt.Sprintf("%d", actorID), ActorType: approval.ActorAgent,
		Command: args.Script, Args: args.Args,
	})
	if err != nil {
		return toolError(proposeOutput{Status: "error", DeviceID: deviceID, Message: err.Error()})
	}
	return toolOK(proposeOutput{
		Status: "pending_approval", Executed: false,
		ApprovalID: res.Approval.ID, DeviceID: deviceID, Command: res.Approval.Command,
		Revision: int32(res.Approval.Revision), ExpiresAtUnix: res.Approval.ExpiresAt.Unix(),
		RestorePointID: res.RestorePoint.ID, RestoreStatus: string(res.RestorePoint.Status),
		RestoreSizeByte: res.RestorePoint.TotalSizeBytes,
		DiffJSON:        string(res.Approval.Diff), ValidationJSON: string(res.Approval.Validation),
		Message: "Configuration was NOT applied. Ask the operator to approve it in the AI Ops approval inbox.",
	})
}

type approvalRefArgs struct {
	ApprovalID string `json:"approval_id" jsonschema:"approval request id"`
}

type approvalOutput struct {
	ID            string `json:"id"`
	DeviceID      string `json:"device_id"`
	Status        string `json:"status"`
	Command       string `json:"command"`
	Revision      int32  `json:"revision"`
	Reason        string `json:"reason,omitempty"`
	RestorePoint  string `json:"restore_point_id,omitempty"`
	ExecutionJSON string `json:"execution_result_json,omitempty"`
	ExpiresAtUnix int64  `json:"expires_at_unix"`
}

func toApprovalOutput(a approval.Request) approvalOutput {
	return approvalOutput{
		ID: a.ID, DeviceID: a.DeviceID, Status: string(a.Status), Command: a.EffectiveCommand(),
		Revision: int32(a.Revision), Reason: a.Reason, RestorePoint: a.RestorePointID,
		ExecutionJSON: string(a.ExecutionResult), ExpiresAtUnix: a.ExpiresAt.Unix(),
	}
}

func (s *Server) getApproval(ctx context.Context, req *mcp.CallToolRequest, args approvalRefArgs) (*mcp.CallToolResult, approvalOutput, error) {
	if _, _, ok := s.actorFrom(ctx, req); !ok {
		return toolError(approvalOutput{})
	}
	if s.approvals == nil {
		return toolError(approvalOutput{})
	}
	a, err := s.approvals.FindByID(ctx, args.ApprovalID)
	if err != nil {
		return toolError(approvalOutput{})
	}
	return toolOK(toApprovalOutput(a))
}

type listApprovalsArgs struct {
	DeviceID string `json:"device_id,omitempty" jsonschema:"filter by device"`
	Status   string `json:"status,omitempty" jsonschema:"filter by status"`
	Limit    int    `json:"limit,omitempty" jsonschema:"max rows (default 20)"`
}

type listApprovalsOutput struct {
	Approvals []approvalOutput `json:"approvals"`
	Total     int              `json:"total"`
}

func (s *Server) listApprovals(ctx context.Context, req *mcp.CallToolRequest, args listApprovalsArgs) (*mcp.CallToolResult, listApprovalsOutput, error) {
	if _, _, ok := s.actorFrom(ctx, req); !ok {
		return toolError(listApprovalsOutput{})
	}
	if s.approvals == nil {
		return toolError(listApprovalsOutput{})
	}
	limit := args.Limit
	if limit <= 0 {
		limit = 20
	}
	items, total, err := s.approvals.List(ctx, port.ApprovalFilter{
		TenantID: "tenant-default", DeviceID: args.DeviceID, Status: args.Status, Limit: limit,
	})
	if err != nil {
		return toolError(listApprovalsOutput{})
	}
	out := make([]approvalOutput, len(items))
	for i := range items {
		out[i] = toApprovalOutput(items[i])
	}
	return toolOK(listApprovalsOutput{Approvals: out, Total: total})
}

func (s *Server) cancelApproval(ctx context.Context, req *mcp.CallToolRequest, args approvalRefArgs) (*mcp.CallToolResult, approvalOutput, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(approvalOutput{})
	}
	if s.approvals == nil {
		return toolError(approvalOutput{})
	}
	a, err := s.approvals.FindByID(ctx, args.ApprovalID)
	if err != nil {
		return toolError(approvalOutput{})
	}
	if a.ActorID != fmt.Sprintf("%d", actorID) && a.RequestedBy != fmt.Sprintf("%d", actorID) {
		return toolError(approvalOutput{ID: a.ID, Status: string(a.Status)})
	}
	if s.decider == nil {
		return toolError(approvalOutput{ID: a.ID, Status: string(a.Status)})
	}
	cancelled, err := s.decider.Cancel(ctx, a.ID, actorID)
	if err != nil {
		return toolError(approvalOutput{ID: a.ID, Status: string(a.Status)})
	}
	return toolOK(toApprovalOutput(cancelled))
}
