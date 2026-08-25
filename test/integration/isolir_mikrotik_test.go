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
