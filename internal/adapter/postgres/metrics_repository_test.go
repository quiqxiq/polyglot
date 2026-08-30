package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/device"
)

func setupMetricsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.DevicePingMetricModel{})
	require.NoError(t, err)

	return db
}

func TestMetricsRepository_SaveAndQuery(t *testing.T) {
	db := setupMetricsTestDB(t)
	repo := postgres.NewMetricsRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	deviceID := "dev-ping-01"

	points := []device.PingMetricPoint{
		{
			RecordedAt: now.Add(-10 * time.Minute),
			DeviceID:   deviceID,
			Target:     "8.8.8.8",
			Seq:        1,
			Size:       56,
			TTL:        116,
			RTTMS:      24.5,
			Status:     "connected",
			Sent:       1,
			Received:   1,
			PacketLoss: 0,
		},
		{
			RecordedAt: now.Add(-5 * time.Minute),
			DeviceID:   deviceID,
			Target:     "8.8.8.8",
			Seq:        2,
			Size:       56,
			TTL:        116,
			RTTMS:      25.5,
			Status:     "connected",
			Sent:       2,
			Received:   2,
			PacketLoss: 0,
		},
		{
			RecordedAt: now.Add(-1 * time.Minute),
			DeviceID:   deviceID,
			Target:     "8.8.8.8",
			Seq:        3,
			Size:       56,
			TTL:        0,
			RTTMS:      0,
			Status:     "timeout",
			Sent:       3,
			Received:   2,
			PacketLoss: 33,
		},
	}

	// 1. Batch Save
	err := repo.SavePingMetricsBatch(ctx, points)
	require.NoError(t, err)

	// 2. Query within time window
	res, summary, err := repo.QueryPingMetrics(ctx, device.PingMetricsFilter{
		DeviceID:       deviceID,
		StartTime:      now.Add(-15 * time.Minute),
		EndTime:        now,
		BucketInterval: "raw",
	})
	require.NoError(t, err)
	assert.Len(t, res, 3)
	assert.Equal(t, int64(3), summary.TotalSamples)
	assert.InDelta(t, 24.5, float64(summary.MinRTT), 0.1)
	assert.InDelta(t, 25.5, float64(summary.MaxRTT), 0.1)
	assert.InDelta(t, 16.67, float64(summary.PacketLossPct), 0.01)

	// 3. Test Cleanup
	err = repo.CleanupExpiredMetrics(ctx, deviceID, 1) // Retention 1 day -> points within 10 min are kept
	require.NoError(t, err)

	resAfter, _, err := repo.QueryPingMetrics(ctx, device.PingMetricsFilter{
		DeviceID:       deviceID,
		StartTime:      now.Add(-15 * time.Minute),
		EndTime:        now,
		BucketInterval: "raw",
	})
	require.NoError(t, err)
	assert.Len(t, resAfter, 3)
}

func TestMetricsRepositoryRejectsInvalidRange(t *testing.T) {
	repo := postgres.NewMetricsRepository(setupMetricsTestDB(t))
	now := time.Now().UTC()

	_, _, err := repo.QueryPingMetrics(context.Background(), device.PingMetricsFilter{
		DeviceID:  "dev-ping-01",
		StartTime: now,
		EndTime:   now,
	})
	require.ErrorIs(t, err, device.ErrInvalidMetricsRange)
}

func TestMetricsRepositoryRejectsUnknownBucket(t *testing.T) {
	repo := postgres.NewMetricsRepository(setupMetricsTestDB(t))
	now := time.Now().UTC()
	require.NoError(t, repo.SavePingMetric(context.Background(), device.PingMetricPoint{
		RecordedAt: now.Add(-time.Minute),
		DeviceID:   "dev-ping-01",
		Target:     "8.8.8.8",
		RTTMS:      10,
		Sent:       1,
		Received:   1,
	}))

	_, _, err := repo.QueryPingMetrics(context.Background(), device.PingMetricsFilter{
		DeviceID:       "dev-ping-01",
		StartTime:      now.Add(-time.Hour),
		EndTime:        now,
		BucketInterval: "unknown",
	})
	require.ErrorIs(t, err, device.ErrInvalidMetricsBucket)
}
