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

	wsAdapter "github.com/quixiq/polyglot/internal/adapter/ws"
)

// TestShutdownCompletesWithSSEClients regression test for clean shutdown with active SSE streams.
func TestShutdownCompletesWithSSEClients(t *testing.T) {
	sseHub := wsAdapter.NewSSEHub()
	mux := http.NewServeMux()
	wsAdapter.RegisterEventRoutes(mux, sseHub, nil)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	httpSrv := &http.Server{Handler: mux}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpSrv.Serve(ln)
	}()

	a := &App{
		httpServer: httpSrv,
		sseHub:     sseHub,
	}

	baseURL := "http://" + ln.Addr().String()

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

	close(stopPing)
	<-pingDone

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	if err := a.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("shutdown took %v with %d SSE clients connected", elapsed, numClients)
	}

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("http server exit: %v, want ErrServerClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("http server tidak berhenti setelah Shutdown")
	}

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
