package memory

import (
	"context"
	"sync"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// MemDeviceRepository is an in-memory implementation of port.DeviceRepository.
type MemDeviceRepository struct {
	mu      sync.RWMutex
	devices map[string]device.Device
}

var _ port.DeviceRepository = (*MemDeviceRepository)(nil)

// NewDeviceRepository constructs a new MemDeviceRepository.
func NewDeviceRepository() *MemDeviceRepository {
	return &MemDeviceRepository{
		devices: make(map[string]device.Device),
	}
}

func (r *MemDeviceRepository) Save(_ context.Context, d device.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[d.ID] = d
	return nil
}

func (r *MemDeviceRepository) FindByID(_ context.Context, id string) (device.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.devices[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}

func (r *MemDeviceRepository) FindAll(_ context.Context) ([]device.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]device.Device, 0, len(r.devices))
	for _, d := range r.devices {
		list = append(list, d)
	}
	return list, nil
}

func (r *MemDeviceRepository) Update(ctx context.Context, d device.Device) error {
	return r.Save(ctx, d)
}

func (r *MemDeviceRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.devices, id)
	return nil
}

// MemCredentialVault is an in-memory implementation of port.CredentialVault.
type MemCredentialVault struct {
	mu    sync.RWMutex
	creds map[string]device.Credentials
}

var _ port.CredentialVault = (*MemCredentialVault)(nil)

// NewCredentialVault constructs a new MemCredentialVault.
func NewCredentialVault() *MemCredentialVault {
	return &MemCredentialVault{
		creds: make(map[string]device.Credentials),
	}
}

func (v *MemCredentialVault) Get(_ context.Context, deviceID string) (device.Credentials, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	c, ok := v.creds[deviceID]
	if !ok {
		return device.Credentials{}, device.ErrNotFound
	}
	return c, nil
}

func (v *MemCredentialVault) Save(_ context.Context, deviceID string, c device.Credentials) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.creds[deviceID] = c
	return nil
}
