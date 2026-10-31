package mocktest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainONT "github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
)

func TestFakeONTRepository_ListFiltersAndTimeline(t *testing.T) {
	ctx := context.Background()
	repo := NewFakeONTRepository()
	customerID := "cust-1"
	now := time.Now()
	repo.Seed(domainONT.ONT{
		ID: "ont-1", OLTDeviceID: "olt-1", PONIndex: 1, ONUIndex: 1,
		MACAddress: "aa:bb:cc:dd:ee:01", SerialNumber: "ZTEG001",
		State: domainONT.StateOnline,
	})
	repo.Seed(domainONT.ONT{
		ID: "ont-2", OLTDeviceID: "olt-1", PONIndex: 1, ONUIndex: 2,
		MACAddress: "aa:bb:cc:dd:ee:02", State: domainONT.StateOffline,
		CustomerID: &customerID,
	})
	repo.Seed(domainONT.ONT{
		ID: "ont-3", OLTDeviceID: "olt-2", PONIndex: 2, ONUIndex: 1,
		MACAddress: "aa:bb:cc:dd:ee:03", State: domainONT.StateUnknown,
	})

	list, err := repo.List(ctx, port.ONTFilter{OLTDeviceID: "olt-1"})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "ont-1", list[0].ID)
	assert.Len(t, mustList(t, repo, port.ONTFilter{State: string(domainONT.StateOffline)}), 1)
	assert.Len(t, mustList(t, repo, port.ONTFilter{Linked: boolPointer(true)}), 1)
	assert.Len(t, mustList(t, repo, port.ONTFilter{Search: "ZTEG"}), 1)
	assert.Len(t, mustList(t, repo, port.ONTFilter{PONIndex: 2}), 1)
	paged := mustList(t, repo, port.ONTFilter{Limit: 2, Offset: 1})
	require.Len(t, paged, 2)
	assert.Equal(t, "ont-2", paged[0].ID)

	byMAC, err := repo.FindByOLTAndMAC(ctx, "olt-1", "AABB.CCDD.EE02")
	require.NoError(t, err)
	assert.Equal(t, "ont-2", byMAC.ID)
	_, err = repo.FindByOLTPosition(ctx, "olt-1", 9, 9)
	assert.ErrorIs(t, err, ErrFakeNotFound)

	counts, err := repo.CountByState(ctx, "olt-1")
	require.NoError(t, err)
	assert.Equal(t, port.ONTStateCounts{Total: 2, Online: 1, Offline: 1}, counts)

	// Timeline: urut terbaru dulu + limit.
	repo.SeedEvent(domainONT.Event{ID: "ev-a", ONTID: "ont-1", EventType: "DISCOVERED",
		CreatedAt: now.Add(-time.Hour)})
	repo.SeedEvent(domainONT.Event{ID: "ev-b", ONTID: "ont-1", EventType: "ONLINE",
		CreatedAt: now})
	require.NoError(t, repo.RecordEvent(ctx, domainONT.Event{ONTID: "ont-1", EventType: "LINK"}))
	events, err := repo.ListEvents(ctx, "ont-1", 2)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "LINK", events[0].EventType)
	assert.NotEmpty(t, events[0].ID, "fake harus membuat ID event seperti repository")

	// Sampel metrik: filter waktu, limit, purge.
	oldTime := now.Add(-2 * time.Hour)
	repo.SeedSample(domainONT.MetricSample{ID: "s-1", ONTID: "ont-1", SampledAt: oldTime})
	repo.SeedSample(domainONT.MetricSample{ID: "s-2", ONTID: "ont-1", SampledAt: now.Add(-time.Hour)})
	require.NoError(t, repo.AddSample(ctx, domainONT.MetricSample{ONTID: "ont-1"}))
	samples, err := repo.ListSamples(ctx, "ont-1", now.Add(-90*time.Minute), 0)
	require.NoError(t, err)
	require.Len(t, samples, 2)
	assert.NotEmpty(t, samples[0].ID)
	purged, err := repo.PurgeSamplesBefore(ctx, now.Add(-90*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, int64(1), purged)

	require.NoError(t, repo.UpdateLink(ctx, "ont-1", &customerID, nil, "user"))
	linked, err := repo.FindByID(ctx, "ont-1")
	require.NoError(t, err)
	assert.True(t, linked.ISOLinked())
	assert.Equal(t, "user", linked.PPPoEUsername)
	assert.ErrorIs(t, repo.UpdateLink(ctx, "missing", nil, nil, ""), ErrFakeNotFound)
	assert.ErrorIs(t, repo.UpdateACS(ctx, "missing", nil, "", nil, nil, 0), ErrFakeNotFound)
}

func TestFakeACSRepository_UpsertMatchAndPresets(t *testing.T) {
	ctx := context.Background()
	repo := NewFakeACSRepository()
	firstSeen := time.Now().Add(-time.Hour)
	dev := domainONT.ACSDevice{
		ACSDeviceID: "acs-1", SerialNumber: "ZTEG001", Tags: []string{"acs"},
		MACCandidates: []string{"aa:bb:cc:dd:ee:01"},
		FirstSeenAt:   firstSeen, LastSeenAt: firstSeen,
	}
	require.NoError(t, repo.UpsertDevice(ctx, dev))
	got, err := repo.FindDeviceByACSID(ctx, "acs-1")
	require.NoError(t, err)
	require.NotEmpty(t, got.ID)

	// Upsert ulang idempoten: ID & first_seen dipertahankan.
	require.NoError(t, repo.UpsertDevice(ctx, domainONT.ACSDevice{ACSDeviceID: "acs-1"}))
	got, err = repo.FindDeviceByACSID(ctx, "acs-1")
	require.NoError(t, err)
	assert.Equal(t, dev.FirstSeenAt, got.FirstSeenAt)
	list, err := repo.ListDevices(ctx, port.ACSDeviceFilter{Search: "ACS-1"})
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Match conflict: device lain yang memegang ONT sama dilepas.
	ontID := "ont-1"
	require.NoError(t, repo.UpsertDevice(ctx, domainONT.ACSDevice{ACSDeviceID: "acs-2"}))
	require.NoError(t, repo.UpdateDeviceMatch(ctx, "acs-1", &ontID, string(domainONT.MatchMAC)))
	require.NoError(t, repo.UpdateDeviceMatch(ctx, "acs-2", &ontID, string(domainONT.MatchSerial)))
	dev1, err := repo.FindDeviceByACSID(ctx, "acs-1")
	require.NoError(t, err)
	assert.Nil(t, dev1.MatchedONTID)
	dev2, err := repo.FindDeviceByACSID(ctx, "acs-2")
	require.NoError(t, err)
	assert.Equal(t, &ontID, dev2.MatchedONTID)
	matched, err := repo.ListDevices(ctx, port.ACSDeviceFilter{Matched: boolPointer(true)})
	require.NoError(t, err)
	assert.Len(t, matched, 1)
	assert.ErrorIs(t, repo.UpdateDeviceMatch(ctx, "missing", &ontID, ""), ErrFakeNotFound)

	// Preset: upsert by (tenant, name) mempertahankan ID.
	require.NoError(t, repo.SavePreset(ctx, domainONT.ACSPreset{
		Name: "polyglot-bootstrap", Configurations: []byte(`[]`), Enabled: true,
	}))
	preset, err := repo.GetPreset(ctx, "polyglot-bootstrap")
	require.NoError(t, err)
	require.NotEmpty(t, preset.ID)
	require.NoError(t, repo.SavePreset(ctx, domainONT.ACSPreset{Name: "polyglot-bootstrap", Enabled: false}))
	preset, err = repo.GetPreset(ctx, "polyglot-bootstrap")
	require.NoError(t, err)
	assert.False(t, preset.Enabled)
	require.NoError(t, repo.SavePreset(ctx, domainONT.ACSPreset{Name: "alpha"}))
	presets, err := repo.ListPresets(ctx)
	require.NoError(t, err)
	require.Len(t, presets, 2)
	assert.Equal(t, "alpha", presets[0].Name)
	_, err = repo.GetPreset(ctx, "missing")
	assert.ErrorIs(t, err, ErrFakeNotFound)
}

func mustList(t *testing.T, repo *FakeONTRepository, filter port.ONTFilter) []domainONT.ONT {
	t.Helper()
	list, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	return list
}

func boolPointer(v bool) *bool { return &v }
