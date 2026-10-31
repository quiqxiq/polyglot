package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
)

// CommandLogFilter narrows command log listings.
type CommandLogFilter struct {
	TenantID string
	DeviceID string
	Status   string
	Limit    int
	Offset   int
}

// CommandLogRepository persists the detailed command audit trail.
type CommandLogRepository interface {
	Write(ctx context.Context, entry audit.CommandLog) error
	List(ctx context.Context, f CommandLogFilter) ([]audit.CommandLog, int, error)
	// DeleteBefore removes log rows older than cutoff (retention job).
	DeleteBefore(ctx context.Context, cutoff time.Time) (int, error)
}
