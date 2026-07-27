//go:build integration

// Test di file ini butuh Mikrotik fisik atau CHR (mis. di GNS3) yang benar-benar
// menyala dan bisa dijangkau — sesuai CLAUDE.md §9: "Driver Mikrotik: uji ke
// Mikrotik CHR di GNS3, bukan mock DeviceDriver." Tidak dijalankan oleh
// `go test ./...` biasa (lihat build tag di atas); jalankan secara eksplisit:
//
//	MIKROTIK_TEST_HOST=192.168.88.1 \
//	MIKROTIK_TEST_USER=admin \
//	MIKROTIK_TEST_PASS=secret \
//	go test -tags=integration ./test/integration/... -run TestMikrotikDriver -v
//
// Kalau MIKROTIK_TEST_HOST kosong, setiap test di file ini di-skip (bukan
// gagal) — supaya CI tanpa akses ke device fisik tetap hijau.
package integration

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/domain/provision"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// mikrotikTestTarget builds a device.Target from MIKROTIK_TEST_* env vars,
// or skips the calling test if MIKROTIK_TEST_HOST is not set.
func mikrotikTestTarget(t *testing.T) device.Target {
	t.Helper()

	host := os.Getenv("MIKROTIK_TEST_HOST")
	if host == "" {
		t.Skip("MIKROTIK_TEST_HOST tidak di-set — skip integration test Mikrotik fisik")
	}

	port := 8728
	if raw := os.Getenv("MIKROTIK_TEST_PORT"); raw != "" {
		p, err := strconv.Atoi(raw)
		require.NoError(t, err, "MIKROTIK_TEST_PORT harus angka")
		port = p
	}

	return device.Target{
		Host:     host,
		Port:     port,
		Username: os.Getenv("MIKROTIK_TEST_USER"),
		Password: os.Getenv("MIKROTIK_TEST_PASS"),
		Timeout:  10 * time.Second,
	}
}

// TestMikrotikDriver_Execute membuktikan koneksi persistent + Execute
// benar-benar bisa menjalankan command one-shot ke device fisik dan
// menerima hasil terstruktur (Rows).
func TestMikrotikDriver_Execute(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik fisik")
	defer func() { assert.NoError(t, drv.Close()) }()

	result, err := drv.Execute(ctx, command.Command{Raw: "/system/resource/print"})
	require.NoError(t, err)
	require.Len(t, result.Rows, 1, "/system/resource/print harus mengembalikan tepat satu baris")
	_, hasUptime := result.Rows[0]["uptime"]
	assert.True(t, hasUptime, "baris hasil harus punya field 'uptime'")
}

// TestMikrotikDriver_Stream membuktikan Stream mengalirkan data /ping tanpa
// polling — data diterima lewat channel sampai device menyelesaikan count
// yang diminta.
func TestMikrotikDriver_Stream(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err)
	defer func() { assert.NoError(t, drv.Close()) }()

	sd, ok := port.DeviceDriver(drv).(port.StreamingDeviceDriver)
	require.True(t, ok, "mikrotik.Driver harus implement port.StreamingDeviceDriver")

	handle, err := sd.Stream(ctx, command.Command{
		Raw:  "/ping",
		Args: map[string]string{"address": target.Host, "count": "5"},
	})
	require.NoError(t, err)
	defer func() { assert.NoError(t, handle.Cancel()) }()

	received := 0
	timeout := time.After(15 * time.Second)
loop:
	for {
		select {
		case _, ok := <-handle.Chan():
			if !ok {
				break loop
			}
			received++
		case <-timeout:
			t.Fatal("timeout menunggu hasil ping streaming")
		}
	}
	assert.Greater(t, received, 0, "harus menerima minimal satu balasan ping")
	assert.NoError(t, handle.Err())
}

// TestMikrotikDriver_ExecuteWhileStreaming adalah test INTI dari requirement
// koneksi ganda: membuktikan Execute() TIDAK terblokir selagi Stream()
// sedang aktif mengalir — dua koneksi persisten yang benar-benar independen.
func TestMikrotikDriver_ExecuteWhileStreaming(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err)
	defer func() { assert.NoError(t, drv.Close()) }()

	sd, ok := port.DeviceDriver(drv).(port.StreamingDeviceDriver)
	require.True(t, ok)

	// Stream ping dengan count tinggi supaya masih mengalir saat Execute diuji.
	handle, err := sd.Stream(ctx, command.Command{
		Raw:  "/ping",
		Args: map[string]string{"address": target.Host, "count": "30"},
	})
	require.NoError(t, err)
	defer func() { assert.NoError(t, handle.Cancel()) }()

	// Beri waktu sebentar supaya stream benar-benar mulai mengalir.
	time.Sleep(500 * time.Millisecond)

	// Selagi stream aktif, Execute() harus tetap cepat selesai — ini bukti
	// utama bahwa exec dan stream tidak berbagi satu koneksi yang sama.
	execCtx, execCancel := context.WithTimeout(ctx, 5*time.Second)
	defer execCancel()

	start := time.Now()
	result, err := drv.Execute(execCtx, command.Command{Raw: "/system/resource/print"})
	elapsed := time.Since(start)

	require.NoError(t, err, "Execute tidak boleh gagal/timeout selagi Stream berjalan")
	assert.Less(t, elapsed, 5*time.Second, "Execute tidak boleh terblokir oleh Stream yang aktif")
	assert.Len(t, result.Rows, 1)

	// Stream tetap harus terus mengalir setelah Execute selesai — buktikan
	// arah sebaliknya juga tidak saling mengganggu.
	select {
	case _, ok := <-handle.Chan():
		assert.True(t, ok, "stream seharusnya belum berhenti hanya karena ada Execute di tengah jalan")
	case <-time.After(10 * time.Second):
		t.Fatal("stream berhenti mengalir setelah Execute — indikasi koneksi bentrok")
	}
}

// TestMikrotikDriver_ListPPPSecrets membuktikan jalur provisioning read-only:
// driver mikrotik implement port.ProvisioningDriver, operasi abstrak
// provision.ListPPPSecrets diterjemahkan jadi /ppp/secret/print, dijalankan
// lewat gate kebijakan (network.ExecuteCommand → Classify/Decide/Execute), dan
// hasilnya di-dump mentah. Tujuan langkah 1: MELIHAT field asli yang device
// kembalikan sebagai dasar desain write path (langkah 2) — bukan menebak.
func TestMikrotikDriver_ListPPPSecrets(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik fisik")
	defer func() { assert.NoError(t, drv.Close()) }()

	pd, ok := port.DeviceDriver(drv).(port.ProvisioningDriver)
	require.True(t, ok, "mikrotik.Driver harus implement port.ProvisioningDriver")

	// Subtest 1: tanpa Fields → RouterOS balikin SEMUA field, apa adanya.
	t.Run("semua field (raw)", func(t *testing.T) {
		cmds, err := pd.TranslateProvision(provision.ListPPPSecrets{})
		require.NoError(t, err)
		require.NotEmpty(t, cmds, "TranslateProvision harus menghasilkan minimal satu command")

		for _, cmd := range cmds {
			res, err := network.ExecuteCommand(ctx, drv, cmd)
			require.NoError(t, err, "read-only /ppp/secret/print harus lolos gate dan sukses")
			t.Logf("%s → %d baris", cmd.Raw, len(res.Rows))
			for i, row := range res.Rows {
				t.Logf("  secret[%d]: %v", i, row)
			}
		}
	})

	// Subtest 2: dengan Fields → RouterOS hanya balikin kolom yang diminta
	// (.proplist). Membuktikan proyeksi field bekerja: field yang tidak diminta
	// (mis. password, caller-id) TIDAK ikut terbawa.
	t.Run("proyeksi field (.proplist)", func(t *testing.T) {
		cmds, err := pd.TranslateProvision(provision.ListPPPSecrets{
			Fields: []string{"name", "profile", "service"},
		})
		require.NoError(t, err)
		require.Len(t, cmds, 1)
		assert.Equal(t, "name,profile,service", cmds[0].Args[".proplist"])

		res, err := network.ExecuteCommand(ctx, drv, cmds[0])
		require.NoError(t, err)
		require.NotEmpty(t, res.Rows, "harus ada minimal satu secret untuk diperiksa")
		for i, row := range res.Rows {
			t.Logf("  secret[%d] (terproyeksi): %v", i, row)
			assert.Contains(t, row, "name", "field yang diminta harus ada")
			assert.NotContains(t, row, "caller-id", "field yang TIDAK diminta tidak boleh terbawa")
			assert.NotContains(t, row, "password", "proyeksi menekan password dari hasil print")
		}
	})
}

// TestMikrotikDriver_ListPPPProfiles menembak /ppp/profile/print lewat operasi
// abstrak provision.ListPPPProfiles — pola yang IDENTIK dengan ListPPPSecrets,
// membuktikan mekanisme provisioning read menggeneralisasi ke operasi lain
// hanya dengan menambah satu struct + satu case di translateProvision. Field
// mentah yang di-dump di sini (mis. rate-limit) jadi dasar desain
// CreatePPPProfile di langkah 2.
func TestMikrotikDriver_ListPPPProfiles(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik fisik")
	defer func() { assert.NoError(t, drv.Close()) }()

	pd, ok := port.DeviceDriver(drv).(port.ProvisioningDriver)
	require.True(t, ok, "mikrotik.Driver harus implement port.ProvisioningDriver")

	cmds, err := pd.TranslateProvision(provision.ListPPPProfiles{})
	require.NoError(t, err)
	require.NotEmpty(t, cmds)

	res, err := network.ExecuteCommand(ctx, drv, cmds[0])
	require.NoError(t, err, "read-only /ppp/profile/print harus lolos gate dan sukses")
	t.Logf("%s → %d baris", cmds[0].Raw, len(res.Rows))
	for i, row := range res.Rows {
		t.Logf("  profile[%d]: %v", i, row)
	}
}

// TestMikrotikDriver_ListActivePPP menembak /ppp/active/print — sesi PPPoE yang
// sedang online SEKARANG (bukan secret tersimpan). Berguna sebelum perubahan
// yang akan memutus pelanggan (langkah 2: ChangeProfile memutus sesi aktif).
func TestMikrotikDriver_ListActivePPP(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik fisik")
	defer func() { assert.NoError(t, drv.Close()) }()

	pd, ok := port.DeviceDriver(drv).(port.ProvisioningDriver)
	require.True(t, ok, "mikrotik.Driver harus implement port.ProvisioningDriver")

	cmds, err := pd.TranslateProvision(provision.ListActivePPP{})
	require.NoError(t, err)
	require.NotEmpty(t, cmds)

	res, err := network.ExecuteCommand(ctx, drv, cmds[0])
	require.NoError(t, err, "read-only /ppp/active/print harus lolos gate dan sukses")
	t.Logf("%s → %d sesi aktif", cmds[0].Raw, len(res.Rows))
	for i, row := range res.Rows {
		t.Logf("  active[%d]: %v", i, row)
	}
}

// TestMikrotikDriver_CreatePPPSecret membuktikan jalur provisioning WRITE
// end-to-end terhadap device asli, dengan dua sifat penting:
//   - Gate kebijakan: /ppp/secret/add diklasifikasi destruktif, jadi
//     network.ExecuteCommand (jalur auto-approve) MENOLAKnya dengan
//     ErrApprovalRequired — write tidak pernah lewat diam-diam.
//   - Jalur ter-approve: network.ExecuteCommandPreApproved (mensimulasikan HITL
//     sudah menyetujui, spt adapter MCP) benar-benar membuat secret. Test lalu
//     memverifikasinya lewat ListPPPSecrets, dan membersihkannya kembali supaya
//     device kembali ke keadaan semula.
func TestMikrotikDriver_CreatePPPSecret(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik fisik")
	defer func() { assert.NoError(t, drv.Close()) }()

	pd, ok := port.DeviceDriver(drv).(port.ProvisioningDriver)
	require.True(t, ok, "mikrotik.Driver harus implement port.ProvisioningDriver")

	const testName = "polyglot-inttest"

	// Jaga-jaga sisa run sebelumnya yang gagal bersih-bersih, lalu pastikan
	// dibersihkan lagi setelah test apa pun hasilnya.
	removePPPSecretByName(ctx, t, drv, testName)
	defer removePPPSecretByName(ctx, t, drv, testName)

	cmds, err := pd.TranslateProvision(provision.CreatePPPSecret{
		Name:     testName,
		Password: "inttest-pw",
		Profile:  "default",
		Service:  "pppoe",
		Comment:  "dibuat oleh integration test polyglot",
	})
	require.NoError(t, err)
	require.Len(t, cmds, 1)
	require.Equal(t, "/ppp/secret/add", cmds[0].Raw)

	// 1. Jalur auto-approve HARUS menolak — /ppp/secret/add itu destruktif.
	_, err = network.ExecuteCommand(ctx, drv, cmds[0])
	require.ErrorIs(t, err, network.ErrApprovalRequired,
		"write tanpa approval harus ditolak gate kebijakan")

	// 2. Jalur ter-approve HARUS membuat secretnya.
	_, err = network.ExecuteCommandPreApproved(ctx, drv, cmds[0])
	require.NoError(t, err, "create secret via jalur ter-approve harus sukses")

	// 3. Verifikasi lewat operasi read: secret baru benar-benar ada di device.
	listCmds, err := pd.TranslateProvision(provision.ListPPPSecrets{
		Fields: []string{"name", "profile", "service", "comment"},
	})
	require.NoError(t, err)
	res, err := network.ExecuteCommand(ctx, drv, listCmds[0])
	require.NoError(t, err)

	var found map[string]string
	for _, row := range res.Rows {
		if row["name"] == testName {
			found = row
			break
		}
	}
	require.NotNil(t, found, "secret %q harus muncul di /ppp/secret/print setelah dibuat", testName)
	t.Logf("secret terbuat & terverifikasi: %v", found)
	assert.Equal(t, "default", found["profile"])
	assert.Equal(t, "pppoe", found["service"])
}

// removePPPSecretByName menghapus secret PPPoE bernama name jika ada.
// Dipakai untuk setup/teardown integration test — best-effort, langsung lewat
// drv.Execute (bukan gate kebijakan) karena ini housekeeping test, bukan bagian
// perilaku yang sedang diuji.
func removePPPSecretByName(ctx context.Context, t *testing.T, drv *mikrotik.Driver, name string) {
	t.Helper()
	removeByName(ctx, t, drv, "/ppp/secret", name)
}

// TestMikrotikDriver_CreatePPPProfile membuktikan write op kedua — pola identik
// dengan CreatePPPSecret. /ppp/profile/add otomatis destruktif berkat Classify
// fail-safe (ADR 0006), tanpa perlu didaftarkan. Field rate-limit yang di-set
// diverifikasi kembali lewat ListPPPProfiles.
func TestMikrotikDriver_CreatePPPProfile(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik fisik")
	defer func() { assert.NoError(t, drv.Close()) }()

	pd, ok := port.DeviceDriver(drv).(port.ProvisioningDriver)
	require.True(t, ok, "mikrotik.Driver harus implement port.ProvisioningDriver")

	const profileName = "polyglot-inttest-profile"

	removeByName(ctx, t, drv, "/ppp/profile", profileName)
	defer removeByName(ctx, t, drv, "/ppp/profile", profileName)

	cmds, err := pd.TranslateProvision(provision.CreatePPPProfile{
		Name:      profileName,
		RateLimit: "5M/5M",
		Comment:   "dibuat oleh integration test polyglot",
	})
	require.NoError(t, err)
	require.Equal(t, "/ppp/profile/add", cmds[0].Raw)

	// Gate: destruktif → auto-approve ditolak; ter-approve membuatnya.
	_, err = network.ExecuteCommand(ctx, drv, cmds[0])
	require.ErrorIs(t, err, network.ErrApprovalRequired)
	_, err = network.ExecuteCommandPreApproved(ctx, drv, cmds[0])
	require.NoError(t, err, "create profile via jalur ter-approve harus sukses")

	// Verifikasi: profil baru ada dengan rate-limit yang diminta.
	listCmds, err := pd.TranslateProvision(provision.ListPPPProfiles{
		Fields: []string{"name", "rate-limit"},
	})
	require.NoError(t, err)
	res, err := network.ExecuteCommand(ctx, drv, listCmds[0])
	require.NoError(t, err)

	var found map[string]string
	for _, row := range res.Rows {
		if row["name"] == profileName {
			found = row
			break
		}
	}
	require.NotNil(t, found, "profil %q harus muncul setelah dibuat", profileName)
	t.Logf("profil terbuat & terverifikasi: %v", found)
	assert.Equal(t, "5M/5M", found["rate-limit"])
}

// TestMikrotikDriver_ChangeProfile membuktikan write op ketiga — sebuah SEKUENS,
// bukan satu command. ChangeProfile: (1) /ppp/secret/set memindah profil secret,
// (2) /ppp/active/remove memutus sesi online supaya profil baru berlaku saat
// dial ulang. Karena secret yang dibuat test ini TIDAK online, langkah kill
// menembak sesi yang tidak ada — dan itu harus jadi no-op sukses (idempotent di
// Driver.Execute), bukan gagal. Verifikasi: profil secret benar-benar berpindah.
func TestMikrotikDriver_ChangeProfile(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek ke Mikrotik fisik")
	defer func() { assert.NoError(t, drv.Close()) }()

	pd, ok := port.DeviceDriver(drv).(port.ProvisioningDriver)
	require.True(t, ok, "mikrotik.Driver harus implement port.ProvisioningDriver")

	const testName = "polyglot-inttest"

	// Siapkan secret uji dengan profil awal "default", lalu bersihkan setelahnya.
	removePPPSecretByName(ctx, t, drv, testName)
	defer removePPPSecretByName(ctx, t, drv, testName)

	createCmds, err := pd.TranslateProvision(provision.CreatePPPSecret{
		Name:     testName,
		Password: "inttest-pw",
		Profile:  "default",
		Service:  "pppoe",
		Comment:  "dibuat oleh integration test polyglot",
	})
	require.NoError(t, err)
	_, err = network.ExecuteCommandPreApproved(ctx, drv, createCmds[0])
	require.NoError(t, err, "setup: create secret harus sukses")

	// Ubah profil ke default-encryption (profil bawaan RouterOS yang pasti ada).
	cmds, err := pd.TranslateProvision(provision.ChangeProfile{
		Username: testName,
		Profile:  "default-encryption",
	})
	require.NoError(t, err)
	require.Len(t, cmds, 2, "change profile harus sekuens set-lalu-kill")
	require.Equal(t, "/ppp/secret/set", cmds[0].Raw)
	require.Equal(t, "/ppp/active/remove", cmds[1].Raw)

	// Gate: kedua command destruktif → jalur auto-approve harus menolak.
	for _, cmd := range cmds {
		_, err := network.ExecuteCommand(ctx, drv, cmd)
		require.ErrorIs(t, err, network.ErrApprovalRequired,
			"tiap langkah sekuens destruktif harus ditolak tanpa approval: %s", cmd.Raw)
	}

	// Jalur ter-approve: jalankan sekuens berurutan, berhenti di error pertama
	// (meniru cara Sync Engine mengeksekusi sekuens per-command). Langkah kill
	// menembak sesi yang tidak ada (secret uji offline) → harus no-op sukses.
	for _, cmd := range cmds {
		_, err := network.ExecuteCommandPreApproved(ctx, drv, cmd)
		require.NoError(t, err, "langkah sekuens ter-approve harus sukses (kill sesi absen = no-op): %s", cmd.Raw)
	}

	// Verifikasi: profil secret benar-benar berpindah ke default-encryption.
	listCmds, err := pd.TranslateProvision(provision.ListPPPSecrets{
		Fields: []string{"name", "profile"},
	})
	require.NoError(t, err)
	res, err := network.ExecuteCommand(ctx, drv, listCmds[0])
	require.NoError(t, err)

	var found map[string]string
	for _, row := range res.Rows {
		if row["name"] == testName {
			found = row
			break
		}
	}
	require.NotNil(t, found, "secret %q harus tetap ada setelah change profile", testName)
	t.Logf("secret setelah change profile: %v", found)
	assert.Equal(t, "default-encryption", found["profile"], "profil harus berpindah ke default-encryption")
}

// removeByName menghapus item RouterOS bernama name di menu menuPath (mis.
// "/ppp/secret", "/ppp/profile") jika ada. Best-effort housekeeping test,
// langsung lewat drv.Execute (bukan gate kebijakan).
func removeByName(ctx context.Context, t *testing.T, drv *mikrotik.Driver, menuPath, name string) {
	t.Helper()

	res, err := drv.Execute(ctx, command.Command{
		Raw:  menuPath + "/print",
		Args: map[string]string{".proplist": ".id,name"},
	})
	require.NoError(t, err)

	for _, row := range res.Rows {
		if row["name"] != name {
			continue
		}
		_, err := drv.Execute(ctx, command.Command{
			Raw:  menuPath + "/remove",
			Args: map[string]string{".id": row[".id"]},
		})
		require.NoError(t, err, "gagal menghapus %s sisa %q (.id=%s)", menuPath, name, row[".id"])
	}
}
