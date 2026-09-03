package port

import "github.com/quixiq/polyglot/internal/domain/device"

// IPPool represents one row returned by /ip/pool/print.
// Aliased to domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type IPPool = device.IPPool
