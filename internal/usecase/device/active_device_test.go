package device_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
)

type memStore struct {
	mu   sync.Mutex
	data map[string]port.ActiveDevice
}

func newMemStore() *memStore    { return &memStore{data: map[string]port.ActiveDevice{}} }
func key(t, u, s string) string { return t + "|" + u + "|" + s }
func (m *memStore) Set(_ context.Context, a port.ActiveDevice, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key(a.TenantID, a.UserID, a.SessionID)] = a
	return nil
}
func (m *memStore) Get(_ context.Context, t, u, s string) (port.ActiveDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.data[key(t, u, s)]
	if !ok {
		return port.ActiveDevice{}, assert.AnError
	}
	return a, nil
}
func (m *memStore) Clear(_ context.Context, t, u, s string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key(t, u, s))
	return nil
}

type devRepo struct{ devices []device.Device }

func (r *devRepo) Save(context.Context, device.Device) error { return nil }
func (r *devRepo) FindByID(_ context.Context, id string) (device.Device, error) {
	for _, d := range r.devices {
		if d.ID == id {
			return d, nil
		}
	}
	return device.Device{}, device.ErrNotFound
}
func (r *devRepo) FindAll(context.Context) ([]device.Device, error) { return r.devices, nil }
func (r *devRepo) Update(context.Context, device.Device) error      { return nil }
func (r *devRepo) Delete(context.Context, string) error             { return nil }
func (r *devRepo) FindByUserScope(context.Context, uint) ([]device.Device, error) {
	return r.devices, nil
}

type allowAuth struct{ allow bool }

func (a allowAuth) CanAccessDevice(context.Context, uint, []string, string) (bool, error) {
	return a.allow, nil
}

func TestActiveDeviceResolvePriority(t *testing.T) {
	store := newMemStore()
	repo := &devRepo{devices: []device.Device{
		{ID: "dev-a", Name: "A", DriverType: "mikrotik", Enabled: true},
		{ID: "dev-b", Name: "B", DriverType: "mikrotik", Enabled: true},
	}}
	uc := deviceUC.NewActiveDeviceUseCase(store, repo, allowAuth{allow: true}, deviceUC.ActiveDeviceConfig{TTL: time.Hour})
	ctx := context.Background()

	// No binding → error, never random fallback.
	_, err := uc.Resolve(ctx, "tenant-default", "1", "")
	require.Error(t, err)

	// User binding.
	_, err = uc.Set(ctx, "tenant-default", "1", []string{"admin"}, "", "dev-a")
	require.NoError(t, err)
	got, err := uc.Resolve(ctx, "tenant-default", "1", "")
	require.NoError(t, err)
	assert.Equal(t, "dev-a", got)

	// Session binding wins over user binding.
	_, err = uc.Set(ctx, "tenant-default", "1", []string{"admin"}, "sess-1", "dev-b")
	require.NoError(t, err)
	got, err = uc.Resolve(ctx, "tenant-default", "1", "sess-1")
	require.NoError(t, err)
	assert.Equal(t, "dev-b", got)

	// Clear session falls back to user binding.
	require.NoError(t, uc.Clear(ctx, "tenant-default", "1", "sess-1"))
	got, err = uc.Resolve(ctx, "tenant-default", "1", "sess-1")
	require.NoError(t, err)
	assert.Equal(t, "dev-a", got)
}

func TestActiveDeviceRejectsNoAccessAndDisabled(t *testing.T) {
	store := newMemStore()
	repo := &devRepo{devices: []device.Device{
		{ID: "dev-a", Name: "A", DriverType: "mikrotik", Enabled: true},
		{ID: "dev-off", Name: "Off", DriverType: "mikrotik", Enabled: false},
	}}
	uc := deviceUC.NewActiveDeviceUseCase(store, repo, allowAuth{allow: false}, deviceUC.ActiveDeviceConfig{})
	ctx := context.Background()
	_, err := uc.Set(ctx, "tenant-default", "1", []string{"teknisi"}, "", "dev-a")
	require.Error(t, err)
	_, err = uc.Set(ctx, "tenant-default", "1", []string{"admin"}, "", "dev-off")
	require.Error(t, err)
}

func TestActiveDeviceSingleFallbackOnlyWhenEnabled(t *testing.T) {
	store := newMemStore()
	repo := &devRepo{devices: []device.Device{{ID: "dev-a", Name: "A", DriverType: "mikrotik", Enabled: true}}}
	uc := deviceUC.NewActiveDeviceUseCase(store, repo, allowAuth{allow: true}, deviceUC.ActiveDeviceConfig{SingleDeviceFallback: true})
	got, err := uc.Resolve(context.Background(), "tenant-default", "1", "")
	require.NoError(t, err)
	assert.Equal(t, "dev-a", got)
}
