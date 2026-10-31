package ont

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
)

// fakeACSManager memenuhi port.ACSManager untuk unit test.
type fakeACSManager struct {
	devices    []port.ACSDeviceSummary
	telemetry  port.ACSTelemetry
	presets    map[string]ont.ACSPreset
	setParams  map[string]map[string]string
	rebooted   []string
	lastInform time.Time
}

func (f *fakeACSManager) Ping(context.Context) error { return nil }

func (f *fakeACSManager) ListDevices(context.Context, string) ([]port.ACSDeviceSummary, error) {
	return f.devices, nil
}

func (f *fakeACSManager) GetDevice(_ context.Context, deviceID string) (port.ACSDeviceSummary, error) {
	for _, d := range f.devices {
		if d.DeviceID == deviceID {
			return d, nil
		}
	}
	return port.ACSDeviceSummary{}, errors.New("not found")
}

func (f *fakeACSManager) FetchTelemetry(context.Context, string, ont.TelemetryCatalog) (port.ACSTelemetry, error) {
	t := f.telemetry
	t.FetchedAt = time.Now().UTC()
	return t, nil
}

func (f *fakeACSManager) SetParameters(_ context.Context, deviceID string, values map[string]string) error {
	if f.setParams == nil {
		f.setParams = map[string]map[string]string{}
	}
	f.setParams[deviceID] = values
	return nil
}

func (f *fakeACSManager) Reboot(_ context.Context, deviceID string) error {
	f.rebooted = append(f.rebooted, deviceID)
	return nil
}

func (f *fakeACSManager) EnsurePreset(_ context.Context, p ont.ACSPreset) error {
	if f.presets == nil {
		f.presets = map[string]ont.ACSPreset{}
	}
	f.presets[p.Name] = p
	return nil
}

func (f *fakeACSManager) DeletePreset(context.Context, string) error { return nil }

func (f *fakeACSManager) ListFaults(context.Context, string) ([]string, error) { return nil, nil }

func newACSTestService(t *testing.T) (*Service, *mocktest.FakeONTRepository, *mocktest.FakeACSRepository, *fakeACSManager) {
	t.Helper()
	onts := mocktest.NewFakeONTRepository()
	acsRepo := mocktest.NewFakeACSRepository()
	acs := &fakeACSManager{}
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"acs.acs_url":                  "http://acs.example:7547",
		"acs.periodic_inform_interval": "300",
		"acs.preset_name":              "polyglot-bootstrap",
	})
	svc := NewService(Deps{ONTs: onts, ACS: acs, ACSRepo: acsRepo, Settings: settings})
	return svc, onts, acsRepo, acs
}

func seedONT(t *testing.T, onts *mocktest.FakeONTRepository, mac, serial string) ont.ONT {
	t.Helper()
	o := ont.ONT{
		ID: "ont-" + mac[len(mac)-2:], TenantID: tenantDefault, OLTDeviceID: "olt-1",
		PONIndex: 1, ONUIndex: 1, MACAddress: mac, SerialNumber: serial,
		State: ont.StateOnline, AdminState: ont.AdminEnable,
	}
	onts.Seed(o)
	return o
}

func TestSyncDevicesMatchesBySerial(t *testing.T) {
	t.Parallel()
	svc, onts, acsRepo, acs := newACSTestService(t)
	seedONT(t, onts, "78:5c:72:50:00:01", "HWTC123")
	acs.devices = []port.ACSDeviceSummary{{
		DeviceID: "AABBCC-HW-123", OUI: "AABBCC", ProductClass: "HW",
		SerialNumber: "HWTC123", LastInform: time.Now().UTC(),
	}}
	stats, err := svc.SyncDevices(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Devices)
	assert.Equal(t, 1, stats.Matched)
	assert.Equal(t, 0, stats.Conflicts)

	o, _ := onts.FindByID(context.Background(), "ont-01")
	require.NotNil(t, o.ACSDeviceID)
	assert.Equal(t, "AABBCC-HW-123", *o.ACSDeviceID)
	assert.Equal(t, string(ont.MatchSerial), o.ACSMatchSource)
	dev, errDev := acsRepo.FindDeviceByACSID(context.Background(), "AABBCC-HW-123")
	require.NoError(t, errDev)
	require.NotNil(t, dev.MatchedONTID)
	assert.Equal(t, o.ID, *dev.MatchedONTID)
}

func TestSyncDevicesMatchesByMACNeighbor(t *testing.T) {
	t.Parallel()
	svc, onts, _, acs := newACSTestService(t)
	seedONT(t, onts, "78:5c:72:50:00:01", "")
	// MAC device ACS berbeda 1 pada oktet terakhir → heuristik.
	acs.devices = []port.ACSDeviceSummary{{
		DeviceID: "AABBCC-HW-124", SerialNumber: "OTHER",
		MACCandidates: []string{"78:5c:72:50:00:02"},
	}}
	stats, err := svc.SyncDevices(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Matched)
	o, _ := onts.FindByID(context.Background(), "ont-01")
	assert.Equal(t, string(ont.MatchMACNeighbor), o.ACSMatchSource)
}

func TestSyncDevicesConflictReleased(t *testing.T) {
	t.Parallel()
	svc, onts, acsRepo, acs := newACSTestService(t)
	seedONT(t, onts, "78:5c:72:50:00:01", "SERIAL1")
	acs.devices = []port.ACSDeviceSummary{
		{DeviceID: "dev-a", SerialNumber: "SERIAL1"},
		{DeviceID: "dev-b", SerialNumber: "SERIAL1"},
	}
	stats, err := svc.SyncDevices(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Conflicts)
	o, _ := onts.FindByID(context.Background(), "ont-01")
	assert.Nil(t, o.ACSDeviceID, "konflik harus melepas kedua device")
	for _, id := range []string{"dev-a", "dev-b"} {
		dev, errDev := acsRepo.FindDeviceByACSID(context.Background(), id)
		require.NoError(t, errDev)
		assert.Nil(t, dev.MatchedONTID)
	}
}

func TestTelemetryUpdatesONT(t *testing.T) {
	t.Parallel()
	svc, onts, _, acs := newACSTestService(t)
	o := seedONT(t, onts, "78:5c:72:50:00:01", "SERIAL1")
	devID := "dev-a"
	require.NoError(t, onts.UpdateACS(context.Background(), o.ID, &devID, string(ont.MatchSerial), nil, nil, 0))
	acs.telemetry = port.ACSTelemetry{
		SerialNumber: "SERIAL1", SSID24: "Home", SSID5: "Home-5G",
		WiFiClients: 3, RxPowerDBM: f64(-23.4), TemperatureC: f64(41.2),
		WANUsername: "cust01",
	}
	tel, err := svc.Telemetry(context.Background(), o.ID)
	require.NoError(t, err)
	assert.Equal(t, "Home", tel.SSID24)
	updated, _ := onts.FindByID(context.Background(), o.ID)
	assert.Equal(t, 2, updated.WLANSlots)
	require.NotNil(t, updated.ACSTelemetry)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(updated.ACSTelemetry, &raw))
	assert.Equal(t, "Home-5G", raw["SSID5"])
}

func TestLinkValidatesAndSets(t *testing.T) {
	t.Parallel()
	svc, onts, _, _ := newACSTestService(t)
	o := seedONT(t, onts, "78:5c:72:50:00:01", "")
	require.Error(t, svc.Link(context.Background(), o.ID, LinkInput{}))
	require.NoError(t, svc.Link(context.Background(), o.ID, LinkInput{CustomerID: "cust-1", SubscriptionID: "sub-1", PPPoEUsername: "cust01"}))
	linked, _ := onts.FindByID(context.Background(), o.ID)
	require.NotNil(t, linked.CustomerID)
	assert.Equal(t, "cust-1", *linked.CustomerID)
	assert.Equal(t, "cust01", linked.PPPoEUsername)
	require.ErrorIs(t, svc.Link(context.Background(), o.ID, LinkInput{SubscriptionID: "sub-2"}), ont.ErrAlreadyLinked)
}

func TestBootstrapPreset(t *testing.T) {
	t.Parallel()
	svc, _, acsRepo, acs := newACSTestService(t)
	require.NoError(t, svc.BootstrapPreset(context.Background()))
	require.Contains(t, acs.presets, "polyglot-bootstrap")
	preset := acs.presets["polyglot-bootstrap"]
	assert.Contains(t, string(preset.Configurations), "http://acs.example:7547")
	assert.Contains(t, string(preset.Configurations), "300")
	stored, err := acsRepo.GetPreset(context.Background(), "polyglot-bootstrap")
	require.NoError(t, err)
	assert.NotNil(t, stored.LastSyncedAt)
}

func f64(v float64) *float64 { return &v }
