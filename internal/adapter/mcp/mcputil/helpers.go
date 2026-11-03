package mcputil

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// ToolOK wraps a successful output into the MCP result/structured-content
// pair the generic AddTool expects.
func ToolOK[T any](out T) (*mcp.CallToolResult, T, error) {
	return &mcp.CallToolResult{}, out, nil
}

// ToolError wraps a failure output with IsError=true so the LLM sees a
// structured error field rather than a bare protocol error string.
func ToolError[T any](out T) (*mcp.CallToolResult, T, error) {
	return &mcp.CallToolResult{IsError: true}, out, nil
}

// AuditBestEffort writes an audit event if the writer is configured. Errors
// are silently discarded — a failed audit log must never block a tool
// response (the command already executed; losing the log is better than
// losing the result).
func AuditBestEffort(ctx context.Context, writer port.AuditWriter, deviceID, cmd string, result command.Result) {
	if writer == nil {
		return
	}
	out := result.Output
	if len(result.Rows) > 0 {
		out = fmt.Sprintf("%s (%d rows)", out, len(result.Rows))
	}
	_ = writer.Write(ctx, port.AuditEvent{
		DeviceID:  deviceID,
		UserID:    "mcp",
		Command:   cmd,
		Result:    out,
		Timestamp: time.Now(),
	})
}
