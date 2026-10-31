package port

import (
	"context"
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/ont"
)

// ACSDeviceSummary ringkasan device dari GenieACS NBI.
type ACSDeviceSummary struct {
	DeviceID        string
	OUI             string
	ProductClass    string
	SerialNumber    string
	SoftwareVersion string
	LastInform      time.Time
	Tags            []string
	MACCandidates   []string
	Raw             json.RawMessage
}

// ACSTelemetry hasil ekstraksi parameter telemetri ONT dari ACS.
type ACSTelemetry struct {
	SerialNumber    string
	ProductClass    string
	SoftwareVersion string
	HardwareVersion string
	UptimeSeconds   int64
	WANIP           string
	WANUsername     string
	WANStatus       string
	SSID24          string
	SSID5           string
	WiFiClients     int
	RxPowerDBM      *float64
	TemperatureC    *float64
	Raw             map[string]string
	FetchedAt       time.Time
}

// ACSManager adalah operasi ke server GenieACS (NBI) — bukan per-CPE driver.
type ACSManager interface {
	Ping(ctx context.Context) error
	ListDevices(ctx context.Context, query string) ([]ACSDeviceSummary, error)
	GetDevice(ctx context.Context, deviceID string) (ACSDeviceSummary, error)
	// FetchTelemetry meminta getParameterValues (best effort) lalu membaca
	// nilai terbaru dari device record GenieACS.
	FetchTelemetry(ctx context.Context, deviceID string, catalog ont.TelemetryCatalog) (ACSTelemetry, error)
	SetParameters(ctx context.Context, deviceID string, values map[string]string) error
	Reboot(ctx context.Context, deviceID string) error
	EnsurePreset(ctx context.Context, p ont.ACSPreset) error
	DeletePreset(ctx context.Context, name string) error
	ListFaults(ctx context.Context, deviceID string) ([]string, error)
}
