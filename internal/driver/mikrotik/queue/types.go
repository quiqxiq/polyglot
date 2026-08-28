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

// QueueStreamParams defines filter and duration parameters for streaming queue statistics.
type QueueStreamParams struct {
	NameFilter   string // filter by queue name, e.g. "sub-budi"
	ParentFilter string // filter by parent queue name, e.g. "parent-total"
	ParentsOnly  bool   // filter non-dynamic static parent queues (?dynamic=false)
	Interval     string // RouterOS duration string (e.g. "1s", "500ms"). Defaults to "1s"
}

