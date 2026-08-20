package network

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestActiveSessionsUseCase(t *testing.T) {
	// Real driver gateway bound to a policy-bypassing executor: the test
	// asserts command construction/parsing through the full seam without
	// needing the policy gate (mock driver classifies everything read-only).
	exec := func(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
		return driver.Execute(ctx, cmd)
	}
	uc := NewActiveSessionsUseCase(mikrotik.NewGateway(exec))
	ctx := context.Background()

	t.Run("GetPPPActiveSessions", func(t *testing.T) {
		driver := &mockDriver{
			executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
				assert.Equal(t, "/ppp/active/print", cmd.Raw)
				return command.Result{
					Rows: []map[string]string{
						{".id": "*1", "name": "user_ppp1", "address": "192.168.10.2", "caller-id": "AA:BB:CC:DD:EE:FF"},
					},
				}, nil
			},
		}
		sessions, err := uc.GetPPPActiveSessions(ctx, driver)
		require.NoError(t, err)
		require.Len(t, sessions, 1)
		assert.Equal(t, "user_ppp1", sessions[0].Name)
		assert.Equal(t, "192.168.10.2", sessions[0].Address)
	})

	t.Run("GetPPPInactiveSessions", func(t *testing.T) {
		driver := &mockDriver{
			executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
				if cmd.Raw == "/ppp/secret/print" {
					return command.Result{
						Rows: []map[string]string{
							{".id": "*1", "name": "user_online", "service": "pppoe"},
							{".id": "*2", "name": "user_offline", "service": "pppoe"},
						},
					}, nil
				}
				if cmd.Raw == "/ppp/active/print" {
					return command.Result{
						Rows: []map[string]string{
							{".id": "*10", "name": "user_online"},
						},
					}, nil
				}
				return command.Result{}, nil
			},
		}
		inactive, err := uc.GetPPPInactiveSessions(ctx, driver)
		require.NoError(t, err)
		require.Len(t, inactive, 1)
		assert.Equal(t, "user_offline", inactive[0].Name)
	})

	t.Run("GetHotspotInactiveUsers", func(t *testing.T) {
		driver := &mockDriver{
			executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
				if cmd.Raw == "/ip/hotspot/user/print" {
					return command.Result{
						Rows: []map[string]string{
							{".id": "*1", "name": "vouc1"},
							{".id": "*2", "name": "vouc2"},
						},
					}, nil
				}
				if cmd.Raw == "/ip/hotspot/active/print" {
					return command.Result{
						Rows: []map[string]string{
							{".id": "*10", "user": "vouc1"},
						},
					}, nil
				}
				return command.Result{}, nil
			},
		}
		inactive, err := uc.GetHotspotInactiveUsers(ctx, driver)
		require.NoError(t, err)
		require.Len(t, inactive, 1)
		assert.Equal(t, "vouc2", inactive[0].Name)
	})

	t.Run("GetDHCPLeases and SetDHCPLeaseBlock", func(t *testing.T) {
		driver := &mockDriver{
			executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
				if cmd.Raw == "/ip/dhcp-server/lease/print" {
					return command.Result{
						Rows: []map[string]string{
							{".id": "*1", "address": "10.0.0.5", "mac-address": "00:11:22:33:44:55", "status": "bound"},
						},
					}, nil
				}
				if cmd.Raw == "/ip/dhcp-server/lease/set" {
					assert.Equal(t, "*1", cmd.Args[".id"])
					assert.Equal(t, "yes", cmd.Args["blocked"])
					return command.Result{}, nil
				}
				return command.Result{}, nil
			},
		}
		leases, err := uc.GetDHCPLeases(ctx, driver, "")
		require.NoError(t, err)
		require.Len(t, leases, 1)
		assert.Equal(t, "00:11:22:33:44:55", leases[0].MACAddress)

		_, err = uc.SetDHCPLeaseBlock(ctx, driver, "*1", true, "SUSPENDED")
		require.NoError(t, err)
	})
}
