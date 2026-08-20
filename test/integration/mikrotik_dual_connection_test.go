//go:build integration

// Integration test untuk skenario DUA KONEKSI KE MIKROTIK SEKALIGUS:
// membuka dua driver independen ke device yang sama dan membuktikan
// keduanya bisa bekerja bersamaan (Execute di keduanya, lalu Stream di
// satu driver selagi Execute di driver lain) tanpa interferensi.
//
// File ini juga memverifikasi Fase B (validasi goros Gate1+Gate2) terhadap
// device nyata — command valid harus lolos di v6 maupun v7; command tidak
// valid hanya ditolak di v7 (di v6 kedua gate degradasi senyap sesuai
// desain goros — bukan error).
//
// Dijalankan dengan:
//
//	MIKROTIK_TEST_HOST=... MIKROTIK_TEST_USER=admin MIKROTIK_TEST_PASS=... \
//	go test -tags=integration ./test/integration/... -run TestMikrotikDual -v
package integration

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// TestMikrotikDual_TwoConnectionsConcurrent membuka DUA driver ke device yang
// sama sekaligus (masing-masing membuka 2 koneksi RouterOS: exec + stream,
// lihat ADR 0003) lalu:
//  1. Execute bersamaan di kedua driver (goroutine paralel),
//  2. Stream /ping di driver A selagi Execute di driver B berjalan.
//
// Tujuannya membuktikan banyak koneksi simultan ke satu device aman dan
// independen — skenario yang persis dialami produksi (registry + beberapa
// client MCP membuka driver ke device yang sama).
func TestMikrotikDual_TwoConnectionsConcurrent(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	drvA, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek driver A")
	defer func() { assert.NoError(t, drvA.Close()) }()

	drvB, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "gagal konek driver B")
	defer func() { assert.NoError(t, drvB.Close()) }()

	t.Logf("Dua driver terbuka ke %s:%d (4 koneksi RouterOS total: 2 exec + 2 stream)", target.Host, target.Port)

	// ─── 1. Execute bersamaan di kedua driver ─────────────────────────────
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		if _, err := drvA.Execute(ctx, command.Command{Raw: "/system/resource/print"}); err != nil {
			errCh <- fmt.Errorf("drvA execute resource: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		if _, err := drvB.Execute(ctx, command.Command{Raw: "/system/identity/print"}); err != nil {
			errCh <- fmt.Errorf("drvB execute identity: %w", err)
		}
	}()
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err, "eksekusi bersamaan antar dua driver gagal")
	}

	// ─── 2. Stream di driver A selagi Execute di driver B ─────────────────
	sdA, ok := port.DeviceDriver(drvA).(port.StreamingDeviceDriver)
	require.True(t, ok, "drvA harus implement port.StreamingDeviceDriver")

	handle, err := sdA.Stream(ctx, command.Command{
		Raw:  "/ping",
		Args: map[string]string{"address": target.Host, "count": "20"},
	})
	require.NoError(t, err, "gagal buka stream /ping di driver A")
	defer func() { assert.NoError(t, handle.Cancel()) }()

	// Beri waktu supaya stream benar-benar mulai mengalir.
	time.Sleep(500 * time.Millisecond)

	// Execute di driver B harus tetap cepat selagi stream A mengalir.
	execCtx, execCancel := context.WithTimeout(ctx, 5*time.Second)
	defer execCancel()

	start := time.Now()
	resB, err := drvB.Execute(execCtx, command.Command{Raw: "/system/resource/print"})
	elapsed := time.Since(start)

	require.NoError(t, err, "Execute di drvB tidak boleh terblokir oleh Stream drvA")
	assert.Less(t, elapsed, 5*time.Second, "Execute drvB terlalu lambat selagi stream drvA aktif")
	assert.Len(t, resB.Rows, 1)

	// Stream di driver A tetap harus mengalir setelah Execute B selesai.
	select {
	case _, streamOK := <-handle.Chan():
		assert.True(t, streamOK, "stream drvA seharusnya masih terbuka")
		t.Log("stream drvA tetap mengalir setelah Execute drvB — kedua koneksi independen")
	case <-time.After(10 * time.Second):
		t.Fatal("stream drvA berhenti setelah Execute drvB — indikasi interferensi antar koneksi")
	}
}

// TestMikrotikDual_ValidateRealDevice memverifikasi integrasi Fase B
// (validasi goros Gate1+Gate2) terhadap device nyata:
//   - Command valid (path & atribut benar) selalu lolos di v6 maupun v7.
//   - Command tidak valid hanya ditolak di v7 (Gate2/Gate1 aktif); di v6
//     kedua gate degradasi senyap — validasi harus TIDAK error.
func TestMikrotikDual_ValidateRealDevice(t *testing.T) {
	drv := newTestDriver(t)
	ctx := context.Background()

	vd, ok := port.DeviceDriver(drv).(port.ValidatingDeviceDriver)
	require.True(t, ok, "mikrotik.Driver harus implement port.ValidatingDeviceDriver")

	// Deteksi versi RouterOS — perilaku Validate berbeda antara v6 dan v7.
	res, err := drv.Execute(ctx, mikrotik.NewPrintSystemResourceCommand())
	require.NoError(t, err)
	sys := mikrotik.ParseSystemResource(res)
	isV6 := strings.HasPrefix(sys.Version, "6.")
	t.Logf("RouterOS version: %s (v6=%v) — Validate aktif penuh hanya di v7", sys.Version, isV6)

	// ─── Command valid: harus lolos di versi mana pun ─────────────────────
	validCommands := []command.Command{
		{Raw: "/system/resource/print"},
		{Raw: "/system/identity/print"},
		{Raw: "/ip/address/print"},
		{Raw: "/interface/print"},
		{Raw: "/ip/address/print", Args: map[string]string{"?interface": "ether1"}},
		{Raw: "/ping", Args: map[string]string{"address": "8.8.8.8", "count": "1"}}, // streaming — harus di-skip bukan error
	}
	for _, cmd := range validCommands {
		t.Run("valid_"+cmd.Raw, func(t *testing.T) {
			require.NoError(t, vd.Validate(ctx, cmd), "command valid tidak boleh ditolak: %q", cmd.Raw)
		})
	}

	// ─── Command tidak valid: hanya diuji di v7 ───────────────────────────
	if isV6 {
		t.Log("v6: gate degradasi senyap — lewati pengujian penolakan command tidak valid (perilaku terverifikasi di lab)")
		return
	}

	// Atribut tak dikenal pada path yang TERKENAL harus ditolak Gate2.
	t.Run("reject_atribut_bogus", func(t *testing.T) {
		err := vd.Validate(ctx, command.Command{Raw: "/system/identity/print", Args: map[string]string{"bogusAttribute": "x"}})
		require.Error(t, err, "di v7, atribut tak dikenal pada path dikenal harus ditolak Gate2")
		t.Logf("ditolak sesuai harapan: %v", err)
	})

	// Path yang sama sekali tidak dikenal: Gate2 me-SKIP (bukan error) karena
	// schema.Discover mengembalikan CategoryUnknown — desain goros untuk
	// menghindari false-positive pada command versi-spesifik. Bukan bug,
	// perilaku terverifikasi di lab (gate2.go: "if sch.Category ==
	// CategoryUnknown { return nil }").
	t.Run("unknown_path_diskips", func(t *testing.T) {
		err := vd.Validate(ctx, command.Command{Raw: "/nonsense/definitely/not/real/print"})
		assert.NoError(t, err, "path tak dikenal di-skip Gate2 (CategoryUnknown), bukan ditolak")
		t.Log("path tak dikenal di-skip sesuai desain goros (anti false-positive)")
	})
}
