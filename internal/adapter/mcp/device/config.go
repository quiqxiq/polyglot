package device

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// pushConfigArgs is the input schema for the push_config tool.
type pushConfigArgs struct {
	DeviceID string            `json:"device_id" jsonschema:"the device ID from inventory"`
	Command  string            `json:"command" jsonschema:"the config command for this vendor (e.g. setParameterValues for GenieACS)"`
	Args     map[string]string `json:"args,omitempty" jsonschema:"configuration parameters as key-value pairs"`
}

// configOutput is the structured result of push_config.
type configOutput struct {
	DeviceID string `json:"device_id"`
	Applied  bool   `json:"applied"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// PushConfig handles the "push_config" tool. Always destructive.
func (h *Handler) PushConfig(ctx context.Context, _ *mcp.CallToolRequest, args pushConfigArgs) (*mcp.CallToolResult, configOutput, error) {
	if h.Registry == nil {
		return mcputil.ToolError(configOutput{DeviceID: args.DeviceID, Error: "device registry not configured"})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(configOutput{DeviceID: args.DeviceID, Error: err.Error()})
	}

	cmd := command.Command{Raw: args.Command, Args: args.Args}
	result, err := network.ExecuteCommandPreApproved(ctx, driver, cmd)
	if err != nil {
		return mcputil.ToolError(configOutput{DeviceID: args.DeviceID, Error: err.Error()})
	}

	mcputil.AuditBestEffort(ctx, h.Audit, args.DeviceID, args.Command, result)

	msg := result.Output
	if msg == "" {
		msg = "configuration applied"
	}
	return mcputil.ToolOK(configOutput{
		DeviceID: args.DeviceID,
		Applied:  true,
		Message:  msg,
	})
}
