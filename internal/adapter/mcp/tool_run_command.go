package mcp

import (
	"context"

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
// populated only when the command failed.
type commandOutput struct {
	DeviceID string              `json:"device_id"`
	Command  string              `json:"command"`
	Executed bool                `json:"executed"`
	Output   string              `json:"output,omitempty"`
	Rows     []map[string]string `json:"rows,omitempty"`
	RowCount int                 `json:"row_count,omitempty"`
	Error    string              `json:"error,omitempty"`
}

// runCommand handles the "run_command" tool. It executes a raw vendor-native
// command via ExecuteCommandPreApproved — the client (LibreChat) gates
// destructive commands via ToolAnnotations before calling, so the server
// executes directly when invoked. DecisionDeny still blocks (hard policy
// deny cannot be overridden by client approval).
func (s *Server) runCommand(ctx context.Context, _ *mcp.CallToolRequest, args runCommandArgs) (*mcp.CallToolResult, commandOutput, error) {
	driver, err := s.registry.Get(ctx, args.DeviceID)
	if err != nil {
		return toolError(commandOutput{DeviceID: args.DeviceID, Command: args.Command, Error: err.Error()})
	}

	cmd := command.Command{Raw: args.Command, Args: args.Args}
	result, err := network.ExecuteCommandPreApproved(ctx, driver, cmd)
	if err != nil {
		return toolError(commandOutput{DeviceID: args.DeviceID, Command: args.Command, Error: err.Error()})
	}

	s.auditBestEffort(ctx, args.DeviceID, args.Command, result)

	out := commandOutput{
		DeviceID: args.DeviceID,
		Command:  args.Command,
		Executed: true,
		Output:   result.Output,
		Rows:     result.Rows,
		RowCount: len(result.Rows),
	}
	if out.Output == "" && out.RowCount == 0 {
		out.Output = "ok"
	}
	return toolOK(out)
}
