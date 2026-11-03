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

// getDeviceStatusArgs is the input schema for the get_device_status tool.
type getDeviceStatusArgs struct {
	DeviceID string `json:"device_id" jsonschema:"the device ID from inventory"`
}

// deviceStatusOutput is the structured result of get_device_status. Concise
// by design: Summary is a 1-line human-readable string the LLM can read
// without parsing Rows; Rows is included for detail when the LLM needs
// specific values.
type deviceStatusOutput struct {
	DeviceID string              `json:"device_id"`
	Status   string              `json:"status"`
	Summary  string              `json:"summary"`
	Rows     []map[string]string `json:"rows,omitempty"`
}

// GetDeviceStatus handles the "get_device_status" tool. It resolves a
// connected driver from the registry, calls usecase/network.GetDeviceStatus
// (which does Translate → ExecuteCommand), and formats the result into a
// concise structured output. Read-only — auto-approved by policy.
func (h *Handler) GetDeviceStatus(ctx context.Context, _ *mcp.CallToolRequest, args getDeviceStatusArgs) (*mcp.CallToolResult, deviceStatusOutput, error) {
	if h.Registry == nil {
		return mcputil.ToolError(deviceStatusOutput{DeviceID: args.DeviceID, Status: "error", Summary: "device registry not configured"})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(deviceStatusOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	result, err := network.GetDeviceStatus(ctx, driver)
	if err != nil {
		return mcputil.ToolError(deviceStatusOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	mcputil.AuditBestEffort(ctx, h.Audit, args.DeviceID, "get_device_status", result)

	return mcputil.ToolOK(FormatStatus(args.DeviceID, result))
}

// FormatStatus builds a concise deviceStatusOutput from a command.Result.
// Summary picks common status fields (uptime, version, last inform) from
// the first row so the LLM gets a quick snapshot without parsing every row.
func FormatStatus(deviceID string, result command.Result) deviceStatusOutput {
	out := deviceStatusOutput{
		DeviceID: deviceID,
		Status:   "online",
		Rows:     result.Rows,
	}
	if len(result.Rows) > 0 {
		row := result.Rows[0]
		parts := make([]string, 0, 4)
		for _, key := range []string{"uptime", "version", "_lastInform", "SerialNumber"} {
			if v, ok := row[key]; ok && v != "" {
				parts = append(parts, fmt.Sprintf("%s: %s", key, v))
			}
		}
		if len(parts) > 0 {
			out.Summary = strings.Join(parts, ", ")
		}
	}
	if out.Summary == "" && result.Output != "" {
		out.Summary = result.Output
	}
	if out.Summary == "" {
		out.Summary = "no status data"
	}
	return out
}
