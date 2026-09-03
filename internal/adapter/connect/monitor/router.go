package monitor

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

// NewNetworkMonitorServiceHandler mounts NetworkMonitorService Connect streaming handlers.
func NewNetworkMonitorServiceHandler(
	hsUC *hotspotUC.UseCase,
	activeUC *networkUC.ActiveSessionsUseCase,
	provider ConnectDriverProvider,
	streamGW port.MonitorStreamGateway,
) (string, http.Handler) {
	handler := NewNetworkMonitorConnectHandler(hsUC, activeUC, provider, streamGW)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.NetworkMonitorService"

	mux.Handle("/"+serviceName+"/StreamTraffic", connect.NewServerStreamHandler("/"+serviceName+"/StreamTraffic", handler.StreamTraffic, opts...))
	mux.Handle("/"+serviceName+"/StreamResource", connect.NewServerStreamHandler("/"+serviceName+"/StreamResource", handler.StreamResource, opts...))
	mux.Handle("/"+serviceName+"/StreamActiveSessions", connect.NewServerStreamHandler("/"+serviceName+"/StreamActiveSessions", handler.StreamActiveSessions, opts...))
	mux.Handle("/"+serviceName+"/StreamActiveStats", connect.NewServerStreamHandler("/"+serviceName+"/StreamActiveStats", handler.StreamActiveStats, opts...))
	mux.Handle("/"+serviceName+"/StreamSystemSnapshot", connect.NewServerStreamHandler("/"+serviceName+"/StreamSystemSnapshot", handler.StreamSystemSnapshot, opts...))
	mux.Handle("/"+serviceName+"/StreamInterfaceEthernet", connect.NewServerStreamHandler("/"+serviceName+"/StreamInterfaceEthernet", handler.StreamInterfaceEthernet, opts...))
	mux.Handle("/"+serviceName+"/StreamQueueStats", connect.NewServerStreamHandler("/"+serviceName+"/StreamQueueStats", handler.StreamQueueStats, opts...))
	mux.Handle("/"+serviceName+"/StreamLogs", connect.NewServerStreamHandler("/"+serviceName+"/StreamLogs", handler.StreamLogs, opts...))
	mux.Handle("/"+serviceName+"/StreamHotspotInactive", connect.NewServerStreamHandler("/"+serviceName+"/StreamHotspotInactive", handler.StreamHotspotInactive, opts...))
	mux.Handle("/"+serviceName+"/StreamPPPActive", connect.NewServerStreamHandler("/"+serviceName+"/StreamPPPActive", handler.StreamPPPActive, opts...))
	mux.Handle("/"+serviceName+"/StreamPPPInactive", connect.NewServerStreamHandler("/"+serviceName+"/StreamPPPInactive", handler.StreamPPPInactive, opts...))

	return "/" + serviceName + "/", mux
}
