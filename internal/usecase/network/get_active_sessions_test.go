package network

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
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

type mockSessionGateway struct {
	port.SessionGateway
	listPPPActiveFn            func(ctx context.Context, driver port.DeviceDriver) ([]port.PPPActiveSession, error)
	listPPPInactiveFn          func(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error)
	listHotspotInactiveUsersFn func(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUser, error)
	listDHCPLeasesFn           func(ctx context.Context, driver port.DeviceDriver, macFilter string) ([]port.DHCPLease, error)
	setDHCPLeaseBlockFn        func(ctx context.Context, driver port.DeviceDriver, rosID string, blocked bool, comment string) (command.Result, error)
}

func (m *mockSessionGateway) ListPPPActive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPActiveSession, error) {
	if m.listPPPActiveFn != nil {
		return m.listPPPActiveFn(ctx, driver)
	}
	return nil, nil
}

func (m *mockSessionGateway) ListPPPInactive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error) {
	if m.listPPPInactiveFn != nil {
		return m.listPPPInactiveFn(ctx, driver)
	}
	return nil, nil
}

func (m *mockSessionGateway) ListHotspotInactiveUsers(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUser, error) {
	if m.listHotspotInactiveUsersFn != nil {
		return m.listHotspotInactiveUsersFn(ctx, driver)
	}
	return nil, nil
}

func (m *mockSessionGateway) ListDHCPLeases(ctx context.Context, driver port.DeviceDriver, macFilter string) ([]port.DHCPLease, error) {
	if m.listDHCPLeasesFn != nil {
		return m.listDHCPLeasesFn(ctx, driver, macFilter)
	}
	return nil, nil
}

func (m *mockSessionGateway) SetDHCPLeaseBlock(ctx context.Context, driver port.DeviceDriver, rosID string, blocked bool, comment string) (command.Result, error) {
	if m.setDHCPLeaseBlockFn != nil {
		return m.setDHCPLeaseBlockFn(ctx, driver, rosID, blocked, comment)
	}
	return command.Result{}, nil
}

func TestActiveSessionsUseCase(t *testing.T) {
	ctx := context.Background()
	driver := &mockDriver{}

	t.Run("GetPPPActiveSessions", func(t *testing.T) {
		gw := &mockSessionGateway{
			listPPPActiveFn: func(ctx context.Context, drv port.DeviceDriver) ([]port.PPPActiveSession, error) {
				return []port.PPPActiveSession{
					{RosID: "*1", Name: "user_ppp1", Address: "192.168.10.2", Profile: "10Mbps_Plan"},
				}, nil
			},
		}
		uc := NewActiveSessionsUseCase(gw)
		sessions, err := uc.GetPPPActiveSessions(ctx, driver)
		require.NoError(t, err)
		require.Len(t, sessions, 1)
		assert.Equal(t, "user_ppp1", sessions[0].Name)
		assert.Equal(t, "192.168.10.2", sessions[0].Address)
		assert.Equal(t, "10Mbps_Plan", sessions[0].Profile)
	})

	t.Run("GetPPPInactiveSessions", func(t *testing.T) {
		gw := &mockSessionGateway{
			listPPPInactiveFn: func(ctx context.Context, drv port.DeviceDriver) ([]port.PPPoESecret, error) {
				return []port.PPPoESecret{
					{RosID: "*2", Name: "user_offline", Service: "pppoe"},
				}, nil
			},
		}
		uc := NewActiveSessionsUseCase(gw)
		inactive, err := uc.GetPPPInactiveSessions(ctx, driver)
		require.NoError(t, err)
		require.Len(t, inactive, 1)
		assert.Equal(t, "user_offline", inactive[0].Name)
	})

	t.Run("GetHotspotInactiveUsers", func(t *testing.T) {
		gw := &mockSessionGateway{
			listHotspotInactiveUsersFn: func(ctx context.Context, drv port.DeviceDriver) ([]port.HotspotUser, error) {
				return []port.HotspotUser{
					{RosID: "*2", Name: "vouc2"},
				}, nil
			},
		}
		uc := NewActiveSessionsUseCase(gw)
		inactive, err := uc.GetHotspotInactiveUsers(ctx, driver)
		require.NoError(t, err)
		require.Len(t, inactive, 1)
		assert.Equal(t, "vouc2", inactive[0].Name)
	})

	t.Run("GetDHCPLeases and SetDHCPLeaseBlock", func(t *testing.T) {
		blockCalled := false
		gw := &mockSessionGateway{
			listDHCPLeasesFn: func(ctx context.Context, drv port.DeviceDriver, macFilter string) ([]port.DHCPLease, error) {
				return []port.DHCPLease{
					{RosID: "*1", Address: "10.0.0.5", MACAddress: "00:11:22:33:44:55", Status: "bound"},
				}, nil
			},
			setDHCPLeaseBlockFn: func(ctx context.Context, drv port.DeviceDriver, rosID string, blocked bool, comment string) (command.Result, error) {
				assert.Equal(t, "*1", rosID)
				assert.True(t, blocked)
				assert.Equal(t, "SUSPENDED", comment)
				blockCalled = true
				return command.Result{}, nil
			},
		}
		uc := NewActiveSessionsUseCase(gw)
		leases, err := uc.GetDHCPLeases(ctx, driver, "")
		require.NoError(t, err)
		require.Len(t, leases, 1)
		assert.Equal(t, "00:11:22:33:44:55", leases[0].MACAddress)

		_, err = uc.SetDHCPLeaseBlock(ctx, driver, "*1", true, "SUSPENDED")
		require.NoError(t, err)
		assert.True(t, blockCalled)
	})
}
