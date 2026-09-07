//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/device"
)

// setupTimescaleTestContainer spins up an ephemeral TimescaleDB PostgreSQL 16 container
// configured with production tuning parameters and executes all SQL migrations up to migration 24.
func setupTimescaleTestContainer(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "timescale/timescaledb:latest-pg16",
		ExposedPorts: []string{"5432/tcp"},
		Cmd: []string{
			"postgres",
			"-c", "shared_preload_libraries=timescaledb",
			"-c", "shared_buffers=512MB",
			"-c", "work_mem=16MB",
			"-c", "maintenance_work_mem=128MB",
			"-c", "max_connections=150",
		},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "netops_load_test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(120 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil && isDockerUnavailable(err.Error()) {
		t.Skip("Docker tidak tersedia — lewati testcontainer load test")
	}
	require.NoError(t, err, "start timescale testcontainer")
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/netops_load_test?sslmode=disable", host, port.Port())

	// Wait for PostgreSQL readiness
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()

	var ready bool
	for i := 0; i < 30; i++ {
		if sqlDB.Ping() == nil {
			ready = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	require.True(t, ready, "timescaledb container must be ready within 30s")

	// Execute migrations
	absMigDir, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
	require.NoError(t, err)
	m, err := migrate.New("file://"+filepath.ToSlash(absMigDir), dsn)
	require.NoError(t, err, "create migrate instance")
	require.NoError(t, m.Up(), "execute migrations up to latest")

	// Configure GORM connection with tuned pool
	gormDB, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "open gorm connection")

	dbPool, err := gormDB.DB()
	require.NoError(t, err)
	dbPool.SetMaxOpenConns(50)
	dbPool.SetMaxIdleConns(25)
	dbPool.SetConnMaxLifetime(30 * time.Minute)
	dbPool.SetConnMaxIdleTime(5 * time.Minute)

	return gormDB, dsn
}

func isDockerUnavailable(msg string) bool {
	lower := strings.ToLower(msg)
	for _, marker := range []string{"docker", "daemon", "cannot connect", "connection refused"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// TestTimescaleDB_100RoutersPingSimulation simulates >= 100 routers emitting ping metrics every second,
// testing high-throughput batch ingestion, continuous aggregates downsampling, concurrent read/write stability,
// and verifying zero deadlock and low-latency queries.
func TestTimescaleDB_100RoutersPingSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 100-router load test in short mode")
	}

	t.Log("=== Starting TimescaleDB Testcontainer (PostgreSQL 16 + TimescaleDB) ===")
	gormDB, _ := setupTimescaleTestContainer(t)
	repo := postgres.NewMetricsRepository(gormDB)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 1. Verify TimescaleDB hypertable & Continuous Aggregate View
	isTimescale, err := repo.IsTimescaleDBAvailable(ctx)
	require.NoError(t, err)
	require.True(t, isTimescale, "TimescaleDB hypertable must be available")

	isCagg, err := isContinuousAggregateReady(gormDB)
	require.NoError(t, err)
	require.True(t, isCagg, "Continuous aggregate device_ping_metrics_1m must be available")
	require.True(t, repo.IsContinuousAggregateAvailable(ctx), "repo continuous aggregate check must be true")

	// 2. Setup 100 distinct routers (devices)
	const numRouters = 100
	routerIDs := make([]string, numRouters)
	for i := 0; i < numRouters; i++ {
		routerIDs[i] = uuid.New().String()
	}
	t.Logf("Initialized %d simulated routers with unique UUIDs", numRouters)

	// -------------------------------------------------------------------------
	// SCENARIO 1: Realistic 100-Router Per-Second Ping Streaming (10 seconds)
	// -------------------------------------------------------------------------
	t.Log("=== Scenario 1: Simulating 100 routers generating 1 ping/second for 10s (1,000 points) ===")
	const streamSeconds = 10
	var totalIngested atomic.Int64
	var totalIngestErrors atomic.Int64
	var readQueriesCount atomic.Int64
	var readQueryErrors atomic.Int64

	simStart := time.Now().UTC()

	var wg sync.WaitGroup
	// Launch 100 concurrent router stream goroutines
	for rIdx := 0; rIdx < numRouters; rIdx++ {
		wg.Add(1)
		go func(deviceID string, routerNum int) {
			defer wg.Done()
			var buffer []device.PingMetricPoint

			for sec := 0; sec < streamSeconds; sec++ {
				// Simulasikan RTT bervariasi antara 15ms - 45ms
				rtt := float32(18.0 + rand.Float64()*22.0)
				loss := 0
				if rand.Float64() < 0.02 { // 2% packet loss simulation
					loss = 100
					rtt = 0
				}
				status := "connected"
				if loss == 100 {
					status = "timeout"
				}

				pt := device.PingMetricPoint{
					RecordedAt: simStart.Add(time.Duration(sec) * time.Second),
					DeviceID:   deviceID,
					Target:     "8.8.8.8",
					Seq:        sec,
					RTTMS:      rtt,
					Status:     status,
					TTL:        56,
					Sent:       sec + 1,
					Received:   sec + 1,
					PacketLoss: loss,
				}
				buffer = append(buffer, pt)

				// Flush buffer setiap 5 item atau di detik terakhir
				if len(buffer) >= 5 || sec == streamSeconds-1 {
					insertCtx, insertCancel := context.WithTimeout(ctx, 5*time.Second)
					err := repo.SavePingMetricsBatch(insertCtx, buffer)
					insertCancel()
					if err != nil {
						totalIngestErrors.Add(1)
						t.Errorf("Router %d SavePingMetricsBatch error: %v", routerNum, err)
					} else {
						totalIngested.Add(int64(len(buffer)))
					}
					buffer = buffer[:0]
				}
			}
		}(routerIDs[rIdx], rIdx)
	}

	// Concurrent Readers: 5 goroutines simulating live dashboard / technician portal queries
	readerStop := make(chan struct{})
	var readerWg sync.WaitGroup
	for readerID := 0; readerID < 5; readerID++ {
		readerWg.Add(1)
		go func(rID int) {
			defer readerWg.Done()
			for {
				select {
				case <-readerStop:
					return
				default:
					randomRouter := routerIDs[rand.Intn(numRouters)]
					queryCtx, queryCancel := context.WithTimeout(ctx, 3*time.Second)
					qStart := time.Now()
					_, summary, err := repo.QueryPingMetrics(queryCtx, device.PingMetricsFilter{
						DeviceID:       randomRouter,
						StartTime:      simStart.Add(-1 * time.Minute),
						EndTime:        simStart.Add(time.Duration(streamSeconds+5) * time.Second),
						BucketInterval: "1m",
					})
					queryCancel()
					elapsed := time.Since(qStart)

					if err != nil {
						readQueryErrors.Add(1)
					} else {
						readQueriesCount.Add(1)
						assert.Less(t, elapsed, 800*time.Millisecond, "Continuous aggregate query should complete under 800ms during heavy concurrent ingestion")
						_ = summary
					}
					time.Sleep(50 * time.Millisecond)
				}
			}
		}(readerID)
	}

	// Wait for 100 router ingestion streams to finish
	wg.Wait()
	close(readerStop)
	readerWg.Wait()

	expectedPoints := int64(numRouters * streamSeconds)
	t.Logf("Scenario 1 Ingestion Complete: Ingested=%d/%d points, IngestErrors=%d, ReadQueries=%d, ReadErrors=%d",
		totalIngested.Load(), expectedPoints, totalIngestErrors.Load(), readQueriesCount.Load(), readQueryErrors.Load())

	assert.Equal(t, int64(0), totalIngestErrors.Load(), "must have 0 ingestion errors during concurrent streaming")
	assert.Equal(t, expectedPoints, totalIngested.Load(), "all 1,000 points must be successfully persisted")
	assert.Equal(t, int64(0), readQueryErrors.Load(), "must have 0 read query errors during streaming")

	// -------------------------------------------------------------------------
	// SCENARIO 2: Peak Burst Batch Ingestion (100 Parallel Flushes of 30 Points = 3,000 points)
	// -------------------------------------------------------------------------
	t.Log("=== Scenario 2: Peak Burst - 100 routers simultaneously flushing full batches of 30 points (3,000 points) ===")
	burstStart := time.Now()
	var burstIngested atomic.Int64
	var burstErrors atomic.Int64

	var burstWg sync.WaitGroup
	for rIdx := 0; rIdx < numRouters; rIdx++ {
		burstWg.Add(1)
		go func(deviceID string, routerNum int) {
			defer burstWg.Done()
			batch := make([]device.PingMetricPoint, 30)
			now := time.Now().UTC()
			for p := 0; p < 30; p++ {
				batch[p] = device.PingMetricPoint{
					RecordedAt: now.Add(time.Duration(p) * time.Second),
					DeviceID:   deviceID,
					Target:     "1.1.1.1",
					Seq:        100 + p,
					RTTMS:      float32(20.0 + float32(p%10)),
					Status:     "connected",
					TTL:        58,
					Sent:       100 + p,
					Received:   100 + p,
					PacketLoss: 0,
				}
			}

			bCtx, bCancel := context.WithTimeout(ctx, 5*time.Second)
			defer bCancel()
			err := repo.SavePingMetricsBatch(bCtx, batch)
			if err != nil {
				burstErrors.Add(1)
				t.Errorf("Burst flush error on router %d: %v", routerNum, err)
			} else {
				burstIngested.Add(int64(len(batch)))
			}
		}(routerIDs[rIdx], rIdx)
	}

	burstWg.Wait()
	burstDuration := time.Since(burstStart)
	throughput := float64(burstIngested.Load()) / burstDuration.Seconds()

	t.Logf("Scenario 2 Burst Ingestion Complete: Ingested=%d points in %v (Throughput: %.2f points/sec), Errors=%d",
		burstIngested.Load(), burstDuration, throughput, burstErrors.Load())

	assert.Equal(t, int64(0), burstErrors.Load(), "burst flush must have zero errors")
	assert.Equal(t, int64(3000), burstIngested.Load(), "burst flush must persist all 3,000 points")
	assert.Greater(t, throughput, float64(500), "throughput should exceed 500 points/sec on batch ingestion")

	// -------------------------------------------------------------------------
	// SCENARIO 3: Downsampling & Continuous Aggregate Validation
	// -------------------------------------------------------------------------
	t.Log("=== Scenario 3: Validating Continuous Aggregate 1m Downsampling & Summary Stats ===")
	// Query random router
	testRouter := routerIDs[0]
	qStart := time.Now()
	points, summary, err := repo.QueryPingMetrics(ctx, device.PingMetricsFilter{
		DeviceID:       testRouter,
		StartTime:      simStart.Add(-1 * time.Minute),
		EndTime:        time.Now().UTC().Add(1 * time.Minute),
		BucketInterval: "1m",
	})
	qElapsed := time.Since(qStart)

	require.NoError(t, err, "downsampled query must succeed")
	assert.Greater(t, len(points), 0, "must return aggregated bucket points")
	assert.Greater(t, summary.TotalSamples, int64(0), "summary total samples must be > 0")
	assert.Greater(t, summary.AvgRTT, float32(0), "summary average RTT must be > 0")
	assert.Less(t, qElapsed, 50*time.Millisecond, "continuous aggregate query should take < 50ms")

	t.Logf("QueryPingMetrics verification passed: router=%s, points=%d, avgRTT=%.2fms, minRTT=%.2fms, maxRTT=%.2fms, queryTime=%v",
		testRouter, len(points), summary.AvgRTT, summary.MinRTT, summary.MaxRTT, qElapsed)

	// -------------------------------------------------------------------------
	// SCENARIO 4: Verify Hypertable Properties & Compression Policy
	// -------------------------------------------------------------------------
	t.Log("=== Scenario 4: Verifying TimescaleDB Hypertable & Compression Policy ===")
	var compressionEnabled bool
	err = gormDB.Raw(`
		SELECT compression_enabled 
		FROM timescaledb_information.hypertables 
		WHERE hypertable_name = 'device_ping_metrics'
	`).Scan(&compressionEnabled).Error
	require.NoError(t, err)
	assert.True(t, compressionEnabled, "compression_enabled must be TRUE on device_ping_metrics")

	var chunkInterval string
	err = gormDB.Raw(`
		SELECT time_interval 
		FROM timescaledb_information.dimensions 
		WHERE hypertable_name = 'device_ping_metrics'
	`).Scan(&chunkInterval).Error
	require.NoError(t, err)
	assert.Contains(t, chunkInterval, "1 day", "chunk interval must be 1 day")

	t.Logf("TimescaleDB verification passed: compression_enabled=%v, chunk_interval=%s",
		compressionEnabled, chunkInterval)
}

func isContinuousAggregateReady(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Raw(`
		SELECT count(1) 
		FROM timescaledb_information.continuous_aggregates 
		WHERE view_name = 'device_ping_metrics_1m'
	`).Scan(&count).Error
	return count > 0, err
}
