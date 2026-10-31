package queue

import (
	"github.com/quixiq/polyglot/internal/port"
)

// SimpleQueue is the vendor-neutral simple queue row.
// Canonical definition lives in internal/port.
type SimpleQueue = port.SimpleQueue

// DedicatedQueueParams is the vendor-neutral dedicated queue params.
// Canonical definition lives in internal/port.
type DedicatedQueueParams = port.DedicatedQueueParams

// SimpleQueueParams holds the parameters for creating a RouterOS simple queue
// (/queue/simple/add).
type SimpleQueueParams struct {
	Name           string
	Target         string
	MaxLimit       string
	LimitAt        string
	BurstLimit     string
	BurstThreshold string
	BurstTime      string
	Priority       string
	Parent         string
	Comment        string
	Disabled       bool
}

// StreamParams defines queue statistics filters and interval.
type StreamParams = port.QueueStreamParams

// QueueTree represents a RouterOS /queue/tree entry.
type QueueTree struct {
	RosID      string `json:"ros_id"`
	Name       string `json:"name"`
	Parent     string `json:"parent"`
	PacketMark string `json:"packet_mark"`
	LimitAt    string `json:"limit_at"`
	MaxLimit   string `json:"max_limit"`
	Priority   string `json:"priority"`
	Bytes      string `json:"bytes"`
	Packets    string `json:"packets"`
	Dropped    string `json:"dropped"`
	Rate       string `json:"rate"`
	PacketRate string `json:"packet_rate"`
	Disabled   bool   `json:"disabled"`
}

// QueueTreeParams holds parameters for adding or setting a queue tree entry.
type QueueTreeParams struct {
	Name       string
	Parent     string
	PacketMark string
	LimitAt    string
	MaxLimit   string
	Priority   string
	Disabled   bool
}
