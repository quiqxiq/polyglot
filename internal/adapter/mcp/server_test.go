package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
)

// testDriver is a configurable port.DeviceDriver for MCP tool tests.
// Each field controls one aspect of the driver's behavior, so individual
// tests can set only what they need.
type testDriver struct {
	executeFn  func(ctx context.Context, cmd command.Command) (command.Result, error)
	classifyFn func(cmd command.Command) command.Classification
	closed     bool
}

func (d *testDriver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	if d.executeFn != nil {
		return d.executeFn(ctx, cmd)
	}
	return command.Result{Output: "ok"}, nil
}

func (d *testDriver) Classify(cmd command.Command) command.Classification {
	if d.classifyFn != nil {
		return d.classifyFn(cmd)
	}
	return command.Classification{Mode: command.ModeRead}
}

func (d *testDriver) Translate(op command.Operation) (command.Command, error) {
	if op == command.OpGetStatus {
		return command.Command{Raw: "get_status"}, nil
	}
	return command.Command{}, errors.New("unsupported operation")
}

func (d *testDriver) Close() error {
	d.closed = true
	return nil
}

// newTestServer builds an MCP Server backed by a registry holding one
// device ("dev1") with the given testDriver. The driver is returned so
// tests can assert on its state (e.g. Close was called).
func newTestServer(t *testing.T, drv port.DeviceDriver) *Server {
	t.Helper()
	return newTestServerWithDevice(t, "dev1", drv)
}

func newTestServerWithDevice(t *testing.T, deviceID string, drv port.DeviceDriver) *Server {
	t.Helper()
	repo := &memDeviceRepo{
		devices: map[string]device.Device{
			deviceID: {ID: deviceID, Name: "test", Vendor: "mikrotik", DriverType: "mikrotik", Host: "10.0.0.1", Port: 8728, Enabled: true},
		},
	}
	vault := &memCredVault{
		creds: map[string]device.Credentials{
			deviceID: {Username: "admin", Password: "secret"},
		},
	}
	factories := map[string]registry.DriverFactory{
		"mikrotik": func(_ context.Context, _ device.Target) (port.DeviceDriver, error) { return drv, nil },
	}
	reg := registry.New(repo, vault, factories)
	return New(reg, nil)
}

// memDeviceRepo and memCredVault are in-memory test implementations of
// port.DeviceRepository and port.CredentialVault.
type memDeviceRepo struct {
	devices map[string]device.Device
}

func (r *memDeviceRepo) Save(_ context.Context, d device.Device) error {
	r.devices[d.ID] = d
	return nil
}

func (r *memDeviceRepo) FindByID(_ context.Context, id string) (device.Device, error) {
	d, ok := r.devices[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}

func (r *memDeviceRepo) FindAll(_ context.Context) ([]device.Device, error) {
	list := make([]device.Device, 0, len(r.devices))
	for _, d := range r.devices {
		list = append(list, d)
	}
	return list, nil
}

func (r *memDeviceRepo) Update(ctx context.Context, d device.Device) error {
	return r.Save(ctx, d)
}

func (r *memDeviceRepo) FindByUserScope(ctx context.Context, userID uint) ([]device.Device, error) {
	return r.FindAll(ctx)
}

func (r *memDeviceRepo) Delete(_ context.Context, id string) error {
	delete(r.devices, id)
	return nil
}

type memCredVault struct {
	creds map[string]device.Credentials
}

func (v *memCredVault) Get(_ context.Context, deviceID string) (device.Credentials, error) {
	c, ok := v.creds[deviceID]
	if !ok {
		return device.Credentials{}, device.ErrNotFound
	}
	return c, nil
}

func (v *memCredVault) Save(_ context.Context, deviceID string, c device.Credentials) error {
	v.creds[deviceID] = c
	return nil
}

func (v *memCredVault) EncryptString(_ context.Context, plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func (v *memCredVault) DecryptString(_ context.Context, ciphertext string) (string, error) {
	if len(ciphertext) > 4 && ciphertext[:4] == "enc:" {
		return ciphertext[4:], nil
	}
	return ciphertext, nil
}

// callToolViaInMemoryTransport creates an in-memory MCP client-server pair,
// calls a tool, and returns the result. This is a real MCP round-trip (not
// a mock) — it exercises tool registration, schema generation, JSON-RPC
// serialization, and structured output, all through the SDK's transport
// layer. The server runs in a goroutine; the test blocks on the client's
// CallTool.
func callToolViaInMemoryTransport(t *testing.T, s *Server, toolName string, args map[string]any) *mcp.CallToolResult {
	t.Helper()

	// The MCP server is JWT-protected in production. In-memory transport has
	// no HTTP headers, so inject the same identity middleware would set.
	baseCtx := iauth.WithIdentity(context.Background(), 1, []string{"admin"})
	ctx, cancel := context.WithCancel(baseCtx)
	t.Cleanup(cancel)

	trServer, trClient := mcp.NewInMemoryTransports()
	go func() { _ = s.mcpServer.Run(ctx, trServer) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.1"}, nil)
	session, err := client.Connect(ctx, trClient, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: args})
	require.NoError(t, err)
	return res
}

func TestMCP_ListTools(t *testing.T) {
	drv := &testDriver{}
	s := newTestServer(t, drv)

	// The MCP server is JWT-protected in production. In-memory transport has
	// no HTTP headers, so inject the same identity middleware would set.
	baseCtx := iauth.WithIdentity(context.Background(), 1, []string{"admin"})
	ctx, cancel := context.WithCancel(baseCtx)
	t.Cleanup(cancel)

	trServer, trClient := mcp.NewInMemoryTransports()
	go func() { _ = s.mcpServer.Run(ctx, trServer) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	session, err := client.Connect(ctx, trClient, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	names := make([]string, len(tools.Tools))
	for i, tool := range tools.Tools {
		names[i] = tool.Name
	}
	assert.ElementsMatch(t, []string{
		"get_device_status", "run_command", "push_config",
		"propose_command", "propose_config", "get_approval", "list_approvals", "cancel_approval",
		"get_active_device", "router_capabilities", "router_read",
		"inspect_children", "inspect_syntax", "inspect_completion", "inspect_validate",
		"mikhmon_get_dashboard",
		"customer_lookup", "list_skills",
	}, names)
}

func TestMCP_GetDeviceStatus_E2E(t *testing.T) {
	drv := &testDriver{
		executeFn: func(_ context.Context, _ command.Command) (command.Result, error) {
			return command.Result{
				Rows: []map[string]string{
					{"uptime": "1d2h", "version": "7.15"},
				},
			}, nil
		},
	}
	s := newTestServer(t, drv)

	res := callToolViaInMemoryTransport(t, s, "get_device_status", map[string]any{
		"device_id": "dev1",
	})

	assert.False(t, res.IsError)
	sc, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok, "StructuredContent must be a map")
	assert.Equal(t, "dev1", sc["device_id"])
	assert.Equal(t, "online", sc["status"])
	assert.Contains(t, sc["summary"].(string), "uptime")
	assert.Contains(t, sc["summary"].(string), "version")
}

func TestMCP_GetDeviceStatus_NotFound_E2E(t *testing.T) {
	drv := &testDriver{}
	s := newTestServer(t, drv)

	res := callToolViaInMemoryTransport(t, s, "get_device_status", map[string]any{
		"device_id": "missing",
	})

	assert.True(t, res.IsError)
	sc, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "error", sc["status"])
}

func TestMCP_RunCommand_E2E(t *testing.T) {
	drv := &testDriver{
		executeFn: func(_ context.Context, _ command.Command) (command.Result, error) {
			return command.Result{
				Rows: []map[string]string{
					{"name": "ether1", "running": "true"},
					{"name": "ether2", "running": "false"},
				},
			}, nil
		},
	}
	s := newTestServer(t, drv)

	res := callToolViaInMemoryTransport(t, s, "run_command", map[string]any{
		"device_id": "dev1",
		"command":   "/interface/print",
	})

	assert.False(t, res.IsError)
	sc, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "dev1", sc["device_id"])
	assert.Equal(t, "/interface/print", sc["command"])
	assert.Equal(t, true, sc["executed"])
	assert.Equal(t, float64(2), sc["row_count"])
}

func TestMCP_PushConfig_E2E(t *testing.T) {
	drv := &testDriver{
		executeFn: func(_ context.Context, _ command.Command) (command.Result, error) {
			return command.Result{Output: "config saved"}, nil
		},
	}
	s := newTestServer(t, drv)

	res := callToolViaInMemoryTransport(t, s, "push_config", map[string]any{
		"device_id": "dev1",
		"command":   "setParameterValues",
		"args":      map[string]any{"parameterValues": `[["Device.X.Y","hello","xsd:string"]]`},
	})

	// M1: configuration changes are proposed, never applied by the tool.
	assert.True(t, res.IsError)
	sc, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "dev1", sc["device_id"])
	assert.Equal(t, false, sc["applied"])
	assert.Equal(t, true, sc["approval_required"])
}

func TestMCP_RunCommand_Error_E2E(t *testing.T) {
	drv := &testDriver{
		executeFn: func(_ context.Context, _ command.Command) (command.Result, error) {
			return command.Result{}, errors.New("device unreachable")
		},
	}
	s := newTestServer(t, drv)

	res := callToolViaInMemoryTransport(t, s, "run_command", map[string]any{
		"device_id": "dev1",
		"command":   "/system/resource/print",
	})

	assert.True(t, res.IsError)
	sc, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok)
	assert.Contains(t, sc["error"].(string), "unreachable")
}

func TestMCP_RunCommand_MutationIsNeverExecuted(t *testing.T) {
	executed := false
	drv := &testDriver{
		classifyFn: func(cmd command.Command) command.Classification {
			if cmd.Raw == "/system/reboot" {
				return command.Classification{Mode: command.ModeMutation}
			}
			return command.Classification{Mode: command.ModeRead}
		},
		executeFn: func(_ context.Context, _ command.Command) (command.Result, error) {
			executed = true
			return command.Result{Output: "rebooted"}, nil
		},
	}
	s := newTestServer(t, drv)

	res := callToolViaInMemoryTransport(t, s, "run_command", map[string]any{
		"device_id": "dev1",
		"command":   "/system/reboot",
	})

	assert.True(t, res.IsError)
	assert.False(t, executed, "a MUTATION must never reach driver.Execute from a tool call")
	sc, ok := res.StructuredContent.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "MUTATION", sc["mode"])
	assert.Equal(t, false, sc["executed"])
	assert.Equal(t, true, sc["approval_required"])
}

func TestMCP_RunCommand_DeniedNeverExecuted(t *testing.T) {
	executed := false
	drv := &testDriver{
		classifyFn: func(cmd command.Command) command.Classification {
			return command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: /system/reset-configuration"}
		},
		executeFn: func(_ context.Context, _ command.Command) (command.Result, error) {
			executed = true
			return command.Result{}, nil
		},
	}
	s := newTestServer(t, drv)

	res := callToolViaInMemoryTransport(t, s, "run_command", map[string]any{
		"device_id": "dev1",
		"command":   "/system/reset-configuration",
	})

	assert.True(t, res.IsError)
	assert.False(t, executed)
	sc, _ := res.StructuredContent.(map[string]any)
	assert.Contains(t, sc["error"].(string), "denied")
}
