package network

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/approval"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// ProposalConfig holds limits for the proposal pipeline.
type ProposalConfig struct {
	TTL            time.Duration
	MaxScriptLines int
	V6Mutations    bool
}

// ProposalDeps wires the proposal usecase.
type ProposalDeps struct {
	Approvals     port.ApprovalRepository
	RestorePoints *RestorePointUseCase
	Devices       port.DeviceRepository
	Driver        func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
	Ops           func(ctx context.Context, deviceID string) (port.RouterOps, error)
	CommandLogs   port.CommandLogRepository
	Config        ProposalConfig
	Now           func() time.Time
}

// ProposeCommandInput is the vendor-neutral proposal request.
type ProposeCommandInput struct {
	TenantID  string
	DeviceID  string
	SessionID string
	ActorID   string
	ActorType approval.ActorType
	Command   string
	Args      map[string]string
}

// Proposal is the result of proposing one mutation.
type Proposal struct {
	Approval     approval.Request          `json:"approval"`
	Diff         Diff                      `json:"diff"`
	Validation   port.ScriptValidation     `json:"validation"`
	RestorePoint restorepoint.RestorePoint `json:"restore_point"`
	Warnings     []string                  `json:"warnings,omitempty"`
}

// ProposalUseCase creates approval requests with diff + restore point.
type ProposalUseCase struct {
	deps ProposalDeps
}

// NewProposalUseCase builds the proposal usecase.
func NewProposalUseCase(deps ProposalDeps) *ProposalUseCase {
	if deps.Now == nil {
		deps.Now = func() time.Time { return time.Now().UTC() }
	}
	if deps.Config.TTL <= 0 {
		deps.Config.TTL = 15 * time.Minute
	}
	if deps.Config.MaxScriptLines <= 0 {
		deps.Config.MaxScriptLines = 50
	}
	return &ProposalUseCase{deps: deps}
}

// ProposeCommand validates a MUTATION, builds a diff, creates a restore point,
// persists a PENDING approval, and returns everything the UI needs.
func (uc *ProposalUseCase) ProposeCommand(ctx context.Context, in ProposeCommandInput) (Proposal, error) {
	if in.Command == "" {
		return Proposal{}, fmt.Errorf("%w: command is required", command.ErrApprovalRequired)
	}
	if in.TenantID == "" {
		in.TenantID = "tenant-default"
	}
	if in.ActorType == "" {
		in.ActorType = approval.ActorAgent
	}
	dev, err := uc.deps.Devices.FindByID(ctx, in.DeviceID)
	if err != nil {
		return Proposal{}, fmt.Errorf("device %s: %w", in.DeviceID, err)
	}
	if !dev.Enabled {
		return Proposal{}, fmt.Errorf("device %s is disabled", in.DeviceID)
	}

	driver, err := uc.deps.Driver(ctx, in.DeviceID)
	if err != nil {
		return Proposal{}, err
	}
	cmd := command.Command{Raw: in.Command, Args: cloneMap(in.Args)}
	class := driver.Classify(cmd)
	switch command.Decide(class) {
	case command.DecisionDeny:
		uc.logAttempt(ctx, in, "MUTATION", "DENIED", class.DenyReason)
		return Proposal{}, fmt.Errorf("%w: %s", command.ErrDenied, class.DenyReason)
	case command.DecisionAutoApprove:
		return Proposal{}, fmt.Errorf("%w: READ commands execute directly and must not be proposed", command.ErrApprovalRequired)
	}

	menu, verb := SplitCommandPath(cmd.Raw)
	if menu == "" || verb == "" {
		return Proposal{}, fmt.Errorf("%w: malformed command path", command.ErrApprovalRequired)
	}
	if (verb == "set" || verb == "remove" || verb == "move") && selectorID(cmd.Args) == "" {
		return Proposal{}, fmt.Errorf("%w: %s requires an explicit selector (.id or numbers) so the diff is unambiguous", command.ErrApprovalRequired, verb)
	}

	ops, err := uc.deps.Ops(ctx, in.DeviceID)
	if err != nil {
		return Proposal{}, err
	}
	caps, err := ops.Capabilities(ctx)
	if err != nil {
		return Proposal{}, err
	}
	if caps.Major > 0 && caps.Major < 7 && !uc.deps.Config.V6Mutations && in.ActorType == approval.ActorAgent {
		return Proposal{}, fmt.Errorf("%w: AI mutations on RouterOS v6 are disabled (AI_V6_MUTATIONS=false)", command.ErrDenied)
	}

	var before []map[string]string
	if verb != "add" {
		before, err = ops.ReadMenu(ctx, menu, selectorFromArgs(cmd.Args))
		if err != nil {
			return Proposal{}, fmt.Errorf("snapshot %s: %w", menu, err)
		}
		if len(before) == 0 {
			return Proposal{}, fmt.Errorf("%w: selector did not match any object in /%s", command.ErrApprovalRequired, menu)
		}
	}

	validation, err := ops.ValidateScript(ctx, renderCommand(cmd))
	if err != nil {
		return Proposal{}, err
	}
	if validation.Supported && !validation.Valid {
		return Proposal{}, fmt.Errorf("%w: syntax invalid: %s", command.ErrApprovalRequired, strings.Join(validation.Errors, "; "))
	}

	diff := BuildCommandDiff(cmd, before)
	inverse := BuildInverse(cmd, before, "")
	if len(inverse.Commands) == 0 {
		return Proposal{}, fmt.Errorf("%w: unable to derive rollback commands; proposal rejected", command.ErrApprovalRequired)
	}
	script := RenderInverseScript(inverse.Commands)

	scopes := []string{menu}
	rp, err := uc.deps.RestorePoints.CreateForMutation(ctx, in.TenantID, in.DeviceID, in.ActorID, scopes, script)
	if err != nil {
		uc.logAttempt(ctx, in, "MUTATION", "FAILED", "restore point failed: "+err.Error())
		return Proposal{}, fmt.Errorf("%w: restore point unavailable: %v", command.ErrApprovalRequired, err)
	}

	now := uc.deps.Now()
	diffJSON, _ := json.Marshal(diff)
	valJSON, _ := json.Marshal(validation)
	key := proposalKey(in)
	if existing, ferr := uc.deps.Approvals.FindByIdempotencyKey(ctx, in.TenantID, key); ferr == nil && existing.Status == approval.StatusPending {
		uc.logAttempt(ctx, in, "MUTATION", "PENDING_APPROVAL", "reused existing approval")
		return Proposal{Approval: existing, Diff: diff, Validation: validation, RestorePoint: rp, Warnings: inverse.Warnings}, nil
	}

	req := approval.Request{
		ID: idgen.New("apr"), TenantID: in.TenantID, DeviceID: in.DeviceID,
		SessionID: in.SessionID, ActorType: in.ActorType, ActorID: in.ActorID,
		RequestedBy: in.ActorID,
		Command:     cmd.Raw, Args: cmd.Args, Revision: 1,
		Diff: diffJSON, Validation: valJSON, RestorePointID: rp.ID,
		Status: approval.StatusPending, IdempotencyKey: key,
		CreatedAt: now, UpdatedAt: now, ExpiresAt: now.Add(uc.deps.Config.TTL),
	}
	if err := uc.deps.Approvals.Create(ctx, req); err != nil {
		return Proposal{}, fmt.Errorf("persist approval: %w", err)
	}
	uc.logAttempt(ctx, in, "MUTATION", "PENDING_APPROVAL", "approval "+req.ID)
	return Proposal{Approval: req, Diff: diff, Validation: validation, RestorePoint: rp, Warnings: inverse.Warnings}, nil
}

// ProposeConfig proposes a multi-line configuration change as one approval.
func (uc *ProposalUseCase) ProposeConfig(ctx context.Context, in ProposeCommandInput) (Proposal, error) {
	lines := splitScript(in.Command, uc.deps.Config.MaxScriptLines)
	if len(lines) == 0 {
		return Proposal{}, fmt.Errorf("%w: empty script", command.ErrApprovalRequired)
	}
	dev, err := uc.deps.Devices.FindByID(ctx, in.DeviceID)
	if err != nil {
		return Proposal{}, fmt.Errorf("device %s: %w", in.DeviceID, err)
	}
	if !dev.Enabled {
		return Proposal{}, fmt.Errorf("device %s is disabled", in.DeviceID)
	}
	driver, err := uc.deps.Driver(ctx, in.DeviceID)
	if err != nil {
		return Proposal{}, err
	}
	ops, err := uc.deps.Ops(ctx, in.DeviceID)
	if err != nil {
		return Proposal{}, err
	}

	cmds := make([]command.Command, 0, len(lines))
	scopes := make([]string, 0, len(lines))
	scopeSeen := map[string]bool{}
	var diffs []CommandDiff
	var allInverse []command.Command
	var warnings []string
	for _, line := range lines {
		cmd := command.Command{Raw: line, Args: cloneMap(in.Args)}
		class := driver.Classify(cmd)
		if command.Decide(class) == command.DecisionDeny {
			return Proposal{}, fmt.Errorf("%w: %s", command.ErrDenied, class.DenyReason)
		}
		if command.Decide(class) != command.DecisionRequireApproval {
			return Proposal{}, fmt.Errorf("%w: line %q is READ; use router_read instead", command.ErrApprovalRequired, line)
		}
		menu, verb := SplitCommandPath(line)
		var before []map[string]string
		if verb != "add" && selectorID(cmd.Args) != "" {
			before, _ = ops.ReadMenu(ctx, menu, selectorFromArgs(cmd.Args))
		}
		diffs = append(diffs, CommandDiff{Command: line, Diff: BuildCommandDiff(cmd, before)})
		inv := BuildInverse(cmd, before, "")
		allInverse = append(allInverse, inv.Commands...)
		warnings = append(warnings, inv.Warnings...)
		if !scopeSeen[menu] {
			scopeSeen[menu] = true
			scopes = append(scopes, menu)
		}
		cmds = append(cmds, cmd)
	}
	if len(allInverse) == 0 {
		return Proposal{}, fmt.Errorf("%w: no deterministic rollback for this script", command.ErrApprovalRequired)
	}
	script := RenderInverseScript(allInverse)
	rp, err := uc.deps.RestorePoints.CreateForMutation(ctx, in.TenantID, in.DeviceID, in.ActorID, scopes, script)
	if err != nil {
		return Proposal{}, fmt.Errorf("%w: restore point unavailable: %v", command.ErrApprovalRequired, err)
	}

	now := uc.deps.Now()
	diffJSON, _ := json.Marshal(diffs)
	valJSON, _ := json.Marshal(map[string]any{"supported": false, "valid": false, "warnings": []string{"multi-line script: per-line validation not available in v1"}})
	key := proposalKey(in)
	req := approval.Request{
		ID: idgen.New("apr"), TenantID: in.TenantID, DeviceID: in.DeviceID,
		SessionID: in.SessionID, ActorType: in.ActorType, ActorID: in.ActorID,
		RequestedBy: in.ActorID,
		Command:     strings.Join(lines, "\n"), Args: in.Args, Revision: 1,
		Diff: diffJSON, Validation: valJSON, RestorePointID: rp.ID,
		Status: approval.StatusPending, IdempotencyKey: key,
		CreatedAt: now, UpdatedAt: now, ExpiresAt: now.Add(uc.deps.Config.TTL),
	}
	if err := uc.deps.Approvals.Create(ctx, req); err != nil {
		return Proposal{}, fmt.Errorf("persist approval: %w", err)
	}
	uc.logAttempt(ctx, in, "MUTATION", "PENDING_APPROVAL", "approval "+req.ID+" (script)")
	return Proposal{Approval: req, Validation: port.ScriptValidation{Supported: false}, RestorePoint: rp, Warnings: warnings}, nil
}

func (uc *ProposalUseCase) logAttempt(ctx context.Context, in ProposeCommandInput, mode, status, summary string) {
	if uc.deps.CommandLogs == nil {
		return
	}
	_ = uc.deps.CommandLogs.Write(ctx, commandLog(in.TenantID, in.DeviceID, string(in.ActorType), in.ActorID, in.SessionID, mode, in.Command, status, summary))
}

func proposalKey(in ProposeCommandInput) string {
	h := sha256.Sum256([]byte(in.TenantID + "|" + in.DeviceID + "|" + in.Command + "|" + fmt.Sprint(in.Args) + "|" + in.ActorID))
	return hex.EncodeToString(h[:16])
}

func splitScript(script string, maxLines int) []string {
	var out []string
	for _, line := range strings.Split(script, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
		if len(out) > maxLines {
			return out
		}
	}
	return out
}

func renderCommand(cmd command.Command) string {
	var b strings.Builder
	b.WriteString(cmd.Raw)
	for _, key := range sortedKeys(cmd.Args) {
		b.WriteString(" ")
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(quoteRouterValue(cmd.Args[key]))
	}
	return b.String()
}

func selectorFromArgs(args map[string]string) map[string]string {
	out := map[string]string{}
	for _, key := range []string{".id", "numbers"} {
		if v := strings.TrimSpace(args[key]); v != "" {
			out[".id"] = v
			break
		}
	}
	for k, v := range args {
		if strings.HasPrefix(k, "?") {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneMap(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
