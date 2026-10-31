package port

import (
	"context"
	"time"
)

// ActiveDevice is the server-side binding of "which router the operator has
// currently selected" for AI/MCP tool calls. It is the source of truth; the
// browser Zustand store is display-only.
type ActiveDevice struct {
	TenantID  string
	UserID    string
	SessionID string
	DeviceID  string
	UpdatedAt time.Time
}

// ActiveDeviceStore persists per-user/per-session active router bindings.
type ActiveDeviceStore interface {
	Set(ctx context.Context, a ActiveDevice, ttl time.Duration) error
	Get(ctx context.Context, tenantID, userID, sessionID string) (ActiveDevice, error)
	Clear(ctx context.Context, tenantID, userID, sessionID string) error
}
