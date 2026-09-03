package port

import "github.com/quixiq/polyglot/internal/domain/device"

// SimpleQueue represents one row returned by /queue/simple/print.
// Aliased to domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type SimpleQueue = device.SimpleQueue
