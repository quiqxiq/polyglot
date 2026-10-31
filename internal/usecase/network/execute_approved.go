package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/approval"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/port"
)

// ExecutionDeps wires the approved-execution pipeline.
type ExecutionDeps struct {
	Approvals       port.ApprovalRepository
	RestorePoints   port.RestorePointRepository
	Devices         port.DeviceRepository
	Driver          func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
	Ops             func(ctx context.Context, deviceID string) (port.RouterOps, error)
	CommandLogs     port.CommandLogRepository
	Restore         *RestorePointUseCase
	WatchdogEnabled bool
	WatchdogSeconds int
	Now             func() time.Time
}

// ExecutionUseCase executes exactly one approved mutation.
type ExecutionUseCase struct {
	deps ExecutionDeps
}

// NewExecutionUseCase builds the execution usecase.
func NewExecutionUseCase(deps ExecutionDeps) *ExecutionUseCase {
	if deps.Now == nil {
		deps.Now = func() time.Time { return time.Now().UTC() }
	}
	return &ExecutionUseCase{deps: deps}
}

// Execute claims an APPROVED request and runs it once. It never retries
// automatically: an execution timeout leaves the row FAILED/UNKNOWN and the
// operator must READ-verify and decide.
func (uc *ExecutionUseCase) Execute(ctx context.Context, approvalID string) (approval.Request, error) {
	now := uc.deps.Now()
	req, ok, err := uc.deps.Approvals.ClaimForExecution(ctx, approvalID, now)
	if err != nil {
		return approval.Request{}, err
	}
	if !ok {
		cur, ferr := uc.deps.Approvals.FindByID(ctx, approvalID)
		if ferr != nil {
			return approval.Request{}, ferr
		}
		if cur.IsExpired(now) && cur.Status == approval.StatusApproved {
			cur.Status = approval.StatusExpired
			cur.UpdatedAt = now
			_, _ = uc.deps.Approvals.UpdateIfRevision(ctx, cur, cur.Revision)
			return cur, approval.ErrExpired
		}
		return cur, fmt.Errorf("%w: approval is %s or already claimed", approval.ErrInvalidTransition, cur.Status)
	}

	fail := func(err error) (approval.Request, error) {
		req.Status = approval.StatusFailed
		req.UpdatedAt = uc.deps.Now()
		req.ExecutionResult = mustJSON(map[string]any{"error": err.Error(), "rollback_recommended": true})
		_, _ = uc.deps.Approvals.UpdateIfRevision(ctx, req, req.Revision)
		uc.log(req, "FAILED", err.Error())
		return req, err
	}

	rp, err := uc.deps.RestorePoints.FindByID(ctx, req.RestorePointID)
	if err != nil {
		return fail(fmt.Errorf("restore point %s: %w", req.RestorePointID, err))
	}
	if !rp.Usable(uc.deps.Now()) {
		return fail(fmt.Errorf("%w: restore point %s is not available", restorepoint.ErrNotAvailable, rp.ID))
	}
	driver, err := uc.deps.Driver(ctx, req.DeviceID)
	if err != nil {
		return fail(err)
	}
	cmd := command.Command{Raw: req.EffectiveCommand(), Args: req.EffectiveArgs()}
	class := driver.Classify(cmd)
	if class.IsDenied() {
		return fail(fmt.Errorf("%w: %s", command.ErrDenied, class.DenyReason))
	}
	if command.Decide(class) != command.DecisionRequireApproval {
		return fail(fmt.Errorf("%w: approved command no longer classifies as MUTATION", approval.ErrInvalidTransition))
	}

	if req.ExecutionNonce == "" {
		req.ExecutionNonce = fmt.Sprintf("%d-%s", uc.deps.Now().UnixNano(), req.ID)
	}
	ops, err := uc.deps.Ops(ctx, req.DeviceID)
	if err != nil {
		return fail(err)
	}
	var watchdog port.RollbackWatchdog
	if uc.deps.WatchdogEnabled && uc.deps.Restore != nil {
		if w, ok := ops.(port.RollbackWatchdog); ok {
			if rc, _, oerr := uc.deps.Restore.OpenArtifact(ctx, rp.ID, restorepoint.ArtifactInverseScript); oerr == nil {
				script, _ := io.ReadAll(io.LimitReader(rc, 1<<20))
				_ = rc.Close()
				seconds := uc.deps.WatchdogSeconds
				if seconds <= 0 {
					seconds = 600
				}
				if aerr := w.ArmWatchdog(ctx, req.ID, script, seconds); aerr != nil {
					return fail(fmt.Errorf("arm watchdog: %w", aerr))
				}
				watchdog = w
			}
		}
	}
	start := uc.deps.Now()
	result, err := ops.ExecuteRaw(ctx, cmd.Raw, cmd.Args)
	if err != nil {
		return fail(fmt.Errorf("execute: %w", err))
	}
	duration := uc.deps.Now().Sub(start).Milliseconds()

	menu, _ := SplitCommandPath(cmd.Raw)
	verify := map[string][]map[string]string{}
	var verifyErr error
	if menu != "" {
		verify, verifyErr = ops.PostVerify(ctx, []string{menu})
	}
	req.Status = approval.StatusExecuted
	req.ExecutedAt = ptrTime(uc.deps.Now())
	req.UpdatedAt = *req.ExecutedAt
	payload := map[string]any{
		"output": result.Output, "rows": result.Rows, "meta": result.Meta,
		"duration_ms": duration, "verify": verify,
	}
	if verifyErr != nil {
		req.Status = approval.StatusFailed
		payload["error"] = "post-verify failed: " + verifyErr.Error()
		payload["rollback_recommended"] = true
	}
	req.ExecutionResult = mustJSON(payload)
	if _, uerr := uc.deps.Approvals.UpdateIfRevision(ctx, req, req.Revision); uerr != nil {
		return req, uerr
	}
	status := "SUCCESS"
	summary := truncate(result.Output, 2000)
	if req.Status == approval.StatusFailed {
		status = "FAILED"
		summary = "post-verify failed"
	}
	if watchdog != nil && verifyErr == nil {
		_ = watchdog.DisarmWatchdog(ctx, req.ID)
	}
	if watchdog != nil && verifyErr != nil {
		payload["watchdog_armed"] = true
		req.ExecutionResult = mustJSON(payload)
		_, _ = uc.deps.Approvals.UpdateIfRevision(ctx, req, req.Revision)
	}
	uc.log(req, status, summary)
	if verifyErr != nil {
		return req, fmt.Errorf("post-verify: %w", verifyErr)
	}
	return req, nil
}

func (uc *ExecutionUseCase) log(req approval.Request, status, summary string) {
	if uc.deps.CommandLogs == nil {
		return
	}
	entry := commandLog(req.TenantID, req.DeviceID, string(req.ActorType), req.ActorID, req.SessionID, "MUTATION", req.EffectiveCommand(), status, summary)
	entry.ApprovalID = req.ID
	entry.RestorePointID = req.RestorePointID
	entry.DurationMS = 0
	_ = uc.deps.CommandLogs.Write(context.Background(), entry)
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func ptrTime(t time.Time) *time.Time { return &t }

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…(truncated)"
}
