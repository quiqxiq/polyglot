package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	domainONT "github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
)

const (
	testOLTDeviceID = "22222222-2222-2222-2222-222222222222"
	testCustomerID  = "cust-ont-1"
)

// newTestDB membuka GORM ke PostgreSQL/Timescale ephemeral yang seluruh
// migrasi SQL sudah diterapkan (lihat startMigratedPostgres).
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("short mode")
	}
	_, _, dsn := startMigratedPostgres(t)
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func seedOLTTestDevice(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO devices (id, tenant_id, name, vendor, driver_type, host)
		 VALUES (?, 'tenant-default', 'OLT Test', 'hioso', 'hioso', '10.10.0.2')`,
		testOLTDeviceID).Error)
}

func seedTestCustomer(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO customers (id, tenant_id, customer_code, name, phone, address, portal_access_code)
		 VALUES (?, 'tenant-default', 'CUST-ONT-1', 'Pelanggan ONT', '081200000001', 'Jl. Uji', '11112222')`,
		testCustomerID).Error)
}

func f64Ptr(v float64) *float64 { return &v }
func intPtr(v int) *int         { return &v }
func strPtr(v string) *string   { return &v }
func boolPtr(v bool) *bool      { return &v }

func assertTimePtr(t *testing.T, want time.Time, got *time.Time) {
	t.Helper()
	require.NotNil(t, got)
	assert.WithinDuration(t, want, *got, time.Microsecond)
}

// TestONTRepository_SaveFindUpsert memverifikasi round-trip seluruh kolom,
// lookup identity, upsert by id, dan partial unique index posisi aktif.
func TestONTRepository_SaveFindUpsert(t *testing.T) {
	db := newTestDB(t)
	seedOLTTestDevice(t, db)
	ctx := context.Background()
	repo := postgres.NewONTRepository(db)

	now := time.Now().UTC().Truncate(time.Microsecond)
	o := domainONT.ONT{
		ID: "ont-test-1", TenantID: "tenant-default", OLTDeviceID: testOLTDeviceID,
		PONIndex: 1, ONUIndex: 3, LLID: intPtr(3), Name: "PELANGGAN-A",
		MACAddress: "aa:bb:cc:dd:ee:01", SerialNumber: "ZTEG123", Model: "HG8245",
		Firmware: "V5R020", State: domainONT.StateOnline, AdminState: domainONT.AdminEnable,
		RxPowerDBM: f64Ptr(-21.4), TxPowerDBM: f64Ptr(2.5), DistanceM: intPtr(1200),
		TemperatureC: f64Ptr(42.5), MissedPolls: 1, OfflineStrikes: 2, RxNAStrikes: 0,
		LastSeenAt: &now, LastPolledAt: &now,
		ACSDeviceID: strPtr("acs-1"), ACSMatchSource: string(domainONT.MatchMAC),
		ACSLastInform: &now, ACSTelemetry: json.RawMessage(`{"ssid":"polyglot"}`), WLANSlots: 2,
		Source: "SNMP",
	}
	require.NoError(t, repo.Save(ctx, o))

	got, err := repo.FindByID(ctx, o.ID)
	require.NoError(t, err)
	assert.Equal(t, o.Name, got.Name)
	assert.Equal(t, o.MACAddress, got.MACAddress)
	assert.Equal(t, o.SerialNumber, got.SerialNumber)
	assert.Equal(t, domainONT.StateOnline, got.State)
	assert.Equal(t, domainONT.AdminEnable, got.AdminState)
	assert.InDelta(t, -21.4, *got.RxPowerDBM, 0.001)
	assert.InDelta(t, 2.5, *got.TxPowerDBM, 0.001)
	assert.Equal(t, 1200, *got.DistanceM)
	assert.InDelta(t, 42.5, *got.TemperatureC, 0.001)
	assert.Equal(t, 1, got.MissedPolls)
	assert.Equal(t, 2, got.OfflineStrikes)
	assertTimePtr(t, now, got.LastSeenAt)
	assert.Equal(t, "acs-1", *got.ACSDeviceID)
	assert.Equal(t, string(domainONT.MatchMAC), got.ACSMatchSource)
	assertTimePtr(t, now, got.ACSLastInform)
	assert.JSONEq(t, `{"ssid":"polyglot"}`, string(got.ACSTelemetry))
	assert.Equal(t, 2, got.WLANSlots)
	assert.Equal(t, "SNMP", got.Source)
	assert.False(t, got.CreatedAt.IsZero())

	// Lookup MAC menerima format non-kanonik (aabb.ccdd.ee01).
	byMAC, err := repo.FindByOLTAndMAC(ctx, testOLTDeviceID, "AABB.CCDD.EE01")
	require.NoError(t, err)
	assert.Equal(t, o.ID, byMAC.ID)
	byPos, err := repo.FindByOLTPosition(ctx, testOLTDeviceID, 1, 3)
	require.NoError(t, err)
	assert.Equal(t, o.ID, byPos.ID)
	list, err := repo.ListByOLT(ctx, testOLTDeviceID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	// Upsert by id: state/counter berubah, tidak membuat baris baru.
	got.State = domainONT.StateOffline
	got.Name = "PELANGGAN-A2"
	got.OfflineStrikes = 0
	got.MissedPolls = 0
	require.NoError(t, repo.Save(ctx, got))
	updated, err := repo.FindByID(ctx, o.ID)
	require.NoError(t, err)
	assert.Equal(t, domainONT.StateOffline, updated.State)
	assert.Equal(t, "PELANGGAN-A2", updated.Name)
	assert.Equal(t, 0, updated.OfflineStrikes)
	all, err := repo.ListByOLT(ctx, testOLTDeviceID)
	require.NoError(t, err)
	require.Len(t, all, 1)

	// Partial unique index: posisi aktif tidak boleh ganda, tetapi baris
	// REMOVED boleh berbagi posisi (histori ONU pengganti).
	replacement := o
	replacement.ID = "ont-test-2"
	replacement.MACAddress = "aa:bb:cc:dd:ee:02"
	require.Error(t, repo.Save(ctx, replacement), "posisi aktif ganda harus ditolak")

	removed, err := repo.FindByID(ctx, o.ID)
	require.NoError(t, err)
	removed.State = domainONT.StateRemoved
	require.NoError(t, repo.Save(ctx, removed))
	require.NoError(t, repo.Save(ctx, replacement), "posisi boleh dipakai ulang setelah REMOVED")
	all, err = repo.ListByOLT(ctx, testOLTDeviceID)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

// TestONTRepository_ListFiltersAndCounts memverifikasi seluruh filter List
// (OLT/PON/state/search/linked/ACS) + paginasi + agregat CountByState.
func TestONTRepository_ListFiltersAndCounts(t *testing.T) {
	db := newTestDB(t)
	seedOLTTestDevice(t, db)
	seedTestCustomer(t, db)
	ctx := context.Background()
	repo := postgres.NewONTRepository(db)

	now := time.Now().UTC()
	linkedID := testCustomerID
	base := domainONT.ONT{
		TenantID: "tenant-default", OLTDeviceID: testOLTDeviceID,
		State: domainONT.StateOnline, AdminState: domainONT.AdminEnable,
		Source: "SNMP", LastSeenAt: &now,
	}
	save := func(id string, pon, onu int, mac, serial, model string, state domainONT.State, customer *string, acs *string) {
		t.Helper()
		o := base
		o.ID, o.PONIndex, o.ONUIndex = id, pon, onu
		o.MACAddress, o.SerialNumber, o.Model = mac, serial, model
		o.State = state
		o.CustomerID, o.ACSDeviceID = customer, acs
		require.NoError(t, repo.Save(ctx, o))
	}
	save("ont-a", 1, 1, "aa:bb:cc:dd:00:01", "ZTEG001", "HG8245", domainONT.StateOnline, nil, nil)
	save("ont-b", 1, 2, "aa:bb:cc:dd:00:02", "ZTEG002", "HG8245", domainONT.StateOffline, nil, strPtr("acs-2"))
	save("ont-c", 2, 1, "aa:bb:cc:dd:00:03", "HWTC003", "F660", domainONT.StateUnknown, &linkedID, nil)

	list := func(f port.ONTFilter) []domainONT.ONT {
		t.Helper()
		out, err := repo.List(ctx, f)
		require.NoError(t, err)
		return out
	}
	assert.Len(t, list(port.ONTFilter{OLTDeviceID: testOLTDeviceID}), 3)
	assert.Len(t, list(port.ONTFilter{PONIndex: 1}), 2)
	assert.Len(t, list(port.ONTFilter{State: string(domainONT.StateOffline)}), 1)
	assert.Len(t, list(port.ONTFilter{Search: "hwtc"}), 1)
	assert.Len(t, list(port.ONTFilter{Search: "HG8245"}), 2)
	assert.Len(t, list(port.ONTFilter{Linked: boolPtr(true)}), 1)
	assert.Len(t, list(port.ONTFilter{Linked: boolPtr(false)}), 2)
	assert.Len(t, list(port.ONTFilter{ACSMatched: boolPtr(true)}), 1)
	assert.Len(t, list(port.ONTFilter{ACSMatched: boolPtr(false)}), 2)

	paged := list(port.ONTFilter{Limit: 2, Offset: 1})
	require.Len(t, paged, 2)
	assert.Equal(t, "ont-b", paged[0].ID) // urut OLT, PON, ONU
	assert.Equal(t, "ont-c", paged[1].ID)
	assert.Empty(t, list(port.ONTFilter{OLTDeviceID: "33333333-3333-3333-3333-333333333333"}))

	counts, err := repo.CountByState(ctx, testOLTDeviceID)
	require.NoError(t, err)
	assert.Equal(t, port.ONTStateCounts{Total: 3, Online: 1, Offline: 1, Unknown: 1}, counts)
	counts, err = repo.CountByState(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, 3, counts.Total)
	counts, err = repo.CountByState(ctx, "33333333-3333-3333-3333-333333333333")
	require.NoError(t, err)
	assert.Equal(t, port.ONTStateCounts{}, counts)
}

// TestONTRepository_UpdateLinkAndACS memverifikasi update kolom link & ACS
// beserta pemetaan not-found ke sentinel domain.
func TestONTRepository_UpdateLinkAndACS(t *testing.T) {
	db := newTestDB(t)
	seedOLTTestDevice(t, db)
	seedTestCustomer(t, db)
	ctx := context.Background()
	repo := postgres.NewONTRepository(db)

	require.NoError(t, repo.Save(ctx, domainONT.ONT{
		ID: "ont-link", OLTDeviceID: testOLTDeviceID, PONIndex: 1, ONUIndex: 1,
		MACAddress: "aa:bb:cc:dd:01:01", State: domainONT.StateOnline,
		AdminState: domainONT.AdminEnable, Source: "SNMP",
	}))

	customerID := testCustomerID
	require.NoError(t, repo.UpdateLink(ctx, "ont-link", &customerID, nil, "user@isp"))
	got, err := repo.FindByID(ctx, "ont-link")
	require.NoError(t, err)
	require.NotNil(t, got.CustomerID)
	assert.Equal(t, testCustomerID, *got.CustomerID)
	assert.Nil(t, got.SubscriptionID)
	assert.Equal(t, "user@isp", got.PPPoEUsername)
	assert.True(t, got.ISOLinked())

	inform := time.Now().UTC().Truncate(time.Microsecond)
	telemetry := []byte(`{"rx":-22.5,"temp":41}`)
	require.NoError(t, repo.UpdateACS(ctx, "ont-link", strPtr("acs-9"),
		string(domainONT.MatchMACNeighbor), &inform, telemetry, 4))
	got, err = repo.FindByID(ctx, "ont-link")
	require.NoError(t, err)
	require.NotNil(t, got.ACSDeviceID)
	assert.Equal(t, "acs-9", *got.ACSDeviceID)
	assert.Equal(t, string(domainONT.MatchMACNeighbor), got.ACSMatchSource)
	assertTimePtr(t, inform, got.ACSLastInform)
	assert.JSONEq(t, string(telemetry), string(got.ACSTelemetry))
	assert.Equal(t, 4, got.WLANSlots)

	// Not-found dipetakan ke sentinel domain.
	_, err = repo.FindByID(ctx, "ont-missing")
	assert.ErrorIs(t, err, domainONT.ErrNotFound)
	_, err = repo.FindByOLTAndMAC(ctx, testOLTDeviceID, "aa:bb:cc:dd:ff:ff")
	assert.ErrorIs(t, err, domainONT.ErrNotFound)
	_, err = repo.FindByOLTPosition(ctx, testOLTDeviceID, 9, 9)
	assert.ErrorIs(t, err, domainONT.ErrNotFound)
	assert.ErrorIs(t, repo.UpdateLink(ctx, "ont-missing", nil, nil, ""), domainONT.ErrNotFound)
	assert.ErrorIs(t, repo.UpdateACS(ctx, "ont-missing", nil, "", nil, nil, 0), domainONT.ErrNotFound)
}

// TestONTRepository_EventsAndSamples memverifikasi timeline event dan sampel
// metrik: ID otomatis, urutan, filter waktu, limit, dan purging retensi.
func TestONTRepository_EventsAndSamples(t *testing.T) {
	db := newTestDB(t)
	seedOLTTestDevice(t, db)
	ctx := context.Background()
	repo := postgres.NewONTRepository(db)
	require.NoError(t, repo.Save(ctx, domainONT.ONT{
		ID: "ont-hist", OLTDeviceID: testOLTDeviceID, PONIndex: 1, ONUIndex: 1,
		MACAddress: "aa:bb:cc:dd:02:01", State: domainONT.StateOnline,
		AdminState: domainONT.AdminEnable, Source: "SNMP",
	}))
	// ONT kedua hanya agar event FK ont_events.ont_id terpenuhi saat menguji
	// filter timeline per ONT.
	require.NoError(t, repo.Save(ctx, domainONT.ONT{
		ID: "ont-other", OLTDeviceID: testOLTDeviceID, PONIndex: 1, ONUIndex: 2,
		MACAddress: "aa:bb:cc:dd:02:02", State: domainONT.StateOnline,
		AdminState: domainONT.AdminEnable, Source: "SNMP",
	}))

	now := time.Now().UTC().Truncate(time.Microsecond)
	events := []domainONT.Event{
		{ONTID: "ont-hist", EventType: "DISCOVERED", NewState: "ONLINE", CreatedAt: now.Add(-2 * time.Hour)},
		{ONTID: "ont-hist", EventType: "OFFLINE", OldState: "ONLINE", NewState: "OFFLINE",
			Message: "los", Metadata: json.RawMessage(`{"los":true}`), CreatedAt: now.Add(-time.Hour)},
		{ONTID: "ont-hist", EventType: "ONLINE", OldState: "OFFLINE", NewState: "ONLINE", CreatedAt: now},
		{ONTID: "ont-other", EventType: "DISCOVERED", CreatedAt: now},
	}
	for _, e := range events {
		require.NoError(t, repo.RecordEvent(ctx, e))
	}

	latest, err := repo.ListEvents(ctx, "ont-hist", 2)
	require.NoError(t, err)
	require.Len(t, latest, 2)
	assert.Equal(t, "ONLINE", latest[0].EventType)
	assert.Equal(t, "OFFLINE", latest[1].EventType)
	assert.NotEmpty(t, latest[0].ID, "event ID harus dibuat repository")
	assert.JSONEq(t, `{"los":true}`, string(latest[1].Metadata))
	assert.WithinDuration(t, now.Add(-time.Hour), latest[1].CreatedAt, time.Microsecond)

	allEvents, err := repo.ListEvents(ctx, "ont-hist", 0)
	require.NoError(t, err)
	require.Len(t, allEvents, 3)
	assert.Equal(t, "DISCOVERED", allEvents[2].EventType)

	old, mid, recent := now.Add(-3*time.Hour), now.Add(-2*time.Hour), now
	samples := []domainONT.MetricSample{
		{ONTID: "ont-hist", SampledAt: old, RxPowerDBM: f64Ptr(-30), Online: false},
		{ONTID: "ont-hist", SampledAt: mid, RxPowerDBM: f64Ptr(-25), Online: true},
		{ONTID: "ont-hist", SampledAt: recent, RxPowerDBM: f64Ptr(-21), Online: true},
	}
	for _, s := range samples {
		require.NoError(t, repo.AddSample(ctx, s))
	}

	from := mid.Add(-time.Minute)
	recentSamples, err := repo.ListSamples(ctx, "ont-hist", from, 1)
	require.NoError(t, err)
	require.Len(t, recentSamples, 1)
	assert.InDelta(t, -21, *recentSamples[0].RxPowerDBM, 0.001)
	assert.NotEmpty(t, recentSamples[0].ID)

	samplesInRange, err := repo.ListSamples(ctx, "ont-hist", from, 0)
	require.NoError(t, err)
	require.Len(t, samplesInRange, 2)
	assert.True(t, samplesInRange[0].SampledAt.After(samplesInRange[1].SampledAt))

	// sampled_at kosong diisi repository dengan waktu sekarang.
	require.NoError(t, repo.AddSample(ctx, domainONT.MetricSample{ONTID: "ont-hist", Online: true}))
	autoSample, err := repo.ListSamples(ctx, "ont-hist", time.Time{}, 1)
	require.NoError(t, err)
	require.Len(t, autoSample, 1)
	assert.NotEmpty(t, autoSample[0].ID)
	assert.False(t, autoSample[0].SampledAt.IsZero(), "sampled_at kosong diisi repository")

	purged, err := repo.PurgeSamplesBefore(ctx, mid)
	require.NoError(t, err)
	assert.Equal(t, int64(1), purged)
	remaining, err := repo.ListSamples(ctx, "ont-hist", time.Time{}, 0)
	require.NoError(t, err)
	assert.Len(t, remaining, 3)
}
