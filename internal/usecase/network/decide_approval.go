package network

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/domain/approval"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// Decision is the human decision applied to a PENDING approval.
type Decision string

const (
	DecisionApprove        Decision = "APPROVE"
	DecisionReject         Decision = "REJECT"
	DecisionEdit           Decision = "EDIT"
	DecisionEditAndApprove Decision = "EDIT_AND_APPROVE"
)

// DecideInput is one human decision request.
type DecideInput struct {
	ApprovalID    string
	Decision      Decision
	ActorID       uint
	EditedCommand string
	EditedArgs    map[string]string
	Reason        string
	Revision      int
}

// DecisionDeps wires the approval decision usecase.
type DecisionDeps struct {
	Approvals           port.ApprovalRepository
	RestorePoints       *RestorePointUseCase
	Devices             port.DeviceRepository
	Driver              func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
	Ops                 func(ctx context.Context, deviceID string) (port.RouterOps, error)
	Execution           *ExecutionUseCase
	RequireRejectReason bool
	Now                 func() time.Time
}

// DecisionUseCase applies approve/reject/edit transitions and triggers
// execution for approvals. It is only reachable from a human JWT session.
type DecisionUseCase struct {
	deps DecisionDeps
}

// NewDecisionUseCase builds the decision usecase.
func NewDecisionUseCase(deps DecisionDeps) *DecisionUseCase {
	if deps.Now == nil {
		deps.Now = func() time.Time { return time.Now().UTC() }
	}
	return &DecisionUseCase{deps: deps}
}

// Decide applies one decision. Approve/edit-and-approve return the final
// approval after the synchronous execution attempt.
func (uc *DecisionUseCase) Decide(ctx context.Context, in DecideInput) (approval.Request, error) {
	if in.ActorID == 0 {
		return approval.Request{}, fmt.Errorf("%w: a human user session is required", approval.ErrSelfApprove)
	}
	req, err := uc.deps.Approvals.FindByID(ctx, in.ApprovalID)
	if err != nil {
		return approval.Request{}, err
	}
	now := uc.deps.Now()
	if req.Status != approval.StatusPending {
		return req, fmt.Errorf("%w: approval is %s", approval.ErrInvalidTransition, req.Status)
	}
	if req.IsExpired(now) {
		req.Status = approval.StatusExpired
		req.UpdatedAt = now
		_, _ = uc.deps.Approvals.UpdateIfRevision(ctx, req, req.Revision)
		return req, approval.ErrExpired
	}
	if in.Revision > 0 && in.Revision != req.Revision {
		return req, approval.ErrRevisionConflict
	}

	switch in.Decision {
	case DecisionReject:
		if req.ActorType == approval.ActorAgent && in.Reason == "" && uc.deps.RequireRejectReason {
			return req, fmt.Errorf("%w: reject reason is required", approval.ErrInvalidTransition)
		}
		if in.Reason == "" && uc.deps.RequireRejectReason {
			return req, fmt.Errorf("%w: reject reason is required", approval.ErrInvalidTransition)
		}
		req.Status = approval.StatusRejected
		req.Reason = in.Reason
		req.DecidedBy = &in.ActorID
		req.DecidedAt = &now
		req.UpdatedAt = now
		if ok, uerr := uc.deps.Approvals.UpdateIfRevision(ctx, req, req.Revision); uerr != nil || !ok {
			if uerr != nil {
				return req, uerr
			}
			return req, approval.ErrRevisionConflict
		}
		return req, nil

	case DecisionEdit, DecisionEditAndApprove:
		edited, updated, err := uc.applyEdit(ctx, req, in, now)
		if err != nil {
			return req, err
		}
		if in.Decision == DecisionEdit {
			return updated, nil
		}
		req = edited
		fallthrough

	case DecisionApprove:
		return uc.approve(ctx, req, in.ActorID, now)
	default:
		return req, fmt.Errorf("%w: unknown decision %q", approval.ErrInvalidTransition, in.Decision)
	}
}

func (uc *DecisionUseCase) applyEdit(ctx context.Context, req approval.Request, in DecideInput, now time.Time) (approval.Request, approval.Request, error) {
	if in.EditedCommand == "" {
		return req, req, fmt.Errorf("%w: edited_command is required for EDIT", approval.ErrInvalidTransition)
	}
	dev, err := uc.deps.Devices.FindByID(ctx, req.DeviceID)
	if err != nil {
		return req, req, fmt.Errorf("device: %w", err)
	}
	if !dev.Enabled {
		return req, req, fmt.Errorf("device is disabled")
	}
	driver, err := uc.deps.Driver(ctx, req.DeviceID)
	if err != nil {
		return req, req, err
	}
	cmd := command.Command{Raw: in.EditedCommand, Args: cloneMap(in.EditedArgs)}
	class := driver.Classify(cmd)
	switch command.Decide(class) {
	case command.DecisionDeny:
		return req, req, fmt.Errorf("%w: %s", command.ErrDenied, class.DenyReason)
	case command.DecisionAutoApprove:
		return req, req, fmt.Errorf("%w: edited command is READ; mutations only", approval.ErrInvalidTransition)
	}
	menu, verb := SplitCommandPath(cmd.Raw)
	if menu == "" || verb == "" {
		return req, req, fmt.Errorf("%w: malformed edited command", approval.ErrInvalidTransition)
	}
	if (verb == "set" || verb == "remove" || verb == "move") && selectorID(cmd.Args) == "" {
		return req, req, fmt.Errorf("%w: edited %s requires .id/numbers selector", approval.ErrInvalidTransition, verb)
	}
	ops, err := uc.deps.Ops(ctx, req.DeviceID)
	if err != nil {
		return req, req, err
	}
	var before []map[string]string
	if verb != "add" {
		before, err = ops.ReadMenu(ctx, menu, selectorFromArgs(cmd.Args))
		if err != nil {
			return req, req, err
		}
		if len(before) == 0 {
			return req, req, fmt.Errorf("%w: selector matched no object", approval.ErrInvalidTransition)
		}
	}
	validation, err := ops.ValidateScript(ctx, renderCommand(cmd))
	if err != nil {
		return req, req, err
	}
	if validation.Supported && !validation.Valid {
		return req, req, fmt.Errorf("%w: edited command syntax invalid", approval.ErrInvalidTransition)
	}
	inverse := BuildInverse(cmd, before, "")
	if len(inverse.Commands) == 0 {
		return req, req, fmt.Errorf("%w: edited command has no deterministic rollback", approval.ErrInvalidTransition)
	}
	rp, err := uc.deps.RestorePoints.CreateForMutation(ctx, req.TenantID, req.DeviceID, fmt.Sprintf("user:%d", in.ActorID), []string{menu}, RenderInverseScript(inverse.Commands))
	if err != nil {
		return req, req, fmt.Errorf("edited restore point: %w", err)
	}
	diffJSON, _ := json.Marshal(BuildCommandDiff(cmd, before))
	valJSON, _ := json.Marshal(validation)
	edited := in.EditedCommand
	req.EditedCommand = &edited
	req.EditedArgs = cloneMap(in.EditedArgs)
	req.EditedBy = &in.ActorID
	req.EditCount++
	req.Revision++
	req.Diff = diffJSON
	req.Validation = valJSON
	req.RestorePointID = rp.ID
	req.UpdatedAt = now
	if ok, uerr := uc.deps.Approvals.UpdateIfRevision(ctx, req, req.Revision-1); uerr != nil || !ok {
		if uerr != nil {
			return req, req, uerr
		}
		return req, req, approval.ErrRevisionConflict
	}
	return req, req, nil
}

func (uc *DecisionUseCase) approve(ctx context.Context, req approval.Request, actorID uint, now time.Time) (approval.Request, error) {
	rp, err := uc.deps.RestorePoints.Get(ctx, req.RestorePointID)
	if err != nil {
		return req, fmt.Errorf("%w: %v", approval.ErrRestorePointUnavailable, err)
	}
	if !rp.Usable(now) {
		return req, approval.ErrRestorePointUnavailable
	}
	req.Status = approval.StatusApproved
	req.ApprovedBy = &actorID
	req.DecidedBy = &actorID
	req.DecidedAt = &now
	req.UpdatedAt = now
	if ok, uerr := uc.deps.Approvals.UpdateIfRevision(ctx, req, req.Revision); uerr != nil || !ok {
		if uerr != nil {
			return req, uerr
		}
		return req, approval.ErrRevisionConflict
	}
	if uc.deps.Execution == nil {
		return req, nil
	}
	return uc.deps.Execution.Execute(ctx, req.ID)
}

// Cancel marks a PENDING approval as CANCELLED. Only the requester or a
// decide-permission holder may cancel; RBAC is enforced by the handler.
func (uc *DecisionUseCase) Cancel(ctx context.Context, approvalID string, actorID uint) (approval.Request, error) {
	req, err := uc.deps.Approvals.FindByID(ctx, approvalID)
	if err != nil {
		return req, err
	}
	if req.Status != approval.StatusPending {
		return req, fmt.Errorf("%w: approval is %s", approval.ErrInvalidTransition, req.Status)
	}
	now := uc.deps.Now()
	req.Status = approval.StatusCancelled
	req.DecidedBy = &actorID
	req.DecidedAt = &now
	req.UpdatedAt = now
	if ok, uerr := uc.deps.Approvals.UpdateIfRevision(ctx, req, req.Revision); uerr != nil || !ok {
		if uerr != nil {
			return req, uerr
		}
		return req, approval.ErrRevisionConflict
	}
	return req, nil
}
