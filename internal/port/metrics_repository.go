package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// MetricsRepository defines the storage contract for high-frequency time-series telemetry.
type MetricsRepository interface {
	// IsTimescaleDBAvailable checks whether the TimescaleDB extension is enabled in PostgreSQL.
	IsTimescaleDBAvailable(ctx context.Context) (bool, error)

	// SavePingMetric stores a single ping metric point.
	SavePingMetric(ctx context.Context, point device.PingMetricPoint) error

	// SavePingMetricsBatch stores a batch of ping metric points in a single transaction/insert.
	SavePingMetricsBatch(ctx context.Context, points []device.PingMetricPoint) error

	// QueryPingMetrics retrieves time-series ping data and statistical summary for a device within a time range.
	QueryPingMetrics(ctx context.Context, filter device.PingMetricsFilter) ([]device.PingMetricPoint, device.PingSummary, error)

	// CleanupExpiredMetrics removes ping telemetry points older than retentionDays for a device.
	CleanupExpiredMetrics(ctx context.Context, deviceID string, retentionDays int) error
}
