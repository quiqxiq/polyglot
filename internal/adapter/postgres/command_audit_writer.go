package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
	"github.com/quixiq/polyglot/internal/port"
)

// CommandAuditWriter adapts the device-command port.AuditWriter onto the
// audit_logs table. It is the M0 audit sink for MCP tool calls; the dedicated
// command_logs table (approval_id/restore_point_id/mode/status) is added by
// the AI Router Ops migration and wired in M1.
type CommandAuditWriter struct {
	repo *AuditLogRepository
}

var _ port.AuditWriter = (*CommandAuditWriter)(nil)

// NewCommandAuditWriter builds the writer from an audit_logs repository.
func NewCommandAuditWriter(repo *AuditLogRepository) *CommandAuditWriter {
	return &CommandAuditWriter{repo: repo}
}

// Write records one device command attempt. output is truncated defensively;
// callers must already have masked secrets.
func (w *CommandAuditWriter) Write(ctx context.Context, event port.AuditEvent) error {
	if w.repo == nil {
		return nil
	}
	actorType := audit.ActorUser
	actorID := event.UserID
	if actorID == "" || actorID == "mcp" {
		actorType = audit.ActorSystem
		actorID = "mcp"
	}
	result := event.Result
	if len(result) > 2000 {
		result = result[:2000] + "…(truncated)"
	}
	desc := fmt.Sprintf("command=%s result=%s", event.Command, result)
	ts := event.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	return w.repo.Write(ctx, audit.AuditLog{
		TenantID:    "tenant-default",
		ActorType:   actorType,
		ActorID:     actorID,
		Action:      "DEVICE_COMMAND",
		EntityType:  "device",
		EntityID:    event.DeviceID,
		Description: desc,
		CreatedAt:   ts,
	})
}
