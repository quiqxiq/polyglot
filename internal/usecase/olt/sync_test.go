package olt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/device"
	domainOLT "github.com/quixiq/polyglot/internal/domain/olt"
	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
)

// --- fakes ---

type fakeDeviceRepo struct {
	byID map[string]device.Device
}

func (f *fakeDeviceRepo) Save(_ context.Context, d device.Device) error {
	f.byID[d.ID] = d
	return nil
}
func (f *fakeDeviceRepo) FindByID(_ context.Context, id string) (device.Device, error) {
	d, ok := f.byID[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}
func (f *fakeDeviceRepo) FindAll(_ context.Context) ([]device.Device, error) {
	out := make([]device.Device, 0, len(f.byID))
	for _, d := range f.byID {
		out = append(out, d)
	}
	return out, nil
}
func (f *fakeDeviceRepo) Update(_ context.Context, d device.Device) error {
	f.byID[d.ID] = d
	return nil
}
func (f *fakeDeviceRepo) Delete(_ context.Context, id string) error { delete(f.byID, id); return nil }
func (f *fakeDeviceRepo) FindByUserScope(context.Context, uint) ([]device.Device, error) {
	return nil, nil
}

type fakeVault struct{ creds map[string]device.Credentials }

func (f *fakeVault) Get(_ context.Context, id string) (device.Credentials, error) {
	c, ok := f.creds[id]
	if !ok {
		return device.Credentials{}, device.ErrNotFound
	}
	return c, nil
}
func (f *fakeVault) Save(_ context.Context, id string, c device.Credentials) error {
	f.creds[id] = c
	return nil
}
func (f *fakeVault) EncryptString(context.Context, string) (string, error) { return "", nil }
func (f *fakeVault) DecryptString(context.Context, string) (string, error) { return "", nil }

type fakeManager struct {
	rows []port.OLTONU
	err  error
}

func (m *fakeManager) Detect(context.Context) (domainOLT.Capabilities, port.OLTHealth, error) {
	return domainOLT.CapabilitiesFor(domainOLT.VariantHA7304), port.OLTHealth{Reachable: true, Variant: domainOLT.VariantHA7304}, nil
}
func (m *fakeManager) ListONUs(context.Context, int) ([]port.OLTONU, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rows, nil
}
func (m *fakeManager) RenameONU(context.Context, int, int, string) error { return nil }
func (m *fakeManager) RebootONU(context.Context, int, int) error         { return nil }
func (m *fakeManager) SetONUAdmin(context.Context, int, int, bool) error { return nil }
func (m *fakeManager) DeleteONU(context.Context, int, int) error         { return nil }
func (m *fakeManager) SaveConfig(context.Context) error                  { return nil }
func (m *fakeManager) RunningConfig(context.Context) (string, error)     { return "config", nil }
func (m *fakeManager) SystemInfo(context.Context) (map[string]string, error) {
	return map[string]string{"Version": "v7.76"}, nil
}

type fakeProvider struct{ mgr port.OLTManager }

func (p *fakeProvider) ForDevice(context.Context, string) (port.OLTManager, error) { return p.mgr, nil }

type fakeAlerts struct{ transitions []string }

func (a *fakeAlerts) ONTStateChanged(_ context.Context, before, after ont.ONT) {
	a.transitions = append(a.transitions, string(before.State)+"->"+string(after.State))
}

func (a *fakeAlerts) ONTRxChanged(_ context.Context, _, _ ont.ONT) {}

func newTestService(t *testing.T) (*Service, *fakeDeviceRepo, *mocktest.FakeONTRepository, *fakeManager, *fakeAlerts) {
	t.Helper()
	devices := &fakeDeviceRepo{byID: map[string]device.Device{
		"olt-1": {ID: "olt-1", TenantID: tenantDefault, Name: "OLT-Pusat", Vendor: vendorHioso, DriverType: driverTypeHioso, Host: "10.0.0.1", Extra: map[string]string{"snmp_port": "161", "variant": "HA7304"}, Enabled: true},
	}}
	onts := mocktest.NewFakeONTRepository()
	mgr := &fakeManager{}
	alerts := &fakeAlerts{}
	svc := NewService(Deps{
		Devices:  devices,
		Vault:    &fakeVault{creds: map[string]device.Credentials{}},
		Managers: &fakeProvider{mgr: mgr},
		ONTs:     onts,
		Settings: mocktest.NewFakeSettingReader(map[string]string{
			"olt.offline_strikes":       "2",
			"olt.roster_miss_threshold": "2",
			"olt.rx_na_strikes":         "3",
		}),
		Alerts: alerts,
	})
	return svc, devices, onts, mgr, alerts
}

func row(mac string, pon, onu int, up bool, rx *float64) port.OLTONU {
	return port.OLTONU{
		PONIndex: pon, ONUIndex: onu, MACAddress: mac, Name: "CUST", LinkUp: up,
		RxPowerDBM: rx, LLID: onu,
	}
}

func f64(v float64) *float64 { return &v }

func TestSyncONUsDiscoversAndUpdates(t *testing.T) {
	t.Parallel()
	svc, _, onts, mgr, _ := newTestService(t)
	ctx := context.Background()
	mgr.rows = []port.OLTONU{
		row("78:5c:72:50:00:01", 1, 1, true, f64(-20.5)),
		row("78:5c:72:50:00:02", 1, 2, false, nil),
	}

	res, err := svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)
	assert.Equal(t, 2, res.Discovered)
	assert.Equal(t, 1, res.Online)
	assert.Equal(t, 1, res.Offline)

	list, err := onts.ListByOLT(ctx, "olt-1")
	require.NoError(t, err)
	require.Len(t, list, 2)
	online := findMAC(t, list, "78:5c:72:50:00:01")
	assert.Equal(t, ont.StateOnline, online.State)
	require.NotNil(t, online.RxPowerDBM)
	assert.InDelta(t, -20.5, *online.RxPowerDBM, 0.001)
	offline := findMAC(t, list, "78:5c:72:50:00:02")
	assert.Equal(t, ont.StateUnknown, offline.State, "poll offline pertama belum cukup strikes")
	assert.Equal(t, 1, offline.OfflineStrikes)

	// Poll kedua: ONU kedua masih offline → OFFLINE, event + alert.
	mgr.rows = []port.OLTONU{
		row("78:5c:72:50:00:01", 1, 1, true, f64(-20.5)),
		row("78:5c:72:50:00:02", 1, 2, false, nil),
	}
	res, err = svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)
	assert.Equal(t, 0, res.Discovered)
	list, _ = onts.ListByOLT(ctx, "olt-1")
	assert.Equal(t, ont.StateOffline, findMAC(t, list, "78:5c:72:50:00:02").State)
}

func TestSyncONUsRosterCarryForward(t *testing.T) {
	t.Parallel()
	svc, _, onts, mgr, _ := newTestService(t)
	ctx := context.Background()
	mgr.rows = []port.OLTONU{
		row("78:5c:72:50:00:01", 1, 1, true, f64(-21)),
		row("78:5c:72:50:00:02", 1, 2, true, f64(-22)),
	}
	_, err := svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)

	// ONU kedua hilang dari walk (mis. packet loss) → state dipertahankan,
	// missed_polls bertambah, belum di-REMOVE.
	mgr.rows = []port.OLTONU{row("78:5c:72:50:00:01", 1, 1, true, f64(-21))}
	res, err := svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)
	assert.Equal(t, 0, res.Removed)
	list, _ := onts.ListByOLT(ctx, "olt-1")
	missing := findMAC(t, list, "78:5c:72:50:00:02")
	assert.Equal(t, ont.StateOnline, missing.State)
	assert.Equal(t, 1, missing.MissedPolls)

	// Poll berikutnya masih hilang → ambang roster (2) tercapai → REMOVED.
	_, err = svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)
	list, _ = onts.ListByOLT(ctx, "olt-1")
	assert.Equal(t, ont.StateRemoved, findMAC(t, list, "78:5c:72:50:00:02").State)
}

func TestSyncONUsReplacesONUAtSamePosition(t *testing.T) {
	t.Parallel()
	svc, _, onts, mgr, _ := newTestService(t)
	ctx := context.Background()
	mgr.rows = []port.OLTONU{row("78:5c:72:50:00:01", 1, 1, true, f64(-21))}
	_, err := svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)

	// ONU diganti unit baru pada posisi PON/ONU yang sama.
	mgr.rows = []port.OLTONU{row("78:5c:72:50:00:0a", 1, 1, true, f64(-19))}
	res, err := svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)
	assert.Equal(t, 1, res.Discovered)
	assert.Equal(t, 1, res.Removed)
	list, _ := onts.ListByOLT(ctx, "olt-1")
	require.Len(t, list, 2, "baris lama dipertahankan sebagai histori REMOVED")
	assert.Equal(t, ont.StateRemoved, findMAC(t, list, "78:5c:72:50:00:01").State)
	assert.Equal(t, ont.StateOnline, findMAC(t, list, "78:5c:72:50:00:0a").State)
}

func TestSyncAllSkipsNonHioso(t *testing.T) {
	t.Parallel()
	svc, devices, _, mgr, _ := newTestService(t)
	ctx := context.Background()
	devices.byID["mt-1"] = device.Device{ID: "mt-1", DriverType: "mikrotik", Enabled: true}
	mgr.rows = []port.OLTONU{row("78:5c:72:50:00:01", 1, 1, true, f64(-21))}

	res, err := svc.SyncAll(ctx)
	require.NoError(t, err)
	assert.Contains(t, res, "olt-1")
	assert.NotContains(t, res, "mt-1")
}

func findMAC(t *testing.T, list []ont.ONT, mac string) ont.ONT {
	t.Helper()
	for _, o := range list {
		if o.MACAddress == mac {
			return o
		}
	}
	t.Fatalf("ONT %s tidak ditemukan", mac)
	return ont.ONT{}
}

func TestCreateValidatesInput(t *testing.T) {
	t.Parallel()
	svc, _, _, _, _ := newTestService(t)
	_, err := svc.Create(context.Background(), Input{Name: "x"})
	require.Error(t, err)
	assert.ErrorIs(t, err, domainOLT.ErrInvalidInput)
}

func TestSyncONUsUsesSettingsThresholds(t *testing.T) {
	t.Parallel()
	svc, _, onts, mgr, alerts := newTestService(t)
	ctx := context.Background()
	mgr.rows = []port.OLTONU{row("78:5c:72:50:00:01", 1, 1, true, f64(-20))}
	_, err := svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)

	mgr.rows = []port.OLTONU{row("78:5c:72:50:00:01", 1, 1, false, nil)}
	_, err = svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)
	list, _ := onts.ListByOLT(ctx, "olt-1")
	assert.Equal(t, ont.StateOnline, list[0].State, "strike 1/2 belum offline")
	assert.Empty(t, alerts.transitions)

	_, err = svc.SyncONUs(ctx, "olt-1")
	require.NoError(t, err)
	list, _ = onts.ListByOLT(ctx, "olt-1")
	assert.Equal(t, ont.StateOffline, list[0].State)
	assert.Equal(t, []string{"ONLINE->OFFLINE"}, alerts.transitions)
	assert.NotZero(t, list[0].LastSeenAt)
	_ = time.Now()
}
