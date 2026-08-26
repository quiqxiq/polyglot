package device

import (
	"net/http"

	"connectrpc.com/connect"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

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
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.DeviceService"
	mux.Handle("/"+serviceName+"/ListDevices", connect.NewUnaryHandler(
		"/"+serviceName+"/ListDevices",
		handler.ListDevices,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetDevice", connect.NewUnaryHandler(
		"/"+serviceName+"/GetDevice",
		handler.GetDevice,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/UpdateDevice", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateDevice",
		handler.UpdateDevice,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/DeleteDevice", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteDevice",
		handler.DeleteDevice,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/TestDeviceConnection", connect.NewUnaryHandler(
		"/"+serviceName+"/TestDeviceConnection",
		handler.TestDeviceConnection,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamDeviceStatus", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamDeviceStatus",
		handler.StreamDeviceStatus,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamPing", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamPing",
		handler.StreamPing,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamInterfaceTraffic", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamInterfaceTraffic",
		handler.StreamInterfaceTraffic,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamTerminal", connect.NewBidiStreamHandler(
		"/"+serviceName+"/StreamTerminal",
		handler.StreamTerminal,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetDevicePingConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/GetDevicePingConfig",
		handler.GetDevicePingConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/UpdateDevicePingConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateDevicePingConfig",
		handler.UpdateDevicePingConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/QueryDevicePingMetrics", connect.NewUnaryHandler(
		"/"+serviceName+"/QueryDevicePingMetrics",
		handler.QueryDevicePingMetrics,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
