package hotspot

import (
	"net/http"

	"connectrpc.com/connect"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// NewHotspotServiceHandler creates the Connect http.Handler for HotspotService.
func NewHotspotServiceHandler(uc *hotspotUC.UseCase, activeUC *network.ActiveSessionsUseCase, provider ConnectDriverProvider) (string, http.Handler) {
	handler := NewHotspotConnectHandler(uc, activeUC, provider)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.HotspotService"
	mux.Handle("/"+serviceName+"/ListProfiles", connect.NewUnaryHandler("/"+serviceName+"/ListProfiles", handler.ListProfiles, opts...))
	mux.Handle("/"+serviceName+"/ListUsers", connect.NewUnaryHandler("/"+serviceName+"/ListUsers", handler.ListUsers, opts...))
	mux.Handle("/"+serviceName+"/ListActiveSessions", connect.NewUnaryHandler("/"+serviceName+"/ListActiveSessions", handler.ListActiveSessions, opts...))
	mux.Handle("/"+serviceName+"/KickActiveSession", connect.NewUnaryHandler("/"+serviceName+"/KickActiveSession", handler.KickActiveSession, opts...))
	mux.Handle("/"+serviceName+"/ListDHCPLeases", connect.NewUnaryHandler("/"+serviceName+"/ListDHCPLeases", handler.ListDHCPLeases, opts...))
	mux.Handle("/"+serviceName+"/BlockDHCPLease", connect.NewUnaryHandler("/"+serviceName+"/BlockDHCPLease", handler.BlockDHCPLease, opts...))
	mux.Handle("/"+serviceName+"/GenerateVouchers", connect.NewUnaryHandler("/"+serviceName+"/GenerateVouchers", handler.GenerateVouchers, opts...))
	mux.Handle("/"+serviceName+"/GetVoucherBatch", connect.NewUnaryHandler("/"+serviceName+"/GetVoucherBatch", handler.GetVoucherBatch, opts...))
	mux.Handle("/"+serviceName+"/GetUser", connect.NewUnaryHandler("/"+serviceName+"/GetUser", handler.GetUser, opts...))
	mux.Handle("/"+serviceName+"/CreateUser", connect.NewUnaryHandler("/"+serviceName+"/CreateUser", handler.CreateUser, opts...))
	mux.Handle("/"+serviceName+"/UpdateUser", connect.NewUnaryHandler("/"+serviceName+"/UpdateUser", handler.UpdateUser, opts...))
	mux.Handle("/"+serviceName+"/ResetUserCounters", connect.NewUnaryHandler("/"+serviceName+"/ResetUserCounters", handler.ResetUserCounters, opts...))
	mux.Handle("/"+serviceName+"/DeleteUser", connect.NewUnaryHandler("/"+serviceName+"/DeleteUser", handler.DeleteUser, opts...))
	mux.Handle("/"+serviceName+"/DeleteHotspotUsers", connect.NewUnaryHandler("/"+serviceName+"/DeleteHotspotUsers", handler.DeleteHotspotUsers, opts...))
	mux.Handle("/"+serviceName+"/CreateProfile", connect.NewUnaryHandler("/"+serviceName+"/CreateProfile", handler.CreateProfile, opts...))
	mux.Handle("/"+serviceName+"/UpdateProfile", connect.NewUnaryHandler("/"+serviceName+"/UpdateProfile", handler.UpdateProfile, opts...))
	mux.Handle("/"+serviceName+"/DeleteProfile", connect.NewUnaryHandler("/"+serviceName+"/DeleteProfile", handler.DeleteProfile, opts...))
	mux.Handle("/"+serviceName+"/ListHosts", connect.NewUnaryHandler("/"+serviceName+"/ListHosts", handler.ListHosts, opts...))
	mux.Handle("/"+serviceName+"/RemoveHost", connect.NewUnaryHandler("/"+serviceName+"/RemoveHost", handler.RemoveHost, opts...))
	mux.Handle("/"+serviceName+"/ListHotspotServers", connect.NewUnaryHandler("/"+serviceName+"/ListHotspotServers", handler.ListHotspotServers, opts...))
	mux.Handle("/"+serviceName+"/ListHotspotIPBindings", connect.NewUnaryHandler("/"+serviceName+"/ListHotspotIPBindings", handler.ListHotspotIPBindings, opts...))
	mux.Handle("/"+serviceName+"/CreateHotspotIPBinding", connect.NewUnaryHandler("/"+serviceName+"/CreateHotspotIPBinding", handler.CreateHotspotIPBinding, opts...))
	mux.Handle("/"+serviceName+"/UpdateHotspotIPBinding", connect.NewUnaryHandler("/"+serviceName+"/UpdateHotspotIPBinding", handler.UpdateHotspotIPBinding, opts...))
	mux.Handle("/"+serviceName+"/DeleteHotspotIPBinding", connect.NewUnaryHandler("/"+serviceName+"/DeleteHotspotIPBinding", handler.DeleteHotspotIPBinding, opts...))
	mux.Handle("/"+serviceName+"/ListHotspotCookies", connect.NewUnaryHandler("/"+serviceName+"/ListHotspotCookies", handler.ListHotspotCookies, opts...))
	mux.Handle("/"+serviceName+"/DeleteHotspotCookie", connect.NewUnaryHandler("/"+serviceName+"/DeleteHotspotCookie", handler.DeleteHotspotCookie, opts...))
	mux.Handle("/"+serviceName+"/CheckVoucherStatus", connect.NewUnaryHandler("/"+serviceName+"/CheckVoucherStatus", handler.CheckVoucherStatus, opts...))
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
	mux.Handle("/"+serviceName+"/ListReports", connect.NewUnaryHandler("/"+serviceName+"/ListReports", handler.ListReports, opts...))
	mux.Handle("/"+serviceName+"/DeleteReport", connect.NewUnaryHandler("/"+serviceName+"/DeleteReport", handler.DeleteReport, opts...))
	mux.Handle("/"+serviceName+"/GetExpireMonitorStatus", connect.NewUnaryHandler("/"+serviceName+"/GetExpireMonitorStatus", handler.GetExpireMonitorStatus, opts...))
	mux.Handle("/"+serviceName+"/SetupExpireMonitor", connect.NewUnaryHandler("/"+serviceName+"/SetupExpireMonitor", handler.SetupExpireMonitor, opts...))
	mux.Handle("/"+serviceName+"/DisableExpireMonitor", connect.NewUnaryHandler("/"+serviceName+"/DisableExpireMonitor", handler.DisableExpireMonitor, opts...))
	mux.Handle("/"+serviceName+"/RemoveExpireMonitor", connect.NewUnaryHandler("/"+serviceName+"/RemoveExpireMonitor", handler.RemoveExpireMonitor, opts...))
	mux.Handle("/"+serviceName+"/ListTemplates", connect.NewUnaryHandler("/"+serviceName+"/ListTemplates", handler.ListTemplates, opts...))
	mux.Handle("/"+serviceName+"/GetTemplateSection", connect.NewUnaryHandler("/"+serviceName+"/GetTemplateSection", handler.GetTemplateSection, opts...))
	mux.Handle("/"+serviceName+"/RenderVouchers", connect.NewUnaryHandler("/"+serviceName+"/RenderVouchers", handler.RenderVouchers, opts...))
	mux.Handle("/"+serviceName+"/ListParentQueues", connect.NewUnaryHandler("/"+serviceName+"/ListParentQueues", handler.ListParentQueues, opts...))
	mux.Handle("/"+serviceName+"/ListIPPools", connect.NewUnaryHandler("/"+serviceName+"/ListIPPools", handler.ListIPPools, opts...))

	return "/" + serviceName + "/", mux
}
