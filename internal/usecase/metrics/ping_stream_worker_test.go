package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
)

type workerDeviceRepo struct{}

func (workerDeviceRepo) Save(context.Context, device.Device) error { return nil }
func (workerDeviceRepo) FindByID(context.Context, string) (device.Device, error) {
	return device.Device{}, device.ErrNotFound
}
func (workerDeviceRepo) FindAll(context.Context) ([]device.Device, error) { return nil, nil }
func (workerDeviceRepo) Update(context.Context, device.Device) error      { return nil }
func (workerDeviceRepo) Delete(context.Context, string) error             { return nil }
func (workerDeviceRepo) FindByUserScope(context.Context, uint) ([]device.Device, error) {
	return nil, nil
}

type workerMetricsRepo struct{}

func (workerMetricsRepo) IsTimescaleDBAvailable(context.Context) (bool, error) {
	return false, nil
}
func (workerMetricsRepo) SavePingMetric(context.Context, device.PingMetricPoint) error {
	return nil
}
func (workerMetricsRepo) SavePingMetricsBatch(context.Context, []device.PingMetricPoint) error {
	return nil
}
func (workerMetricsRepo) QueryPingMetrics(context.Context, device.PingMetricsFilter) ([]device.PingMetricPoint, device.PingSummary, error) {
	return nil, device.PingSummary{}, nil
}
func (workerMetricsRepo) CleanupExpiredMetrics(context.Context, string, int) error { return nil }

func TestPingStreamWorker_StartStopIsIdempotent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	worker := NewPingStreamWorker(workerDeviceRepo{}, workerMetricsRepo{}, nil)

	worker.Start(ctx)
	worker.Start(ctx)
	cancel()
	worker.Stop()
	worker.Stop()

	time.Sleep(100 * time.Millisecond)
}
