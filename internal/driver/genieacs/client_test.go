package genieacs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
)

// newTestDriver builds a Driver pointed at an httptest.Server, bypassing
// NewDriver's reachability ping so tests can assert Execute behavior in
// isolation. The server is a real HTTP server (in-process), not a mock
// interface — this matches CLAUDE.md §9's "real logic over mocks" carve-out
// for external HTTP APIs the driver genuinely cannot reach in CI (the
// GenieACS NBI is an external Node.js service).
func newTestDriver(t *testing.T, srv *httptest.Server, deviceID string) *Driver {
	t.Helper()
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return &Driver{
		baseURL:      *u,
		deviceID:     deviceID,
		httpClient:   srv.Client(),
		pollInterval: 5 * time.Millisecond,
		faultChannel: defaultFaultChannel,
	}
}

// writeBody writes body to w in a test handler. Errors are ignored — an
// httptest.ResponseWriter never fails in practice, and errcheck flags bare
// io.WriteString return values.
func writeBody(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	_, _ = io.WriteString(w, body)
}

// pingHandler responds to NewDriver's reachability ping with an empty array.
func pingHandler(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	if r.Method == http.MethodGet && r.URL.Path == "/devices/" {
		writeBody(t, w, `[]`)
		return
	}
	t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
}

func TestNewDriver_MissingDeviceID(t *testing.T) {
	_, err := NewDriver(context.Background(), device.Target{Host: "localhost"})
	require.ErrorIs(t, err, ErrDeviceIDMissing)
}

func TestNewDriver_UnreachableNBI(t *testing.T) {
	// Point at a port that is effectively guaranteed closed.
	_, err := NewDriver(context.Background(), device.Target{
		Host:  "127.0.0.1",
		Port:  1,
		Extra: map[string]string{"device_id": "dev1"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NBI unreachable")
}

// mustPort extracts the port from a httptest server URL as an int.
func mustPort(t *testing.T, u *url.URL) int {
	t.Helper()
	p, err := strconv.Atoi(u.Port())
	require.NoError(t, err)
	return p
}

func TestNewDriver_ReachableDefaultsFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pingHandler(t, w, r)
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	d, err := NewDriver(context.Background(), device.Target{
		Host:  u.Hostname(),
		Port:  mustPort(t, u),
		Extra: map[string]string{"device_id": "dev1"},
	})
	require.NoError(t, err)
	assert.Equal(t, "http", d.baseURL.Scheme)
	assert.Equal(t, u.Host, d.baseURL.Host)
	assert.Equal(t, "dev1", d.deviceID)
	assert.Equal(t, DefaultPollInterval, d.pollInterval)
}

func TestNewDriver_CustomPollInterval(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pingHandler(t, w, r)
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	d, err := NewDriver(context.Background(), device.Target{
		Host:  u.Hostname(),
		Port:  mustPort(t, u),
		Extra: map[string]string{"device_id": "dev1", "poll_interval": "10ms"},
	})
	require.NoError(t, err)
	assert.Equal(t, 10*time.Millisecond, d.pollInterval)
}

func TestExecute_GetDevice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/devices/", r.URL.Path)
		assert.Contains(t, r.URL.Query().Get("query"), "dev1")
		writeBody(t, w, `[{"_id":"dev1","_lastInform":"2024-01-01T00:00:00Z","Device":{"DeviceInfo":{"SerialNumber":{"_value":"SN001"}}}}]`)
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	res, err := d.Execute(context.Background(), command.Command{Raw: "getDevice"})
	require.NoError(t, err)
	require.Len(t, res.Rows, 1)
	assert.Equal(t, "dev1", res.Rows[0]["_id"])
	assert.Equal(t, "SN001", res.Rows[0]["Device.DeviceInfo.SerialNumber._value"])
}

func TestExecute_GetDevice_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[]`)
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "missing")
	_, err := d.Execute(context.Background(), command.Command{Raw: "getDevice"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestExecute_UnsupportedTaskType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	_, err := d.Execute(context.Background(), command.Command{Raw: "bogus"})
	require.ErrorIs(t, err, ErrUnsupportedTaskType)
}

func TestExecute_Reboot_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/devices/dev1/tasks":
			w.WriteHeader(http.StatusOK)
			writeBody(t, w, `{"_id":"task-1","name":"reboot","device":"dev1"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/faults/":
			writeBody(t, w, `[]`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	res, err := d.Execute(context.Background(), command.Command{Raw: "reboot"})
	require.NoError(t, err)
	assert.Equal(t, "task-1", res.Output)
}

func TestExecute_Reboot_202_PollThenDone(t *testing.T) {
	taskPolls := atomic.Int32{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/devices/dev1/tasks":
			w.WriteHeader(http.StatusAccepted)
			writeBody(t, w, `{"_id":"task-7","name":"reboot","device":"dev1"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/tasks/":
			n := taskPolls.Add(1)
			if n <= 1 {
				writeBody(t, w, `[{"_id":"task-7"}]`)
			} else {
				writeBody(t, w, `[]`)
			}
		case r.Method == http.MethodGet && r.URL.Path == "/faults/":
			writeBody(t, w, `[]`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	res, err := d.Execute(context.Background(), command.Command{Raw: "reboot"})
	require.NoError(t, err)
	assert.Equal(t, "task-7", res.Output)
	assert.GreaterOrEqual(t, taskPolls.Load(), int32(2))
}

func TestExecute_202_CtxTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/devices/dev1/tasks":
			w.WriteHeader(http.StatusAccepted)
			writeBody(t, w, `{"_id":"task-9","name":"reboot"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/tasks/":
			writeBody(t, w, `[{"_id":"task-9"}]`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	res, err := d.Execute(ctx, command.Command{Raw: "reboot"})
	require.ErrorIs(t, err, ErrTaskQueued)
	assert.Equal(t, "task-9", res.Output)
}

func TestExecute_TaskFault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/devices/dev1/tasks":
			w.WriteHeader(http.StatusOK)
			writeBody(t, w, `{"_id":"task-f","name":"reboot"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/faults/":
			writeBody(t, w, `[{"_id":"dev1:default","message":"device rejected reboot"}]`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	_, err := d.Execute(context.Background(), command.Command{Raw: "reboot"})
	require.ErrorIs(t, err, ErrTaskFault)
	assert.Contains(t, err.Error(), "device rejected reboot")
}

func TestExecute_GetParameterValues_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/devices/dev1/tasks":
			w.WriteHeader(http.StatusOK)
			writeBody(t, w, `{"_id":"task-pv","name":"getParameterValues"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/faults/":
			writeBody(t, w, `[]`)
		case r.Method == http.MethodGet && r.URL.Path == "/devices/":
			assert.Contains(t, r.URL.Query().Get("projection"), "Device.X.Y")
			writeBody(t, w, `[{"_id":"dev1","Device":{"X":{"Y":{"_value":"hello"}}}}]`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	cmd := command.Command{
		Raw:  "getParameterValues",
		Args: map[string]string{"parameterNames": `["Device.X.Y"]`},
	}
	res, err := d.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "task-pv", res.Output)
	require.Len(t, res.Rows, 1)
	assert.Equal(t, "hello", res.Rows[0]["Device.X.Y._value"])
}

func TestExecute_ConnectionRequestFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/devices/dev1/tasks":
			_, ok := r.URL.Query()["connection_request"]
			assert.True(t, ok, "connection_request flag must be present when requested")
			w.WriteHeader(http.StatusOK)
			writeBody(t, w, `{"_id":"task-cr","name":"reboot"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/faults/":
			writeBody(t, w, `[]`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	cmd := command.Command{
		Raw:  "reboot",
		Args: map[string]string{"connection_request": ""},
	}
	res, err := d.Execute(context.Background(), cmd)
	require.NoError(t, err)
	assert.Equal(t, "task-cr", res.Output)
}

func TestExecute_AuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "secret-key", r.Header.Get("x-api-key"))
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/devices/dev1/tasks":
			w.WriteHeader(http.StatusOK)
			writeBody(t, w, `{"_id":"task-a","name":"reboot"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/faults/":
			writeBody(t, w, `[]`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	d.apiKey = "secret-key"
	_, err := d.Execute(context.Background(), command.Command{Raw: "reboot"})
	require.NoError(t, err)
}

func TestClose_ReleasesIdleConnections(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[]`)
	}))
	defer srv.Close()

	d := newTestDriver(t, srv, "dev1")
	// One request to open an idle connection.
	_, _ = d.Execute(context.Background(), command.Command{Raw: "getDevice"})
	// Close should not error and should release idle sockets.
	require.NoError(t, d.Close())
}
