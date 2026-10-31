// Package acsmanager menjembatani GenieACS NBI ke port.ACSManager:
// usecase/ont hanya melihat operasi telemetri/aksi, bukan detail HTTP NBI.
package acsmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/driver/genieacs"
	"github.com/quixiq/polyglot/internal/port"
)

// SettingsReader membaca konfigurasi ACS (server device id).
type SettingsReader interface {
	GetValue(ctx context.Context, key, fallback string) string
}

// GenieACS mengimplementasikan port.ACSManager di atas NBI GenieACS.
type GenieACS struct {
	devices  port.DeviceRepository
	vault    port.CredentialVault
	settings SettingsReader

	mu     sync.Mutex
	cached *genieacs.NBI
	key    string
}

var _ port.ACSManager = (*GenieACS)(nil)

// New membangun adapter ACS manager.
func New(devices port.DeviceRepository, vault port.CredentialVault, settings SettingsReader) *GenieACS {
	return &GenieACS{devices: devices, vault: vault, settings: settings}
}

// client me-resolve konfigurasi server ACS lalu membangun NBI client.
func (g *GenieACS) client(ctx context.Context) (*genieacs.NBI, error) {
	serverID := g.settings.GetValue(ctx, "acs.server_device_id", "")
	if strings.TrimSpace(serverID) == "" {
		return nil, fmt.Errorf("%w: acs.server_device_id kosong", ont.ErrACSNotConfigured)
	}
	g.mu.Lock()
	cacheKey := serverID
	if g.cached != nil && g.key == cacheKey {
		cached := g.cached
		g.mu.Unlock()
		return cached, nil
	}
	g.mu.Unlock()

	dev, err := g.devices.FindByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("acsmanager: server NBI %s: %w", serverID, err)
	}
	scheme := "http"
	if dev.Extra["use_tls"] == "true" {
		scheme = "https"
	}
	nbiPort := dev.Port
	if nbiPort <= 0 {
		nbiPort = 7557
	}
	rawURL := fmt.Sprintf("%s://%s:%d", scheme, dev.Host, nbiPort)
	apiKey := ""
	creds, errCred := g.vault.Get(ctx, serverID)
	if errCred == nil {
		apiKey = creds.Extra["api_key"]
	}
	timeout := time.Duration(dev.TimeoutMS) * time.Millisecond
	client, err := genieacs.NewNBI(rawURL, apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("acsmanager: %w", err)
	}
	g.mu.Lock()
	g.cached = client
	g.key = cacheKey
	g.mu.Unlock()
	return client, nil
}

// Ping memeriksa konektivitas NBI.
func (g *GenieACS) Ping(ctx context.Context) error {
	client, err := g.client(ctx)
	if err != nil {
		return err
	}
	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("%w: %v", ont.ErrACSNotConfigured, err)
	}
	return nil
}

// ListDevices mengambil cache device dari NBI.
func (g *GenieACS) ListDevices(ctx context.Context, query string) ([]port.ACSDeviceSummary, error) {
	client, err := g.client(ctx)
	if err != nil {
		return nil, err
	}
	catalog := ont.DefaultTelemetryCatalog()
	records, err := client.ListDevices(ctx, genieacs.NBIListOptions{
		Query:      query,
		Projection: catalog.AllPaths(),
		Limit:      500,
	})
	if err != nil {
		return nil, fmt.Errorf("acsmanager: list devices: %w", err)
	}
	out := make([]port.ACSDeviceSummary, 0, len(records))
	for _, rec := range records {
		out = append(out, port.ACSDeviceSummary{
			DeviceID:        rec.DeviceID,
			OUI:             rec.OUI,
			ProductClass:    rec.ProductClass,
			SerialNumber:    rec.SerialNumber,
			SoftwareVersion: rec.SoftwareVersion,
			LastInform:      rec.LastInform,
			Tags:            rec.Tags,
			MACCandidates:   collectMACs(rec.Params),
			Raw:             rec.Raw,
		})
	}
	return out, nil
}

// GetDevice mengambil satu device dari NBI.
func (g *GenieACS) GetDevice(ctx context.Context, deviceID string) (port.ACSDeviceSummary, error) {
	client, err := g.client(ctx)
	if err != nil {
		return port.ACSDeviceSummary{}, err
	}
	rec, err := client.GetDevice(ctx, deviceID, nil)
	if err != nil {
		return port.ACSDeviceSummary{}, fmt.Errorf("acsmanager: get device: %w", err)
	}
	return port.ACSDeviceSummary{
		DeviceID:        rec.DeviceID,
		OUI:             rec.OUI,
		ProductClass:    rec.ProductClass,
		SerialNumber:    rec.SerialNumber,
		SoftwareVersion: rec.SoftwareVersion,
		LastInform:      rec.LastInform,
		Tags:            rec.Tags,
		MACCandidates:   collectMACs(rec.Params),
		Raw:             rec.Raw,
	}, nil
}

// FetchTelemetry meminta getParameterValues lalu membaca nilai terbaru dari
// device record (nilai tidak ada di response task GenieACS).
func (g *GenieACS) FetchTelemetry(ctx context.Context, deviceID string, catalog ont.TelemetryCatalog) (port.ACSTelemetry, error) {
	client, err := g.client(ctx)
	if err != nil {
		return port.ACSTelemetry{}, err
	}
	paths := catalog.AllPaths()
	body := map[string]any{"name": "getParameterValues", "parameterNames": paths}
	taskID, immediate, err := client.PostTask(ctx, deviceID, body, true)
	if err != nil {
		return port.ACSTelemetry{}, fmt.Errorf("acsmanager: telemetry task: %w", err)
	}
	if !immediate && !client.WaitTask(ctx, taskID, 2*time.Second) {
		return port.ACSTelemetry{}, fmt.Errorf("%w: task %s belum selesai", ont.ErrACSTaskFailed, taskID)
	}
	rec, err := client.GetDevice(ctx, deviceID, paths)
	if err != nil {
		return port.ACSTelemetry{}, fmt.Errorf("acsmanager: baca telemetry: %w", err)
	}
	return extractTelemetry(rec, catalog), nil
}

// SetParameters mengirim setParameterValues untuk parameter string.
func (g *GenieACS) SetParameters(ctx context.Context, deviceID string, values map[string]string) error {
	client, err := g.client(ctx)
	if err != nil {
		return err
	}
	params := make([][]string, 0, len(values))
	for name, value := range values {
		params = append(params, []string{name, value, "xsd:string"})
	}
	body := map[string]any{"name": "setParameterValues", "parameterValues": params}
	taskID, immediate, err := client.PostTask(ctx, deviceID, body, true)
	if err != nil {
		return fmt.Errorf("acsmanager: set parameters: %w", err)
	}
	if !immediate && !client.WaitTask(ctx, taskID, 2*time.Second) {
		return fmt.Errorf("%w: task %s queued", ont.ErrACSTaskFailed, taskID)
	}
	return nil
}

// Reboot mengirim task reboot TR-069.
func (g *GenieACS) Reboot(ctx context.Context, deviceID string) error {
	client, err := g.client(ctx)
	if err != nil {
		return err
	}
	taskID, immediate, err := client.PostTask(ctx, deviceID, map[string]any{"name": "reboot"}, true)
	if err != nil {
		return fmt.Errorf("acsmanager: reboot: %w", err)
	}
	if !immediate && !client.WaitTask(ctx, taskID, 2*time.Second) {
		return fmt.Errorf("%w: reboot %s queued", ont.ErrACSTaskFailed, taskID)
	}
	return nil
}

// EnsurePreset membuat/memperbarui preset bootstrap di GenieACS (idempoten).
func (g *GenieACS) EnsurePreset(ctx context.Context, p ont.ACSPreset) error {
	client, err := g.client(ctx)
	if err != nil {
		return err
	}
	payload := map[string]any{}
	if len(p.Precondition) > 0 {
		var pre any
		if errUnmarshal := json.Unmarshal(p.Precondition, &pre); errUnmarshal == nil {
			payload["precondition"] = pre
		}
	}
	var configs any
	if errUnmarshal := json.Unmarshal(p.Configurations, &configs); errUnmarshal == nil {
		payload["configurations"] = configs
	}
	if err := client.PutPreset(ctx, p.Name, payload); err != nil {
		return fmt.Errorf("acsmanager: put preset: %w", err)
	}
	return nil
}

// DeletePreset menghapus preset.
func (g *GenieACS) DeletePreset(ctx context.Context, name string) error {
	client, err := g.client(ctx)
	if err != nil {
		return err
	}
	if err := client.DeletePreset(ctx, name); err != nil {
		return fmt.Errorf("acsmanager: delete preset: %w", err)
	}
	return nil
}

// ListFaults mengembalikan fault message milik device.
func (g *GenieACS) ListFaults(ctx context.Context, deviceID string) ([]string, error) {
	client, err := g.client(ctx)
	if err != nil {
		return nil, err
	}
	faults, err := client.ListFaults(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("acsmanager: list faults: %w", err)
	}
	return faults, nil
}

// --- helpers ---

// collectMACs mengumpulkan kandidat MAC dari parameter flatten device record.
func collectMACs(params map[string]string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, 4)
	for path, raw := range params {
		lower := strings.ToLower(path)
		if !strings.Contains(lower, "mac") {
			continue
		}
		mac, ok := ont.NormalizeMAC(raw)
		if !ok || seen[mac] {
			continue
		}
		seen[mac] = true
		out = append(out, mac)
	}
	return out
}

// extractTelemetry memilih kandidat path pertama yang tersedia untuk tiap field.
func extractTelemetry(rec genieacs.DeviceRecord, catalog ont.TelemetryCatalog) port.ACSTelemetry {
	t := port.ACSTelemetry{FetchedAt: time.Now().UTC(), Raw: rec.Params}
	t.SerialNumber = firstParam(rec.Params, catalog.SerialNumber)
	t.ProductClass = firstParam(rec.Params, catalog.ProductClass)
	t.SoftwareVersion = firstParam(rec.Params, catalog.SoftwareVersion)
	t.HardwareVersion = firstParam(rec.Params, catalog.HardwareVersion)
	t.WANIP = firstParam(rec.Params, catalog.WANIP)
	t.WANUsername = firstParam(rec.Params, catalog.WANUsername)
	t.WANStatus = firstParam(rec.Params, catalog.WANStatus)
	t.SSID24 = firstParam(rec.Params, catalog.SSID24)
	t.SSID5 = firstParam(rec.Params, catalog.SSID5)
	if v := firstParam(rec.Params, catalog.WiFiClients); v != "" {
		t.WiFiClients, _ = strconv.Atoi(strings.TrimSpace(v))
	}
	if v := firstParam(rec.Params, catalog.RxPowerDBM); v != "" {
		if rx, ok := ont.ParseOpticalPower(v); ok {
			t.RxPowerDBM = rx
		}
	}
	if v := firstParam(rec.Params, catalog.TemperatureC); v != "" {
		if temp, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(v, " C")), 64); err == nil {
			t.TemperatureC = &temp
		}
	}
	if v := firstParam(rec.Params, catalog.Uptime); v != "" {
		t.UptimeSeconds, _ = strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	}
	return t
}

func firstParam(params map[string]string, paths []string) string {
	for _, path := range paths {
		if v, ok := params[path]; ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
