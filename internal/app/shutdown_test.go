package app

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	wsAdapter "github.com/quixiq/polyglot/internal/adapter/ws"
)

// TestShutdownCompletesWithSSEClients adalah regression test untuk bug
// "Error shutting down HTTP server: context deadline exceeded" saat Ctrl+C.
//
// Dulu SSE /events (c.Stream) memblokir http.Server.Shutdown sampai grace
// timeout habis karena EventSource browser tidak pernah menutup koneksi.
// Sekarang App.Shutdown memanggil sseHub.Close() lebih dulu sehingga stream
// berakhir dan shutdown harus selesai jauh di bawah grace period.
//
// Test ini self-contained: server HTTP + SSEHub + gin router asli, koneksi
// SSE via HTTP client sungguhan (meniru EventSource), lalu shutdown dengan
// context ber-grace 3 detik. Tanpa fix, elapsed >= 3s dan test ini gagal.
func TestShutdownCompletesWithSSEClients(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sseHub := wsAdapter.NewSSEHub()
	// openTermUC nil aman — route /ws/devices tidak dipanggil di test ini,
	// hanya /events yang dipakai.
	r := gin.New()
	wsAdapter.RegisterEventRoutes(r, sseHub, nil)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	httpSrv := &http.Server{Handler: r}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpSrv.Serve(ln)
	}()

	// Konstruksi App parsial — hanya komponen yang disentuh Shutdown:
	// httpServer + sseHub. waManager/registry/pgStore nil, di-skip aman.
	a := &App{
		httpServer: httpSrv,
		sseHub:     sseHub,
	}

	baseURL := "http://" + ln.Addr().String()

	// gin c.Stream baru mem-flush header setelah event PERTAMA dikirim ke
	// klien tersebut (flush terjadi setelah step menghasilkan output). Jadi
	// alirkan event ping berkala meniru produksi (session_status/chat_update
	// terus mengalir) supaya Do() mendapat header dan tidak deadlock.
	stopPing := make(chan struct{})
	pingDone := make(chan struct{})
	go func() {
		defer close(pingDone)
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopPing:
				return
			case <-ticker.C:
				sseHub.Broadcast("ping", map[string]string{"k": "v"})
			}
		}
	}()

	// Buka 3 koneksi SSE sekaligus — meniru browser dengan beberapa tab
	// (halaman chats + whatsapp + indikator Live) yang memicu bug aslinya.
	const numClients = 3
	readers := make([]*bufio.Reader, 0, numClients)
	for i := 0; i < numClients; i++ {
		req, err := http.NewRequest(http.MethodGet, baseURL+"/events", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("open SSE connection %d: %v", i, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("SSE client %d: status = %d, want 200", i, resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
			t.Fatalf("SSE client %d: content-type = %q, want text/event-stream", i, ct)
		}
		readers = append(readers, bufio.NewReader(resp.Body))
	}

	// Pastikan stream benar-benar hidup: baca event line dari tiap klien.
	// Sejak RegisterClient menulis event `ready` saat koneksi dibuka (agar
	// EventSource langsung open), baris pertama adalah `event: ready` —
	// lewati event selain ping sampai event ping (aliran produksi) tiba.
	for i, rdr := range readers {
		gotPing := false
		deadline := time.Now().Add(2 * time.Second)
		for !gotPing && time.Now().Before(deadline) {
			line, err := rdr.ReadString('\n')
			if err != nil {
				t.Fatalf("SSE client %d: read event line: %v", i, err)
			}
			if strings.HasPrefix(line, "event: ping") {
				gotPing = true
			}
		}
		if !gotPing {
			t.Fatalf("SSE client %d: event ping tidak diterima dalam 2s", i)
		}
	}

	// Hentikan aliran ping sebelum shutdown supaya pengukuran elapsed bersih.
	// Catatan: test memanggil App.Shutdown langsung (in-process) — ini jalur
	// yang sama yang dijalankan cmd/server/main.go setelah signal.NotifyContext
	// menerima SIGINT/SIGTERM; bagian "kirim sinyal" hanyalah ~5 baris glue
	// yang tidak perlu infra DB/Redis/WhatsApp untuk diuji.
	close(stopPing)
	<-pingDone

	// Shutdown dengan grace pendek (3s) — harus selesai cepat tanpa error.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	if err := a.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	elapsed := time.Since(start)

	// Margin lebar untuk CI noise, tapi jauh di bawah grace 3s — kalau SSE
	// masih memblokir (regresi), elapsed >= 3s dan test ini gagal.
	if elapsed > 2*time.Second {
		t.Fatalf("shutdown took %v with %d SSE clients connected — SSE stream masih memblokir graceful shutdown", elapsed, numClients)
	}

	// Server HTTP harus berhenti bersih dengan ErrServerClosed.
	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("http server exit: %v, want ErrServerClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("http server tidak berhenti setelah Shutdown")
	}

	// Semua koneksi SSE harus sudah tertutup oleh server. Reader bisa masih
	// punya event ping yang ter-buffer sebelum koneksi ditutup, jadi drain
	// dulu — kuncinya kita harus mencapai EOF/error dalam jendela pendek.
	for i, rdr := range readers {
		sawEOF := false
		for {
			_, err := readByteWithTimeout(rdr, 2*time.Second)
			if err != nil {
				sawEOF = true
				break
			}
		}
		if !sawEOF {
			t.Errorf("SSE client %d: koneksi masih terbuka setelah shutdown", i)
		}
	}
}

// readByteWithTimeout membungkus ReadByte dengan timeout supaya pembacaan
// tidak bisa memblokir tanpa batas kalau koneksi tidak pernah ditutup.
func readByteWithTimeout(rdr *bufio.Reader, timeout time.Duration) (byte, error) {
	type result struct {
		b   byte
		err error
	}
	ch := make(chan result, 1)
	go func() {
		b, err := rdr.ReadByte()
		ch <- result{b: b, err: err}
	}()
	select {
	case r := <-ch:
		return r.b, r.err
	case <-time.After(timeout):
		return 0, errors.New("read timed out")
	}
}
