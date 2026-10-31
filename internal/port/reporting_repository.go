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

// AdvancedReportRepository menghitung laporan operasional tingkat lanjut
// (planning v2 Fase 5): AR Aging, rekonsiliasi gateway, dan statistik PSB.
type AdvancedReportRepository interface {
	ARAging(ctx context.Context, tenantID string, asOf time.Time) (reporting.ARAgingReport, error)
	GatewayReconciliation(ctx context.Context, tenantID string, from, to time.Time) (reporting.GatewayReconciliationReport, error)
	PSBStats(ctx context.Context, tenantID string, from, to time.Time) (reporting.PSBStatsReport, error)
	TaxReport(ctx context.Context, tenantID, period string) (reporting.TaxReport, error)
	SalesReport(ctx context.Context, tenantID, period string) (reporting.SalesReport, error)
	CollectionReport(ctx context.Context, tenantID, period string) (reporting.CollectionReport, error)
	CashflowReport(ctx context.Context, tenantID string, from, to time.Time) (reporting.CashflowReport, error)
	AgingDetail(ctx context.Context, tenantID string, asOf time.Time, limit int) (reporting.AgingDetailReport, error)
	ClosingReport(ctx context.Context, tenantID, period string, closingDate time.Time) (reporting.ClosingReport, error)
}
