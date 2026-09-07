//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	mikrotiksystem "github.com/quixiq/polyglot/internal/driver/mikrotik/system"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/ping"
)

// TestPingTelemetryE2E_FullPipeline tests the end-to-end telemetry pipeline:
// Streaming ping from physical MikroTik -> Parsing -> Batch Insertion to TimescaleDB -> Continuous Aggregate Query.
func TestPingTelemetryE2E_FullPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	_ = godotenv.Load(filepath.Join("..", "..", ".env"))

	target := mikrotikTestTarget(t)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:netops@localhost:5432/netops?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// 1. Connect to TimescaleDB
	store, err := postgres.NewStore(dbURL)
	require.NoError(t, err, "failed to connect to database")
	repo := postgres.NewMetricsRepository(store.DB())

	isTimescale, err := repo.IsTimescaleDBAvailable(ctx)
	require.NoError(t, err)
	if !isTimescale {
		t.Skip("TimescaleDB not available on target database")
	}

	// 2. Connect to MikroTik router
	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "failed to connect to physical mikrotik")
	defer func() { assert.NoError(t, drv.Close()) }()

	sd, ok := port.DeviceDriver(drv).(port.StreamingDeviceDriver)
	require.True(t, ok, "mikrotik driver must implement StreamingDeviceDriver")

	// 3. Stream ping from router
	pingCmd := mikrotiksystem.NewPingStreamCommand("8.8.8.8")
	handle, err := sd.Stream(ctx, pingCmd)
	require.NoError(t, err, "failed to initiate ping stream")
	defer func() { assert.NoError(t, handle.Cancel()) }()

	deviceID := uuid.New().String()
	var points []device.PingMetricPoint

	startTime := time.Now().UTC().Add(-1 * time.Minute)

	t.Logf("Streaming 5 ping responses from MikroTik %s...", target.Host)
	for i := 0; i < 5; i++ {
		select {
		case res, ok := <-handle.Chan():
			require.True(t, ok, "handle chan closed unexpectedly")
			if len(res.Rows) > 0 {
				row := res.Rows[0]
				latency, status := ping.ParsePingLatency(row)
				seq, _ := strconv.Atoi(row["seq"])
				ttl, _ := strconv.Atoi(row["ttl"])
				sent, _ := strconv.Atoi(row["sent"])
				recv, _ := strconv.Atoi(row["received"])
				loss := ping.ParsePacketLoss(row["packet-loss"])

				pt := device.PingMetricPoint{
					RecordedAt: time.Now().UTC(),
					DeviceID:   deviceID,
					Target:     "8.8.8.8",
					Seq:        seq,
					RTTMS:      float32(latency),
					Status:     status,
					TTL:        ttl,
					Sent:       sent,
					Received:   recv,
					PacketLoss: int(loss),
				}
				points = append(points, pt)
				t.Logf("Captured ping frame %d: RTT=%.1fms, Loss=%d%%", i+1, pt.RTTMS, pt.PacketLoss)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for ping frame %d", i+1)
		}
	}
	require.NotEmpty(t, points, "must have collected ping points")

	// 4. Ingest batch into TimescaleDB
	err = repo.SavePingMetricsBatch(ctx, points)
	require.NoError(t, err, "failed to save ping metrics batch to timescaledb")

	endTime := time.Now().UTC().Add(1 * time.Minute)

	// 5. Query back raw metrics
	rawPoints, rawSummary, err := repo.QueryPingMetrics(ctx, device.PingMetricsFilter{
		DeviceID:       deviceID,
		StartTime:      startTime,
		EndTime:        endTime,
		BucketInterval: "raw",
	})
	require.NoError(t, err, "failed to query raw ping metrics")
	assert.Len(t, rawPoints, len(points))
	assert.Equal(t, int64(len(points)), rawSummary.TotalSamples)
	assert.Greater(t, rawSummary.AvgRTT, float32(0))

	// 6. Test Continuous Aggregate / Downsampling query
	caggAvailable := repo.IsContinuousAggregateAvailable(ctx)
	t.Logf("Continuous aggregate available: %v", caggAvailable)

	downsampledPoints, downsampledSummary, err := repo.QueryPingMetrics(ctx, device.PingMetricsFilter{
		DeviceID:       deviceID,
		StartTime:      startTime,
		EndTime:        endTime,
		BucketInterval: "1m",
	})
	require.NoError(t, err, "failed to query 1m downsampled metrics")
	assert.GreaterOrEqual(t, downsampledSummary.TotalSamples, int64(1))
	t.Logf("Downsampled query result count: %d, summary avg RTT: %.2fms", len(downsampledPoints), downsampledSummary.AvgRTT)

	// 7. Cleanup test data
	err = repo.CleanupExpiredMetrics(ctx, deviceID, 0)
	assert.NoError(t, err, "failed to cleanup test device metrics")
}
