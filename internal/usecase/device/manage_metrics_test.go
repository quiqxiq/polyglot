package device

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/device"
)

type mockMetricsRepo struct {
	available   bool
	savedPoints []device.PingMetricPoint
	returnQuery []device.PingMetricPoint
	summary     device.PingSummary
}

func (m *mockMetricsRepo) IsTimescaleDBAvailable(ctx context.Context) (bool, error) {
	return m.available, nil
}

func (m *mockMetricsRepo) SavePingMetric(ctx context.Context, point device.PingMetricPoint) error {
	m.savedPoints = append(m.savedPoints, point)
	return nil
}

func (m *mockMetricsRepo) SavePingMetricsBatch(ctx context.Context, points []device.PingMetricPoint) error {
	m.savedPoints = append(m.savedPoints, points...)
	return nil
}

func (m *mockMetricsRepo) QueryPingMetrics(ctx context.Context, filter device.PingMetricsFilter) ([]device.PingMetricPoint, device.PingSummary, error) {
	return m.returnQuery, m.summary, nil
}

func (m *mockMetricsRepo) CleanupExpiredMetrics(ctx context.Context, deviceID string, retentionDays int) error {
	return nil
}

type mockAuthorizer struct {
	allowed bool
}

func (a *mockAuthorizer) CanAccessDevice(ctx context.Context, userID uint, roles []string, deviceID string) (bool, error) {
	return a.allowed, nil
}

func TestManageMetricsUseCase_PingConfig(t *testing.T) {
	ctx := context.Background()
	devRepo := newMockDeviceRepo()
	metricsRepo := &mockMetricsRepo{available: true}
	authorizer := &mockAuthorizer{allowed: true}

	dev := device.Device{
		ID:      "router-test-01",
		Name:    "Main Router",
		Host:    "192.168.1.1",
		Enabled: true,
	}
	err := devRepo.Save(ctx, dev)
	require.NoError(t, err)

	uc := NewManageMetricsUseCase(devRepo, metricsRepo, authorizer)

	// 1. Get default ping config
	cfg, available, err := uc.GetPingConfig(ctx, "router-test-01", 1, []string{"admin"})
	require.NoError(t, err)
	assert.True(t, available)
	assert.False(t, cfg.Enabled)
	assert.Equal(t, "8.8.8.8", cfg.Target)
	assert.Equal(t, 7, cfg.RetentionDays)

	// 2. Update ping config
	newCfg := device.DevicePingConfig{
		Enabled:       true,
		Target:        "1.1.1.1",
		RetentionDays: 14,
	}
	updatedCfg, err := uc.UpdatePingConfig(ctx, "router-test-01", newCfg, 1, []string{"admin"})
	require.NoError(t, err)
	assert.True(t, updatedCfg.Enabled)
	assert.Equal(t, "1.1.1.1", updatedCfg.Target)
	assert.Equal(t, 14, updatedCfg.RetentionDays)

	// 3. Re-get and verify persistence
	cfgAfter, _, err := uc.GetPingConfig(ctx, "router-test-01", 1, []string{"admin"})
	require.NoError(t, err)
	assert.True(t, cfgAfter.Enabled)
	assert.Equal(t, "1.1.1.1", cfgAfter.Target)
	assert.Equal(t, 14, cfgAfter.RetentionDays)
}

func TestManageMetricsUseCase_QueryMetrics(t *testing.T) {
	ctx := context.Background()
	devRepo := newMockDeviceRepo()
	metricsRepo := &mockMetricsRepo{
		available: true,
		returnQuery: []device.PingMetricPoint{
			{
				RecordedAt: time.Now().UTC(),
				DeviceID:   "router-01",
				Target:     "8.8.8.8",
				RTTMS:      22.4,
				Status:     "connected",
			},
		},
		summary: device.PingSummary{
			MinRTT:        22.4,
			AvgRTT:        22.4,
			MaxRTT:        22.4,
			TotalSamples:  1,
			PacketLossPct: 0,
		},
	}
	authorizer := &mockAuthorizer{allowed: true}

	uc := NewManageMetricsUseCase(devRepo, metricsRepo, authorizer)

	points, summary, available, err := uc.QueryPingMetrics(ctx, device.PingMetricsFilter{
		DeviceID: "router-01",
	}, 1, []string{"owner"})

	require.NoError(t, err)
	assert.True(t, available)
	assert.Len(t, points, 1)
	assert.Equal(t, float32(22.4), summary.AvgRTT)
}
