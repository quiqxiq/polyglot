package ppp

import (
	"context"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	pppUC "github.com/quixiq/polyglot/internal/usecase/ppp"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceId.
type ConnectDriverProvider = iconnect.DriverProvider

// PPPConnectHandler orchestrates ConnectRPC procedures for the PPPService.
type PPPConnectHandler struct {
	useCase        *pppUC.UseCase
	driverProvider ConnectDriverProvider
}

// NewPPPConnectHandler constructs a new PPPConnectHandler.
func NewPPPConnectHandler(uc *pppUC.UseCase, provider ConnectDriverProvider) *PPPConnectHandler {
	return &PPPConnectHandler{
		useCase:        uc,
		driverProvider: provider,
	}
}

func (h *PPPConnectHandler) getDriver(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
	if deviceID == "" {
		return nil, response.InvalidArgument("device_id is required")
	}
	driver, err := h.driverProvider(ctx, deviceID)
	if err != nil {
		return nil, response.MapDomainError(fault.Wrap(fault.KindUnavailable, err))
	}
	return driver, nil
}
