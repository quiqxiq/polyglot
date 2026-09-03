package network_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	connectNetwork "github.com/quixiq/polyglot/internal/adapter/connect/network"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

type mockDriver struct{}

func (m *mockDriver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
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

type mockHotspotGateway struct {
	port.HotspotGateway
}

type mockSessionGateway struct {
	port.SessionGateway
	leases []port.DHCPLease
}

func (m *mockSessionGateway) ListDHCPLeases(ctx context.Context, driver port.DeviceDriver, server string) ([]port.DHCPLease, error) {
	return m.leases, nil
}

func TestNetworkConnectHandler_ListDHCPLeases(t *testing.T) {
	fakeDriver := &mockDriver{}
	activeGW := &mockSessionGateway{
		leases: []port.DHCPLease{
			{
				RosID:      "*1",
				Address:    "192.168.88.10",
				MACAddress: "AA:BB:CC:DD:EE:FF",
				HostName:   "android-phone",
				Status:     "bound",
			},
		},
	}

	hotUC := hotspotUC.New(&mockHotspotGateway{})
	activeUC := networkUC.NewActiveSessionsUseCase(activeGW)

	provider := func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
		return fakeDriver, nil
	}

	handler := connectNetwork.NewNetworkConnectHandler(hotUC, activeUC, provider)
	ctx := context.Background()

	resp, err := handler.ListDHCPLeases(ctx, connect.NewRequest(&devicepb.ListDHCPLeasesRequest{
		DeviceId: "dev-1",
	}))
	require.NoError(t, err)
	require.Len(t, resp.Msg.Leases, 1)
	assert.Equal(t, "192.168.88.10", resp.Msg.Leases[0].Address)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", resp.Msg.Leases[0].MacAddress)
}
