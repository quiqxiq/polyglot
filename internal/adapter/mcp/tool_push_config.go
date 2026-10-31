package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// pushConfigArgs is the input schema for the push_config tool.
type pushConfigArgs struct {
	DeviceID string            `json:"device_id" jsonschema:"the device ID from inventory"`
	Command  string            `json:"command" jsonschema:"the config command for this vendor (e.g. setParameterValues for GenieACS)"`
	Args     map[string]string `json:"args,omitempty" jsonschema:"configuration parameters as key-value pairs"`
}

// configOutput is the structured result of push_config.
type configOutput struct {
	DeviceID         string `json:"device_id"`
	Mode             string `json:"mode"`
	Applied          bool   `json:"applied"`
	ApprovalRequired bool   `json:"approval_required,omitempty"`
	Message          string `json:"message,omitempty"`
	Error            string `json:"error,omitempty"`
}

// pushConfig handles the "push_config" tool. Configuration changes are ALWAYS
// mutations: this tool never executes them. It fails closed until the
// approval pipeline is wired, so the LLM cannot bypass human approval.
func (s *Server) pushConfig(ctx context.Context, req *mcp.CallToolRequest, args pushConfigArgs) (*mcp.CallToolResult, configOutput, error) {
	if _, _, ok := s.actorFrom(ctx, req); !ok {
		return toolError(configOutput{DeviceID: args.DeviceID, Error: "unauthenticated: a valid user session is required"})
	}
	if args.DeviceID == "" || args.Command == "" {
		return toolError(configOutput{DeviceID: args.DeviceID, Error: "device_id and command are required"})
	}

	// Still classify so the response mode is accurate and hard-denied
	// commands get a specific message (driver lookup is best-effort: the
	// request is refused regardless).
	mode := "MUTATION"
	if driver, err := s.registry.Get(ctx, args.DeviceID); err == nil {
		class := driver.Classify(command.Command{Raw: args.Command, Args: args.Args})
		if class.IsDenied() {
			return toolError(configOutput{DeviceID: args.DeviceID, Mode: mode, Error: "denied by policy: " + class.DenyReason})
		}
	}

	if s.proposals != nil {
		res, out, perr := s.proposeConfig(ctx, req, proposeConfigArgs{
			DeviceID: args.DeviceID, Script: args.Command, Args: args.Args,
		})
		if perr != nil {
			return toolError(configOutput{DeviceID: args.DeviceID, Mode: mode, ApprovalRequired: true, Error: perr.Error()})
		}
		if out.Status == "error" || out.ApprovalID == "" {
			return res, configOutput{DeviceID: args.DeviceID, Mode: mode, ApprovalRequired: true, Error: out.Message}, nil
		}
		_ = res
		return toolOK(configOutput{
			DeviceID: args.DeviceID, Mode: mode, ApprovalRequired: true, Applied: false,
			Message: "pending approval " + out.ApprovalID + "; configuration NOT applied",
		})
	}
	return toolError(configOutput{
		DeviceID: args.DeviceID, Mode: mode, ApprovalRequired: true,
		Error: "configuration was NOT applied: this tool can only propose mutations, and the approval pipeline is not configured.",
	})
}
