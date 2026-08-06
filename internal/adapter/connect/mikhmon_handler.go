package connectadapter

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceId.
type ConnectDriverProvider func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

// MikhmonConnectHandler orchestrates Mikhmon ConnectRPC procedures across modular handler files.
type MikhmonConnectHandler struct {
	useCase               *network.MikhmonUseCase
	activeSessionsUseCase *network.ActiveSessionsUseCase
	driverProvider        ConnectDriverProvider
}

// NewMikhmonConnectHandler constructs a new MikhmonConnectHandler.
func NewMikhmonConnectHandler(uc *network.MikhmonUseCase, provider ConnectDriverProvider) *MikhmonConnectHandler {
	return &MikhmonConnectHandler{
		useCase:               uc,
		activeSessionsUseCase: network.NewActiveSessionsUseCase(),
		driverProvider:        provider,
	}
}

func (h *MikhmonConnectHandler) getDriver(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
	if deviceID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}
	driver, err := h.driverProvider(ctx, deviceID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get device driver for %s: %w", deviceID, err))
	}
	return driver, nil
}

// NewMikhmonServiceHandler creates the Connect http.Handler and registers procedures.
func NewMikhmonServiceHandler(uc *network.MikhmonUseCase, provider ConnectDriverProvider) (string, http.Handler) {
	handler := NewMikhmonConnectHandler(uc, provider)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(connectJSONCodec{})

	serviceName := "polyglot.v1.MikhmonService"
	mux.Handle("/"+serviceName+"/GetDashboard", connect.NewUnaryHandler(
		"/"+serviceName+"/GetDashboard",
		handler.GetDashboard,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListProfiles", connect.NewUnaryHandler(
		"/"+serviceName+"/ListProfiles",
		handler.ListProfiles,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListUsers", connect.NewUnaryHandler(
		"/"+serviceName+"/ListUsers",
		handler.ListUsers,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListActiveSessions", connect.NewUnaryHandler(
		"/"+serviceName+"/ListActiveSessions",
		handler.ListActiveSessions,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/KickActiveSession", connect.NewUnaryHandler(
		"/"+serviceName+"/KickActiveSession",
		handler.KickActiveSession,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListDHCPLeases", connect.NewUnaryHandler(
		"/"+serviceName+"/ListDHCPLeases",
		handler.ListDHCPLeases,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/BlockDHCPLease", connect.NewUnaryHandler(
		"/"+serviceName+"/BlockDHCPLease",
		handler.BlockDHCPLease,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GenerateVouchers", connect.NewUnaryHandler(
		"/"+serviceName+"/GenerateVouchers",
		handler.GenerateVouchers,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
