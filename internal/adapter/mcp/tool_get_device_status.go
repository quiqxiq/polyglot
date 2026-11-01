package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// getDeviceStatusArgs is the input schema for the get_device_status tool.
// The generic mcp.AddTool auto-generates the JSON schema from this struct's
// tags, so the LLM sees exactly what it needs to call the tool — no manual
// schema maintenance.
type getDeviceStatusArgs struct {
	DeviceID string `json:"device_id" jsonschema:"the device ID from inventory"`
}

// deviceStatusOutput is the structured result of get_device_status. Concise
// by design: Summary is a 1-line human-readable string the LLM can read
// without parsing Rows; Rows is included for detail when the LLM needs
// specific values. This keeps token usage low for cheap LLMs while
// preserving full data for expensive ones.
type deviceStatusOutput struct {
	DeviceID string              `json:"device_id"`
	Status   string              `json:"status"`
	Summary  string              `json:"summary"`
	Rows     []map[string]string `json:"rows,omitempty"`
}

// getDeviceStatus handles the "get_device_status" tool. It resolves a
// connected driver from the registry, calls usecase/network.GetDeviceStatus
// (which does Translate → ExecuteCommand), and formats the result into a
// concise structured output. Read-only — auto-approved by policy.
func (s *Server) getDeviceStatus(ctx context.Context, _ *mcp.CallToolRequest, args getDeviceStatusArgs) (*mcp.CallToolResult, deviceStatusOutput, error) {
	driver, err := s.registry.Get(ctx, args.DeviceID)
	if err != nil {
		return toolError(deviceStatusOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	result, err := network.GetDeviceStatus(ctx, driver)
	if err != nil {
		return toolError(deviceStatusOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	s.auditBestEffort(ctx, args.DeviceID, "get_device_status", result)

	return toolOK(formatStatus(args.DeviceID, result))
}

// formatStatus builds a concise deviceStatusOutput from a command.Result.
// Summary picks common status fields (uptime, version, last inform) from
// the first row so the LLM gets a quick snapshot without parsing every row.
func formatStatus(deviceID string, result command.Result) deviceStatusOutput {
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

// auditBestEffort writes an audit event if the writer is configured. Errors
// are silently discarded — a failed audit log must never block a tool
// response (the command already executed; losing the log is better than
// losing the result).
func (s *Server) auditBestEffort(ctx context.Context, deviceID, cmd string, result command.Result) {
	if s.audit == nil {
		return
	}
	out := result.Output
	if len(result.Rows) > 0 {
		out = fmt.Sprintf("%s (%d rows)", out, len(result.Rows))
	}
	_ = s.audit.Write(ctx, port.AuditEvent{
		DeviceID:  deviceID,
		UserID:    "mcp",
		Command:   cmd,
		Result:    out,
		Timestamp: time.Now(),
	})
}

// toolOK wraps a successful output into the MCP result/structured-content
// pair the generic AddTool expects.
func toolOK[T any](out T) (*mcp.CallToolResult, T, error) {
	return &mcp.CallToolResult{}, out, nil
}

// toolError wraps a failure output with IsError=true so the LLM sees a
// structured error field rather than a bare protocol error string.
func toolError[T any](out T) (*mcp.CallToolResult, T, error) {
	return &mcp.CallToolResult{IsError: true}, out, nil
}
