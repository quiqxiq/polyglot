//go:build integration

// Integration test infrastruktur isolir terhadap Mikrotik FISIK.
// Membuktikan: NAT add/print/remove, EnsureIsolirInfrastructure idempotent,
// dan siklus profil PPP secret normal ↔ isolir — lalu membersihkan semua
// artifact yang dibuat.
//
// Jalankan:
//
//	set -a; source .env; set +a
//	go test -tags=integration ./test/integration/ -run TestMikrotikIsolir -v
package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

const (
	isolirTestUsername = "polyglot-it-sub"
	isolirTestProfile  = "polyglot-it-plan"
)

func mikrotikGateway(t *testing.T) (*mikrotik.Gateway, port.DeviceDriver) {
	t.Helper()
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik fisik")
	t.Cleanup(func() { _ = drv.Close() })

	exec := func(ctx context.Context, d port.DeviceDriver, cmd command.Command) (command.Result, error) {
		return d.Execute(ctx, cmd)
	}
	return mikrotik.NewGateway(exec), drv
}

// TestMikrotikIsolir_NATRoundtrip membuktikan command builder + parser NAT
// benar-benar bekerja pada RouterOS nyata: tambah rule dst-nat, lihat
// muncul di print, lalu hapus lagi.
func TestMikrotikIsolir_NATRoundtrip(t *testing.T) {
	gw, driver := mikrotikGateway(t)
	ctx := context.Background()

	comment := "ISOLIR_REDIRECT_IT " + strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	_, err := gw.AddNATRule(ctx, driver, port.FirewallNATRuleParams{
		Chain:     "dstnat",
		Action:    "redirect",
		SrcAddress: "172.16.99.0/24",
		Protocol:  "tcp",
		DstPort:   "8088",
		ToPorts:   "9999",
		Comment:   comment,
	})
	require.NoError(t, err, "add nat rule harus sukses")

	rules, err := gw.ListNATRules(ctx, driver)
	require.NoError(t, err)

	var found *port.FirewallNATRule
	for i := range rules {
		if rules[i].Comment == comment {
			found = &rules[i]
			break
		}
	}
	require.NotNil(t, found, "rule baru harus terlihat di print")
	assert.Equal(t, "dstnat", found.Chain)
	assert.Equal(t, "redirect", found.Action)
	assert.Equal(t, "8088", found.DstPort)

	_, err = gw.RemoveNATRule(ctx, driver, found.RosID)
	require.NoError(t, err)

	rules, err = gw.ListNATRules(ctx, driver)
	require.NoError(t, err)
	for _, r := range rules {
		assert.NotEqual(t, comment, r.Comment, "rule harus sudah terhapus")
	}
}

// TestMikrotikIsolir_EnsureInfrastructureIdempotent memanggil
// EnsureIsolirInfrastructure dua kali — panggilan kedua tidak boleh membuat
// duplikat pool/profil/rule.
func TestMikrotikIsolir_EnsureInfrastructureIdempotent(t *testing.T) {
	gw, driver := mikrotikGateway(t)
	ctx := context.Background()

	cfg := port.IsolirConfig{
		ProfileName:    "polyglot-it-isolir",
		PoolName:       "pool-polyglot-it",
		PoolRange:      "172.16.98.10-172.16.98.20",
		PortalIP:       "",
		PortalHTTPPort: "8099",
		RedirectPorts:  "8091,8092", // port unik agar tak menabrak rule produksi
	}

	first, err := gw.EnsureIsolirInfrastructure(ctx, driver, cfg)
	require.NoError(t, err, "pemanggilan pertama harus sukses")

	second, err := gw.EnsureIsolirInfrastructure(ctx, driver, cfg)
	require.NoError(t, err, "pemanggilan kedua (idempotent) harus sukses")

	assert.True(t, second.PoolExisted, "pool harus sudah ada dari pemanggilan pertama")
	assert.True(t, second.ProfileExisted, "profil harus sudah ada dari pemanggilan pertama")
	assert.Len(t, second.NATRuleIDs, 2, "dua redirect port → dua rule")
	assert.Empty(t, second.CreatedNATIDs, "tidak boleh ada rule baru dibuat ulang")
	assert.Len(t, first.CreatedNATIDs, 2, "pemanggilan pertama membuat dua rule")

	profiles, err := gw.ListProfiles(ctx, driver, cfg.ProfileName)
	require.NoError(t, err)
	require.Len(t, profiles, 1, "tepat satu profile isolir")
	assert.Equal(t, "512k/512k", profiles[0].RateLimit)

	pools, err := func() ([]mikrotik.IPPool, error) {
		res, err := driver.Execute(ctx, mikrotik.NewPrintIPPoolsCommand(cfg.PoolName))
		if err != nil {
			return nil, err
		}
		return mikrotik.ParseIPPools(res), nil
	}()
	require.NoError(t, err)
	assert.NotEmpty(t, pools, "pool isolir harus ada")

	require.NoError(t, gw.RemoveIsolirInfrastructure(ctx, driver, cfg), "teardown harus bersih")

	rules, err := gw.ListNATRules(ctx, driver)
	require.NoError(t, err)
	for _, r := range rules {
		for _, id := range second.NATRuleIDs {
			assert.NotEqual(t, id, r.RosID, "rule isolir harus terhapus setelah teardown")
		}
	}
}

// TestMikrotikIsolir_PPPSecretCycle menyimulasikan siklus provisioning +
// isolir pada satu PPP secret sungguhan, lalu menghapusnya.
func TestMikrotikIsolir_PPPSecretCycle(t *testing.T) {
	gw, driver := mikrotikGateway(t)
	ctx := context.Background()

	cfg := port.IsolirConfig{
		ProfileName:    "polyglot-it-isolir",
		PoolName:       "pool-polyglot-it",
		PoolRange:      "172.16.98.10-172.16.98.20",
		PortalHTTPPort: "8099",
		RedirectPorts:  "8091",
	}
	require.NoError(t, cleanupPPPSecret(gw, driver, cfg, ctx))
	t.Cleanup(func() { _ = cleanupPPPSecret(gw, driver, cfg, context.Background()) })
	t.Cleanup(func() { _ = gw.RemoveIsolirInfrastructure(ctx, driver, cfg) })

	_, err := gw.EnsureIsolirInfrastructure(ctx, driver, cfg)
	require.NoError(t, err)

	// 1. Provisioning: buat secret dengan profile plan sementara.
	planProfile := "polyglot-it-plan"
	if _, err := gw.AddProfile(ctx, driver, port.PPPProfileParams{
		Name: planProfile, RateLimit: "10M/10M", Comment: "IT_TEST",
	}); err != nil {
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		profiles, ferr := gw.ListProfiles(context.Background(), driver, planProfile)
		if ferr == nil {
			for _, pr := range profiles {
				_, _ = gw.RemoveProfile(context.Background(), driver, pr.RosID)
			}
		}
	})

	_, err = gw.AddSecret(ctx, driver, port.PPPoESecretParams{
		Name: isolirTestUsername, Password: "it-secret-pass",
		Profile: planProfile, Service: "pppoe", Comment: "POLYGLOT_IT",
	})
	require.NoError(t, err, "create secret harus sukses")

	secrets, err := gw.ListSecrets(ctx, driver, isolirTestUsername)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	secretID := secrets[0].RosID

	// 2. Isolate: switch profile ke isolir.
	_, err = gw.UpdateSecret(ctx, driver, secretID, port.PPPoESecretParams{Profile: cfg.ProfileName})
	require.NoError(t, err)
	secrets, err = gw.ListSecrets(ctx, driver, isolirTestUsername)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	assert.Equal(t, cfg.ProfileName, secrets[0].Profile, "profile secret harus jadi isolir")

	// 3. Unisolate: kembalikan profile plan.
	_, err = gw.UpdateSecret(ctx, driver, secretID, port.PPPoESecretParams{Profile: planProfile})
	require.NoError(t, err)
	secrets, err = gw.ListSecrets(ctx, driver, isolirTestUsername)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	assert.Equal(t, planProfile, secrets[0].Profile, "profile secret dikembalikan")

	// 4. Deprovision: hapus secret.
	_, err = gw.RemoveSecret(ctx, driver, secretID)
	require.NoError(t, err)
	secrets, err = gw.ListSecrets(ctx, driver, isolirTestUsername)
	require.NoError(t, err)
	assert.Empty(t, secrets, "secret harus terhapus")
}

func cleanupPPPSecret(gw *mikrotik.Gateway, driver port.DeviceDriver, cfg port.IsolirConfig, ctx context.Context) error {
	secrets, err := gw.ListSecrets(ctx, driver, isolirTestUsername)
	if err != nil {
		return err
	}
	for _, s := range secrets {
		if _, err := gw.RemoveSecret(ctx, driver, s.RosID); err != nil {
			return err
		}
	}
	return nil
}

// stubSettings menyediakan nilai isolir.* untuk test endpoint tanpa
// menyentuh Postgres (hanya GetValue yang dipakai usecase).
type stubSettings struct {
	port.SettingRepository
	values map[string]string
}

func (s *stubSettings) GetValue(ctx context.Context, key, fallback string) string {
	if v, ok := s.values[key]; ok && v != "" {
		return v
	}
	return fallback
}

// TestMikrotikIsolir_EndpointFlow menjalankan alur endpoint isolir secara
// penuh terhadap device fisik: Setup → Status (hadir semua) → Setup ulang
// (idempotent) → Remove → Status (kosong). Nama pool/profil/port khusus
// test agar tidak menyentuh infrastruktur isolir produksi di router.
func TestMikrotikIsolir_EndpointFlow(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik fisik")
	t.Cleanup(func() { _ = drv.Close() })

	exec := func(ctx context.Context, d port.DeviceDriver, cmd command.Command) (command.Result, error) {
		return d.Execute(ctx, cmd)
	}
	gw := mikrotik.NewGateway(exec)

	settings := &stubSettings{values: map[string]string{
		"isolir.profile_name":     "polyglot-it-ep-iso",
		"isolir.pool_name":        "pool-polyglot-it-ep",
		"isolir.pool_range":       "172.16.96.10-172.16.96.20",
		"isolir.portal_ip":        "192.0.2.10",
		"isolir.portal_http_port": "8099",
		"isolir.redirect_ports":   "8095,8096",
	}}
	uc := networkUC.NewManageIsolationUseCase(
		settings, gw,
		func(ctx context.Context, deviceID string) (port.DeviceDriver, error) { return drv, nil },
	)
	t.Cleanup(func() {
		cctx, cc := context.WithTimeout(context.Background(), 15*time.Second)
		defer cc()
		_ = uc.Remove(cctx, "integration-test")
	})

	res, cfg, err := uc.Setup(ctx, "integration-test", networkUC.IsolirConfigOverride{})
	require.NoError(t, err, "Setup harus sukses")
	assert.False(t, res.PoolExisted)
	assert.False(t, res.ProfileExisted)
	assert.Len(t, res.CreatedNATIDs, 2)
	assert.Equal(t, "polyglot-it-ep-iso", cfg.ProfileName)

	ins, cfg, warnings, err := uc.Status(ctx, "integration-test")
	require.NoError(t, err, "Status harus sukses")
	assert.True(t, ins.PoolExists, "pool harus terdeteksi")
	assert.Equal(t, "172.16.96.10-172.16.96.20", ins.PoolRanges)
	assert.True(t, ins.ProfileExists, "profile isolir harus terdeteksi")
	assert.Equal(t, "512k/512k", ins.ProfileRateLimit)
	assert.Equal(t, cfg.PoolName, ins.ProfileRemoteAddress, "profile harus menunjuk pool isolir")
	require.Len(t, ins.NATRules, 2)
	for _, r := range ins.NATRules {
		assert.True(t, r.Exists, "rule port %s harus ada", r.Port)
	}
	assert.Empty(t, warnings, "tidak boleh ada warning saat semua lengkap")

	res2, _, err := uc.Setup(ctx, "integration-test", networkUC.IsolirConfigOverride{})
	require.NoError(t, err, "Setup ulang (idempotent) harus sukses")
	assert.True(t, res2.PoolExisted)
	assert.True(t, res2.ProfileExisted)
	assert.Empty(t, res2.CreatedNATIDs, "tidak boleh ada rule baru dibuat ulang")

	require.NoError(t, uc.Remove(ctx, "integration-test"), "Remove harus sukses")

	ins, _, warnings, err = uc.Status(ctx, "integration-test")
	require.NoError(t, err)
	assert.False(t, ins.PoolExists)
	assert.False(t, ins.ProfileExists)
	for _, r := range ins.NATRules {
		assert.False(t, r.Exists)
	}
	assert.NotEmpty(t, warnings, "warnings harus muncul saat infrastruktur kosong")
}
