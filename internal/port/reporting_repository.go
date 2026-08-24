package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/reporting"
)

// ReportingRepository defines persistence operations for pre-aggregated
// daily financial snapshots (DATABASE-SCHEMA-ISP.md §2.9).
type ReportingRepository interface {
	// UpsertSnapshot inserts or replaces the snapshot for its
	// (tenant_id, snapshot_date) — refreshed by the daily aggregation job.
	UpsertSnapshot(ctx context.Context, s reporting.DailyFinancialSnapshot) error
	GetByDate(ctx context.Context, tenantID string, date time.Time) (reporting.DailyFinancialSnapshot, error)
	ListRange(ctx context.Context, tenantID string, from, to time.Time) ([]reporting.DailyFinancialSnapshot, error)
}

// SnapshotComputer menghitung ulang snapshot harian dari data transaksional
// (dipanggil job cron; implementasi memakai agregasi SQL).
type SnapshotComputer interface {
	RecomputeDaily(ctx context.Context, tenantID string, date time.Time) error
}
