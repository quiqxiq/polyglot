//go:build integration

// Integration test untuk typed command builders dan streaming commands.
// Dijalankan dengan:
//
//	go test -tags=integration ./test/integration/... -run TestMikrotikBuilders -v
//
// Membutuhkan MIKROTIK_TEST_HOST (dan opsional MIKROTIK_TEST_PORT/USER/PASS)
// di environment. Semua operasi write (add/set/remove) menggunakan nama/IP
// test yang tidak konflik dengan konfigurasi router production:
//   - PPPoE secret: "polyglot-test-user"
//   - Simple Queue: "polyglot-test-queue"
// Semua entry yang dibuat di-cleanup di t.Cleanup, jadi test bersih
// meskipun gagal di tengah jalan.
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// newTestDriver membuat koneksi ke router test dan meregistrasi cleanup.
func newTestDriver(t *testing.T) *mikrotik.Driver {
	t.Helper()
	target := mikrotikTestTarget(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	drv, err := mikrotik.NewDriver(ctx, device.Target{
		Host:     target.Host,
		Port:     target.Port,
		Username: target.Username,
		Password: target.Password,
		Timeout:  target.Timeout,
	})
	require.NoError(t, err, "gagal konek ke Mikrotik test")
	t.Cleanup(func() { assert.NoError(t, drv.Close()) })
	return drv
}

// ─── /system/resource — one-shot ─────────────────────────────────────────

func TestMikrotikBuilders_SystemResource(t *testing.T) {
	drv := newTestDriver(t)
	ctx := context.Background()

	cmd := mikrotik.NewPrintSystemResourceCommand()
	result, err := drv.Execute(ctx, cmd)
	require.NoError(t, err)
	require.NotEmpty(t, result.Rows)

	res := mikrotik.ParseSystemResource(result)
	t.Logf("RouterOS version: %s", res.Version)
	t.Logf("Uptime: %s", res.Uptime)
	t.Logf("CPU load: %d%%", res.CPULoad)
	t.Logf("Board: %s", res.BoardName)

	assert.NotEmpty(t, res.Version, "version harus ada")
	assert.NotEmpty(t, res.Uptime, "uptime harus ada")
	assert.GreaterOrEqual(t, res.CPULoad, 0)
	assert.LessOrEqual(t, res.CPULoad, 100)
}

// ─── /system/identity ─────────────────────────────────────────────────────

func TestMikrotikBuilders_SystemIdentity(t *testing.T) {
	drv := newTestDriver(t)
	ctx := context.Background()

	result, err := drv.Execute(ctx, mikrotik.NewPrintSystemIdentityCommand())
	require.NoError(t, err)

	identity := mikrotik.ParseSystemIdentity(result)
	t.Logf("Router identity: %s", identity.Name)
	assert.NotEmpty(t, identity.Name)
}

// ─── /interface/print + monitor-traffic once ──────────────────────────────

func TestMikrotikBuilders_Interfaces(t *testing.T) {
	drv := newTestDriver(t)
	ctx := context.Background()

	// List interfaces
	result, err := drv.Execute(ctx, mikrotik.NewPrintInterfacesCommand(""))
	require.NoError(t, err)

	ifaces := mikrotik.ParseInterfaces(result)
	require.NotEmpty(t, ifaces, "router harus punya minimal satu interface")
	t.Logf("Interface pertama: %s (type=%s running=%v)", ifaces[0].Name, ifaces[0].Type, ifaces[0].Running)

	// Monitor traffic one-shot pada interface pertama
	ifaceName := ifaces[0].Name
	onceResult, err := drv.Execute(ctx, mikrotik.NewMonitorTrafficOnceCommand(ifaceName))
	require.NoError(t, err)

	stats := mikrotik.ParseInterfaceTrafficStats(onceResult)
	t.Logf("Traffic stats %s: rx=%s bps tx=%s bps", ifaceName, stats.RxBitsPerSecond, stats.TxBitsPerSecond)
	// Hanya validasi field ada — nilai bisa nol jika interface idle
	assert.NotNil(t, stats)
}

// ─── /ppp/active/print ────────────────────────────────────────────────────

func TestMikrotikBuilders_PPPActive(t *testing.T) {
	drv := newTestDriver(t)
	ctx := context.Background()

	result, err := drv.Execute(ctx, mikrotik.NewPrintPPPActiveCommand(""))
	require.NoError(t, err)

	sessions := mikrotik.ParsePPPActiveSessions(result)
	t.Logf("Sesi PPPoE aktif: %d", len(sessions))
	// Tidak assert jumlah — mungkin memang tidak ada yang online
}

// ─── /ppp/secret CRUD cycle ───────────────────────────────────────────────

func TestMikrotikBuilders_PPPoESecretCRUD(t *testing.T) {
	const testUsername = "polyglot-test-user"
	drv := newTestDriver(t)
	ctx := context.Background()

	// Cleanup: pastikan test user tidak ada di akhir, baik sukses atau gagal.
	t.Cleanup(func() {
		printResult, err := drv.Execute(context.Background(), mikrotik.NewPrintPPPoESecretsCommand(testUsername))
		if err != nil {
			return
		}
		rosID, err := mikrotik.FindPPPoESecretRosID(printResult, testUsername)
		if err != nil {
			return // tidak ada — fine
		}
		_ = func() error {
			_, e := drv.Execute(context.Background(), mikrotik.NewRemovePPPoESecretCommand(rosID))
			return e
		}()
	})

	// 1. Pastikan test user belum ada
	printResult, err := drv.Execute(ctx, mikrotik.NewPrintPPPoESecretsCommand(testUsername))
	require.NoError(t, err)
	_, errFind := mikrotik.FindPPPoESecretRosID(printResult, testUsername)
	if errFind == nil {
		// User sudah ada dari test sebelumnya yang tidak bersih — remove dulu
		rosID, _ := mikrotik.FindPPPoESecretRosID(printResult, testUsername)
		_, _ = drv.Execute(ctx, mikrotik.NewRemovePPPoESecretCommand(rosID))
	}

	// 2. Add
	addCmd := mikrotik.NewAddPPPoESecretCommand(mikrotik.PPPoESecretParams{
		Name:     testUsername,
		Password: "testpass123",
		Profile:  "default",
		Comment:  "polyglot-integration-test",
	})
	_, err = drv.Execute(ctx, addCmd)
	require.NoError(t, err, "gagal add PPPoE secret")

	// 3. Print — verifikasi ada
	printResult, err = drv.Execute(ctx, mikrotik.NewPrintPPPoESecretsCommand(testUsername))
	require.NoError(t, err)
	secrets := mikrotik.ParsePPPoESecrets(printResult)
	require.NotEmpty(t, secrets, "secret harus ditemukan setelah add")
	found := secrets[0]
	assert.Equal(t, testUsername, found.Name)
	assert.Equal(t, "polyglot-integration-test", found.Comment)
	t.Logf("Secret dibuat: name=%s rosID=%s", found.Name, found.RosID)

	// 4. Set — ganti password
	setCmd := mikrotik.NewSetPPPoESecretCommand(found.RosID, mikrotik.PPPoESecretParams{
		Password: "newpass456",
	})
	_, err = drv.Execute(ctx, setCmd)
	require.NoError(t, err, "gagal set PPPoE secret")

	// 5. Verifikasi set berhasil
	printResult2, err := drv.Execute(ctx, mikrotik.NewPrintPPPoESecretsCommand(testUsername))
	require.NoError(t, err)
	secrets2 := mikrotik.ParsePPPoESecrets(printResult2)
	require.NotEmpty(t, secrets2)
	t.Logf("Secret setelah set: name=%s profile=%s", secrets2[0].Name, secrets2[0].Profile)

	// 6. Remove
	removeCmd := mikrotik.NewRemovePPPoESecretCommand(found.RosID)
	_, err = drv.Execute(ctx, removeCmd)
	require.NoError(t, err, "gagal remove PPPoE secret")

	// 7. Verifikasi sudah tidak ada
	printResult3, err := drv.Execute(ctx, mikrotik.NewPrintPPPoESecretsCommand(testUsername))
	require.NoError(t, err)
	_, errNotFound := mikrotik.FindPPPoESecretRosID(printResult3, testUsername)
	assert.ErrorIs(t, errNotFound, mikrotik.ErrSecretNotFound, "secret harus sudah terhapus")
}

// ─── /queue/simple CRUD ───────────────────────────────────────────────────

func TestMikrotikBuilders_SimpleQueueCRUD(t *testing.T) {
	const testQueueName = "polyglot-test-queue"
	const testTarget = "192.0.2.99" // TEST-NET (RFC 5737) — tidak akan pernah konflik
	drv := newTestDriver(t)
	ctx := context.Background()

	// Cleanup
	t.Cleanup(func() {
		printResult, err := drv.Execute(context.Background(), mikrotik.NewPrintSimpleQueuesCommand(testQueueName))
		if err != nil {
			return
		}
		queues := mikrotik.ParseSimpleQueues(printResult)
		for _, q := range queues {
			_, _ = drv.Execute(context.Background(), mikrotik.NewRemoveSimpleQueueCommand(q.RosID))
		}
	})

	// Pre-cleanup jika queue sudah ada dari test sebelumnya yang gagal
	if printResult, err := drv.Execute(ctx, mikrotik.NewPrintSimpleQueuesCommand(testQueueName)); err == nil {
		for _, q := range mikrotik.ParseSimpleQueues(printResult) {
			_, _ = drv.Execute(ctx, mikrotik.NewRemoveSimpleQueueCommand(q.RosID))
		}
	}

	// Add queue
	addCmd := mikrotik.NewAddSimpleQueueCommand(mikrotik.SimpleQueueParams{
		Name:     testQueueName,
		Target:   testTarget,
		MaxLimit: "1M/1M",
		Comment:  "polyglot-integration-test",
	})
	_, err := drv.Execute(ctx, addCmd)
	require.NoError(t, err, "gagal add simple queue")

	// Verifikasi ada
	printResult, err := drv.Execute(ctx, mikrotik.NewPrintSimpleQueuesCommand(testQueueName))
	require.NoError(t, err)
	queues := mikrotik.ParseSimpleQueues(printResult)
	require.NotEmpty(t, queues)
	q := queues[0]
	assert.Equal(t, testQueueName, q.Name)
	assert.Contains(t, q.Target, testTarget, "target harus mengandung IP test (RouterOS menambahkan /32)")
	t.Logf("Queue dibuat: name=%s rosID=%s maxLimit=%s", q.Name, q.RosID, q.MaxLimit)

	// Remove
	_, err = drv.Execute(ctx, mikrotik.NewRemoveSimpleQueueCommand(q.RosID))
	require.NoError(t, err, "gagal remove simple queue")
}

func TestMikrotikBuilders_StreamQueueStats(t *testing.T) {
	drv := newTestDriver(t)
	ctx := context.Background()

	testQueueName := "polyglot-stream-queue"
	addCmd := mikrotik.NewAddSimpleQueueCommand(mikrotik.SimpleQueueParams{
		Name:     testQueueName,
		Target:   "10.88.99.1/32",
		MaxLimit: "2M/2M",
		Comment:  "polyglot-stream-test",
	})
	_, err := drv.Execute(ctx, addCmd)
	require.NoError(t, err)

	defer func() {
		if res, err := drv.Execute(ctx, mikrotik.NewPrintSimpleQueuesCommand(testQueueName)); err == nil && len(res.Rows) > 0 {
			rosID := res.Rows[0][".id"]
			_, _ = drv.Execute(ctx, mikrotik.NewRemoveSimpleQueueCommand(rosID))
		}
	}()

	sd, ok := port.DeviceDriver(drv).(port.StreamingDeviceDriver)
	require.True(t, ok)

	streamCmd := mikrotik.NewStreamQueueStatsCommand(mikrotik.QueueStreamParams{
		NameFilter:  testQueueName,
		ParentsOnly: true,
		Interval:    "1s",
	})

	handle, err := sd.Stream(ctx, streamCmd)
	require.NoError(t, err)
	defer handle.Cancel()

	select {
	case res, ok := <-handle.Chan():
		require.True(t, ok)
		queues := mikrotik.ParseSimpleQueues(res)
		require.NotEmpty(t, queues)
		assert.Equal(t, testQueueName, queues[0].Name)
		t.Logf("Queue stream tick: name=%s rate=%s packetRate=%s", queues[0].Name, queues[0].Rate, queues[0].PacketRate)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for queue stats stream tick")
	}
}

// ─── /interface/monitor-traffic streaming ────────────────────────────────

func TestMikrotikBuilders_MonitorTrafficStream(t *testing.T) {
	drv := newTestDriver(t)
	ctx := context.Background()

	// Ambil interface pertama
	ifaceResult, err := drv.Execute(ctx, mikrotik.NewPrintInterfacesCommand(""))
	require.NoError(t, err)
	ifaces := mikrotik.ParseInterfaces(ifaceResult)
	require.NotEmpty(t, ifaces)
	ifaceName := ifaces[0].Name

	// Cast ke StreamingDeviceDriver
	sd, ok := port.DeviceDriver(drv).(port.StreamingDeviceDriver)
	require.True(t, ok)

	// Stream monitor-traffic (tanpa once = streaming)
	streamCtx, streamCancel := context.WithTimeout(ctx, 10*time.Second)
	defer streamCancel()

	handle, err := sd.Stream(streamCtx, mikrotik.NewMonitorTrafficStreamCommand(ifaceName))
	require.NoError(t, err)
	defer func() { assert.NoError(t, handle.Cancel()) }()

	// Tunggu minimal 2 tick (RouterOS kirim setiap ~1 detik)
	received := 0
	timeout := time.After(5 * time.Second)
loop:
	for {
		select {
		case result, ok := <-handle.Chan():
			if !ok {
				break loop
			}
			stats := mikrotik.ParseInterfaceTrafficStats(result)
			t.Logf("tick %d: %s rx=%s tx=%s bps", received+1, ifaceName, stats.RxBitsPerSecond, stats.TxBitsPerSecond)
			received++
			if received >= 2 {
				break loop // cukup 2 tick sebagai bukti streaming berjalan
			}
		case <-timeout:
			t.Fatal("timeout menunggu data monitor-traffic streaming")
		}
	}
	assert.GreaterOrEqual(t, received, 1, "harus terima minimal 1 tick data")
	assert.NoError(t, handle.Err())
}

// ─── /ppp/active streaming (follow) ──────────────────────────────────────

func TestMikrotikBuilders_PPPActiveStream(t *testing.T) {
	drv := newTestDriver(t)

	sd, ok := port.DeviceDriver(drv).(port.StreamingDeviceDriver)
	require.True(t, ok)

	streamCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stream PPP active — test hanya membuktikan stream bisa dibuka dan
	// tidak langsung error. Jumlah row 0 sah jika tidak ada yang online.
	handle, err := sd.Stream(streamCtx, mikrotik.NewStreamPPPActiveCommand(""))
	require.NoError(t, err, "gagal buka stream ppp active")

	// Tunggu sebentar lalu cancel — stream yang valid harus masih terbuka
	time.Sleep(1 * time.Second)
	assert.NoError(t, handle.Cancel())
	assert.NoError(t, handle.Err())
	t.Log("stream /ppp/active/print follow berhasil dibuka dan di-cancel")
}

// ─── /log/print ──────────────────────────────────────────────────────────

func TestMikrotikBuilders_LogPrint(t *testing.T) {
	drv := newTestDriver(t)
	ctx := context.Background()

	result, err := drv.Execute(ctx, mikrotik.NewPrintLogCommand(""))
	require.NoError(t, err)

	entries := mikrotik.ParseLogEntries(result)
	t.Logf("Log entries: %d", len(entries))
	// Log mungkin kosong di CHR baru — tidak di-assert jumlahnya
	if len(entries) > 0 {
		t.Logf("Entry pertama: [%s] %s", entries[0].Topics, entries[0].Message)
	}
}
