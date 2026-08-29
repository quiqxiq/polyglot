package hotspot

import (
	"context"

	"github.com/quixiq/polyglot/internal/port"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	"github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceId.
type ConnectDriverProvider func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

// HotspotConnectHandler orchestrates Hotspot ConnectRPC procedures across modular handler files.
type HotspotConnectHandler struct {
	useCase               *hotspotUC.UseCase
	activeSessionsUseCase *network.ActiveSessionsUseCase
	driverProvider        ConnectDriverProvider
}

// NewHotspotConnectHandler constructs a new HotspotConnectHandler. activeUC
// is provided by the composition root — it is not constructed here so the
// adapter never builds usecases or driver gateways itself.
func NewHotspotConnectHandler(uc *hotspotUC.UseCase, activeUC *network.ActiveSessionsUseCase, provider ConnectDriverProvider) *HotspotConnectHandler {
	return &HotspotConnectHandler{
		useCase:               uc,
		activeSessionsUseCase: activeUC,
		driverProvider:        provider,
	}
}

func (h *HotspotConnectHandler) getDriver(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
	if deviceID == "" {
		return nil, response.InvalidArgument("device_id is required")
	}
	driver, err := h.driverProvider(ctx, deviceID)
	if err != nil {
		return nil, response.MapDomainError(fault.Wrap(fault.KindUnavailable, err))
	}
	return driver, nil
}
