package device

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/mcp/mcptest"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

func TestListDevices_Success(t *testing.T) {
	repo := mcptest.NewMemDeviceRepo(map[string]device.Device{
		"dev1": {
			ID:         "dev1",
			Name:       "Router-Core",
			Vendor:     "MikroTik",
			DriverType: "mikrotik",
			Host:       "10.0.0.1",
			Port:       8728,
			Enabled:    true,
			Tags:       []string{"core", "gateway"},
		},
		"dev2": {
			ID:         "dev2",
			Name:       "OLT-ZTE-01",
			Vendor:     "ZTE",
			DriverType: "zteolt",
			Host:       "192.168.1.10",
			Port:       23,
			Enabled:    false,
		},
	})

	h := NewHandler(nil, func() port.DeviceRepository { return repo }, nil)

	// List all
	res, out, err := h.ListDevices(context.Background(), nil, ListDevicesArgs{})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, "success", out.Status)
	assert.Len(t, out.Devices, 2)

	// Filter by vendor
	res, out, err = h.ListDevices(context.Background(), nil, ListDevicesArgs{Vendor: "mikrotik"})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Len(t, out.Devices, 1)
	assert.Equal(t, "dev1", out.Devices[0].ID)
	assert.Equal(t, "active", out.Devices[0].Status)

	// Filter by non-existent
	res, out, err = h.ListDevices(context.Background(), nil, ListDevicesArgs{Vendor: "cisco"})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Empty(t, out.Devices)
	assert.Equal(t, "No devices found matching criteria", out.Summary)
}

func TestListDevices_NoRepoConfigured(t *testing.T) {
	h := NewHandler(nil, nil, nil)
	res, out, err := h.ListDevices(context.Background(), nil, ListDevicesArgs{})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, "Device repository not configured", out.Summary)
}
