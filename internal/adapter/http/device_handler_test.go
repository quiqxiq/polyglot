package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/business"
	"github.com/stretchr/testify/assert"
)

type mockDevRepo struct {
	devices map[string]device.Device
}

func newMockDevRepo() *mockDevRepo {
	return &mockDevRepo{devices: make(map[string]device.Device)}
}

func (m *mockDevRepo) Save(ctx context.Context, d device.Device) error {
	m.devices[d.ID] = d
	return nil
}

func (m *mockDevRepo) FindByID(ctx context.Context, id string) (device.Device, error) {
	d, ok := m.devices[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}

func (m *mockDevRepo) FindAll(ctx context.Context) ([]device.Device, error) {
	list := make([]device.Device, 0, len(m.devices))
	for _, d := range m.devices {
		list = append(list, d)
	}
	return list, nil
}

func (m *mockDevRepo) Update(ctx context.Context, d device.Device) error {
	m.devices[d.ID] = d
	return nil
}

func (m *mockDevRepo) Delete(ctx context.Context, id string) error {
	delete(m.devices, id)
	return nil
}

type mockVaultImpl struct {
	creds map[string]device.Credentials
}

func newMockVaultImpl() *mockVaultImpl {
	return &mockVaultImpl{creds: make(map[string]device.Credentials)}
}

func (m *mockVaultImpl) Get(ctx context.Context, deviceID string) (device.Credentials, error) {
	c, ok := m.creds[deviceID]
	if !ok {
		return device.Credentials{}, device.ErrNotFound
	}
	return c, nil
}

func (m *mockVaultImpl) Save(ctx context.Context, deviceID string, creds device.Credentials) error {
	m.creds[deviceID] = creds
	return nil
}

func setupDeviceTestRouter(driver port.DeviceDriver) (*gin.Engine, *mockDevRepo) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := newMockDevRepo()
	vault := newMockVaultImpl()
	uc := business.NewManageDeviceUseCase(repo, vault)
	handler := NewDeviceHandler(uc, func(ctx *gin.Context, deviceID string) (port.DeviceDriver, error) {
		return driver, nil
	})
	RegisterDeviceRoutes(r, handler)
	return r, repo
}

func TestDeviceHandler_Routes(t *testing.T) {
	driver := &dummyDriver{
		executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
			if cmd.Raw == "/system/resource/print" {
				return command.Result{Rows: []map[string]string{
					{"uptime": "5d", "version": "7.10", "board-name": "hEX"},
				}}, nil
			}
			return command.Result{}, nil
		},
	}

	r, _ := setupDeviceTestRouter(driver)

	t.Run("POST Create Device", func(t *testing.T) {
		payload := DevicePayload{
			ID:         "dev-100",
			Name:       "Router-Branch-1",
			Vendor:     "mikrotik",
			DriverType: "routeros_api",
			Host:       "192.168.10.1",
			Port:       8728,
			Username:   "admin",
			Password:   "secret",
		}
		jsonBody, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/devices", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Router-Branch-1")
	})

	t.Run("GET List Devices", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "dev-100")
	})

	t.Run("GET Device Details", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/dev-100", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Router-Branch-1")
	})

	t.Run("POST Test Connection", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/devices/dev-100/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "connected")
		assert.Contains(t, w.Body.String(), "hEX")
	})

	t.Run("DELETE Device", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/devices/dev-100", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "deleted")
	})
}
