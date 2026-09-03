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
