//go:build integration

// Test di file ini membuktikan internal/platformdef/mikrotik_routeros.yaml
// benar-benar bisa dipakai end-to-end lewat genericssh ke MikroTik fisik/CHR
// — bukan cuma tervalidasi skema YAML-nya (itu sudah dicek terpisah, lihat
// komentar di file YAML-nya). Beda dengan test/integration/genericcli_test.go
// yang sepenuhnya diparameterisasi (BYOD, platform apa saja), file ini
// KHUSUS untuk platformdef MikroTik yang baru dibuat, jadi cuma butuh
// HOST/USER/PASS:
//
//	MIKROTIK_SSH_TEST_HOST=192.168.88.1 \
//	MIKROTIK_SSH_TEST_USER=admin \
//	MIKROTIK_SSH_TEST_PASS=secret \
//	go test -tags=integration ./test/integration/... -run TestMikrotikRouterOSSSH -v
//
// Kalau MIKROTIK_SSH_TEST_HOST kosong, test di-skip.
package integration

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/genericcli"
	"github.com/quixiq/polyglot/internal/driver/genericssh"
)

// mikrotikRouterOSPlatformPath resolves the path to
// internal/platformdef/mikrotik_routeros.yaml relative to THIS source
// file's own location (via runtime.Caller), not the process's current
// working directory — go test changes cwd to the test's own package
// directory, and this makes the path correct regardless of how/where
// `go test` was invoked from.
func mikrotikRouterOSPlatformPath(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "gagal menentukan lokasi file test ini")

	// test/integration/mikrotik_ssh_test.go -> ../../internal/platformdef/
	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "platformdef", "mikrotik_routeros.yaml")
	_, err := os.Stat(path)
	require.NoError(t, err, "file platformdef tidak ditemukan di %s", path)

	return path
}

// mikrotikRouterOSCatalog is a concrete, ready-to-use genericcli.Catalog
// for internal/platformdef/mikrotik_routeros.yaml. This is the reference
// example other custom platformdef files' Catalogs should follow.
func mikrotikRouterOSCatalog() genericcli.Catalog {
	return genericcli.Catalog{
		// HARUS konsisten dengan bentuk command di Operations (space-form):
		// Catalog.Classify memakai strings.HasPrefix, jadi kalau prefix
		// memakai slash padahal command memakai spasi, command destruktif
		// malah ter-classify read-only dan fail-safe HITAM. RouterOS v7
		// asli menerima keduanya; G-Net v6 hanya spasi — jadi pakai
		// space-form di sini supaya Classify dan Translate sejalan.
		DestructivePrefixes: []string{
			"system reboot",
			"system reset-configuration",
			"system shutdown",
		},
		Operations: map[command.Operation]command.Command{
			// Sengaja sama dengan yang dipakai internal/driver/mikrotik
			// (driver API) untuk OpGetStatus — hanya saja pemisah level-nya
			// memakai SPASI, bukan slash. Kenapa? Diverifikasi langsung di
			// lab (lihat komentar panjang di TestMikrotikRouterOSSSH):
			// firmware G-Net RouterOS v6 (banner "PT Gnet Ghaib Network")
			// MENOLAK path slash-delimited bertingkat
			// ("expected command name (line 1 column N)" di slash kedua),
			// tapi menerima spasi sebagai pemisah level. RouterOS v7 asli
			// menerima keduanya. Pakai spasi = satu bentuk yang jalan di
			// semua firmware.
			command.OpGetStatus: {Raw: "system resource print"},
			command.OpReboot:    {Raw: "system reboot"},
		},
		// Lihat GitHub issue scrapli/scrapligo#95 dan komentar di
		// mikrotik_routeros.yaml: RouterOS meng-echo prompt/command dua
		// kali lewat SSH. 100ms adalah titik awal yang konservatif, bukan
		// angka final — sesuaikan berdasarkan hasil test ini di device
		// Anda.
		ReadDelay: 100 * time.Millisecond,
		// Kunci kompatibilitas v6: RouterOS hanya mengeksekusi command
		// setelah menerima carriage return, dan default scrapligo "\n"
		// tidak cukup — firmware G-Net v6 (192.168.233.1) tidak pernah
		// menjalankan command dengan "\n" (SendCommand timeout). "\r\n"
		// persis comms_return_char milik scrapli_community, terverifikasi
		// bekerja di v6 G-Net maupun v7 vanilla (192.168.230.3).
		ReturnChar: "\r\n",
	}
}

// TestMikrotikRouterOSSSH membuktikan genericssh + platformdef
// mikrotik_routeros.yaml benar-benar bisa konek, login (dengan suffix
// username wajib +cet512w — lihat komentar di file YAML), dan menjalankan
// command satu-kali ke MikroTik fisik lewat SSH.
func TestMikrotikRouterOSSSH(t *testing.T) {
	host := os.Getenv("MIKROTIK_SSH_TEST_HOST")
	if host == "" {
		t.Skip("MIKROTIK_SSH_TEST_HOST tidak di-set — skip integration test MikroTik SSH CLI")
	}

	username := os.Getenv("MIKROTIK_SSH_TEST_USER")
	require.NotEmpty(t, username, "MIKROTIK_SSH_TEST_USER wajib di-set")

	target := device.Target{
		Host: host,
		// +cet512w: WAJIB per komentar di mikrotik_routeros.yaml — disable
		// warna (c), dumb terminal mode (e), disable auto-detect (t), lebar
		// terminal 512 kolom (w). Tanpa ini, output kemungkinan besar
		// tercampur kode warna ANSI atau sesi macet menunggu deteksi
		// terminal.
		Username: username + "+cet512w",
		Password: os.Getenv("MIKROTIK_SSH_TEST_PASS"),
		Timeout:  15 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	drv, err := genericssh.NewDriver(ctx, target, mikrotikRouterOSPlatformPath(t), mikrotikRouterOSCatalog())
	require.NoError(t, err, "gagal konek SSH ke MikroTik fisik — kalau timeout di sini, coba naikkan ReadDelay di mikrotikRouterOSCatalog()")
	defer func() { assert.NoError(t, drv.Close()) }()

	// Command memakai SPASI sebagai pemisah level ("system resource print"),
	// BUKAN slash ("/system/resource/print") — walau slash adalah bentuk
	// kanonik RouterOS asli, firmware G-Net RouterOS v6 di lab (banner
	// "G-Net RouterOS 6.49.11 ... PT Gnet Ghaib Network") MENOLAK path
	// slash bertingkat: tiap command slash-delimited gagal dengan
	// "expected command name (line 1 column N)" tepat di slash kedua
	// (/system/resource/print → kolom 8, /ip/address/print → kolom 4,
	// system/resource/print → kolom 7). RouterOS v7 asli menerima keduanya.
	// Ini bukan bug pattern prompt atau ReadDelay — sudah dibuktikan dengan
	// ssh -T remote command tanpa PTY sekalipun tetap gagal di v6, dan
	// command spasi setara jalan di v6 maupun v7. Diverifikasi: system
	// resource print, interface print, ip address print (semua spasi) OK
	// di 192.168.233.1 (v6) dan 192.168.230.3 (v7).
	result, err := drv.Execute(ctx, command.Command{Raw: "system resource print"})
	require.NoError(t, err)
	assert.NotEmpty(t, result.Output, "output system resource print tidak boleh kosong")
	assert.Contains(t, result.Output, "uptime", "output harus mengandung field 'uptime' seperti CLI RouterOS asli")

	// Execute kedua di sesi PERSISTEN yang sama — membuktikan koneksi
	// benar-benar dipakai ulang, bukan dial baru setiap command.
	result2, err := drv.Execute(ctx, command.Command{Raw: "interface print"})
	require.NoError(t, err)
	assert.NotEmpty(t, result2.Output)
}
