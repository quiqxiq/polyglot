package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	domainONT "github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
)

// TestACSRepository_UpsertDeviceIdempotent memverifikasi upsert device cache
// ACS: idempoten by acs_device_id, mempertahankan first_seen_at & match lama,
// serta filter/pencarian ListDevices.
func TestACSRepository_UpsertDeviceIdempotent(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := postgres.NewACSRepository(db)

	firstSeen := time.Now().UTC().Add(-time.Hour).Truncate(time.Microsecond)
	lastInform := time.Now().UTC().Truncate(time.Microsecond)
	dev := domainONT.ACSDevice{
		ID: "acsd-1", TenantID: "tenant-default", ACSDeviceID: "acs-100",
		OUI: "001122", ProductClass: "HG8245", SerialNumber: "ZTEG123",
		SoftwareVersion: "V5R020", LastInformAt: &lastInform,
		Tags: []string{"polyglot", "acs"}, MACCandidates: []string{"aa:bb:cc:dd:ee:01", "aa:bb:cc:dd:ee:02"},
		Raw:         json.RawMessage(`{"vendor":"huawei"}`),
		FirstSeenAt: firstSeen, LastSeenAt: firstSeen,
	}
	require.NoError(t, repo.UpsertDevice(ctx, dev))

	got, err := repo.FindDeviceByACSID(ctx, "acs-100")
	require.NoError(t, err)
	assert.Equal(t, "acsd-1", got.ID)
	assert.Equal(t, "HG8245", got.ProductClass)
	assert.Equal(t, "ZTEG123", got.SerialNumber)
	assert.ElementsMatch(t, dev.Tags, got.Tags)
	assert.ElementsMatch(t, dev.MACCandidates, got.MACCandidates)
	assert.JSONEq(t, `{"vendor":"huawei"}`, string(got.Raw))
	assertTimePtr(t, firstSeen, &got.FirstSeenAt)
	assertTimePtr(t, lastInform, got.LastInformAt)
	assert.Nil(t, got.MatchedONTID)

	// Sync ulang tanpa ID/timestamp: baris yang sama diperbarui, first_seen
	// dan last_inform lama dipertahankan.
	resync := dev
	resync.ID = ""
	resync.FirstSeenAt = time.Time{}
	resync.LastSeenAt = time.Time{}
	resync.LastInformAt = nil
	resync.SoftwareVersion = "V5R021"
	resync.Tags = []string{"polyglot", "acs", "synced"}
	require.NoError(t, repo.UpsertDevice(ctx, resync))
	got, err = repo.FindDeviceByACSID(ctx, "acs-100")
	require.NoError(t, err)
	assert.Equal(t, "acsd-1", got.ID)
	assert.Equal(t, "V5R021", got.SoftwareVersion)
	assert.ElementsMatch(t, []string{"polyglot", "acs", "synced"}, got.Tags)
	assertTimePtr(t, firstSeen, &got.FirstSeenAt)
	assertTimePtr(t, lastInform, got.LastInformAt)
	list, err := repo.ListDevices(ctx, port.ACSDeviceFilter{})
	require.NoError(t, err)
	require.Len(t, list, 1)

	// Device kedua + filter matched/search/limit/offset.
	require.NoError(t, repo.UpsertDevice(ctx, domainONT.ACSDevice{
		ACSDeviceID: "acs-200", OUI: "445566", ProductClass: "F660",
		SerialNumber: "HWTC999", FirstSeenAt: firstSeen, LastSeenAt: lastInform,
	}))
	devices, err := repo.ListDevices(ctx, port.ACSDeviceFilter{Search: "f660"})
	require.NoError(t, err)
	require.Len(t, devices, 1)
	assert.Equal(t, "acs-200", devices[0].ACSDeviceID)
	devices, err = repo.ListDevices(ctx, port.ACSDeviceFilter{Search: "acs-100"})
	require.NoError(t, err)
	require.Len(t, devices, 1)
	devices, err = repo.ListDevices(ctx, port.ACSDeviceFilter{Matched: boolPtr(false)})
	require.NoError(t, err)
	assert.Len(t, devices, 2)
	devices, err = repo.ListDevices(ctx, port.ACSDeviceFilter{Matched: boolPtr(true)})
	require.NoError(t, err)
	assert.Empty(t, devices)
	devices, err = repo.ListDevices(ctx, port.ACSDeviceFilter{Limit: 1, Offset: 1})
	require.NoError(t, err)
	require.Len(t, devices, 1)

	_, err = repo.FindDeviceByACSID(ctx, "acs-missing")
	assert.ErrorIs(t, err, domainONT.ErrNotFound)
}

// TestACSRepository_UpdateDeviceMatch memverifikasi penetapan match dan
// pelepasan device lain yang masih memegang ONT yang sama (D10).
func TestACSRepository_UpdateDeviceMatch(t *testing.T) {
	db := newTestDB(t)
	seedOLTTestDevice(t, db)
	ctx := context.Background()
	ontRepo := postgres.NewONTRepository(db)
	repo := postgres.NewACSRepository(db)

	for i, mac := range []string{"aa:bb:cc:dd:03:01", "aa:bb:cc:dd:03:02"} {
		require.NoError(t, ontRepo.Save(ctx, domainONT.ONT{
			ID: []string{"ont-acs-1", "ont-acs-2"}[i], OLTDeviceID: testOLTDeviceID,
			PONIndex: 1, ONUIndex: i + 1, MACAddress: mac,
			State: domainONT.StateOnline, AdminState: domainONT.AdminEnable, Source: "SNMP",
		}))
	}
	for _, id := range []string{"acs-a", "acs-b"} {
		require.NoError(t, repo.UpsertDevice(ctx, domainONT.ACSDevice{
			ACSDeviceID: id, FirstSeenAt: time.Now().UTC(), LastSeenAt: time.Now().UTC(),
		}))
	}

	ont1 := "ont-acs-1"
	require.NoError(t, repo.UpdateDeviceMatch(ctx, "acs-a", &ont1, string(domainONT.MatchMAC)))
	devA, err := repo.FindDeviceByACSID(ctx, "acs-a")
	require.NoError(t, err)
	assert.Equal(t, &ont1, devA.MatchedONTID)
	assert.Equal(t, string(domainONT.MatchMAC), devA.MatchSource)

	// Device B mengambil ONT yang sama → match device A dilepas.
	require.NoError(t, repo.UpdateDeviceMatch(ctx, "acs-b", &ont1, string(domainONT.MatchSerial)))
	devA, err = repo.FindDeviceByACSID(ctx, "acs-a")
	require.NoError(t, err)
	assert.Nil(t, devA.MatchedONTID)
	assert.Empty(t, devA.MatchSource)
	devB, err := repo.FindDeviceByACSID(ctx, "acs-b")
	require.NoError(t, err)
	assert.Equal(t, &ont1, devB.MatchedONTID)
	assert.Equal(t, string(domainONT.MatchSerial), devB.MatchSource)

	// Pin pindah ke ONT lain: hanya device yang memegang ONT tujuan yang
	// dilepas, device lain tidak tersentuh.
	ont2 := "ont-acs-2"
	require.NoError(t, repo.UpdateDeviceMatch(ctx, "acs-a", &ont2, string(domainONT.MatchManual)))
	devA, err = repo.FindDeviceByACSID(ctx, "acs-a")
	require.NoError(t, err)
	assert.Equal(t, &ont2, devA.MatchedONTID)
	devB, err = repo.FindDeviceByACSID(ctx, "acs-b")
	require.NoError(t, err)
	assert.Equal(t, &ont1, devB.MatchedONTID, "device B tetap memegang ONT-nya")

	// Device B dipindah ke ONT-2 (dipegang A) → match A dilepas.
	require.NoError(t, repo.UpdateDeviceMatch(ctx, "acs-b", &ont2, string(domainONT.MatchMAC)))
	devA, err = repo.FindDeviceByACSID(ctx, "acs-a")
	require.NoError(t, err)
	assert.Nil(t, devA.MatchedONTID, "match lama device A harus dilepas")
	devB, err = repo.FindDeviceByACSID(ctx, "acs-b")
	require.NoError(t, err)
	assert.Equal(t, &ont2, devB.MatchedONTID)

	require.NoError(t, repo.UpdateDeviceMatch(ctx, "acs-b", nil, ""))
	devB, err = repo.FindDeviceByACSID(ctx, "acs-b")
	require.NoError(t, err)
	assert.Nil(t, devB.MatchedONTID)
	assert.Empty(t, devB.MatchSource)

	matched, err := repo.ListDevices(ctx, port.ACSDeviceFilter{Matched: boolPtr(true)})
	require.NoError(t, err)
	assert.Empty(t, matched)
	assert.ErrorIs(t, repo.UpdateDeviceMatch(ctx, "acs-missing", &ont1, ""), domainONT.ErrNotFound)
}

// TestACSRepository_Presets memverifikasi upsert preset by (tenant, name),
// termasuk persistensi enabled=false dan round-trip JSONB.
func TestACSRepository_Presets(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := postgres.NewACSRepository(db)

	configs := json.RawMessage(`[{"path":"ManagementServer.URL","value":"http://acs:7547"}]`)
	preset := domainONT.ACSPreset{
		Name: "polyglot-bootstrap", Precondition: json.RawMessage(`{}`),
		Configurations: configs, Enabled: true,
	}
	require.NoError(t, repo.SavePreset(ctx, preset))

	got, err := repo.GetPreset(ctx, "polyglot-bootstrap")
	require.NoError(t, err)
	assert.Equal(t, "polyglot-bootstrap", got.Name)
	assert.True(t, got.Enabled)
	assert.JSONEq(t, string(configs), string(got.Configurations))
	assert.JSONEq(t, `{}`, string(got.Precondition))
	assert.False(t, got.CreatedAt.IsZero())
	firstID := got.ID
	require.NotEmpty(t, firstID)

	// Upsert nama sama: ID/created_at dipertahankan, enabled=false tersimpan.
	require.NoError(t, repo.SavePreset(ctx, domainONT.ACSPreset{
		Name: "polyglot-bootstrap", Configurations: json.RawMessage(`[]`), Enabled: false,
	}))
	got, err = repo.GetPreset(ctx, "polyglot-bootstrap")
	require.NoError(t, err)
	assert.Equal(t, firstID, got.ID)
	assert.False(t, got.Enabled)
	assert.JSONEq(t, `[]`, string(got.Configurations))
	assert.Nil(t, got.Precondition)

	require.NoError(t, repo.SavePreset(ctx, domainONT.ACSPreset{Name: "alpha", Enabled: true}))
	presets, err := repo.ListPresets(ctx)
	require.NoError(t, err)
	require.Len(t, presets, 2)
	assert.Equal(t, "alpha", presets[0].Name)
	assert.Equal(t, "polyglot-bootstrap", presets[1].Name)

	_, err = repo.GetPreset(ctx, "alpha-missing")
	assert.ErrorIs(t, err, domainONT.ErrNotFound)
}
