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
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	mikrotiksystem "github.com/quixiq/polyglot/internal/driver/mikrotik/system"
	"github.com/quixiq/polyglot/internal/port"
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

// TestMikrotikDriver_PingStreamDetails menguji secara langsung output row dari /ping streaming
// MikroTik fisik dan mencetak seluruh key-value map untuk menganalisis format respons RTT/time.
func TestMikrotikDriver_PingStreamDetails(t *testing.T) {
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	drv, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err)
	defer func() { assert.NoError(t, drv.Close()) }()

	sd, ok := port.DeviceDriver(drv).(port.StreamingDeviceDriver)
	require.True(t, ok)

	cmd := mikrotiksystem.NewPingStreamCommand("8.8.8.8")
	handle, err := sd.Stream(ctx, cmd)
	require.NoError(t, err)
	defer func() { assert.NoError(t, handle.Cancel()) }()

	t.Logf("=== MEMULAI INTEGRATION TEST STREAM PING KE 8.8.8.8 VIA %s:%d ===", target.Host, target.Port)

	for i := 0; i < 5; i++ {
		select {
		case res, ok := <-handle.Chan():
			require.True(t, ok, "channel stream ping harus tetap terbuka")
			if len(res.Rows) > 0 {
				row := res.Rows[0]
				t.Logf("[Frame %d] Received Row Map from MikroTik: %+v", i+1, row)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("[Frame %d] Timeout menunggu response /ping dari MikroTik", i+1)
		}
	}
}
