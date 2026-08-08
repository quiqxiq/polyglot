package device

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDeviceRepo struct {
	devices map[string]device.Device
}

func newMockDeviceRepo() *mockDeviceRepo {
	return &mockDeviceRepo{devices: make(map[string]device.Device)}
}

func (m *mockDeviceRepo) Save(ctx context.Context, d device.Device) error {
	m.devices[d.ID] = d
	return nil
}

func (m *mockDeviceRepo) FindByID(ctx context.Context, id string) (device.Device, error) {
	d, ok := m.devices[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}

func (m *mockDeviceRepo) FindAll(ctx context.Context) ([]device.Device, error) {
	list := make([]device.Device, 0, len(m.devices))
	for _, d := range m.devices {
		list = append(list, d)
	}
	return list, nil
}

func (m *mockDeviceRepo) Update(ctx context.Context, d device.Device) error {
	m.devices[d.ID] = d
	return nil
}

func (m *mockDeviceRepo) Delete(ctx context.Context, id string) error {
	delete(m.devices, id)
	return nil
}

type mockVault struct {
	creds map[string]device.Credentials
}

func newMockVault() *mockVault {
	return &mockVault{creds: make(map[string]device.Credentials)}
}

func (m *mockVault) Save(ctx context.Context, deviceID string, c device.Credentials) error {
	m.creds[deviceID] = c
	return nil
}

func (m *mockVault) Get(ctx context.Context, deviceID string) (device.Credentials, error) {
	c, ok := m.creds[deviceID]
	if !ok {
		return device.Credentials{}, device.ErrNotFound
	}
	return c, nil
}

func (m *mockVault) Delete(ctx context.Context, deviceID string) error {
	delete(m.creds, deviceID)
	return nil
}

type mockDriver struct {
	executeFn func(ctx context.Context, cmd command.Command) (command.Result, error)
}

func (m *mockDriver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, cmd)
	}
	return command.Result{}, nil
}

func (m *mockDriver) Classify(cmd command.Command) command.Class {
	return command.ClassReadOnly
}

func (m *mockDriver) Translate(op command.Operation) (command.Command, error) {
	return command.Command{}, nil
}

func (m *mockDriver) Close() error {
	return nil
}

func TestManageDeviceUseCase(t *testing.T) {
	repo := newMockDeviceRepo()
	vault := newMockVault()
	uc := NewManageDeviceUseCase(repo, vault, nil)
	ctx := context.Background()

	t.Run("Create and Get Device", func(t *testing.T) {
		dev := device.Device{
			ID:         "dev-1",
			Name:       "Router-Core",
			Vendor:     "mikrotik",
			DriverType: "routeros_api",
			Host:       "192.168.1.1",
			Port:       8728,
		}
		creds := device.Credentials{
			Username: "admin",
			Password: "pass",
		}

		err := uc.CreateDevice(ctx, dev, creds)
		require.NoError(t, err)

		fetched, err := uc.GetDevice(ctx, "dev-1")
		require.NoError(t, err)
		assert.Equal(t, "Router-Core", fetched.Name)

		savedCreds, err := vault.Get(ctx, "dev-1")
		require.NoError(t, err)
		assert.Equal(t, "admin", savedCreds.Username)
	})

	t.Run("List Devices", func(t *testing.T) {
		list, err := uc.ListDevices(ctx)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("TestConnection", func(t *testing.T) {
		driver := &mockDriver{
			executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
				if cmd.Raw == "/system/resource/print" {
					return command.Result{Rows: []map[string]string{
						{"uptime": "3d", "version": "7.10", "board-name": "hEX"},
					}}, nil
				}
				if cmd.Raw == "/system/identity/print" {
					return command.Result{Rows: []map[string]string{
						{"name": "Router-Sim"},
					}}, nil
				}
				return command.Result{}, nil
			},
		}

		res, err := uc.TestConnection(ctx, driver, "dev-1")
		require.NoError(t, err)
		assert.Equal(t, "connected", res.Status)
		assert.Equal(t, "hEX", res.BoardName)
		assert.Equal(t, "Router-Sim", res.Identity)
	})

	t.Run("Delete Device", func(t *testing.T) {
		err := uc.DeleteDevice(ctx, "dev-1")
		require.NoError(t, err)

		_, err = uc.GetDevice(ctx, "dev-1")
		assert.ErrorIs(t, err, device.ErrNotFound)
	})
}
