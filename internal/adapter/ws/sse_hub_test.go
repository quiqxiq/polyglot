package ws

import (
	"testing"
)

func TestSSEHubCloseClosesAllClients(t *testing.T) {
	h := NewSSEHub()

	// Simulasikan 3 client terdaftar tanpa melibatkan gin.Context — akses
	// langsung ke map internal sudah cukup untuk menguji kontrak Close().
	chans := make([]chan SSEEvent, 0, 3)
	h.mutex.Lock()
	for i := 0; i < 3; i++ {
		ch := make(chan SSEEvent, 1)
		h.clients[ch] = true
		chans = append(chans, ch)
	}
	h.mutex.Unlock()

	h.Close()

	// Semua channel client harus tertutup — inilah yang membuat c.Stream
	// di RegisterClient berhenti (ok=false) saat shutdown.
	for i, ch := range chans {
		select {
		case _, ok := <-ch:
			if ok {
				t.Errorf("client channel %d: expected closed, got open", i)
			}
		default:
			t.Errorf("client channel %d: expected closed state, got open", i)
		}
	}

	// Hub tidak boleh menyimpan client lagi setelah Close.
	h.mutex.RLock()
	n := len(h.clients)
	h.mutex.RUnlock()
	if n != 0 {
		t.Errorf("expected 0 registered clients after Close, got %d", n)
	}

	// Close harus idempoten dan Broadcast aman dipanggil setelahnya.
	h.Close()
	h.Broadcast("test", map[string]string{"k": "v"})
}

func TestSSEHubCloseThenCleanupNoPanic(t *testing.T) {
	// Simulasi skenario shutdown: Close() menutup channel lebih dulu, lalu
	// deferred cleanup di RegisterClient (delete + close) berjalan. Cleanup
	// harus idempoten — tidak boleh panic "close of closed channel".
	h := NewSSEHub()
	ch := make(chan SSEEvent, 1)

	h.mutex.Lock()
	h.clients[ch] = true
	h.mutex.Unlock()

	h.Close()

	// Jalur yang sama dengan deferred cleanup di RegisterClient.
	h.mutex.Lock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
	h.mutex.Unlock()

	// Cleanup kedua kali juga aman (idempoten).
	h.mutex.Lock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
	h.mutex.Unlock()
}

func TestSSEHubCloseUnblocksStream(t *testing.T) {
	h := NewSSEHub()
	ch := make(chan SSEEvent, 1)

	h.mutex.Lock()
	h.clients[ch] = true
	h.mutex.Unlock()

	h.Close()

	// Setelah Close, Broadcast tidak boleh panic (map sudah diganti baru).
	h.Broadcast("event", "data")

	// Simulasi langkah stream: read dari channel tertutup harus mengembalikan
	// ok=false, sama seperti yang dicek RegisterClient untuk mengakhiri stream.
	_, ok := <-ch
	if ok {
		t.Errorf("expected ok=false reading from closed client channel")
	}
}
