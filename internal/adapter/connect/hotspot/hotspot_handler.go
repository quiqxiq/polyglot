package hotspot

import (
	"context"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceId.
type ConnectDriverProvider = iconnect.DriverProvider

// HotspotConnectHandler orchestrates Hotspot ConnectRPC procedures across modular handler files.
type HotspotConnectHandler struct {
	useCase        *hotspotUC.UseCase
	driverProvider ConnectDriverProvider
}

// NewHotspotConnectHandler constructs a new HotspotConnectHandler.
func NewHotspotConnectHandler(uc *hotspotUC.UseCase, provider ConnectDriverProvider) *HotspotConnectHandler {
	return &HotspotConnectHandler{
		useCase:        uc,
		driverProvider: provider,
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
