package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

type SSEEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type SSEHub struct {
	clients map[chan SSEEvent]bool
	mutex   sync.RWMutex
}

var _ port.EventPublisher = (*SSEHub)(nil)

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan SSEEvent]bool),
	}
}

func (h *SSEHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := make(chan SSEEvent, 10)

	h.mutex.Lock()
	h.clients[clientChan] = true
	h.mutex.Unlock()

	defer func() {
		h.mutex.Lock()
		if _, ok := h.clients[clientChan]; ok {
			delete(h.clients, clientChan)
			close(clientChan)
		}
		h.mutex.Unlock()
		logger.WithComponent("SSEHub").Debug("client disconnected")
	}()

	logger.WithComponent("SSEHub").Debug("new client connected")

	// Event awal (ready) langsung ditulis + di-flush
	if _, err := fmt.Fprintf(w, "event: ready\ndata: {}\n\n"); err != nil {
		return
	}
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case event, ok := <-clientChan:
			if !ok {
				return
			}
			dataBytes, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, string(dataBytes)); err != nil {
				return
			}
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
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

func (h *SSEHub) Publish(eventName string, data any) {
	h.Broadcast(eventName, data)
}

func (h *SSEHub) PublishEvent(eventType string, data any) {
	h.Broadcast(eventType, data)
}

func (h *SSEHub) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	for clientChan := range h.clients {
		close(clientChan)
		delete(h.clients, clientChan)
	}
}
