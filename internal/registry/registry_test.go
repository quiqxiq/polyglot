package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// mockDriver is a minimal port.DeviceDriver for testing. It returns a
// fixed Result from Execute and predictable Classify/Translate results.
type mockDriver struct {
	closed bool
}

func (m *mockDriver) Execute(_ context.Context, _ command.Command) (command.Result, error) {
	return command.Result{Output: "ok"}, nil
}
func (m *mockDriver) Classify(_ command.Command) command.Class { return command.ClassReadOnly }
func (m *mockDriver) Translate(op command.Operation) (command.Command, error) {
	if op == command.OpGetStatus {
		return command.Command{Raw: "mock-status"}, nil
	}
	return command.Command{}, errors.New("unsupported")
}
func (m *mockDriver) Close() error {
	m.closed = true
	return nil
}

// memRepo is an in-memory port.DeviceRepository for testing.
type memRepo struct {
	devices map[string]device.Device
}

func (r *memRepo) Save(_ context.Context, d device.Device) error {
	r.devices[d.ID] = d
	return nil
}

func (r *memRepo) FindByID(_ context.Context, id string) (device.Device, error) {
	d, ok := r.devices[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}

func (r *memRepo) FindAll(_ context.Context) ([]device.Device, error) {
	list := make([]device.Device, 0, len(r.devices))
	for _, d := range r.devices {
		list = append(list, d)
	}
	return list, nil
}

func (r *memRepo) Update(ctx context.Context, d device.Device) error {
	return r.Save(ctx, d)
}

func (r *memRepo) FindByUserScope(ctx context.Context, userID uint) ([]device.Device, error) {
	return r.FindAll(ctx)
}

func (r *memRepo) Delete(_ context.Context, id string) error {
	delete(r.devices, id)
	return nil
}

// memVault is an in-memory port.CredentialVault for testing.
type memVault struct {
	creds map[string]device.Credentials
}

func (v *memVault) Get(_ context.Context, deviceID string) (device.Credentials, error) {
	c, ok := v.creds[deviceID]
	if !ok {
		return device.Credentials{}, device.ErrNotFound
	}
	return c, nil
}

func (v *memVault) Save(_ context.Context, deviceID string, c device.Credentials) error {
	v.creds[deviceID] = c
	return nil
}

func (v *memVault) EncryptString(_ context.Context, plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func (v *memVault) DecryptString(_ context.Context, ciphertext string) (string, error) {
	if len(ciphertext) > 4 && ciphertext[:4] == "enc:" {
		return ciphertext[4:], nil
	}
	return ciphertext, nil
}

func newTestRegistry(t *testing.T) (*Registry, *mockDriver) {
	t.Helper()
	drv := &mockDriver{}
	repo := &memRepo{
		devices: map[string]device.Device{
			"dev1": {ID: "dev1", Name: "test", Vendor: "mikrotik", DriverType: "mikrotik", Host: "10.0.0.1", Port: 8728, Enabled: true},
		},
	}
	vault := &memVault{
		creds: map[string]device.Credentials{
			"dev1": {Username: "admin", Password: "secret"},
		},
	}
	factories := map[string]DriverFactory{
		"mikrotik": func(_ context.Context, _ device.Target) (port.DeviceDriver, error) {
			return drv, nil
		},
	}
	return New(repo, vault, factories), drv
}

func TestRegistry_Get_CacheHit(t *testing.T) {
	reg, _ := newTestRegistry(t)
	defer func() { _ = reg.Close() }()

	ctx := context.Background()
	d1, err := reg.Get(ctx, "dev1")
	require.NoError(t, err)
	d2, err := reg.Get(ctx, "dev1")
	require.NoError(t, err)
	assert.Same(t, d1, d2, "second Get must return the same cached driver instance")
}

func TestRegistry_Get_NotFound(t *testing.T) {
	reg, _ := newTestRegistry(t)
	defer func() { _ = reg.Close() }()

	_, err := reg.Get(context.Background(), "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, device.ErrNotFound)
}

func TestRegistry_Get_DisabledDevice(t *testing.T) {
	drv := &mockDriver{}
	repo := &memRepo{
		devices: map[string]device.Device{
			"disabled1": {ID: "disabled1", DriverType: "mikrotik", Enabled: false},
		},
	}
	vault := &memVault{
		creds: map[string]device.Credentials{
			"disabled1": {Username: "admin"},
		},
	}
	reg := New(repo, vault, map[string]DriverFactory{
		"mikrotik": func(_ context.Context, _ device.Target) (port.DeviceDriver, error) { return drv, nil },
	})
	defer func() { _ = reg.Close() }()

	_, err := reg.Get(context.Background(), "disabled1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestRegistry_Get_UnknownDriverType(t *testing.T) {
	repo := &memRepo{
		devices: map[string]device.Device{
			"dev2": {ID: "dev2", DriverType: "unknown_vendor", Enabled: true},
		},
	}
	vault := &memVault{
		creds: map[string]device.Credentials{
			"dev2": {Username: "admin"},
		},
	}
	reg := New(repo, vault, map[string]DriverFactory{})
	defer func() { _ = reg.Close() }()

	_, err := reg.Get(context.Background(), "dev2")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no driver factory")
}

func TestRegistry_Close_ReleasesDrivers(t *testing.T) {
	reg, drv := newTestRegistry(t)

	_, _ = reg.Get(context.Background(), "dev1")
	require.NoError(t, reg.Close())
	assert.True(t, drv.closed, "Close must call driver.Close on cached drivers")
}
