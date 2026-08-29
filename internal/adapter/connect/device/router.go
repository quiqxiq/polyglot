package device

import (
	"net/http"

	"connectrpc.com/connect"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// NewDeviceServiceHandler creates the device ConnectRPC service handler.
func NewDeviceServiceHandler(
	uc *deviceUC.ManageDeviceUseCase,
	openTermUC *network.OpenTerminalUseCase,
	getter DriverGetter,
	metricsUC *deviceUC.ManageMetricsUseCase,
) (string, http.Handler) {
	handler := &DeviceConnectHandler{
		useCase:      uc,
		openTermUC:   openTermUC,
		driverGetter: getter,
		metricsUC:    metricsUC,
	}
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.DeviceService"
	mux.Handle("/"+serviceName+"/ListDevices", connect.NewUnaryHandler(
		"/"+serviceName+"/ListDevices",
		handler.ListDevices,
		opts...,
	))
	mux.Handle("/"+serviceName+"/GetDevice", connect.NewUnaryHandler(
		"/"+serviceName+"/GetDevice",
		handler.GetDevice,
		opts...,
	))
	mux.Handle("/"+serviceName+"/UpdateDevice", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateDevice",
		handler.UpdateDevice,
		opts...,
	))
	mux.Handle("/"+serviceName+"/DeleteDevice", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteDevice",
		handler.DeleteDevice,
		opts...,
	))
	mux.Handle("/"+serviceName+"/TestDeviceConnection", connect.NewUnaryHandler(
		"/"+serviceName+"/TestDeviceConnection",
		handler.TestDeviceConnection,
		opts...,
	))
	mux.Handle("/"+serviceName+"/StreamDeviceStatus", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamDeviceStatus",
		handler.StreamDeviceStatus,
		opts...,
	))
	mux.Handle("/"+serviceName+"/StreamPing", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamPing",
		handler.StreamPing,
		opts...,
	))
	mux.Handle("/"+serviceName+"/StreamInterfaceTraffic", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamInterfaceTraffic",
		handler.StreamInterfaceTraffic,
		opts...,
	))
	mux.Handle("/"+serviceName+"/StreamTerminal", connect.NewBidiStreamHandler(
		"/"+serviceName+"/StreamTerminal",
		handler.StreamTerminal,
		opts...,
	))
	mux.Handle("/"+serviceName+"/GetDevicePingConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/GetDevicePingConfig",
		handler.GetDevicePingConfig,
		opts...,
	))
	mux.Handle("/"+serviceName+"/UpdateDevicePingConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateDevicePingConfig",
		handler.UpdateDevicePingConfig,
		opts...,
	))
	mux.Handle("/"+serviceName+"/QueryDevicePingMetrics", connect.NewUnaryHandler(
		"/"+serviceName+"/QueryDevicePingMetrics",
		handler.QueryDevicePingMetrics,
		opts...,
	))

	return "/" + serviceName + "/", mux
}
