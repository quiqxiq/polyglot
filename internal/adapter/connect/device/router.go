package device

import (
	"net/http"

	"connectrpc.com/connect"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// NewDeviceServiceHandler creates the device ConnectRPC service handler.
func NewDeviceServiceHandler(
	uc *deviceUC.ManageDeviceUseCase,
	openTermUC *network.OpenTerminalUseCase,
	getter DriverGetter,
	metricsUC *deviceUC.ManageMetricsUseCase,
	isolationUC *deviceUC.ManageIsolationUseCase,
	streamGW port.MonitorStreamGateway,
) (string, http.Handler) {
	handler := &DeviceConnectHandler{
		useCase:      uc,
		openTermUC:   openTermUC,
		driverGetter: getter,
		metricsUC:    metricsUC,
		isolationUC:  isolationUC,
		streamGW:     streamGW,
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

	// Profil Isolir & Integrasi Script Router
	mux.Handle("/"+serviceName+"/GetIsolationStatus", connect.NewUnaryHandler(
		"/"+serviceName+"/GetIsolationStatus",
		handler.GetIsolationStatus,
		opts...,
	))
	mux.Handle("/"+serviceName+"/CreateIsolationProfile", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateIsolationProfile",
		handler.CreateIsolationProfile,
		opts...,
	))
	mux.Handle("/"+serviceName+"/UpdateIsolationProfile", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateIsolationProfile",
		handler.UpdateIsolationProfile,
		opts...,
	))
	mux.Handle("/"+serviceName+"/DeleteIsolationProfile", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteIsolationProfile",
		handler.DeleteIsolationProfile,
		opts...,
	))
	mux.Handle("/"+serviceName+"/GetRouterIntegrationScript", connect.NewUnaryHandler(
		"/"+serviceName+"/GetRouterIntegrationScript",
		handler.GetRouterIntegrationScript,
		opts...,
	))
	mux.Handle("/"+serviceName+"/ApplyRouterIntegrationScript", connect.NewUnaryHandler(
		"/"+serviceName+"/ApplyRouterIntegrationScript",
		handler.ApplyRouterIntegrationScript,
		opts...,
	))

	return "/" + serviceName + "/", mux
}
