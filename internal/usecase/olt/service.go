// Package olt mengorkestrasi inventori OLT EPON HiOSO: CRUD perangkat,
// sync ONU (read SNMP), dan aksi tulis via job queue. Lihat
// PLAN-ACS-OLT-HIOSO.md §7.
package olt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quixiq/polyglot/internal/domain/device"
	domainOLT "github.com/quixiq/polyglot/internal/domain/olt"
	"github.com/quixiq/polyglot/internal/domain/ont"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
)

const (
	tenantDefault   = domainTenant.DefaultID
	vendorHioso     = "hioso"
	driverTypeHioso = "hioso"
	tagOLT          = "olt"
	defaultTelnet   = 23
	defaultSNMP     = 161
	defaultTimeout  = 30 * time.Second
)

// DriverEvicter membersihkan cache koneksi driver setelah device diubah/dihapus.
type DriverEvicter interface {
	Evict(deviceID string) error
}

// AlertHook menerima transisi state/Rx ONT dari sync (diimplementasikan
// usecase/ont untuk notifikasi WA).
type AlertHook interface {
	ONTStateChanged(ctx context.Context, before, after ont.ONT)
	ONTRxChanged(ctx context.Context, before, after ont.ONT)
}

// Deps adalah dependensi Service.
type Deps struct {
	Devices  port.DeviceRepository
	Vault    port.CredentialVault
	Managers port.OLTManagerProvider
	ONTs     port.ONTRepository
	Settings port.SettingReader
	Jobs     port.RouterJobRepository
	Evicter  DriverEvicter
	Alerts   AlertHook
}

// Service adalah usecase OLT/ONT inventori.
type Service struct {
	deps Deps
}

// NewService membangun usecase OLT.
func NewService(deps Deps) *Service {
	return &Service{deps: deps}
}

// Input adalah input create/update OLT.
type Input struct {
	Name        string
	Host        string
	TelnetPort  int
	SSHPort     int
	SNMPPort    int
	TimeoutMS   int
	Transport   string
	Variant     string
	PONCount    int
	Username    string
	Password    string
	CommunityRO string
	CommunityRW string
	Enabled     bool
}

// OLT adalah ringkasan OLT untuk API/UI.
type OLT struct {
	DeviceID   string
	Name       string
	Host       string
	TelnetPort int
	SSHPort    int
	SNMPPort   int
	Transport  string
	Variant    string
	PONCount   int
	Enabled    bool
	Counts     port.ONTStateCounts
}

// Health adalah status koneksi OLT.
type Health struct {
	Reachable bool
	Variant   string
	Firmware  string
	SysDescr  string
	Uptime    string
	LatencyMS int64
}

// Create menyimpan OLT + kredensial lalu memverifikasi konektivitas.
func (s *Service) Create(ctx context.Context, in Input) (OLT, error) {
	dev, err := s.buildDevice(uuid.NewString(), in)
	if err != nil {
		return OLT{}, err
	}
	if err := s.deps.Devices.Save(ctx, dev); err != nil {
		return OLT{}, fmt.Errorf("olt: simpan device: %w", err)
	}
	if err := s.saveCredentials(ctx, dev.ID, in, nil); err != nil {
		return OLT{}, err
	}
	// Verifikasi koneksi setelah tersimpan (registry butuh baris DB).
	_, caps, err := s.probe(ctx, dev.ID)
	if err != nil {
		return OLT{}, fmt.Errorf("olt tersimpan, tetapi verifikasi koneksi gagal: %w", err)
	}
	dev.Extra["variant"] = string(caps.Variant)
	dev.Extra["pon_count"] = strconv.Itoa(caps.PONCount)
	if err := s.deps.Devices.Update(ctx, dev); err != nil {
		return OLT{}, fmt.Errorf("olt: simpan hasil deteksi: %w", err)
	}
	return s.toOLT(dev), nil
}

// Update mengubah konfigurasi OLT; kredensial kosong dipertahankan.
func (s *Service) Update(ctx context.Context, deviceID string, in Input) (OLT, error) {
	dev, err := s.deps.Devices.FindByID(ctx, deviceID)
	if err != nil {
		return OLT{}, err
	}
	if dev.DriverType != driverTypeHioso {
		return OLT{}, fmt.Errorf("%w: device bukan OLT hioso", domainOLT.ErrUnsupportedVariant)
	}
	updated, err := s.buildDevice(deviceID, in)
	if err != nil {
		return OLT{}, err
	}
	updated.TenantID = dev.TenantID
	updated.Tags = mergeTag(dev.Tags, tagOLT)
	if err := s.deps.Devices.Update(ctx, updated); err != nil {
		return OLT{}, fmt.Errorf("olt: update device: %w", err)
	}
	existing, _ := s.deps.Vault.Get(ctx, deviceID)
	if err := s.saveCredentials(ctx, deviceID, in, &existing); err != nil {
		return OLT{}, err
	}
	if s.deps.Evicter != nil {
		_ = s.deps.Evicter.Evict(deviceID)
	}
	if _, _, errProbe := s.probe(ctx, deviceID); errProbe != nil {
		return OLT{}, fmt.Errorf("olt terupdate, tetapi verifikasi koneksi gagal: %w", errProbe)
	}
	return s.toOLT(updated), nil
}

// Delete menghapus OLT (ONT terkait ikut terhapus via FK cascade).
func (s *Service) Delete(ctx context.Context, deviceID string) error {
	if _, err := s.deps.Devices.FindByID(ctx, deviceID); err != nil {
		return err
	}
	if err := s.deps.Devices.Delete(ctx, deviceID); err != nil {
		return fmt.Errorf("olt: hapus device: %w", err)
	}
	if s.deps.Evicter != nil {
		_ = s.deps.Evicter.Evict(deviceID)
	}
	return nil
}

// List mengembalikan semua OLT beserta ringkasan ONT.
func (s *Service) List(ctx context.Context) ([]OLT, error) {
	devices, err := s.deps.Devices.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]OLT, 0)
	for _, dev := range devices {
		if dev.DriverType != driverTypeHioso {
			continue
		}
		out = append(out, s.toOLT(dev))
	}
	return out, nil
}

// Get mengembalikan satu OLT + ringkasan ONT.
func (s *Service) Get(ctx context.Context, deviceID string) (OLT, error) {
	dev, err := s.deps.Devices.FindByID(ctx, deviceID)
	if err != nil {
		return OLT{}, err
	}
	if dev.DriverType != driverTypeHioso {
		return OLT{}, fmt.Errorf("%w: device bukan OLT hioso", domainOLT.ErrUnsupportedVariant)
	}
	return s.toOLT(dev), nil
}

// TestConnection melakukan probe SNMP ke OLT.
func (s *Service) TestConnection(ctx context.Context, deviceID string) (Health, error) {
	health, caps, err := s.probe(ctx, deviceID)
	if err != nil {
		return health, err
	}
	if caps.Variant == domainOLT.VariantUnknown {
		return health, domainOLT.ErrUnsupportedVariant
	}
	return health, nil
}

// ListONTs mengembalikan inventori ONT dari DB (bukan live).
func (s *Service) ListONTs(ctx context.Context, filter port.ONTFilter) ([]ont.ONT, error) {
	return s.deps.ONTs.List(ctx, filter)
}

// GetONT mengembalikan satu ONT.
func (s *Service) GetONT(ctx context.Context, ontID string) (ont.ONT, error) {
	return s.deps.ONTs.FindByID(ctx, ontID)
}

// ListONTEvents mengembalikan timeline satu ONT.
func (s *Service) ListONTEvents(ctx context.Context, ontID string, limit int) ([]ont.Event, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.deps.ONTs.ListEvents(ctx, ontID, limit)
}

// ListONTSamples mengembalikan riwayat metrik optik satu ONT.
func (s *Service) ListONTSamples(ctx context.Context, ontID string, from time.Time, limit int) ([]ont.MetricSample, error) {
	if limit <= 0 {
		limit = 500
	}
	return s.deps.ONTs.ListSamples(ctx, ontID, from, limit)
}

// probe me-resolve manager dan membaca identitas OLT.
func (s *Service) probe(ctx context.Context, deviceID string) (Health, domainOLT.Capabilities, error) {
	mgr, err := s.deps.Managers.ForDevice(ctx, deviceID)
	if err != nil {
		return Health{}, domainOLT.Capabilities{}, err
	}
	caps, h, err := mgr.Detect(ctx)
	if err != nil {
		return Health{}, domainOLT.Capabilities{}, err
	}
	return Health{
		Reachable: h.Reachable,
		Variant:   string(h.Variant),
		Firmware:  h.Firmware,
		SysDescr:  h.SysDescr,
		Uptime:    h.Uptime,
		LatencyMS: h.LatencyMS,
	}, caps, nil
}

func (s *Service) buildDevice(id string, in Input) (device.Device, error) {
	name := strings.TrimSpace(in.Name)
	host := strings.TrimSpace(in.Host)
	if host == "" {
		return device.Device{}, fmt.Errorf("%w: host wajib diisi", domainOLT.ErrInvalidInput)
	}
	if name == "" {
		name = host
	}
	if in.CommunityRO == "" {
		return device.Device{}, fmt.Errorf("%w: SNMP community read-only wajib diisi", domainOLT.ErrInvalidInput)
	}
	variant := strings.ToUpper(strings.TrimSpace(in.Variant))
	switch variant {
	case "", "AUTO":
		variant = ""
	case string(domainOLT.VariantHA7302), string(domainOLT.VariantHA7304):
	default:
		return device.Device{}, fmt.Errorf("%w: varian %q tidak dikenal", domainOLT.ErrInvalidInput, in.Variant)
	}
	telnetPort := in.TelnetPort
	if telnetPort <= 0 {
		telnetPort = defaultTelnet
	}
	snmpPort := in.SNMPPort
	if snmpPort <= 0 {
		snmpPort = defaultSNMP
	}
	transport := strings.ToLower(strings.TrimSpace(in.Transport))
	if transport == "" {
		transport = "telnet"
	}
	if transport != "telnet" && transport != "ssh" {
		return device.Device{}, fmt.Errorf("%w: transport %q tidak dikenal", domainOLT.ErrInvalidInput, in.Transport)
	}
	timeoutMS := in.TimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = int(defaultTimeout / time.Millisecond)
	}
	extra := map[string]string{
		"snmp_port":     strconv.Itoa(snmpPort),
		"cli_transport": transport,
	}
	if variant != "" {
		extra["variant"] = variant
	}
	if in.PONCount > 0 {
		extra["pon_count"] = strconv.Itoa(in.PONCount)
	}
	sshPort := in.SSHPort
	if sshPort <= 0 {
		sshPort = 22
	}
	return device.Device{
		ID:         id,
		TenantID:   tenantDefault,
		Name:       name,
		Vendor:     vendorHioso,
		DriverType: driverTypeHioso,
		Host:       host,
		Port:       telnetPort,
		SSHPort:    sshPort,
		TimeoutMS:  timeoutMS,
		Extra:      extra,
		Tags:       []string{tagOLT},
		Enabled:    in.Enabled,
	}, nil
}

func (s *Service) saveCredentials(ctx context.Context, deviceID string, in Input, existing *device.Credentials) error {
	creds := device.Credentials{Username: in.Username, Password: in.Password}
	if existing != nil {
		if creds.Username == "" {
			creds.Username = existing.Username
		}
		if creds.Password == "" {
			creds.Password = existing.Password
		}
		creds.Extra = existing.Extra
	}
	if creds.Extra == nil {
		creds.Extra = map[string]string{}
	}
	if in.CommunityRO != "" {
		creds.Extra["community_ro"] = in.CommunityRO
	}
	if in.CommunityRW != "" {
		creds.Extra["community_rw"] = in.CommunityRW
	}
	if err := s.deps.Vault.Save(ctx, deviceID, creds); err != nil {
		return fmt.Errorf("olt: simpan kredensial: %w", err)
	}
	return nil
}

func (s *Service) toOLT(dev device.Device) OLT {
	counts, _ := s.deps.ONTs.CountByState(context.Background(), dev.ID)
	telnetPort := dev.Port
	if v, err := strconv.Atoi(dev.Extra["telnet_port"]); err == nil && v > 0 {
		telnetPort = v
	}
	snmpPort, _ := strconv.Atoi(dev.Extra["snmp_port"])
	ponCount, _ := strconv.Atoi(dev.Extra["pon_count"])
	return OLT{
		DeviceID:   dev.ID,
		Name:       dev.Name,
		Host:       dev.Host,
		TelnetPort: telnetPort,
		SSHPort:    dev.SSHPort,
		SNMPPort:   snmpPort,
		Transport:  dev.Extra["cli_transport"],
		Variant:    dev.Extra["variant"],
		PONCount:   ponCount,
		Enabled:    dev.Enabled,
		Counts:     counts,
	}
}

func mergeTag(tags []string, tag string) []string {
	for _, t := range tags {
		if t == tag {
			return tags
		}
	}
	return append(tags, tag)
}
