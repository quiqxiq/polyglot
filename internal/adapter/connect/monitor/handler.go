package monitor

import (
	"context"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceId.
type ConnectDriverProvider = iconnect.DriverProvider

// NetworkMonitorConnectHandler implements the NetworkMonitorService ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type NetworkMonitorConnectHandler struct {
	hotspotUC             *hotspotUC.UseCase
	activeSessionsUseCase *networkUC.ActiveSessionsUseCase
	driverProvider        ConnectDriverProvider
}

// NewNetworkMonitorConnectHandler constructs a NetworkMonitorConnectHandler.
func NewNetworkMonitorConnectHandler(
	hsUC *hotspotUC.UseCase,
	activeUC *networkUC.ActiveSessionsUseCase,
	provider ConnectDriverProvider,
) *NetworkMonitorConnectHandler {
	return &NetworkMonitorConnectHandler{
		hotspotUC:             hsUC,
		activeSessionsUseCase: activeUC,
		driverProvider:        provider,
	}
}

func (h *NetworkMonitorConnectHandler) getDriver(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
	if deviceID == "" {
		return nil, response.InvalidArgument("device_id is required")
	}
	driver, err := h.driverProvider(ctx, deviceID)
	if err != nil {
		return nil, response.MapDomainError(fault.Wrap(fault.KindUnavailable, err))
	}
	return driver, nil
}
