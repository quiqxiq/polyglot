package mikrotik

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
)

type mikhmonKickSessionArgs struct {
	DeviceID  string `json:"device_id" jsonschema:"the Mikrotik device ID from inventory"`
	SessionID string `json:"session_id,omitempty" jsonschema:"the RouterOS active session .id"`
	User      string `json:"user,omitempty" jsonschema:"username or IP address of active session to kick"`
}

type mikhmonKickSessionOutput struct {
	DeviceID string `json:"device_id"`
	Status   string `json:"status"`
	Summary  string `json:"summary"`
}

// MikhmonKickSession handles the "mikhmon_kick_session" tool.
func (h *Handler) MikhmonKickSession(ctx context.Context, _ *mcp.CallToolRequest, args mikhmonKickSessionArgs) (*mcp.CallToolResult, mikhmonKickSessionOutput, error) {
	if args.DeviceID == "" || (args.SessionID == "" && args.User == "") {
		return mcputil.ToolError(mikhmonKickSessionOutput{Status: "error", Summary: "device_id and either session_id or user are required"})
	}

	if h.Registry == nil {
		return mcputil.ToolError(mikhmonKickSessionOutput{DeviceID: args.DeviceID, Status: "error", Summary: "device registry not configured"})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(mikhmonKickSessionOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	uc := h.uc()
	if uc == nil {
		return mcputil.ToolError(mikhmonKickSessionOutput{DeviceID: args.DeviceID, Status: "error", Summary: "mikhmon use case not configured"})
	}

	targetRosID := args.SessionID
	if targetRosID == "" && args.User != "" {
		activeSessions, err := uc.GetActiveSessions(ctx, driver)
		if err != nil {
			return mcputil.ToolError(mikhmonKickSessionOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
		}
		for _, sess := range activeSessions {
			if sess.User == args.User || sess.Address == args.User {
				targetRosID = sess.RosID
				break
			}
		}
		if targetRosID == "" {
			return mcputil.ToolError(mikhmonKickSessionOutput{DeviceID: args.DeviceID, Status: "error", Summary: fmt.Sprintf("active session for user %q not found", args.User)})
		}
	}

	res, err := uc.RemoveActiveSession(ctx, driver, targetRosID)
	if err != nil {
		return mcputil.ToolError(mikhmonKickSessionOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	summary := fmt.Sprintf("Successfully disconnected session %s on device %s. Result: %s", targetRosID, args.DeviceID, res.Output)

	return mcputil.ToolOK(mikhmonKickSessionOutput{
		DeviceID: args.DeviceID,
		Status:   "success",
		Summary:  summary,
	})
}
