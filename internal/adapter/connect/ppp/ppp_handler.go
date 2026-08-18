package ppp

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/quixiq/polyglot/internal/port"
	pppUC "github.com/quixiq/polyglot/internal/usecase/ppp"
)

// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceId.
type ConnectDriverProvider func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

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
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}
	driver, err := h.driverProvider(ctx, deviceID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get device driver for %s: %w", deviceID, err))
	}
	return driver, nil
}
