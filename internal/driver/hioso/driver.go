// Package hioso implements port.DeviceDriver untuk OLT EPON HiOSO
// (HA7302/HA7304): baca inventori ONU via SNMP v2c + tulis via Telnet CLI.
// Lihat PLAN-ACS-OLT-HIOSO.md §5.
package hioso

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/domain/olt"
	"github.com/quixiq/polyglot/internal/port"
)

const (
	defaultTelnetPort = 23
	defaultSNMPPort   = 161
)

// Config adalah parameter koneksi OLT hasil ekstraksi device.Target.
type Config struct {
	Host        string
	TelnetPort  int
	SSHPort     int
	SNMPPort    int
	CommunityRO string
	CommunityRW string
	Username    string
	Password    string
	Timeout     time.Duration
	Transport   string // "telnet" (default) | "ssh"
	Variant     olt.Variant
	PONCount    int
}

// configFromTarget memetakan Target (device.Extra + credentials.Extra) ke Config.
func configFromTarget(t device.Target) Config {
	cfg := Config{
		Host:        t.Host,
		TelnetPort:  t.Port,
		SNMPPort:    defaultSNMPPort,
		CommunityRO: t.Extra["community_ro"],
		CommunityRW: t.Extra["community_rw"],
		Username:    t.Username,
		Password:    t.Password,
		Timeout:     t.Timeout,
		Transport:   t.Extra["cli_transport"],
		Variant:     olt.Variant(t.Extra["variant"]),
	}
	if p, err := strconv.Atoi(t.Extra["ssh_port"]); err == nil && p > 0 {
		cfg.SSHPort = p
	}
	if cfg.SNMPPort <= 0 {
		if p, err := strconv.Atoi(t.Extra["snmp_port"]); err == nil && p > 0 {
			cfg.SNMPPort = p
		}
	}
	if cfg.TelnetPort <= 0 {
		cfg.TelnetPort = defaultTelnetPort
	}
	if cfg.Transport == "" {
		cfg.Transport = "telnet"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Variant != olt.VariantHA7302 && cfg.Variant != olt.VariantHA7304 {
		cfg.Variant = olt.VariantUnknown
	}
	if p, err := strconv.Atoi(t.Extra["pon_count"]); err == nil && p > 0 {
		cfg.PONCount = p
	}
	return cfg
}

// Driver mengimplementasikan port.DeviceDriver untuk OLT HiOSO.
type Driver struct {
	cfg  Config
	snmp *SNMPClient

	mu        sync.Mutex // serialisasi akses CLI (OLT membatasi sesi konkuren)
	cli       *CLIClient
	caps      olt.Capabilities
	health    port.OLTHealth
	writeMu   sync.Mutex
	writeSNMP *SNMPClient
}

var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver membuat driver yang sudah terhubung ke OLT (SNMP) dan mendeteksi
// varian. Sesi CLI dibuka lazy saat aksi tulis/fallback read dibutuhkan.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	cfg := configFromTarget(target)
	if cfg.Host == "" {
		return nil, fmt.Errorf("%w: host kosong", olt.ErrInvalidInput)
	}
	if cfg.CommunityRO == "" {
		return nil, fmt.Errorf("%w: community_ro wajib diisi", olt.ErrInvalidInput)
	}
	tr, err := newSNMPTransport(cfg.Host, cfg.SNMPPort, cfg.CommunityRO, cfg.Timeout, 2)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", olt.ErrUnreachable, err)
	}
	d := &Driver{cfg: cfg, snmp: newSNMPClient(tr)}
	caps, health, err := d.probe(ctx)
	if err != nil {
		_ = d.snmp.Close()
		return nil, err
	}
	d.caps = caps
	d.health = health
	return d, nil
}

// probe membaca identitas OLT dan mendeteksi varian (sysObjectID/firmware).
func (d *Driver) probe(ctx context.Context) (olt.Capabilities, port.OLTHealth, error) {
	start := time.Now()
	sysDescr, sysObj, uptime, err := d.snmp.SysInfo(ctx)
	if err != nil {
		return olt.Capabilities{}, port.OLTHealth{}, fmt.Errorf("%w: %v", olt.ErrUnreachable, err)
	}
	fw, err := d.snmp.FirmwareSignature(ctx)
	if err != nil {
		fw = ""
	}
	variant := d.cfg.Variant
	if variant == olt.VariantUnknown {
		variant = olt.DetectVariant(sysObj, sysDescr, fw)
	}
	caps := olt.CapabilitiesFor(variant)
	if d.cfg.PONCount > 0 && !caps.FlatLLID {
		caps.PONCount = d.cfg.PONCount
	}
	health := port.OLTHealth{
		Reachable:   true,
		Variant:     variant,
		SysDescr:    sysDescr,
		SysObjectID: sysObj,
		Firmware:    fw,
		Uptime:      uptime,
		LatencyMS:   time.Since(start).Milliseconds(),
	}
	return caps, health, nil
}

// Detect mengembalikan capability varian + kesehatan OLT (refresh).
func (d *Driver) Detect(ctx context.Context) (olt.Capabilities, port.OLTHealth, error) {
	caps, health, err := d.probe(ctx)
	if err != nil {
		return caps, health, err
	}
	d.mu.Lock()
	d.caps = caps
	d.health = health
	d.mu.Unlock()
	return caps, health, nil
}

// Capabilities mengembalikan capability hasil deteksi terakhir.
func (d *Driver) Capabilities() olt.Capabilities {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.caps
}

// ListONUs membaca inventori ONU: SNMP (primary), fallback CLI bila SNMP gagal.
func (d *Driver) ListONUs(ctx context.Context, ponIndex int) ([]port.OLTONU, error) {
	caps := d.Capabilities()
	rows, err := d.snmp.ListONUs(ctx, caps.Variant, ponIndex)
	if err == nil {
		return rows, nil
	}
	var cliRows []port.OLTONU
	cliErr := d.withCLI(ctx, func(cli *CLIClient) error {
		parsed, parseErr := cli.ListONUs(ctx, caps, caps.PONCount)
		cliRows = parsed
		return parseErr
	})
	if cliErr != nil {
		return nil, fmt.Errorf("%w: snmp: %v; cli: %v", olt.ErrUnreachable, err, cliErr)
	}
	return cliRows, nil
}

// RenameONU mengganti nama ONU (HA7304 via CLI, HA7302 via SNMP SET).
func (d *Driver) RenameONU(ctx context.Context, ponIndex, onuIndex int, name string) error {
	if err := olt.ValidateONUName(name); err != nil {
		return err
	}
	if err := olt.ValidatePosition(d.Capabilities(), ponIndex, onuIndex); err != nil {
		return err
	}
	caps := d.Capabilities()
	if caps.RenameViaSNMP {
		writer, err := d.ensureWriteSNMP()
		if err != nil {
			return err
		}
		return writer.SetONUName(ctx, caps.Variant, ponIndex, onuIndex, name)
	}
	return d.withCLI(ctx, func(cli *CLIClient) error {
		return cli.RenameONU(ctx, caps, ponIndex, onuIndex, name)
	})
}

// RebootONU me-reboot ONU via CLI.
func (d *Driver) RebootONU(ctx context.Context, ponIndex, onuIndex int) error {
	caps := d.Capabilities()
	return d.withCLI(ctx, func(cli *CLIClient) error {
		return cli.RebootONU(ctx, caps, ponIndex, onuIndex)
	})
}

// SetONUAdmin mengaktifkan/menonaktifkan ONU via CLI.
func (d *Driver) SetONUAdmin(ctx context.Context, ponIndex, onuIndex int, enabled bool) error {
	caps := d.Capabilities()
	return d.withCLI(ctx, func(cli *CLIClient) error {
		return cli.SetONUAdmin(ctx, caps, ponIndex, onuIndex, enabled)
	})
}

// DeleteONU menghapus ONU via CLI.
func (d *Driver) DeleteONU(ctx context.Context, ponIndex, onuIndex int) error {
	caps := d.Capabilities()
	return d.withCLI(ctx, func(cli *CLIClient) error {
		return cli.DeleteONU(ctx, caps, ponIndex, onuIndex)
	})
}

// SaveConfig menyimpan konfigurasi OLT.
func (d *Driver) SaveConfig(ctx context.Context) error {
	return d.withCLI(ctx, func(cli *CLIClient) error { return cli.SaveConfig(ctx) })
}

// RunningConfig membaca konfigurasi berjalan.
func (d *Driver) RunningConfig(ctx context.Context) (string, error) {
	var out string
	err := d.withCLI(ctx, func(cli *CLIClient) error {
		raw, err := cli.ShowRunningConfig(ctx)
		out = raw
		return err
	})
	return out, err
}

// SystemInfo membaca ringkasan sistem (show version/show system).
func (d *Driver) SystemInfo(ctx context.Context) (map[string]string, error) {
	var info map[string]string
	err := d.withCLI(ctx, func(cli *CLIClient) error {
		raw, err := cli.SystemInfo(ctx)
		info = raw
		return err
	})
	return info, err
}

// Execute menjalankan perintah CLI mentah (generic DeviceDriver path).
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	var out string
	err := d.withCLI(ctx, func(cli *CLIClient) error {
		raw, err := cli.run(ctx, cmd.Raw)
		out = raw
		return err
	})
	if err != nil {
		return command.Result{}, err
	}
	return command.Result{Output: out}, nil
}

// Classify menandai klasifikasi AI Router Ops untuk perintah CLI HiOSO.
func (d *Driver) Classify(cmd command.Command) command.Classification {
	return Classify(cmd)
}

// Translate memetakan operation abstrak ke perintah CLI HiOSO.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close menutup sesi SNMP + CLI.
func (d *Driver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cli != nil {
		_ = d.cli.t.Close()
		d.cli = nil
	}
	if d.writeSNMP != nil {
		_ = d.writeSNMP.Close()
		d.writeSNMP = nil
	}
	return d.snmp.Close()
}

// withCLI mengeksekusi fn sambil memegang lock sesi CLI (serialisasi).
func (d *Driver) withCLI(ctx context.Context, fn func(*CLIClient) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := d.connectCLILocked(ctx); err != nil {
		return err
	}
	err := fn(d.cli)
	if err != nil {
		// Sesi bisa dalam keadaan kotor (timeout/putus) — tutup agar
		// percobaan berikutnya dial ulang.
		_ = d.cli.t.Close()
		d.cli = nil
	}
	return err
}

func (d *Driver) connectCLILocked(ctx context.Context) error {
	if d.cli != nil {
		return nil
	}
	if d.cfg.Username == "" {
		return fmt.Errorf("%w: kredensial CLI kosong", olt.ErrInvalidInput)
	}
	if d.cfg.Transport == "ssh" {
		return fmt.Errorf("%w: transport SSH belum diimplementasikan (gunakan telnet)", olt.ErrUnsupportedVariant)
	}
	t, err := newTelnetTransport(ctx, d.cfg.Host, d.cfg.TelnetPort, d.cfg.Username, d.cfg.Password, d.cfg.Timeout)
	if err != nil {
		return fmt.Errorf("%w: %v", olt.ErrUnreachable, err)
	}
	d.cli = newCLIClient(t)
	return nil
}

func (d *Driver) ensureWriteSNMP() (*SNMPClient, error) {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	if d.writeSNMP != nil {
		return d.writeSNMP, nil
	}
	if d.cfg.CommunityRW == "" {
		return nil, fmt.Errorf("%w: community_rw wajib untuk rename HA7302", olt.ErrInvalidInput)
	}
	tr, err := newSNMPTransport(d.cfg.Host, d.cfg.SNMPPort, d.cfg.CommunityRW, d.cfg.Timeout, 2)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", olt.ErrUnreachable, err)
	}
	d.writeSNMP = newSNMPClient(tr)
	return d.writeSNMP, nil
}
