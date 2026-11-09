package ont

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/domain/routerjob"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// EnqueueReboot menjadwalkan reboot ONT via TR-069.
func (s *Service) EnqueueReboot(ctx context.Context, ontID string) (routerjob.Job, error) {
	return s.enqueue(ctx, ontID, routerjob.TypeACSReboot, map[string]any{"ont_id": ontID}, "acs-reboot")
}

// EnqueueSetWiFi menjadwalkan perubahan SSID/password.
func (s *Service) EnqueueSetWiFi(ctx context.Context, ontID, ssid, password, band string) (routerjob.Job, error) {
	return s.enqueue(ctx, ontID, routerjob.TypeACSSetWiFi, map[string]any{
		"ont_id": ontID, "ssid": ssid, "password": password, "band": band,
	}, "acs-wifi")
}

// EnqueuePushPPPoE menjadwalkan push kredensial WAN PPPoE.
func (s *Service) EnqueuePushPPPoE(ctx context.Context, ontID, username, password string) (routerjob.Job, error) {
	return s.enqueue(ctx, ontID, routerjob.TypeACSPushPPPoE, map[string]any{
		"ont_id": ontID, "username": username, "password": password,
	}, "acs-pppoe")
}

// EnqueueRefreshTelemetry menjadwalkan refresh telemetri satu ONT.
func (s *Service) EnqueueRefreshTelemetry(ctx context.Context, ontID string) (routerjob.Job, error) {
	return s.enqueue(ctx, ontID, routerjob.TypeACSRefreshTelemetry, map[string]any{"ont_id": ontID}, "acs-telemetry")
}

// EnqueueSyncDevices menjadwalkan sinkronisasi katalog device + matching.
// Saat oltDeviceID kosong, device_id job memakai server ACS (FK devices).
func (s *Service) EnqueueSyncDevices(ctx context.Context, oltDeviceID string) (routerjob.Job, error) {
	if oltDeviceID == "" {
		oltDeviceID = s.deps.Settings.GetValue(ctx, "acs.server_device_id", "")
	}
	if oltDeviceID == "" {
		return routerjob.Job{}, fmt.Errorf("%w: acs.server_device_id kosong", ont.ErrACSNotConfigured)
	}
	return s.enqueueWithDevice(ctx, oltDeviceID, routerjob.TypeACSSyncMatch, map[string]any{}, "acs-sync")
}

func (s *Service) enqueue(ctx context.Context, ontID string, typ routerjob.Type, payload map[string]any, prefix string) (routerjob.Job, error) {
	o, err := s.deps.ONTs.FindByID(ctx, ontID)
	if err != nil {
		return routerjob.Job{}, err
	}
	if o.ACSDeviceID == nil || *o.ACSDeviceID == "" {
		return routerjob.Job{}, fmt.Errorf("%w: ONT belum tertaut ke device ACS", ont.ErrNotLinked)
	}
	if o.OLTDeviceID == "" {
		return routerjob.Job{}, fmt.Errorf("%w: ONT tidak memiliki OLT", ont.ErrInvalidInput)
	}
	return s.enqueueWithDevice(ctx, o.OLTDeviceID, typ, payload, prefix+":"+ontID)
}

func (s *Service) enqueueWithDevice(ctx context.Context, deviceID string, typ routerjob.Type, payload map[string]any, keySuffix string) (routerjob.Job, error) {
	if s.deps.Jobs == nil {
		return routerjob.Job{}, fmt.Errorf("%w: job queue not configured", ont.ErrQueueNotConfigured)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return routerjob.Job{}, fmt.Errorf("%w: marshal payload: %w", ont.ErrInvalidInput, err)
	}
	now := time.Now().UTC()
	job := routerjob.Job{
		ID:             idgen.New("rj"),
		TenantID:       tenantDefault,
		DeviceID:       deviceID,
		JobType:        typ,
		Payload:        raw,
		IdempotencyKey: "ont:" + keySuffix + ":" + now.Format("200601021504"),
		Status:         routerjob.StatusQueued,
		MaxAttempts:    jobMaxAttempt,
		NextRunAt:      now,
	}
	return s.deps.Jobs.Enqueue(ctx, job)
}

// --- aksi langsung (dipanggil worker) ---

// RebootDevice mengirim task reboot TR-069.
func (s *Service) RebootDevice(ctx context.Context, ontID string) error {
	o, err := s.deps.ONTs.FindByID(ctx, ontID)
	if err != nil {
		return err
	}
	if o.ACSDeviceID == nil || *o.ACSDeviceID == "" {
		return fmt.Errorf("%w: ONT belum tertaut ke device ACS", ont.ErrNotLinked)
	}
	return s.deps.ACS.Reboot(ctx, *o.ACSDeviceID)
}

// RefreshTelemetry adalah wrapper worker untuk Telemetry (Fase D job).
func (s *Service) RefreshTelemetry(ctx context.Context, ontID string) error {
	_, err := s.Telemetry(ctx, ontID)
	return err
}

// SetWiFiPath adalah path object WLAN TR-098 per band.
func wifiPaths(band string, slots int) []string {
	paths := []string{}
	add := func(idx int) {
		paths = append(paths,
			fmt.Sprintf("InternetGatewayDevice.LANDevice.1.WLANConfiguration.%d.SSID", idx),
			fmt.Sprintf("InternetGatewayDevice.LANDevice.1.WLANConfiguration.%d.PreSharedKey.1.KeyPassphrase", idx),
		)
	}
	switch strings.ToLower(strings.TrimSpace(band)) {
	case "2.4", "24", "2g":
		add(1)
	case "5", "5g":
		add(2)
	default:
		add(1)
		if slots >= 2 {
			add(2)
		}
	}
	return paths
}

// SetWiFi mengubah SSID/password melalui setParameterValues.
func (s *Service) SetWiFi(ctx context.Context, ontID, ssid, password, band string) error {
	o, err := s.deps.ONTs.FindByID(ctx, ontID)
	if err != nil {
		return err
	}
	if o.ACSDeviceID == nil || *o.ACSDeviceID == "" {
		return fmt.Errorf("%w: ONT belum tertaut ke device ACS", ont.ErrNotLinked)
	}
	if strings.TrimSpace(ssid) == "" {
		return fmt.Errorf("%w: SSID wajib diisi", ont.ErrInvalidInput)
	}
	if len(password) > 0 && len(password) < 8 {
		return fmt.Errorf("%w: password WiFi minimal 8 karakter", ont.ErrInvalidInput)
	}
	values := map[string]string{}
	for _, path := range wifiPaths(band, o.WLANSlots) {
		if strings.HasSuffix(path, ".SSID") {
			values[path] = ssid
		} else {
			values[path] = password
		}
	}
	return s.deps.ACS.SetParameters(ctx, *o.ACSDeviceID, values)
}

// PushPPPoE mengubah kredensial WAN PPPoE ONT.
func (s *Service) PushPPPoE(ctx context.Context, ontID, username, password string) error {
	o, err := s.deps.ONTs.FindByID(ctx, ontID)
	if err != nil {
		return err
	}
	if o.ACSDeviceID == nil || *o.ACSDeviceID == "" {
		return fmt.Errorf("%w: ONT belum tertaut ke device ACS", ont.ErrNotLinked)
	}
	if username == "" {
		return fmt.Errorf("%w: username PPPoE wajib diisi", ont.ErrInvalidInput)
	}
	catalog := ont.DefaultTelemetryCatalog()
	base := ""
	for _, candidate := range catalog.WANUsername {
		if idx := strings.LastIndex(candidate, ".Username"); idx > 0 {
			base = candidate[:idx]
			break
		}
	}
	if base == "" {
		return fmt.Errorf("%w: path WAN PPPoE tidak dikenal", ont.ErrInvalidInput)
	}
	return s.deps.ACS.SetParameters(ctx, *o.ACSDeviceID, map[string]string{
		base + ".Username": username,
		base + ".Password": password,
	})
}
