package inventory

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

const (
	// probeTimeout bounds one management-endpoint connection attempt.
	probeTimeout = 1200 * time.Millisecond
	// probeBudget bounds the complete connectivity pass, including queueing.
	probeBudget = 2400 * time.Millisecond
	// probeConcurrency caps simultaneous dials so a large inventory cannot
	// exhaust file descriptors or the caller's connection budget.
	probeConcurrency = 16

	connectivityOnline  = "online"
	connectivityOffline = "offline"
	connectivityUnknown = "unknown"
)

// DeviceDTO represents the non-sensitive public representation of an inventory device.
type DeviceDTO struct {
	ID         string   `json:"id"`
	TenantID   string   `json:"tenant_id,omitempty"`
	Name       string   `json:"name"`
	Vendor     string   `json:"vendor"`
	DriverType string   `json:"driver_type"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	SSHPort    int      `json:"ssh_port,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Enabled    bool     `json:"enabled"`
	// Status is the inventory registration state: "active" or "disabled".
	Status string `json:"status"`
	// Connectivity is the live management-endpoint reachability: "online",
	// "offline", or "unknown" when the device is disabled and was not probed.
	Connectivity string `json:"connectivity"`
	// LatencyMS is the successful dial duration in whole milliseconds; nil when
	// the probe failed, so a fast successful dial still reports 0.
	LatencyMS *int64 `json:"latency_ms,omitempty"`
}

// Handler serves inventory HTTP endpoints for workbench and UI integrations.
type Handler struct {
	devices port.DeviceRepository
}

// NewHandler constructs an inventory HTTP handler.
func NewHandler(devices port.DeviceRepository) *Handler {
	return &Handler{devices: devices}
}

// RegisterProtected mounts device inventory endpoints on the authenticated mux.
func (h *Handler) RegisterProtected(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/devices", h.handleListDevices)
	mux.HandleFunc("GET /api/v1/devices", h.handleListDevices)
}

func (h *Handler) handleListDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.devices == nil {
		_ = json.NewEncoder(w).Encode(map[string]any{"devices": []DeviceDTO{}})
		return
	}

	devices, err := h.devices.FindAll(r.Context())
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	probes := probeDevices(r.Context(), devices)

	dtos := make([]DeviceDTO, 0, len(devices))
	for i, d := range devices {
		status := "disabled"
		if d.Enabled {
			status = "active"
		}
		dtos = append(dtos, DeviceDTO{
			ID:           d.ID,
			TenantID:     d.TenantID,
			Name:         d.Name,
			Vendor:       d.Vendor,
			DriverType:   d.DriverType,
			Host:         d.Host,
			Port:         d.Port,
			SSHPort:      d.SSHPort,
			Tags:         d.Tags,
			Enabled:      d.Enabled,
			Status:       status,
			Connectivity: probes[i].connectivity,
			LatencyMS:    probes[i].latencyMS,
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"devices": dtos})
}

// probeResult is the connectivity fact derived for one inventory device.
type probeResult struct {
	connectivity string
	latencyMS    *int64
}

// probeDevices dials every enabled device's management endpoint concurrently.
// Disabled devices are not probed: their registration state already answers the
// question, and a disabled device is not expected to accept connections.
func probeDevices(ctx context.Context, devices []device.Device) []probeResult {
	results := make([]probeResult, len(devices))
	ctx, cancel := context.WithTimeout(ctx, probeBudget)
	defer cancel()

	sem := make(chan struct{}, probeConcurrency)
	var wg sync.WaitGroup
	for i := range devices {
		if !devices[i].Enabled {
			results[i].connectivity = connectivityUnknown
			continue
		}
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[index].connectivity = connectivityOffline
				return
			}
			results[index] = probeDevice(ctx, devices[index])
		}(i)
	}
	wg.Wait()
	return results
}

// probeDevice reports whether the management endpoint accepts a TCP connection.
func probeDevice(ctx context.Context, d device.Device) probeResult {
	host := strings.TrimSpace(d.Host)
	if host == "" {
		return probeResult{connectivity: connectivityOffline}
	}
	port := d.Port
	if port <= 0 {
		port = d.SSHPort
	}
	if port <= 0 {
		port = 22
	}

	dialer := net.Dialer{Timeout: probeTimeout}
	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return probeResult{connectivity: connectivityOffline}
	}
	_ = conn.Close()
	return probeResult{connectivity: connectivityOnline, latencyMS: &latency}
}
