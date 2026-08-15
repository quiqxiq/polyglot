package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/quixiq/polyglot/pkg/logger"
)

type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
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

func (h *SSEHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan SSEEvent, 10)

	h.mutex.Lock()
	h.clients[clientChan] = true
	h.mutex.Unlock()

	defer func() {
		h.mutex.Lock()
		delete(h.clients, clientChan)
		close(clientChan)
		h.mutex.Unlock()
		logger.FromContext(r.Context()).Info("SSE client disconnected")
	}()

	logger.FromContext(r.Context()).Info("New SSE client connected")

	// Send initial ping to establish connection
	fmt.Fprintf(w, ": ping\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-clientChan:
			if !ok {
				return
			}
			dataBytes, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			if event.Event != "" {
				fmt.Fprintf(w, "event: %s\n", event.Event)
			}
			fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
			flusher.Flush()
		}
	}
}

// Broadcast sends an event payload to all connected clients.
func (h *SSEHub) Broadcast(event string, data interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	payload := SSEEvent{
		Event: event,
		Data:  data,
	}

	for clientChan := range h.clients {
		select {
		case clientChan <- payload:
		default:
			logger.WithField("event", event).Warn("SSE client buffer full, dropping event")
		}
	}
}
