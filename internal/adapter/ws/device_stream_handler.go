package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

// DeviceStatusItem holds a device inventory record along with its live diagnostic result.
type DeviceStatusItem struct {
	Device device.Device             `json:"device"`
	Test   business.DeviceTestResult `json:"test"`
}

// DeviceStreamHandler streams live device status and inventory updates over WebSockets / SSE using native wire streaming.
type DeviceStreamHandler struct {
	useCase        *business.ManageDeviceUseCase
	driverProvider StreamDriverProvider
}

// NewDeviceStreamHandler constructs a new DeviceStreamHandler.
func NewDeviceStreamHandler(uc *business.ManageDeviceUseCase, provider StreamDriverProvider) *DeviceStreamHandler {
	return &DeviceStreamHandler{
		useCase:        uc,
		driverProvider: provider,
	}
}

// StreamSingleDeviceStatus uses MikroTik native wire streaming (driver.Stream) on /system/resource/print follow-only=true.
// Zero polling, zero tickers — it listens directly on the TCP socket channel push from RouterOS.
func (h *DeviceStreamHandler) StreamSingleDeviceStatus(ctx context.Context, deviceID string, outChan chan<- []byte) error {
	dev, err := h.useCase.GetDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	if !dev.Enabled {
		item := DeviceStatusItem{
			Device: dev,
			Test: business.DeviceTestResult{
				DeviceID: dev.ID,
				Status:   "disabled",
				Message:  "Device is disabled",
			},
		}
		data, _ := json.Marshal(item)
		select {
		case outChan <- data:
		case <-ctx.Done():
		}
		return nil
	}

	driver, err := h.driverProvider(ctx, deviceID)
	if err != nil {
		item := DeviceStatusItem{
			Device: dev,
			Test: business.DeviceTestResult{
				DeviceID: dev.ID,
				Status:   "failed",
				Message:  err.Error(),
			},
		}
		data, _ := json.Marshal(item)
		select {
		case outChan <- data:
		case <-ctx.Done():
		}
		return err
	}

	// Issue native wire streaming command (/system/resource/print follow-only=true interval=1s)
	cmd := mikrotik.NewStreamSystemResourceCommand("1s")
	handle, err := driver.Stream(ctx, cmd)
	if err != nil {
		item := DeviceStatusItem{
			Device: dev,
			Test: business.DeviceTestResult{
				DeviceID: dev.ID,
				Status:   "failed",
				Message:  fmt.Sprintf("failed to start native stream: %v", err),
			},
		}
		data, _ := json.Marshal(item)
		select {
		case outChan <- data:
		case <-ctx.Done():
		}
		return err
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-handle.Chan():
			if !ok {
				item := DeviceStatusItem{
					Device: dev,
					Test: business.DeviceTestResult{
						DeviceID: dev.ID,
						Status:   "failed",
						Message:  "native stream disconnected",
					},
				}
				data, _ := json.Marshal(item)
				select {
				case outChan <- data:
				case <-ctx.Done():
				}
				return handle.Err()
			}

			sysRes := mikrotik.ParseSystemResource(res)
			latestDev, errGet := h.useCase.GetDevice(ctx, deviceID)
			if errGet == nil {
				dev = latestDev
			}

			item := DeviceStatusItem{
				Device: dev,
				Test: business.DeviceTestResult{
					DeviceID:  dev.ID,
					Status:    "connected",
					LatencyMS: 0,
					Identity:  dev.Name,
					Version:   sysRes.Version,
					BoardName: sysRes.BoardName,
					Uptime:    sysRes.Uptime,
					Message:   "Streaming live from MikroTik socket",
				},
			}
			data, err := json.Marshal(item)
			if err == nil {
				select {
				case outChan <- data:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

// StreamDevicesStatus subscribes to native wire streams for all active devices and pushes aggregated updates on every incoming frame.
func (h *DeviceStreamHandler) StreamDevicesStatus(ctx context.Context, outChan chan<- []byte) error {
	devices, err := h.useCase.ListDevices(ctx)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		data, _ := json.Marshal([]DeviceStatusItem{})
		select {
		case outChan <- data:
		case <-ctx.Done():
		}
		return nil
	}

	statusMap := make(map[string]DeviceStatusItem)
	var mu sync.Mutex

	pushSnapshot := func() {
		mu.Lock()
		items := make([]DeviceStatusItem, 0, len(statusMap))
		for _, dev := range devices {
			if item, exists := statusMap[dev.ID]; exists {
				items = append(items, item)
			} else {
				items = append(items, DeviceStatusItem{
					Device: dev,
					Test: business.DeviceTestResult{
						DeviceID: dev.ID,
						Status:   "connecting",
						Message:  "Connecting native stream...",
					},
				})
			}
		}
		data, err := json.Marshal(items)
		mu.Unlock()

		if err == nil {
			select {
			case outChan <- data:
			case <-ctx.Done():
			}
		}
	}

	ctxChild, cancelChild := context.WithCancel(ctx)
	defer cancelChild()

	for _, dev := range devices {
		devCopy := dev
		go func() {
			singleChan := make(chan []byte, 10)
			go func() {
				_ = h.StreamSingleDeviceStatus(ctxChild, devCopy.ID, singleChan)
				close(singleChan)
			}()

			for msg := range singleChan {
				var item DeviceStatusItem
				if err := json.Unmarshal(msg, &item); err == nil {
					mu.Lock()
					statusMap[devCopy.ID] = item
					mu.Unlock()
					pushSnapshot()
				}
			}
		}()
	}

	<-ctx.Done()
	return ctx.Err()
}
