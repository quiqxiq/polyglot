package device

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
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

func (m *mockDeviceRepo) FindByUserScope(ctx context.Context, userID uint) ([]device.Device, error) {
	if userID == 2 {
		if d, ok := m.devices["dev-1"]; ok {
			return []device.Device{d}, nil
		}
	}
	return []device.Device{}, nil
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

func (m *mockVault) EncryptString(_ context.Context, plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func (m *mockVault) DecryptString(_ context.Context, ciphertext string) (string, error) {
	if len(ciphertext) > 4 && ciphertext[:4] == "enc:" {
		return ciphertext[4:], nil
	}
	return ciphertext, nil
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

type mockDeviceDiagnostics struct {
	getSystemResourceFn  func(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error)
	getSystemIdentityFn  func(ctx context.Context, driver port.DeviceDriver) (string, error)
	listInterfacesFn     func(ctx context.Context, driver port.DeviceDriver, typeFilter, nameFilter string) ([]port.Interface, error)
	monitorTrafficOnceFn func(ctx context.Context, driver port.DeviceDriver, ifaceName string) (port.InterfaceTrafficStats, error)
}

func (m *mockDeviceDiagnostics) GetSystemResource(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error) {
	if m.getSystemResourceFn != nil {
		return m.getSystemResourceFn(ctx, driver)
	}
	return port.SystemResource{}, nil
}

func (m *mockDeviceDiagnostics) GetSystemIdentity(ctx context.Context, driver port.DeviceDriver) (string, error) {
	if m.getSystemIdentityFn != nil {
		return m.getSystemIdentityFn(ctx, driver)
	}
	return "", nil
}

func (m *mockDeviceDiagnostics) ListInterfaces(ctx context.Context, driver port.DeviceDriver, typeFilter, nameFilter string) ([]port.Interface, error) {
	if m.listInterfacesFn != nil {
		return m.listInterfacesFn(ctx, driver, typeFilter, nameFilter)
	}
	return nil, nil
}

func (m *mockDeviceDiagnostics) MonitorTrafficOnce(ctx context.Context, driver port.DeviceDriver, ifaceName string) (port.InterfaceTrafficStats, error) {
	if m.monitorTrafficOnceFn != nil {
		return m.monitorTrafficOnceFn(ctx, driver, ifaceName)
	}
	return port.InterfaceTrafficStats{}, nil
}

func TestManageDeviceUseCase(t *testing.T) {
	repo := newMockDeviceRepo()
	vault := newMockVault()
	diag := &mockDeviceDiagnostics{
		getSystemResourceFn: func(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error) {
			return port.SystemResource{
				Uptime:    "3d",
				Version:   "7.10",
				BoardName: "hEX",
			}, nil
		},
		getSystemIdentityFn: func(ctx context.Context, driver port.DeviceDriver) (string, error) {
			return "Router-Sim", nil
		},
	}
	uc := NewManageDeviceUseCase(repo, vault, nil, diag)
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

	t.Run("List Devices For User Scoping", func(t *testing.T) {
		// Owner sees all
		ownerList, err := uc.ListDevicesForUser(ctx, 1, []string{"owner"})
		require.NoError(t, err)
		assert.Len(t, ownerList, 1)

		// Admin with assigned router sees dev-1
		adminList, err := uc.ListDevicesForUser(ctx, 2, []string{"admin"})
		require.NoError(t, err)
		assert.Len(t, adminList, 1)

		// Teknisi with 0 assigned routers sees 0
		tekList, err := uc.ListDevicesForUser(ctx, 3, []string{"teknisi"})
		require.NoError(t, err)
		assert.Empty(t, tekList)
	})

	t.Run("TestConnection", func(t *testing.T) {
		driver := &mockDriver{}

		res, err := uc.TestConnection(ctx, driver, "dev-1", "", "ether", "")
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
