package inventory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/device"
)

type mockDeviceRepo struct {
	devices []device.Device
}

func (m *mockDeviceRepo) Save(_ context.Context, _ device.Device) error { return nil }
func (m *mockDeviceRepo) FindByID(_ context.Context, _ string) (device.Device, error) {
	return device.Device{}, device.ErrNotFound
}
func (m *mockDeviceRepo) FindAll(_ context.Context) ([]device.Device, error) {
	return m.devices, nil
}
func (m *mockDeviceRepo) Update(_ context.Context, _ device.Device) error { return nil }
func (m *mockDeviceRepo) Delete(_ context.Context, _ string) error        { return nil }
func (m *mockDeviceRepo) FindByUserScope(_ context.Context, _ uint) ([]device.Device, error) {
	return m.devices, nil
}

func TestListDevicesHTTP(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: []device.Device{
			{
				ID:         "dev-01",
				Name:       "Core-Router",
				Vendor:     "MikroTik",
				DriverType: "mikrotik",
				Host:       "10.0.0.1",
				Port:       8728,
				Enabled:    true,
				Tags:       []string{"core"},
			},
		},
	}

	handler := NewHandler(repo)
	mux := http.NewServeMux()
	handler.RegisterPublic(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Devices []DeviceDTO `json:"devices"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Devices, 1)
	assert.Equal(t, "dev-01", resp.Devices[0].ID)
	assert.Equal(t, "active", resp.Devices[0].Status)
	assert.Equal(t, "Core-Router", resp.Devices[0].Name)
}
