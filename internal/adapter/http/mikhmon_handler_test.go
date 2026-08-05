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
	"github.com/quixiq/polyglot/internal/driver/mikrotik/mikhmon"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/stretchr/testify/assert"
)

type dummyDriver struct {
	executeFn func(ctx context.Context, cmd command.Command) (command.Result, error)
}

func (d *dummyDriver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	if d.executeFn != nil {
		return d.executeFn(ctx, cmd)
	}
	return command.Result{}, nil
}

func (d *dummyDriver) Classify(cmd command.Command) command.Class {
	return command.ClassReadOnly
}

func (d *dummyDriver) Translate(op command.Operation) (command.Command, error) {
	return command.Command{}, nil
}

func (d *dummyDriver) Close() error {
	return nil
}

func setupTestRouter(driver port.DeviceDriver) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	uc := network.NewMikhmonUseCase()
	handler := NewMikhmonHandler(uc, func(ctx *gin.Context, deviceID string) (port.DeviceDriver, error) {
		return driver, nil
	})
	RegisterMikhmonRoutes(r, handler)
	return r
}

func TestMikhmonHandler_Routes(t *testing.T) {
	driver := &dummyDriver{
		executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
			switch cmd.Raw {
			case "/system/resource/print":
				return command.Result{Rows: []map[string]string{
					{"cpu-load": "10", "uptime": "2d", "version": "7.10", "board-name": "hEX"},
				}}, nil
			case "/system/identity/print":
				return command.Result{Rows: []map[string]string{
					{"name": "RouterOS-Test"},
				}}, nil
			case "/ip/pool/print":
				return command.Result{Rows: []map[string]string{
					{".id": "*1", "name": "hs-pool", "ranges": "192.168.1.10-192.168.1.100"},
				}}, nil
			case "/queue/simple/print":
				return command.Result{Rows: []map[string]string{
					{".id": "*1", "name": "parent-q1", "target": "192.168.1.0/24"},
				}}, nil
			default:
				return command.Result{}, nil
			}
		},
	}

	r := setupTestRouter(driver)

	t.Run("GET Dashboard", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/dev-123/mikhmon/dashboard", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "hEX")
		assert.Contains(t, w.Body.String(), "RouterOS-Test")
	})

	t.Run("GET IP Pools", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/dev-123/mikhmon/pools", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "hs-pool")
	})

	t.Run("GET Parent Queues", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/dev-123/mikhmon/parent-queues", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "parent-q1")
	})

	t.Run("POST Generate Vouchers", func(t *testing.T) {
		body := map[string]interface{}{
			"profile": "1Day_10K",
			"count":   2,
		}
		jsonBody, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/devices/dev-123/mikhmon/vouchers/generate", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "vouchers")
	})

	t.Run("POST Create Profile", func(t *testing.T) {
		params := mikhmon.MikhmonProfileParams{
			Name:     "Profile_50K",
			Price:    "50000",
			Validity: "30d",
		}
		jsonBody, _ := json.Marshal(params)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/devices/dev-123/mikhmon/profiles", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})

	t.Run("GET PPP Active & Inactive", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/dev-123/mikhmon/ppp/active", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/dev-123/mikhmon/ppp/inactive", nil)
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})

	t.Run("GET DHCP Leases", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/devices/dev-123/mikhmon/dhcp/leases", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
