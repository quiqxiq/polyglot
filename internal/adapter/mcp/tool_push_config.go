package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// pushConfigArgs is the input schema for the push_config tool.
type pushConfigArgs struct {
	DeviceID string            `json:"device_id" jsonschema:"the device ID from inventory"`
	Command  string            `json:"command" jsonschema:"the config command for this vendor (e.g. setParameterValues for GenieACS)"`
	Args     map[string]string `json:"args,omitempty" jsonschema:"configuration parameters as key-value pairs"`
}

// configOutput is the structured result of push_config. Concise: Applied
// signals success, Message gives a 1-line detail, Error on failure.
type configOutput struct {
	DeviceID string `json:"device_id"`
	Applied  bool   `json:"applied"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// pushConfig handles the "push_config" tool. Always destructive (the
// ToolAnnotations signal this to the client). Uses the same
// ExecuteCommandPreApproved pipeline as run_command — the semantic
// distinction is for the LLM: push_config signals "I am writing config",
// run_command signals "I am running a general command".
func (s *Server) pushConfig(ctx context.Context, _ *mcp.CallToolRequest, args pushConfigArgs) (*mcp.CallToolResult, configOutput, error) {
	driver, err := s.registry.Get(ctx, args.DeviceID)
	if err != nil {
		return toolError(configOutput{DeviceID: args.DeviceID, Error: err.Error()})
	}

	cmd := command.Command{Raw: args.Command, Args: args.Args}
	result, err := network.ExecuteCommandPreApproved(ctx, driver, cmd)
	if err != nil {
		return toolError(configOutput{DeviceID: args.DeviceID, Error: err.Error()})
	}

	s.auditBestEffort(ctx, args.DeviceID, args.Command, result)

	msg := result.Output
	if msg == "" {
		msg = "configuration applied"
	}
	return toolOK(configOutput{
		DeviceID: args.DeviceID,
		Applied:  true,
		Message:  msg,
	})
}
