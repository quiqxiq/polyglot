package hotspot

import (
	"net/http"

	"connectrpc.com/connect"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
)

// NewHotspotServiceHandler creates the Connect http.Handler for HotspotService.
func NewHotspotServiceHandler(uc *hotspotUC.UseCase, provider ConnectDriverProvider) (string, http.Handler) {
	handler := NewHotspotConnectHandler(uc, provider)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.HotspotService"
	mux.Handle("/"+serviceName+"/GetDashboard", connect.NewUnaryHandler("/"+serviceName+"/GetDashboard", handler.GetDashboard, codecOpt))
	mux.Handle("/"+serviceName+"/ListProfiles", connect.NewUnaryHandler("/"+serviceName+"/ListProfiles", handler.ListProfiles, codecOpt))
	mux.Handle("/"+serviceName+"/ListUsers", connect.NewUnaryHandler("/"+serviceName+"/ListUsers", handler.ListUsers, codecOpt))
	mux.Handle("/"+serviceName+"/ListActiveSessions", connect.NewUnaryHandler("/"+serviceName+"/ListActiveSessions", handler.ListActiveSessions, codecOpt))
	mux.Handle("/"+serviceName+"/KickActiveSession", connect.NewUnaryHandler("/"+serviceName+"/KickActiveSession", handler.KickActiveSession, codecOpt))
	mux.Handle("/"+serviceName+"/ListDHCPLeases", connect.NewUnaryHandler("/"+serviceName+"/ListDHCPLeases", handler.ListDHCPLeases, codecOpt))
	mux.Handle("/"+serviceName+"/BlockDHCPLease", connect.NewUnaryHandler("/"+serviceName+"/BlockDHCPLease", handler.BlockDHCPLease, codecOpt))
	mux.Handle("/"+serviceName+"/GenerateVouchers", connect.NewUnaryHandler("/"+serviceName+"/GenerateVouchers", handler.GenerateVouchers, codecOpt))
	mux.Handle("/"+serviceName+"/StreamTraffic", connect.NewServerStreamHandler("/"+serviceName+"/StreamTraffic", handler.StreamTraffic, codecOpt))
	mux.Handle("/"+serviceName+"/StreamResource", connect.NewServerStreamHandler("/"+serviceName+"/StreamResource", handler.StreamResource, codecOpt))
	mux.Handle("/"+serviceName+"/StreamActiveSessions", connect.NewServerStreamHandler("/"+serviceName+"/StreamActiveSessions", handler.StreamActiveSessions, codecOpt))

	return "/" + serviceName + "/", mux
}
