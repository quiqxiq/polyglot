package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// DedicatedQueueParams berisi parameter simple queue khusus pelanggan
// DEDICATED (bandwidth didedikasi / CIR dijamin).
type DedicatedQueueParams struct {
	Name           string // "dq-<username>" — deterministik per pelanggan
	Target         string // username pppoe (RouterOS me-resolve ke IP sesi)
	MaxLimit       string // rate maksimal ("10M/5M" atau 8-segmen burst)
	LimitAt        string // CIR yang dijamin ("10M/5M")
	BurstLimit     string // kosong bila tanpa burst
	BurstThreshold string
	BurstTime      string
	Comment        string
}

// QueueGateway mengelola /queue/simple di router — dipakai untuk menegakkan
// CIR langganan DEDICATED. Implementasi vendor: internal/driver/mikrotik.
type QueueGateway interface {
	// ListQueues mengambil daftar simple queue, opsional difilter nama.
	ListQueues(ctx context.Context, driver DeviceDriver, nameFilter string) ([]SimpleQueue, error)
	// AddQueue membuat simple queue baru dari parameter dedicated.
	AddQueue(ctx context.Context, driver DeviceDriver, p DedicatedQueueParams) (command.Result, error)
	// RemoveQueue menghapus queue by RouterOS .id.
	RemoveQueue(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// SetQueueEnabled mengaktifkan/menonaktifkan queue by RouterOS .id.
	SetQueueEnabled(ctx context.Context, driver DeviceDriver, rosID string, enabled bool) (command.Result, error)
}
