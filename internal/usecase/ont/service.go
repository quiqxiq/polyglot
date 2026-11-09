// Package ont mengorkestrasi ONT/ONU pelanggan: pencocokan device GenieACS
// (TR-069), telemetri, aksi remote, dan penautan ke langganan/pelanggan.
// Lihat PLAN-ACS-OLT-HIOSO.md Fase D/E.
package ont

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/ont"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

const (
	tenantDefault = domainTenant.DefaultID
	jobMaxAttempt = 5
)

// Deps adalah dependensi usecase ONT/ACS.
type Deps struct {
	ONTs      port.ONTRepository
	ACS       port.ACSManager
	ACSRepo   port.ACSRepository
	Settings  port.SettingReader
	Jobs      port.RouterJobRepository
	Customers port.CustomerRepository
	Subs      port.SubscriptionRepository
}

// Service adalah usecase ONT/ACS.
type Service struct {
	deps Deps
}

// NewService membangun usecase ONT.
func NewService(deps Deps) *Service {
	return &Service{deps: deps}
}

// SyncStats meringkas satu siklus sinkronisasi ACS.
type SyncStats struct {
	Devices   int `json:"devices"`
	Matched   int `json:"matched"`
	Unmatched int `json:"unmatched"`
	Conflicts int `json:"conflicts"`
}

// SyncDevices menarik katalog device dari GenieACS, menyimpannya, lalu
// menjalankan pencocokan ACS↔ONT.
func (s *Service) SyncDevices(ctx context.Context) (SyncStats, error) {
	stats := SyncStats{}
	devices, err := s.deps.ACS.ListDevices(ctx, "")
	if err != nil {
		return stats, err
	}
	now := time.Now().UTC()
	for _, d := range devices {
		raw := d.Raw
		if len(raw) == 0 {
			raw = json.RawMessage("{}")
		}
		item := ont.ACSDevice{
			ID:              idgen.New("acsdev"),
			TenantID:        tenantDefault,
			ACSDeviceID:     d.DeviceID,
			OUI:             d.OUI,
			ProductClass:    d.ProductClass,
			SerialNumber:    ont.NormalizeSerial(d.SerialNumber),
			SoftwareVersion: d.SoftwareVersion,
			Tags:            d.Tags,
			MACCandidates:   d.MACCandidates,
			Raw:             raw,
			LastSeenAt:      now,
		}
		if !d.LastInform.IsZero() {
			li := d.LastInform
			item.LastInformAt = &li
		}
		if err := s.deps.ACSRepo.UpsertDevice(ctx, item); err != nil {
			return stats, fmt.Errorf("ont: simpan device ACS %s: %w", d.DeviceID, err)
		}
	}
	stats.Devices = len(devices)
	matched, conflicts, err := s.matchDevices(ctx)
	if err != nil {
		return stats, err
	}
	stats.Matched = matched
	stats.Conflicts = conflicts
	stats.Unmatched = len(devices) - matched
	return stats, nil
}

// matchDevices mencocokkan device ACS ke ONT: pin manual → serial → MAC →
// MAC±1; konflik (dua device memperebutkan ONT yang sama) dilepas.
func (s *Service) matchDevices(ctx context.Context) (int, int, error) {
	devices, err := s.deps.ACSRepo.ListDevices(ctx, port.ACSDeviceFilter{})
	if err != nil {
		return 0, 0, err
	}
	onts, err := s.deps.ONTs.List(ctx, port.ONTFilter{Limit: 10000})
	if err != nil {
		return 0, 0, err
	}
	bySerial := make(map[string]ont.ONT, len(onts))
	byMAC := make(map[string]ont.ONT, len(onts))
	byID := make(map[string]ont.ONT, len(onts))
	for _, o := range onts {
		byID[o.ID] = o
		if o.SerialNumber != "" {
			bySerial[ont.NormalizeSerial(o.SerialNumber)] = o
		}
		byMAC[o.MACAddress] = o
	}

	type claim struct {
		device ont.ACSDevice
		ontID  string
		source ont.MatchSource
	}
	claims := make(map[string][]claim)
	matched := 0
	conflicts := 0

	for _, d := range devices {
		if d.MatchSource == string(ont.MatchManual) && d.MatchedONTID != nil {
			matched++
			continue
		}
		target, source, ok := s.pickONT(d, bySerial, byMAC)
		if !ok {
			continue
		}
		claims[target.ID] = append(claims[target.ID], claim{device: d, ontID: target.ID, source: source})
	}

	released := map[string]bool{}
	for ontID, list := range claims {
		if _, exists := byID[ontID]; !exists {
			continue
		}
		if len(list) > 1 {
			// Konflik: sengaja dilepas semua, bukan dipilih salah satu.
			for _, c := range list {
				if err := s.deps.ACSRepo.UpdateDeviceMatch(ctx, c.device.ACSDeviceID, nil, ""); err != nil {
					return matched, conflicts, err
				}
				released[c.device.ACSDeviceID] = true
			}
			if err := s.deps.ONTs.UpdateACS(ctx, ontID, nil, "", nil, nil, 0); err != nil {
				return matched, conflicts, err
			}
			conflicts++
			continue
		}
		c := list[0]
		// Lepas device lama yang masih memegang ONT ini.
		if cur, okCur := byID[ontID]; okCur && cur.ACSDeviceID != nil && *cur.ACSDeviceID != c.device.ACSDeviceID {
			_ = s.deps.ACSRepo.UpdateDeviceMatch(ctx, *cur.ACSDeviceID, nil, "")
		}
		var lastInform *time.Time
		if c.device.LastInformAt != nil {
			li := *c.device.LastInformAt
			lastInform = &li
		}
		if err := s.deps.ONTs.UpdateACS(ctx, ontID, &c.device.ACSDeviceID, string(c.source), lastInform, nil, 0); err != nil {
			return matched, conflicts, err
		}
		if err := s.deps.ACSRepo.UpdateDeviceMatch(ctx, c.device.ACSDeviceID, &c.ontID, string(c.source)); err != nil {
			return matched, conflicts, err
		}
		s.recordEvent(ctx, ontID, "ACS_MATCHED", "ACS "+c.device.ACSDeviceID+" tertaut via "+string(c.source))
		matched++
	}

	// Device lama yang tidak lagi cocok dilepas (kecuali pin manual).
	for _, d := range devices {
		if d.MatchSource == string(ont.MatchManual) || d.MatchedONTID == nil || released[d.ACSDeviceID] {
			continue
		}
		if _, still := claims[*d.MatchedONTID]; !still {
			if err := s.deps.ACSRepo.UpdateDeviceMatch(ctx, d.ACSDeviceID, nil, ""); err != nil {
				return matched, conflicts, err
			}
			_ = s.deps.ONTs.UpdateACS(ctx, *d.MatchedONTID, nil, "", nil, nil, 0)
		}
	}
	return matched, conflicts, nil
}

func (s *Service) pickONT(d ont.ACSDevice, bySerial, byMAC map[string]ont.ONT) (ont.ONT, ont.MatchSource, bool) {
	if d.SerialNumber != "" {
		if o, ok := bySerial[ont.NormalizeSerial(d.SerialNumber)]; ok {
			return o, ont.MatchSerial, true
		}
	}
	for _, mac := range d.MACCandidates {
		if o, ok := byMAC[mac]; ok {
			return o, ont.MatchMAC, true
		}
	}
	for _, mac := range d.MACCandidates {
		for ontMAC, o := range byMAC {
			if ont.MACNeighbor(ontMAC, mac) {
				return o, ont.MatchMACNeighbor, true
			}
		}
	}
	return ont.ONT{}, "", false
}

// Telemetry menarik telemetri ONT dari ACS dan menyimpannya.
func (s *Service) Telemetry(ctx context.Context, ontID string) (port.ACSTelemetry, error) {
	o, err := s.deps.ONTs.FindByID(ctx, ontID)
	if err != nil {
		return port.ACSTelemetry{}, err
	}
	if o.ACSDeviceID == nil || *o.ACSDeviceID == "" {
		return port.ACSTelemetry{}, fmt.Errorf("%w: ONT belum tertaut ke device ACS", ont.ErrNotLinked)
	}
	catalog, err := ont.CatalogFromJSON(s.deps.Settings.GetValue(ctx, "acs.param_catalog", "{}"))
	if err != nil {
		return port.ACSTelemetry{}, fmt.Errorf("%w: katalog parameter: %v", ont.ErrInvalidInput, err)
	}
	telemetry, err := s.deps.ACS.FetchTelemetry(ctx, *o.ACSDeviceID, catalog)
	if err != nil {
		return port.ACSTelemetry{}, err
	}
	// last inform diambil dari katalog device (bukan response task).
	lastInform := telemetry.FetchedAt
	if dev, errDev := s.deps.ACS.GetDevice(ctx, *o.ACSDeviceID); errDev == nil && !dev.LastInform.IsZero() {
		lastInform = dev.LastInform
	}
	raw, _ := json.Marshal(telemetry)
	slots := 0
	if telemetry.SSID24 != "" {
		slots = 1
	}
	if telemetry.SSID5 != "" {
		slots = 2
	}
	source := o.ACSMatchSource
	if err := s.deps.ONTs.UpdateACS(ctx, o.ID, o.ACSDeviceID, source, &lastInform, raw, slots); err != nil {
		return telemetry, err
	}
	s.recordEvent(ctx, o.ID, "ACS_TELEMETRY", "telemetri ACS diperbarui")
	return telemetry, nil
}

// RefreshTelemetryAll memperbarui telemetri semua ONT yang sudah tertaut.
func (s *Service) RefreshTelemetryAll(ctx context.Context) (int, error) {
	list, err := s.deps.ONTs.List(ctx, port.ONTFilter{ACSMatched: boolPtr(true), Limit: 5000})
	if err != nil {
		return 0, err
	}
	ok := 0
	var firstErr error
	for _, o := range list {
		if _, errTel := s.Telemetry(ctx, o.ID); errTel != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("ont %s: %w", o.MACAddress, errTel)
			}
			continue
		}
		ok++
	}
	return ok, firstErr
}

// BootstrapPreset membuat/memperbarui preset ACS URL + periodic inform.
func (s *Service) BootstrapPreset(ctx context.Context) error {
	acsURL := strings.TrimSpace(s.deps.Settings.GetValue(ctx, "acs.acs_url", ""))
	if acsURL == "" {
		return fmt.Errorf("%w: acs.acs_url belum diisi", ont.ErrACSNotConfigured)
	}
	interval := s.deps.Settings.GetValue(ctx, "acs.periodic_inform_interval", "300")
	if _, err := strconv.Atoi(interval); err != nil {
		interval = "300"
	}
	name := s.deps.Settings.GetValue(ctx, "acs.preset_name", "polyglot-bootstrap")
	configurations := json.RawMessage(fmt.Sprintf(`[
		{"type":"setParameterValues","parameterValues":[
			["InternetGatewayDevice.ManagementServer.URL",%q,"xsd:string"],
			["InternetGatewayDevice.ManagementServer.PeriodicInformInterval",%s,"xsd:unsignedInt"]
		]}
	]`, acsURL, interval))
	preset := ont.ACSPreset{
		ID:             idgen.New("acspreset"),
		TenantID:       tenantDefault,
		Name:           name,
		Configurations: configurations,
		Enabled:        true,
	}
	if err := s.deps.ACS.EnsurePreset(ctx, preset); err != nil {
		return err
	}
	now := time.Now().UTC()
	preset.LastSyncedAt = &now
	if err := s.deps.ACSRepo.SavePreset(ctx, preset); err != nil {
		return fmt.Errorf("ont: simpan preset ACS: %w", err)
	}
	return nil
}

// AcsConfig adalah konfigurasi ACS yang terbaca (read-only; edit lewat Settings).
type AcsConfig struct {
	Enabled           bool
	ServerDeviceID    string
	ACSURL            string
	PeriodicInform    int
	PresetName        string
	TelemetryInterval int
}

// Config membaca konfigurasi ACS dari settings.
func (s *Service) Config(ctx context.Context) AcsConfig {
	interval, _ := strconv.Atoi(s.deps.Settings.GetValue(ctx, "acs.periodic_inform_interval", "300"))
	return AcsConfig{
		Enabled:        s.deps.Settings.GetValue(ctx, "acs.enabled", "false") == "true",
		ServerDeviceID: s.deps.Settings.GetValue(ctx, "acs.server_device_id", ""),
		ACSURL:         s.deps.Settings.GetValue(ctx, "acs.acs_url", ""),
		PeriodicInform: interval,
		PresetName:     s.deps.Settings.GetValue(ctx, "acs.preset_name", "polyglot-bootstrap"),
	}
}

func (s *Service) recordEvent(ctx context.Context, ontID, eventType, message string) {
	_ = s.deps.ONTs.RecordEvent(ctx, ont.Event{
		ID:        idgen.New("ontev"),
		TenantID:  tenantDefault,
		ONTID:     ontID,
		EventType: eventType,
		Message:   message,
	})
}

func boolPtr(v bool) *bool { return &v }
