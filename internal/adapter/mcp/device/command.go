package device

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// runCommandArgs is the input schema for the run_command tool.
type runCommandArgs struct {
	DeviceID string            `json:"device_id" jsonschema:"the device ID from inventory"`
	Command  string            `json:"command" jsonschema:"the vendor-native command to execute (e.g. 'show version' or 'show ip interface brief'). For MikroTik RouterOS devices, DO NOT use this tool — use routeros_run instead."`
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

// RunCommand handles the "run_command" tool. It executes a raw vendor-native
// command via ExecuteCommandPreApproved — the client (LibreChat) gates
// destructive commands via ToolAnnotations before calling, so the server
// executes directly when invoked. DecisionDeny still blocks (hard policy
// deny cannot be overridden by client approval).
func (h *Handler) RunCommand(ctx context.Context, _ *mcp.CallToolRequest, args runCommandArgs) (*mcp.CallToolResult, commandOutput, error) {
	if h.Registry == nil {
		return mcputil.ToolError(commandOutput{DeviceID: args.DeviceID, Command: args.Command, Error: "device registry not configured"})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(commandOutput{DeviceID: args.DeviceID, Command: args.Command, Error: err.Error()})
	}

	cmd := command.Command{Raw: args.Command, Args: args.Args}
	result, err := network.ExecuteCommandPreApproved(ctx, driver, cmd)
	if err != nil {
		errText := err.Error()
		if h.isMikroTikDevice(ctx, args.DeviceID) {
			errText = fmt.Sprintf("COMMAND EXECUTION FAILED: %s\n"+
				"NOTICE: For MikroTik RouterOS devices, please use 'routeros_run' with 'routeros_validate' instead of 'run_command'.\n"+
				"ACTION REQUIRED: Do NOT retry 'run_command'. Use 'routeros_suggest' with device_id and input=%q to inspect valid commands and parameter names.",
				err.Error(), args.Command)
		}
		return mcputil.ToolError(commandOutput{DeviceID: args.DeviceID, Command: args.Command, Error: errText})
	}

	mcputil.AuditBestEffort(ctx, h.Audit, args.DeviceID, args.Command, result)

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
	return mcputil.ToolOK(out)
}

func (h *Handler) isMikroTikDevice(ctx context.Context, deviceID string) bool {
	if r := h.repo(); r != nil {
		if dev, err := r.FindByID(ctx, deviceID); err == nil {
			return strings.EqualFold(dev.Vendor, "mikrotik") || strings.EqualFold(dev.DriverType, "mikrotik")
		}
	}
	return false
}
