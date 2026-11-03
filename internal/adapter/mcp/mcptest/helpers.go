package mcptest

import (
	"context"
	"errors"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
)

// TestDriver is a configurable port.DeviceDriver for MCP tool tests.
type TestDriver struct {
	ExecuteFn  func(ctx context.Context, cmd command.Command) (command.Result, error)
	ClassifyFn func(cmd command.Command) command.Classification
	Closed     bool
}

// Execute runs the test driver's configured execute function.
func (d *TestDriver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	if d.ExecuteFn != nil {
		return d.ExecuteFn(ctx, cmd)
	}
	return command.Result{Output: "ok"}, nil
}

// Classify runs the test driver's configured classify function.
func (d *TestDriver) Classify(cmd command.Command) command.Classification {
	if d.ClassifyFn != nil {
		return d.ClassifyFn(cmd)
	}
	return command.Classification{Mode: command.ModeRead}
}

// Translate handles operation translation for the test driver.
func (d *TestDriver) Translate(op command.Operation) (command.Command, error) {
	if op == command.OpGetStatus {
		return command.Command{Raw: "get_status"}, nil
	}
	return command.Command{}, errors.New("unsupported operation")
}

// Close marks the test driver as closed.
func (d *TestDriver) Close() error {
	d.Closed = true
	return nil
}

// MemDeviceRepo is an in-memory implementation of port.DeviceRepository.
type MemDeviceRepo struct {
	Devices map[string]device.Device
}

// NewMemDeviceRepo creates a new MemDeviceRepo with the given device map.
func NewMemDeviceRepo(devices map[string]device.Device) *MemDeviceRepo {
	if devices == nil {
		devices = make(map[string]device.Device)
	}
	return &MemDeviceRepo{Devices: devices}
}

// Save stores or updates a device in memory.
func (r *MemDeviceRepo) Save(_ context.Context, d device.Device) error {
	r.Devices[d.ID] = d
	return nil
}

// FindByID returns a device by its ID from memory.
func (r *MemDeviceRepo) FindByID(_ context.Context, id string) (device.Device, error) {
	d, ok := r.Devices[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}

// FindAll returns all devices stored in memory.
func (r *MemDeviceRepo) FindAll(_ context.Context) ([]device.Device, error) {
	list := make([]device.Device, 0, len(r.Devices))
	for _, d := range r.Devices {
		list = append(list, d)
	}
	return list, nil
}

// Update updates an existing device in memory.
func (r *MemDeviceRepo) Update(ctx context.Context, d device.Device) error {
	return r.Save(ctx, d)
}

// FindByUserScope returns all devices in scope for the given user ID.
func (r *MemDeviceRepo) FindByUserScope(ctx context.Context, _ uint) ([]device.Device, error) {
	return r.FindAll(ctx)
}

// Delete removes a device from memory.
func (r *MemDeviceRepo) Delete(_ context.Context, id string) error {
	delete(r.Devices, id)
	return nil
}

// MemCredVault is an in-memory implementation of port.CredentialVault.
type MemCredVault struct {
	Creds map[string]device.Credentials
}

// NewMemCredVault creates a new MemCredVault with the given credential map.
func NewMemCredVault(creds map[string]device.Credentials) *MemCredVault {
	if creds == nil {
		creds = make(map[string]device.Credentials)
	}
	return &MemCredVault{Creds: creds}
}

// Get returns credentials for the specified device ID from memory.
func (v *MemCredVault) Get(_ context.Context, deviceID string) (device.Credentials, error) {
	c, ok := v.Creds[deviceID]
	if !ok {
		return device.Credentials{}, device.ErrNotFound
	}
	return c, nil
}

// Save stores credentials for the specified device ID in memory.
func (v *MemCredVault) Save(_ context.Context, deviceID string, c device.Credentials) error {
	v.Creds[deviceID] = c
	return nil
}

// EncryptString provides mock encryption for tests.
func (v *MemCredVault) EncryptString(_ context.Context, plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

// DecryptString provides mock decryption for tests.
func (v *MemCredVault) DecryptString(_ context.Context, ciphertext string) (string, error) {
	if len(ciphertext) > 4 && ciphertext[:4] == "enc:" {
		return ciphertext[4:], nil
	}
	return ciphertext, nil
}
