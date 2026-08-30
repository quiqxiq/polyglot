package metrics

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/ping"
)

// DriverGetter fetches an active port.DeviceDriver by device ID.
type DriverGetter func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

// PingStreamWorker oversees continuous background streaming ping collection for enabled routers.
type PingStreamWorker struct {
	deviceRepo   port.DeviceRepository
	metricsRepo  port.MetricsRepository
	driverGetter DriverGetter

	mu       sync.Mutex
	cancelFn map[string]context.CancelFunc
	running  bool
}

// NewPingStreamWorker constructs a new PingStreamWorker.
func NewPingStreamWorker(
	deviceRepo port.DeviceRepository,
	metricsRepo port.MetricsRepository,
	getter DriverGetter,
) *PingStreamWorker {
	return &PingStreamWorker{
		deviceRepo:   deviceRepo,
		metricsRepo:  metricsRepo,
		driverGetter: getter,
		cancelFn:     make(map[string]context.CancelFunc),
	}
}

// Start launches the supervisor loop and retention cleanup worker.
func (m *PingStreamWorker) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	go m.supervisorLoop(ctx)
	go m.cleanupLoop(ctx)
}

// Stop terminates all active ping streams.
func (m *PingStreamWorker) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	for id, cancel := range m.cancelFn {
		cancel()
		delete(m.cancelFn, id)
	}
}

func (m *PingStreamWorker) supervisorLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Initial sync
	m.syncActiveStreams(ctx)

	for {
		select {
		case <-ctx.Done():
			m.Stop()
			return
		case <-ticker.C:
			m.syncActiveStreams(ctx)
		}
	}
}

func (m *PingStreamWorker) syncActiveStreams(ctx context.Context) {
	if m.deviceRepo == nil || m.metricsRepo == nil {
		return
	}

	available, err := m.metricsRepo.IsTimescaleDBAvailable(ctx)
	if err != nil || !available {
		return
	}

	devices, err := m.deviceRepo.FindAll(ctx)
	if err != nil {
		return
	}

	activeDeviceMap := make(map[string]device.Device)
	for _, dev := range devices {
		if dev.Enabled && dev.PingConfig().Enabled {
			activeDeviceMap[dev.ID] = dev
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Stop removed or disabled streams
	for id, cancel := range m.cancelFn {
		if _, stillActive := activeDeviceMap[id]; !stillActive {
			cancel()
			delete(m.cancelFn, id)
			logger.WithComponent("PingStreamWorker").WithField("device_id", id).Info("background ping stream stopped")
		}
	}

	// 2. Start newly enabled streams
	for id, dev := range activeDeviceMap {
		if _, alreadyRunning := m.cancelFn[id]; !alreadyRunning {
			streamCtx, cancel := context.WithCancel(ctx)
			m.cancelFn[id] = cancel
			go m.runDeviceStream(streamCtx, dev)
			logger.WithComponent("PingStreamWorker").WithFields(map[string]any{
				"device_id": id,
				"target":    dev.PingConfig().Target,
			}).Info("background ping stream started")
		}
	}
}

func (m *PingStreamWorker) runDeviceStream(ctx context.Context, dev device.Device) {
	cfg := dev.PingConfig()
	target := cfg.Target
	if target == "" {
		target = "8.8.8.8"
	}
	if hostOnly, _, err := net.SplitHostPort(target); err == nil {
		target = hostOnly
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if m.driverGetter == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		drv, err := m.driverGetter(ctx, dev.ID)
		if err != nil || drv == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		sDrv, ok := drv.(port.StreamingDeviceDriver)
		if !ok {
			time.Sleep(5 * time.Second)
			continue
		}

		pingCmd := command.Command{
			Raw: "/ping",
			Args: map[string]string{
				"address":  target,
				"interval": "1s",
			},
		}
		handle, err := sDrv.Stream(ctx, pingCmd)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}

		m.consumeStream(ctx, dev.ID, target, handle)
		_ = handle.Cancel()

		// Short backoff before reconnecting stream
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (m *PingStreamWorker) consumeStream(
	ctx context.Context,
	deviceID string,
	target string,
	handle port.StreamHandle,
) {
	buffer := make([]device.PingMetricPoint, 0, 10)
	flushTicker := time.NewTicker(3 * time.Second)
	defer flushTicker.Stop()
	streamSeq := 0

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		_ = m.metricsRepo.SavePingMetricsBatch(ctx, buffer)
		buffer = buffer[:0]
	}
	defer flush()

	for {
		select {
		case <-ctx.Done():
			return
		case <-flushTicker.C:
			flush()
		case res, ok := <-handle.Chan():
			if !ok {
				return
			}
			if len(res.Rows) > 0 {
				row := res.Rows[0]
				latency, status := ping.ParsePingLatency(row)
				seq, _ := strconv.Atoi(row["seq"])
				if seq == 0 {
					if s, ok := row["sequence"]; ok {
						seq, _ = strconv.Atoi(s)
					}
					if seq == 0 {
						seq = streamSeq
					}
				}
				streamSeq = seq + 1
				ttl, _ := strconv.Atoi(row["ttl"])
				size, _ := strconv.Atoi(row["size"])
				sent, _ := strconv.Atoi(row["sent"])
				recv, _ := strconv.Atoi(row["received"])
				loss := ping.ParsePacketLoss(row["packet-loss"])
				minRTT := ping.ParseDurationMs(row["min-rtt"])
				avgRTT := ping.ParseDurationMs(row["avg-rtt"])
				maxRTT := ping.ParseDurationMs(row["max-rtt"])

				var minPtr, avgPtr, maxPtr *float32
				if minRTT > 0 {
					v := float32(minRTT)
					minPtr = &v
				}
				if avgRTT > 0 {
					v := float32(avgRTT)
					avgPtr = &v
				}
				if maxRTT > 0 {
					v := float32(maxRTT)
					maxPtr = &v
				}

				buffer = append(buffer, device.PingMetricPoint{
					RecordedAt: time.Now().UTC(),
					DeviceID:   deviceID,
					Target:     target,
					Seq:        seq,
					Size:       size,
					TTL:        ttl,
					RTTMS:      float32(latency),
					Status:     status,
					Sent:       sent,
					Received:   recv,
					PacketLoss: int(loss),
					MinRTTMS:   minPtr,
					AvgRTTMS:   avgPtr,
					MaxRTTMS:   maxPtr,
				})

				if len(buffer) >= 10 {
					flush()
				}
			}
		}
	}
}

func (m *PingStreamWorker) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			devices, err := m.deviceRepo.FindAll(ctx)
			if err != nil {
				continue
			}
			for _, dev := range devices {
				retentionDays := dev.PingConfig().RetentionDays
				if retentionDays <= 0 {
					retentionDays = 7
				}
				if err := m.metricsRepo.CleanupExpiredMetrics(ctx, dev.ID, retentionDays); err != nil {
					logger.WithComponent("PingStreamWorker").WithField("device_id", dev.ID).Warn(fmt.Sprintf("cleanup metrics error: %v", err))
				}
			}
		}
	}
}
