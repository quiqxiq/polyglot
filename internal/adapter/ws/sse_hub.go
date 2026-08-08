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
		delete(h.clients, clientChan)
		close(clientChan)
		h.mutex.Unlock()
		log.Println("[SSEHub] Client disconnected")
	}()

	log.Println("[SSEHub] New client connected")

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

func (h *SSEHub) Broadcast(eventName string, data interface{}) {
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

func (h *SSEHub) PublishEvent(eventType string, data interface{}) {
	h.Broadcast(eventType, data)
}
