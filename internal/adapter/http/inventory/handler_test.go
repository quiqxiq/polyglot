package inventory

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/device"
)

type mockDeviceRepo struct {
	devices    []device.Device
	findAllErr error
}

func (m *mockDeviceRepo) Save(_ context.Context, _ device.Device) error { return nil }
func (m *mockDeviceRepo) FindByID(_ context.Context, _ string) (device.Device, error) {
	return device.Device{}, device.ErrNotFound
}
func (m *mockDeviceRepo) FindAll(_ context.Context) ([]device.Device, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	return m.devices, nil
}
func (m *mockDeviceRepo) Update(_ context.Context, _ device.Device) error { return nil }
func (m *mockDeviceRepo) Delete(_ context.Context, _ string) error        { return nil }
func (m *mockDeviceRepo) FindByUserScope(_ context.Context, _ uint) ([]device.Device, error) {
	return m.devices, nil
}

// startTCPListener binds an ephemeral loopback port that accepts connections
// for the lifetime of the test.
func startTCPListener(t *testing.T) (string, int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	addr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)
	return addr.IP.String(), addr.Port
}

// closedTCPPort reserves an ephemeral loopback port, then releases it so a
// dial is refused for the lifetime of the test.
func closedTCPPort(t *testing.T) (string, int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)
	require.NoError(t, listener.Close())
	return addr.IP.String(), addr.Port
}

func listDevices(t *testing.T, repo *mockDeviceRepo) []DeviceDTO {
	t.Helper()
	handler := NewHandler(repo)
	mux := http.NewServeMux()
	handler.RegisterProtected(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Devices []DeviceDTO `json:"devices"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp.Devices
}

func TestListDevicesHTTP(t *testing.T) {
	host, port := startTCPListener(t)
	repo := &mockDeviceRepo{
		devices: []device.Device{
			{
				ID:         "dev-01",
				Name:       "Core-Router",
				Vendor:     "MikroTik",
				DriverType: "mikrotik",
				Host:       host,
				Port:       port,
				Enabled:    true,
				Tags:       []string{"core"},
			},
		},
	}

	devices := listDevices(t, repo)

	require.Len(t, devices, 1)
	assert.Equal(t, "dev-01", devices[0].ID)
	assert.Equal(t, "active", devices[0].Status)
	assert.Equal(t, "Core-Router", devices[0].Name)
	assert.Equal(t, "online", devices[0].Connectivity)
	require.NotNil(t, devices[0].LatencyMS)
	assert.GreaterOrEqual(t, *devices[0].LatencyMS, int64(0))
}

func TestListDevicesHTTP_ConnectivityStates(t *testing.T) {
	onlineHost, onlinePort := startTCPListener(t)
	offlineHost, offlinePort := closedTCPPort(t)

	repo := &mockDeviceRepo{
		devices: []device.Device{
			{ID: "online", Host: onlineHost, Port: onlinePort, Enabled: true},
			{ID: "offline", Host: offlineHost, Port: offlinePort, Enabled: true},
			{ID: "disabled", Host: onlineHost, Port: onlinePort, Enabled: false},
			{ID: "no-host", Enabled: true},
		},
	}

	devices := listDevices(t, repo)
	require.Len(t, devices, 4)

	byID := make(map[string]DeviceDTO, len(devices))
	for _, d := range devices {
		byID[d.ID] = d
	}

	assert.Equal(t, "online", byID["online"].Connectivity)
	require.NotNil(t, byID["online"].LatencyMS)
	assert.GreaterOrEqual(t, *byID["online"].LatencyMS, int64(0))
	assert.Equal(t, "offline", byID["offline"].Connectivity)
	assert.Nil(t, byID["offline"].LatencyMS)

	// A disabled device is not probed: its registration state answers the
	// question, so connectivity stays unknown even when the port is open.
	assert.Equal(t, "disabled", byID["disabled"].Status)
	assert.Equal(t, "unknown", byID["disabled"].Connectivity)

	// A device with no management host cannot be reached.
	assert.Equal(t, "active", byID["no-host"].Status)
	assert.Equal(t, "offline", byID["no-host"].Connectivity)
}

func TestListDevicesHTTP_Error(t *testing.T) {
	repo := &mockDeviceRepo{
		findAllErr: device.ErrNotFound,
	}

	handler := NewHandler(repo)
	mux := http.NewServeMux()
	handler.RegisterProtected(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	assert.Equal(t, "resource not found", resp.Error.Message)
}
