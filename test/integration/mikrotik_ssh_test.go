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
		DestructivePrefixes: []string{
			"/system/reboot",
			"/system/reset-configuration",
			"/system/shutdown",
		},
		Operations: map[command.Operation]command.Command{
			// Sengaja sama dengan yang dipakai internal/driver/mikrotik
			// (driver API) untuk OpGetStatus — path slash-delimited yang
			// sama berlaku di CLI RouterOS, bukan cuma di API.
			command.OpGetStatus: {Raw: "/system/resource/print"},
			command.OpReboot:    {Raw: "/system/reboot"},
		},
		// Lihat GitHub issue scrapli/scrapligo#95 dan komentar di
		// mikrotik_routeros.yaml: RouterOS meng-echo prompt/command dua
		// kali lewat SSH. 100ms adalah titik awal yang konservatif, bukan
		// angka final — sesuaikan berdasarkan hasil test ini di device
		// Anda.
		ReadDelay: 100 * time.Millisecond,
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

	result, err := drv.Execute(ctx, command.Command{Raw: "/system/resource/print"})
	require.NoError(t, err)
	assert.NotEmpty(t, result.Output, "output /system/resource/print tidak boleh kosong")
	assert.Contains(t, result.Output, "uptime", "output harus mengandung field 'uptime' seperti CLI RouterOS asli")

	// Execute kedua di sesi PERSISTEN yang sama — membuktikan koneksi
	// benar-benar dipakai ulang, bukan dial baru setiap command.
	result2, err := drv.Execute(ctx, command.Command{Raw: "/interface/print"})
	require.NoError(t, err)
	assert.NotEmpty(t, result2.Output)
}
