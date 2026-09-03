package network

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

// NewNetworkServiceHandler mounts NetworkService Connect handlers.
func NewNetworkServiceHandler(
	hsUC *hotspotUC.UseCase,
	activeUC *networkUC.ActiveSessionsUseCase,
	provider ConnectDriverProvider,
) (string, http.Handler) {
	handler := NewNetworkConnectHandler(hsUC, activeUC, provider)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.NetworkService"

	mux.Handle("/"+serviceName+"/ListDHCPLeases", connect.NewUnaryHandler("/"+serviceName+"/ListDHCPLeases", handler.ListDHCPLeases, opts...))
	mux.Handle("/"+serviceName+"/BlockDHCPLease", connect.NewUnaryHandler("/"+serviceName+"/BlockDHCPLease", handler.BlockDHCPLease, opts...))
	mux.Handle("/"+serviceName+"/ListParentQueues", connect.NewUnaryHandler("/"+serviceName+"/ListParentQueues", handler.ListParentQueues, opts...))
	mux.Handle("/"+serviceName+"/ListIPPools", connect.NewUnaryHandler("/"+serviceName+"/ListIPPools", handler.ListIPPools, opts...))

	return "/" + serviceName + "/", mux
}
