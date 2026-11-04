package mikrotik

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/mcp/mcptest"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
)

type mockAuditWriter struct {
	events []port.AuditEvent
}

func (m *mockAuditWriter) Write(_ context.Context, ev port.AuditEvent) error {
	m.events = append(m.events, ev)
	return nil
}

func setupTestRegistry(t *testing.T, deviceID string, drv port.DeviceDriver) *registry.Registry {
	t.Helper()
	repo := mcptest.NewMemDeviceRepo(map[string]device.Device{
		deviceID: {
			ID:         deviceID,
			Name:       "Test-Router",
			Vendor:     "mikrotik",
			DriverType: "mikrotik",
			Host:       "192.168.88.1",
			Port:       8728,
			Enabled:    true,
		},
	})
	vault := mcptest.NewMemCredVault(map[string]device.Credentials{
		deviceID: {Username: "admin", Password: "password"},
	})
	factories := map[string]registry.DriverFactory{
		"mikrotik": func(_ context.Context, _ device.Target) (port.DeviceDriver, error) {
			return drv, nil
		},
	}
	return registry.New(repo, vault, factories)
}

func TestInspectClient_NormalizeHelpers(t *testing.T) {
	// normalizePath
	assert.Equal(t, "", normalizePath(""))
	assert.Equal(t, "", normalizePath("/"))
	assert.Equal(t, "ip,address,add", normalizePath("/ip/address/add"))
	assert.Equal(t, "ip,address,add", normalizePath("ip address add"))
	assert.Equal(t, "ip,address,add", normalizePath("ip,address,add"))
	assert.Equal(t, "ip,firewall,filter", normalizePath("  /ip/firewall/filter/  "))
	assert.Equal(t, "system,resource", normalizePath("system resource"))

	// normalizeConsoleInput
	assert.Equal(t, "/", normalizeConsoleInput(""))
	assert.Equal(t, "/ip address add", normalizeConsoleInput("/ip address add"))
	assert.Equal(t, "/ip address add", normalizeConsoleInput("ip address add"))
	assert.Equal(t, "/interface print", normalizeConsoleInput("interface print"))
	assert.Equal(t, "/system reboot", normalizeConsoleInput("system reboot"))
	assert.Equal(t, "custom test", normalizeConsoleInput("custom test"))

	// buildConsoleLine
	assert.Equal(t, "/ip address print", buildConsoleLine("/ip/address/print", nil))
	assert.Equal(t, "/ip address add address=10.0.0.1/24 interface=ether1",
		buildConsoleLine("/ip/address/add", map[string]string{"address": "10.0.0.1/24", "interface": "ether1"}))
	assert.Equal(t, "/ip hotspot user add comment=\"vip guest\" name=user1",
		buildConsoleLine("/ip/hotspot/user/add", map[string]string{"name": "user1", "comment": "vip guest"}))

	// parseConsoleCommand
	raw, args := parseConsoleCommand("/ip/address/print")
	assert.Equal(t, "/ip/address/print", raw)
	assert.Empty(t, args)

	raw, args = parseConsoleCommand("/ip address add address=192.168.1.1/24 interface=ether1 comment=\"test lan\" disabled")
	assert.Equal(t, "/ip/address/add", raw)
	assert.Equal(t, "192.168.1.1/24", args["address"])
	assert.Equal(t, "ether1", args["interface"])
	assert.Equal(t, "test lan", args["comment"])
	assert.Equal(t, "", args["disabled"])
}

func TestInspectCache(t *testing.T) {
	cache := NewInspectCache(50 * time.Millisecond)

	cache.Set("key1", "val1")
	val, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "val1", val)

	time.Sleep(60 * time.Millisecond)
	_, ok = cache.Get("key1")
	assert.False(t, ok, "key1 should be expired")

	cache.Set("key2", "val2")
	cache.Clear()
	_, ok = cache.Get("key2")
	assert.False(t, ok, "key2 should be cleared")
}

func TestRouterOSListMenu(t *testing.T) {
	executedCommands := 0
	drv := &mcptest.TestDriver{
		ExecuteFn: func(_ context.Context, cmd command.Command) (command.Result, error) {
			executedCommands++
			assert.Equal(t, "/console/inspect", cmd.Raw)
			assert.Equal(t, "child", cmd.Args["request"])
			assert.Equal(t, "ip,address", cmd.Args["path"])

			return command.Result{
				Rows: []map[string]string{
					{"name": "add", "type": "cmd", "help": "add an address"},
					{"name": "print", "type": "cmd", "help": "print addresses"},
					{"name": "pool", "type": "dir", "help": "address pool menu"},
				},
			}, nil
		},
	}

	reg := setupTestRegistry(t, "dev1", drv)
	h := NewHandler(reg, nil)

	// Validation: missing device_id
	res, _, err := h.RouterOSListMenu(context.Background(), nil, routerosListMenuArgs{})
	require.NoError(t, err)
	assert.True(t, res.IsError)

	// First call: device queried, not cached
	res, out, err := h.RouterOSListMenu(context.Background(), nil, routerosListMenuArgs{
		DeviceID: "dev1",
		Path:     "/ip/address",
	})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, "dev1", out.DeviceID)
	assert.Equal(t, "ip,address", out.Path)
	assert.Equal(t, 3, out.Count)
	assert.False(t, out.Cached)
	assert.Equal(t, 1, executedCommands)

	// Second call: served from cache, driver not called
	res2, out2, err := h.RouterOSListMenu(context.Background(), nil, routerosListMenuArgs{
		DeviceID: "dev1",
		Path:     "ip,address",
	})
	require.NoError(t, err)
	assert.False(t, res2.IsError)
	assert.True(t, out2.Cached)
	assert.Equal(t, 3, out2.Count)
	assert.Equal(t, 1, executedCommands, "driver should not be called on cache hit")
}

func TestRouterOSDescribeCommand(t *testing.T) {
	syntaxQueries := 0
	childQueries := 0

	drv := &mcptest.TestDriver{
		ExecuteFn: func(_ context.Context, cmd command.Command) (command.Result, error) {
			assert.Equal(t, "/console/inspect", cmd.Raw)
			req := cmd.Args["request"]
			switch req {
			case "syntax":
				syntaxQueries++
				assert.Equal(t, "ip,address,add", cmd.Args["path"])
				return command.Result{
					Rows: []map[string]string{
						{
							"name":        "address",
							"type":        "ip/netmask",
							"required":    "yes",
							"help":        "IP address with mask",
							"description": "IP address with mask",
						},
						{
							"name":     "interface",
							"type":     "string",
							"required": "yes",
							"help":     "Interface name",
						},
						{
							"name":    "disabled",
							"type":    "enum",
							"values":  "yes,no",
							"default": "no",
						},
					},
				}, nil
			case "child":
				childQueries++
				return command.Result{
					Rows: []map[string]string{
						{"name": "address", "type": "arg"},
						{"name": "interface", "type": "arg"},
						{"name": "comment", "type": "arg", "help": "optional comment"},
					},
				}, nil
			default:
				return command.Result{}, errors.New("unexpected request")
			}
		},
	}

	reg := setupTestRegistry(t, "dev1", drv)
	h := NewHandler(reg, nil)

	// Validation: missing path
	res, _, err := h.RouterOSDescribeCommand(context.Background(), nil, routerosDescribeCommandArgs{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.True(t, res.IsError)

	// First call: queries driver
	res, out, err := h.RouterOSDescribeCommand(context.Background(), nil, routerosDescribeCommandArgs{
		DeviceID: "dev1",
		Path:     "/ip/address/add",
	})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, "ip,address,add", out.Path)
	assert.Equal(t, 4, out.Count) // address, interface, disabled from syntax + comment from child
	assert.False(t, out.Cached)
	assert.Equal(t, 1, syntaxQueries)
	assert.Equal(t, 1, childQueries)

	// Second call: served from cache
	res2, out2, err := h.RouterOSDescribeCommand(context.Background(), nil, routerosDescribeCommandArgs{
		DeviceID: "dev1",
		Path:     "ip/address/add",
	})
	require.NoError(t, err)
	assert.False(t, res2.IsError)
	assert.True(t, out2.Cached)
	assert.Equal(t, 4, out2.Count)
	assert.Equal(t, 1, syntaxQueries)
}

func TestRouterOSSuggest(t *testing.T) {
	drv := &mcptest.TestDriver{
		ExecuteFn: func(_ context.Context, cmd command.Command) (command.Result, error) {
			assert.Equal(t, "/console/inspect", cmd.Raw)
			assert.Equal(t, "completion", cmd.Args["request"])
			assert.Equal(t, "/ip address add interface=", cmd.Args["input"])

			return command.Result{
				Rows: []map[string]string{
					{"completion": "ether1", "style": "none", "preference": "10", "show": "true"},
					{"completion": "ether2", "style": "none", "preference": "10", "show": "true"},
					{"completion": "internal_hidden", "style": "none", "show": "false"},
				},
			}, nil
		},
	}

	reg := setupTestRegistry(t, "dev1", drv)
	h := NewHandler(reg, nil)

	// Validation
	res, _, err := h.RouterOSSuggest(context.Background(), nil, routerosSuggestArgs{})
	require.NoError(t, err)
	assert.True(t, res.IsError)

	// Successful suggestion
	res, out, err := h.RouterOSSuggest(context.Background(), nil, routerosSuggestArgs{
		DeviceID: "dev1",
		Input:    "/ip address add interface=",
	})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, 2, out.Count, "hidden entry must be filtered out")
	assert.Equal(t, "ether1", out.Suggestions[0].Completion)
	assert.Equal(t, "ether2", out.Suggestions[1].Completion)
}

func TestRouterOSValidate(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		driverRows    []map[string]string
		driverErr     error
		expectedValid bool
		expectedError string
	}{
		{
			name:  "valid command syntax",
			input: "/ip address add address=192.168.88.1/24 interface=ether1",
			driverRows: []map[string]string{
				{"error": ""},
			},
			expectedValid: true,
		},
		{
			name:  "invalid netmask syntax error",
			input: "/ip address add address=10.0.0.1/99 interface=ether1",
			driverRows: []map[string]string{
				{"error": "value of netmask out of range (1..32)", "position": "31"},
			},
			expectedValid: false,
			expectedError: "value of netmask out of range (1..32)",
		},
		{
			name:          "routeros v6 fallback",
			input:         "/ip address print",
			driverErr:     errors.New("no such command: /console/inspect"),
			expectedValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			drv := &mcptest.TestDriver{
				ExecuteFn: func(_ context.Context, cmd command.Command) (command.Result, error) {
					if tc.driverErr != nil {
						return command.Result{}, tc.driverErr
					}
					assert.Equal(t, "/console/inspect", cmd.Raw)
					assert.Equal(t, "error", cmd.Args["request"])
					return command.Result{Rows: tc.driverRows}, nil
				},
			}

			reg := setupTestRegistry(t, "dev1", drv)
			h := NewHandler(reg, nil)

			res, out, err := h.RouterOSValidate(context.Background(), nil, routerosValidateArgs{
				DeviceID: "dev1",
				Input:    tc.input,
			})
			require.NoError(t, err)
			assert.False(t, res.IsError)
			assert.Equal(t, tc.expectedValid, out.Valid)
			if !tc.expectedValid {
				assert.Equal(t, tc.expectedError, out.ErrorMessage)
				assert.Equal(t, 31, out.ErrorPosition)
			}
		})
	}
}

func TestRouterOSRun(t *testing.T) {
	t.Run("auto-validate catches error and rejects execution", func(t *testing.T) {
		executedActual := false
		drv := &mcptest.TestDriver{
			ExecuteFn: func(_ context.Context, cmd command.Command) (command.Result, error) {
				if cmd.Raw == "/console/inspect" {
					return command.Result{
						Rows: []map[string]string{
							{"error": "expected valid IP address", "offset": "25"},
						},
					}, nil
				}
				executedActual = true
				return command.Result{Output: "should not be reached"}, nil
			},
		}

		reg := setupTestRegistry(t, "dev1", drv)
		h := NewHandler(reg, nil)

		res, out, err := h.RouterOSRun(context.Background(), nil, routerosRunArgs{
			DeviceID: "dev1",
			Command:  "/ip/address/add",
			Args:     map[string]string{"address": "bad-ip", "interface": "ether1"},
		})
		require.NoError(t, err)
		assert.True(t, res.IsError)
		assert.False(t, out.Validated)
		assert.False(t, out.Executed)
		assert.False(t, executedActual, "driver must not execute when validation fails")
		assert.Contains(t, out.Error, "validation failed: expected valid IP address")
		assert.Contains(t, out.Error, "STOP: Do NOT retry 'routeros_run' or 'run_command' blindly")
		assert.Contains(t, out.Error, "routeros_suggest")
		assert.Equal(t, out.Error, out.Remediation)
	})

	t.Run("auto-validate succeeds and command executes", func(t *testing.T) {
		auditWriter := &mockAuditWriter{}
		actualExecuted := false

		drv := &mcptest.TestDriver{
			ExecuteFn: func(_ context.Context, cmd command.Command) (command.Result, error) {
				if cmd.Raw == "/console/inspect" {
					return command.Result{Rows: []map[string]string{{"error": ""}}}, nil
				}
				actualExecuted = true
				assert.Equal(t, "/ip/address/add", cmd.Raw)
				assert.Equal(t, "192.168.1.1/24", cmd.Args["address"])
				return command.Result{
					Rows: []map[string]string{
						{".id": "*1", "address": "192.168.1.1/24"},
					},
				}, nil
			},
		}

		reg := setupTestRegistry(t, "dev1", drv)
		h := NewHandler(reg, nil, WithAudit(auditWriter))

		res, out, err := h.RouterOSRun(context.Background(), nil, routerosRunArgs{
			DeviceID: "dev1",
			Command:  "/ip/address/add",
			Args:     map[string]string{"address": "192.168.1.1/24"},
		})
		require.NoError(t, err)
		assert.False(t, res.IsError)
		assert.True(t, out.Validated)
		assert.True(t, out.Executed)
		assert.True(t, actualExecuted)
		assert.Equal(t, 1, out.RowCount)
		assert.Len(t, auditWriter.events, 1)
		assert.Equal(t, "dev1", auditWriter.events[0].DeviceID)
	})

	t.Run("console command string is parsed and executed", func(t *testing.T) {
		actualExecuted := false
		autoVal := false

		drv := &mcptest.TestDriver{
			ExecuteFn: func(_ context.Context, cmd command.Command) (command.Result, error) {
				actualExecuted = true
				assert.Equal(t, "/ip/address/add", cmd.Raw)
				assert.Equal(t, "10.0.0.1/24", cmd.Args["address"])
				assert.Equal(t, "ether1", cmd.Args["interface"])
				assert.Equal(t, "lan gateway", cmd.Args["comment"])
				return command.Result{Output: "ok"}, nil
			},
		}

		reg := setupTestRegistry(t, "dev1", drv)
		h := NewHandler(reg, nil)

		res, out, err := h.RouterOSRun(context.Background(), nil, routerosRunArgs{
			DeviceID:     "dev1",
			Command:      "/ip address add address=10.0.0.1/24 interface=ether1 comment=\"lan gateway\"",
			AutoValidate: &autoVal,
		})
		require.NoError(t, err)
		assert.False(t, res.IsError)
		assert.False(t, out.Validated)
		assert.True(t, out.Executed)
		assert.True(t, actualExecuted)
	})
}
