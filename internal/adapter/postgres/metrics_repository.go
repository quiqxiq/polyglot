package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// ErrTimescaleDBNotAvailable is returned when TimescaleDB extension is missing.
var ErrTimescaleDBNotAvailable = errors.New("timescaledb extension is not active on database")

// MetricsRepository implements port.MetricsRepository on top of PostgreSQL/TimescaleDB.
type MetricsRepository struct {
	db              *gorm.DB
	timescaleCached atomic.Bool
	checkOnce       sync.Once
}

var _ port.MetricsRepository = (*MetricsRepository)(nil)

// NewMetricsRepository creates a new MetricsRepository.
func NewMetricsRepository(db *gorm.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

// IsTimescaleDBAvailable checks whether the TimescaleDB extension is active.
func (r *MetricsRepository) IsTimescaleDBAvailable(ctx context.Context) (bool, error) {
	if r.db.Dialector.Name() != "postgres" {
		// In-memory SQLite or test environments allow mock execution
		return true, nil
	}

	if r.timescaleCached.Load() {
		return true, nil
	}

	var count int64
	err := r.db.WithContext(ctx).
		Raw("SELECT count(1) FROM pg_extension WHERE extname = 'timescaledb'").
		Scan(&count).Error
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return false, ctx.Err()
		}
		return false, fmt.Errorf("check timescaledb extension: %w", err)
	}
	if count == 0 {
		return false, nil
	}

	count = 0
	err = r.db.WithContext(ctx).
		Raw(`SELECT count(1)
			FROM timescaledb_information.hypertables
			WHERE hypertable_name = 'device_ping_metrics'`).
		Scan(&count).Error
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return false, ctx.Err()
		}
		return false, fmt.Errorf("check ping metrics hypertable: %w", err)
	}

	available := count > 0
	if available {
		r.timescaleCached.Store(true)
	}
	return available, nil
}

// SavePingMetric persists a single ping metric point to TimescaleDB.
func (r *MetricsRepository) SavePingMetric(ctx context.Context, point device.PingMetricPoint) error {
	return r.SavePingMetricsBatch(ctx, []device.PingMetricPoint{point})
}

// SavePingMetricsBatch persists multiple ping metric points in batches.
func (r *MetricsRepository) SavePingMetricsBatch(ctx context.Context, points []device.PingMetricPoint) error {
	if len(points) == 0 {
		return nil
	}

	available, err := r.IsTimescaleDBAvailable(ctx)
	if err != nil || !available {
		return ErrTimescaleDBNotAvailable
	}

	models := make([]model.DevicePingMetricModel, len(points))
	for i, p := range points {
		models[i] = model.PingMetricModelFromDomain(p)
	}

	return r.db.WithContext(ctx).CreateInBatches(models, 100).Error
}

// QueryPingMetrics retrieves time-series ping records and statistics within the filter window.
func (r *MetricsRepository) QueryPingMetrics(ctx context.Context, filter device.PingMetricsFilter) ([]device.PingMetricPoint, device.PingSummary, error) {
	if filter.DeviceID == "" {
		return nil, device.PingSummary{}, errors.New("device_id is required")
	}

	available, err := r.IsTimescaleDBAvailable(ctx)
	if err != nil || !available {
		return nil, device.PingSummary{}, ErrTimescaleDBNotAvailable
	}

	if filter.StartTime.IsZero() || filter.EndTime.IsZero() || !filter.StartTime.Before(filter.EndTime) {
		return nil, device.PingSummary{}, device.ErrInvalidMetricsRange
	}

	bucket := strings.TrimSpace(strings.ToLower(filter.BucketInterval))
	if err := validateMetricsRange(filter.StartTime, filter.EndTime, bucket); err != nil {
		return nil, device.PingSummary{}, err
	}

	// 1. Calculate overall summary statistics for the time window
	type statResult struct {
		MinRTT        *float32
		AvgRTT        *float32
		MaxRTT        *float32
		TotalSent     int64
		TotalReceived int64
		TotalSamples  int64
	}

	var stat statResult
	summaryQuery := r.db.WithContext(ctx).Model(&model.DevicePingMetricModel{}).
		Select("MIN(NULLIF(rtt_ms, 0)) AS min_rtt, AVG(NULLIF(rtt_ms, 0)) AS avg_rtt, MAX(NULLIF(rtt_ms, 0)) AS max_rtt, COALESCE(SUM(sent), 0) AS total_sent, COALESCE(SUM(received), 0) AS total_received, COUNT(*) AS total_samples").
		Where("device_id = ? AND recorded_at >= ? AND recorded_at < ?", filter.DeviceID, filter.StartTime, filter.EndTime)

	if err := summaryQuery.Scan(&stat).Error; err != nil {
		return nil, device.PingSummary{}, fmt.Errorf("calculate ping summary: %w", err)
	}

	summary := device.PingSummary{
		TotalSamples: stat.TotalSamples,
	}
	if stat.MinRTT != nil {
		summary.MinRTT = *stat.MinRTT
	}
	if stat.AvgRTT != nil {
		summary.AvgRTT = float32(math.Round(float64(*stat.AvgRTT)*100) / 100)
	}
	if stat.MaxRTT != nil {
		summary.MaxRTT = *stat.MaxRTT
	}
	if stat.TotalSent > 0 {
		loss := float64(stat.TotalSent-stat.TotalReceived) / float64(stat.TotalSent) * 100
		summary.PacketLossPct = float32(math.Round(loss*100) / 100)
	}

	if stat.TotalSamples == 0 {
		return []device.PingMetricPoint{}, summary, nil
	}

	// 2. Query metric points (raw or downsampled)
	supportedBucket := map[string]bool{
		"1m": true, "1min": true, "1minute": true,
		"5m": true, "5min": true, "5minutes": true,
		"15m": true, "15min": true, "15minutes": true,
		"1h": true, "1hour": true,
		"1d": true, "1day": true,
	}
	if bucket != "" && bucket != "raw" && !supportedBucket[bucket] {
		return nil, summary, device.ErrInvalidMetricsBucket
	}
	if bucket == "" || bucket == "raw" || r.db.Dialector.Name() != "postgres" {
		var list []model.DevicePingMetricModel
		err := r.db.WithContext(ctx).
			Where("device_id = ? AND recorded_at >= ? AND recorded_at < ?", filter.DeviceID, filter.StartTime, filter.EndTime).
			Order("recorded_at ASC").
			Limit(5000).
			Find(&list).Error
		if err != nil {
			return nil, summary, fmt.Errorf("fetch raw ping metrics: %w", err)
		}

		points := make([]device.PingMetricPoint, len(list))
		for i, m := range list {
			pt := m.ToDomain()
			if pt.Seq == 0 {
				pt.Seq = i
			}
			points[i] = pt
		}
		return points, summary, nil
	}

	// Downsampling via TimescaleDB time_bucket
	intervalSQL := "1 minute"
	switch bucket {
	case "1m", "1min", "1minute":
		intervalSQL = "1 minute"
	case "5m", "5min", "5minutes":
		intervalSQL = "5 minutes"
	case "15m", "15min", "15minutes":
		intervalSQL = "15 minutes"
	case "1h", "1hour":
		intervalSQL = "1 hour"
	case "1d", "1day":
		intervalSQL = "1 day"
	}

	type bucketRow struct {
		BucketTime time.Time `gorm:"column:bucket_time"`
		Target     string    `gorm:"column:target"`
		RTTMS      float32   `gorm:"column:rtt_ms"`
		PacketLoss float32   `gorm:"column:packet_loss"`
		MinRTTMS   *float32  `gorm:"column:min_rtt_ms"`
		AvgRTTMS   *float32  `gorm:"column:avg_rtt_ms"`
		MaxRTTMS   *float32  `gorm:"column:max_rtt_ms"`
		Sent       int       `gorm:"column:sent"`
		Received   int       `gorm:"column:received"`
	}

	var rows []bucketRow
	err = r.db.WithContext(ctx).
		Table("device_ping_metrics").
		Select(`
			time_bucket(?, recorded_at) AS bucket_time,
			target,
			AVG(NULLIF(rtt_ms, 0)) AS rtt_ms,
			CASE WHEN SUM(sent) > 0 THEN ((SUM(sent) - SUM(received)) * 100.0 / SUM(sent)) ELSE 0 END AS packet_loss,
			MIN(NULLIF(rtt_ms, 0)) AS min_rtt_ms,
			AVG(NULLIF(rtt_ms, 0)) AS avg_rtt_ms,
			MAX(NULLIF(rtt_ms, 0)) AS max_rtt_ms,
			SUM(sent) AS sent,
			SUM(received) AS received
		`, intervalSQL).
		Where("device_id = ? AND recorded_at >= ? AND recorded_at < ?", filter.DeviceID, filter.StartTime, filter.EndTime).
		Group("bucket_time, target").
		Order("bucket_time ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, summary, fmt.Errorf("fetch downsampled ping metrics: %w", err)
	}

	points := make([]device.PingMetricPoint, len(rows))
	for i, r := range rows {
		lossInt := int(math.Round(float64(r.PacketLoss)))
		status := "connected"
		if lossInt >= 100 {
			status = "timeout"
		}
		points[i] = device.PingMetricPoint{
			RecordedAt: r.BucketTime,
			DeviceID:   filter.DeviceID,
			Target:     r.Target,
			Seq:        i,
			RTTMS:      r.RTTMS,
			Status:     status,
			Sent:       r.Sent,
			Received:   r.Received,
			PacketLoss: lossInt,
			MinRTTMS:   r.MinRTTMS,
			AvgRTTMS:   r.AvgRTTMS,
			MaxRTTMS:   r.MaxRTTMS,
		}
	}

	return points, summary, nil
}

func validateMetricsRange(start, end time.Time, bucket string) error {
	var maxRange time.Duration
	switch bucket {
	case "", "raw":
		maxRange = 24 * time.Hour
	case "1m", "1min", "1minute":
		maxRange = 7 * 24 * time.Hour
	case "5m", "5min", "5minutes":
		maxRange = 30 * 24 * time.Hour
	case "15m", "15min", "15minutes":
		maxRange = 90 * 24 * time.Hour
	case "1h", "1hour":
		maxRange = 365 * 24 * time.Hour
	case "1d", "1day":
		maxRange = 5 * 365 * 24 * time.Hour
	default:
		return device.ErrInvalidMetricsBucket
	}
	if end.Sub(start) > maxRange {
		return device.ErrInvalidMetricsRange
	}
	return nil
}

// CleanupExpiredMetrics removes ping points older than retentionDays for a device.
func (r *MetricsRepository) CleanupExpiredMetrics(ctx context.Context, deviceID string, retentionDays int) error {
	if deviceID == "" || retentionDays <= 0 {
		return nil
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
	return r.db.WithContext(ctx).
		Where("device_id = ? AND recorded_at < ?", deviceID, cutoff).
		Delete(&model.DevicePingMetricModel{}).Error
}
