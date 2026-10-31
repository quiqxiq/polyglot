package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// runCommandArgs is the input schema for the run_command tool.
type runCommandArgs struct {
	DeviceID string            `json:"device_id" jsonschema:"the device ID from inventory"`
	Command  string            `json:"command" jsonschema:"the vendor-native command to execute (e.g. /interface/print for Mikrotik, show version for Cisco)"`
	Args     map[string]string `json:"args,omitempty" jsonschema:"optional key-value parameters for the command"`
}

// commandOutput is the structured result of run_command. Rows and RowCount
// give the LLM structured data; Output is the raw text fallback. Error is
// populated only when the command failed or was refused.
//
// Executed is always false for MUTATION commands; the LLM must never claim a
// mutation was applied before the approval pipeline reports EXECUTED.
type commandOutput struct {
	DeviceID         string              `json:"device_id"`
	Command          string              `json:"command"`
	Mode             string              `json:"mode"` // READ | MUTATION
	Executed         bool                `json:"executed"`
	ApprovalRequired bool                `json:"approval_required,omitempty"`
	Output           string              `json:"output,omitempty"`
	Rows             []map[string]string `json:"rows,omitempty"`
	RowCount         int                 `json:"row_count,omitempty"`
	Error            string              `json:"error,omitempty"`
}

// runCommand handles the "run_command" tool. READ commands execute directly;
// MUTATION commands are never executed here — they must be proposed through
// the approval pipeline (PLAN-AI-ROUTER-OPS v2). Until that pipeline is
// wired, a mutation returns a clear "approval required" failure instead of
// silently executing.
func (s *Server) runCommand(ctx context.Context, req *mcp.CallToolRequest, args runCommandArgs) (*mcp.CallToolResult, commandOutput, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(commandOutput{DeviceID: args.DeviceID, Command: args.Command, Error: "unauthenticated: a valid user session is required"})
	}

	driver, err := s.registry.Get(ctx, args.DeviceID)
	if err != nil {
		return toolError(commandOutput{DeviceID: args.DeviceID, Command: args.Command, Error: err.Error()})
	}

	cmd := command.Command{Raw: args.Command, Args: args.Args}
	class := driver.Classify(cmd)
	mode := "READ"
	if class.Mode == command.ModeMutation {
		mode = "MUTATION"
	}
	if class.IsDenied() {
		return toolError(commandOutput{
			DeviceID: args.DeviceID, Command: args.Command, Mode: mode,
			Error: "denied by policy: " + class.DenyReason,
		})
	}
	if class.Mode == command.ModeMutation {
		// Delegate to the proposal pipeline: the tool NEVER executes the
		// mutation; it creates an approval request with diff + restore point.
		if s.proposals != nil {
			res, out, perr := s.proposeCommand(ctx, req, proposeCommandArgs{
				DeviceID: args.DeviceID, Command: args.Command, Args: args.Args,
			})
			if perr != nil {
				return toolError(commandOutput{DeviceID: args.DeviceID, Command: args.Command, Mode: mode, ApprovalRequired: true, Error: perr.Error()})
			}
			if out.Status == "error" || out.ApprovalID == "" {
				return res, commandOutput{DeviceID: args.DeviceID, Command: args.Command, Mode: mode, ApprovalRequired: true, Error: out.Message}, nil
			}
			_ = res
			return toolOK(commandOutput{
				DeviceID: args.DeviceID, Command: args.Command, Mode: mode,
				ApprovalRequired: true, Executed: false,
				Output: fmt.Sprintf("pending approval %s (restore point %s); NOT executed", out.ApprovalID, out.RestorePointID),
			})
		}
		return toolError(commandOutput{
			DeviceID: args.DeviceID, Command: args.Command, Mode: mode,
			ApprovalRequired: true,
			Error:            "mutation was NOT executed: this tool can only propose mutations, and the approval pipeline is not configured.",
		})
	}

	result, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return toolError(commandOutput{DeviceID: args.DeviceID, Command: args.Command, Mode: mode, Error: err.Error()})
	}

	s.auditBestEffort(ctx, actorID, args.DeviceID, args.Command, result)

	out := commandOutput{
		DeviceID: args.DeviceID, Command: args.Command, Mode: mode,
		Executed: true, Output: result.Output, Rows: result.Rows, RowCount: len(result.Rows),
	}
	if out.Output == "" && out.RowCount == 0 {
		out.Output = "ok"
	}
	return toolOK(out)
}
