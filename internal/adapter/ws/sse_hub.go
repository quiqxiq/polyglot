package ws

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

type SSEEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type SSEHub struct {
	clients map[chan SSEEvent]bool
	mutex   sync.RWMutex
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan SSEEvent]bool),
	}
}

func (h *SSEHub) RegisterClient(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	clientChan := make(chan SSEEvent, 10)

	h.mutex.Lock()
	h.clients[clientChan] = true
	h.mutex.Unlock()

	defer func() {
		h.mutex.Lock()
		// Idempoten: hanya tutup channel kalau masih terdaftar. Tanpa guard ini,
		// Close() saat shutdown bisa menutup channel lebih dulu lalu cleanup ini
		// memanggil close lagi -> panic "close of closed channel".
		if _, ok := h.clients[clientChan]; ok {
			delete(h.clients, clientChan)
			close(clientChan)
		}
		h.mutex.Unlock()
		log.Println("[SSEHub] Client disconnected")
	}()

	log.Println("[SSEHub] New client connected")

	// Event awal (ready) langsung ditulis + di-flush SEBELUM loop stream.
	// Tanpa ini gin c.Stream baru mengirim header HTTP setelah event pertama
	// masuk, sehingga EventSource browser tetap di state CONNECTING
	// ("Menghubungkan…") sampai event berikutnya tiba — dan semua broadcast
	// (session_status, chat_update, conversation_status) yang terjadi di
	// antara terlewat tanpa pernah sampai ke browser. Dengan flush segera,
	// EventSource membuka koneksi (onopen) dan siap menerima broadcast.
	fmt.Fprintf(c.Writer, "event: ready\ndata: {}\n\n")
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-clientChan:
			if !ok {
				return false
			}
			dataBytes, err := json.Marshal(event.Data)
			if err != nil {
				return true
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, string(dataBytes))
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

func (h *SSEHub) Broadcast(eventName string, data any) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	event := SSEEvent{
		Event: eventName,
		Data:  data,
	}

	for clientChan := range h.clients {
		select {
		case clientChan <- event:
		default:
		}
	}
}

func (h *SSEHub) PublishEvent(eventType string, data any) {
	h.Broadcast(eventType, data)
}

// Close disconnects every registered SSE client by closing their channels.
// Dipanggil saat graceful shutdown supaya http.Server.Shutdown tidak menunggu
// koneksi streaming yang tidak pernah selesai (EventSource browser tetap
// terbuka) sampai context deadline habis lalu gagal. Setelah Close, stream di
// RegisterClient menerima ok=false dan handler selesai sehingga koneksi jadi
// idle dan shutdown bisa selesai segera.
func (h *SSEHub) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for clientChan := range h.clients {
		close(clientChan)
	}
	h.clients = make(map[chan SSEEvent]bool)
}
