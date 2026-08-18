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
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.HotspotService"
	mux.Handle("/"+serviceName+"/ListProfiles", connect.NewUnaryHandler("/"+serviceName+"/ListProfiles", handler.ListProfiles, codecOpt))
	mux.Handle("/"+serviceName+"/ListUsers", connect.NewUnaryHandler("/"+serviceName+"/ListUsers", handler.ListUsers, codecOpt))
	mux.Handle("/"+serviceName+"/ListActiveSessions", connect.NewUnaryHandler("/"+serviceName+"/ListActiveSessions", handler.ListActiveSessions, codecOpt))
	mux.Handle("/"+serviceName+"/KickActiveSession", connect.NewUnaryHandler("/"+serviceName+"/KickActiveSession", handler.KickActiveSession, codecOpt))
	mux.Handle("/"+serviceName+"/ListDHCPLeases", connect.NewUnaryHandler("/"+serviceName+"/ListDHCPLeases", handler.ListDHCPLeases, codecOpt))
	mux.Handle("/"+serviceName+"/BlockDHCPLease", connect.NewUnaryHandler("/"+serviceName+"/BlockDHCPLease", handler.BlockDHCPLease, codecOpt))
	mux.Handle("/"+serviceName+"/GenerateVouchers", connect.NewUnaryHandler("/"+serviceName+"/GenerateVouchers", handler.GenerateVouchers, codecOpt))
	mux.Handle("/"+serviceName+"/GetVoucherBatch", connect.NewUnaryHandler("/"+serviceName+"/GetVoucherBatch", handler.GetVoucherBatch, codecOpt))
	mux.Handle("/"+serviceName+"/GetUser", connect.NewUnaryHandler("/"+serviceName+"/GetUser", handler.GetUser, codecOpt))
	mux.Handle("/"+serviceName+"/CreateUser", connect.NewUnaryHandler("/"+serviceName+"/CreateUser", handler.CreateUser, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateUser", connect.NewUnaryHandler("/"+serviceName+"/UpdateUser", handler.UpdateUser, codecOpt))
	mux.Handle("/"+serviceName+"/ResetUserCounters", connect.NewUnaryHandler("/"+serviceName+"/ResetUserCounters", handler.ResetUserCounters, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteUser", connect.NewUnaryHandler("/"+serviceName+"/DeleteUser", handler.DeleteUser, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteHotspotUsers", connect.NewUnaryHandler("/"+serviceName+"/DeleteHotspotUsers", handler.DeleteHotspotUsers, codecOpt))
	mux.Handle("/"+serviceName+"/CreateProfile", connect.NewUnaryHandler("/"+serviceName+"/CreateProfile", handler.CreateProfile, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateProfile", connect.NewUnaryHandler("/"+serviceName+"/UpdateProfile", handler.UpdateProfile, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteProfile", connect.NewUnaryHandler("/"+serviceName+"/DeleteProfile", handler.DeleteProfile, codecOpt))
	mux.Handle("/"+serviceName+"/ListHosts", connect.NewUnaryHandler("/"+serviceName+"/ListHosts", handler.ListHosts, codecOpt))
	mux.Handle("/"+serviceName+"/RemoveHost", connect.NewUnaryHandler("/"+serviceName+"/RemoveHost", handler.RemoveHost, codecOpt))
	mux.Handle("/"+serviceName+"/ListHotspotServers", connect.NewUnaryHandler("/"+serviceName+"/ListHotspotServers", handler.ListHotspotServers, codecOpt))
	mux.Handle("/"+serviceName+"/ListHotspotIPBindings", connect.NewUnaryHandler("/"+serviceName+"/ListHotspotIPBindings", handler.ListHotspotIPBindings, codecOpt))
	mux.Handle("/"+serviceName+"/CreateHotspotIPBinding", connect.NewUnaryHandler("/"+serviceName+"/CreateHotspotIPBinding", handler.CreateHotspotIPBinding, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateHotspotIPBinding", connect.NewUnaryHandler("/"+serviceName+"/UpdateHotspotIPBinding", handler.UpdateHotspotIPBinding, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteHotspotIPBinding", connect.NewUnaryHandler("/"+serviceName+"/DeleteHotspotIPBinding", handler.DeleteHotspotIPBinding, codecOpt))
	mux.Handle("/"+serviceName+"/ListHotspotCookies", connect.NewUnaryHandler("/"+serviceName+"/ListHotspotCookies", handler.ListHotspotCookies, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteHotspotCookie", connect.NewUnaryHandler("/"+serviceName+"/DeleteHotspotCookie", handler.DeleteHotspotCookie, codecOpt))
	mux.Handle("/"+serviceName+"/CheckVoucherStatus", connect.NewUnaryHandler("/"+serviceName+"/CheckVoucherStatus", handler.CheckVoucherStatus, codecOpt))
	mux.Handle("/"+serviceName+"/StreamTraffic", connect.NewServerStreamHandler("/"+serviceName+"/StreamTraffic", handler.StreamTraffic, codecOpt))
	mux.Handle("/"+serviceName+"/StreamResource", connect.NewServerStreamHandler("/"+serviceName+"/StreamResource", handler.StreamResource, codecOpt))
	mux.Handle("/"+serviceName+"/StreamActiveSessions", connect.NewServerStreamHandler("/"+serviceName+"/StreamActiveSessions", handler.StreamActiveSessions, codecOpt))
	mux.Handle("/"+serviceName+"/StreamSystemSnapshot", connect.NewServerStreamHandler("/"+serviceName+"/StreamSystemSnapshot", handler.StreamSystemSnapshot, codecOpt))
	mux.Handle("/"+serviceName+"/StreamInterfaceEthernet", connect.NewServerStreamHandler("/"+serviceName+"/StreamInterfaceEthernet", handler.StreamInterfaceEthernet, codecOpt))
	mux.Handle("/"+serviceName+"/StreamQueueStats", connect.NewServerStreamHandler("/"+serviceName+"/StreamQueueStats", handler.StreamQueueStats, codecOpt))
	mux.Handle("/"+serviceName+"/StreamLogs", connect.NewServerStreamHandler("/"+serviceName+"/StreamLogs", handler.StreamLogs, codecOpt))
	mux.Handle("/"+serviceName+"/StreamHotspotInactive", connect.NewServerStreamHandler("/"+serviceName+"/StreamHotspotInactive", handler.StreamHotspotInactive, codecOpt))
	mux.Handle("/"+serviceName+"/StreamPPPActive", connect.NewServerStreamHandler("/"+serviceName+"/StreamPPPActive", handler.StreamPPPActive, codecOpt))
	mux.Handle("/"+serviceName+"/StreamPPPInactive", connect.NewServerStreamHandler("/"+serviceName+"/StreamPPPInactive", handler.StreamPPPInactive, codecOpt))
	mux.Handle("/"+serviceName+"/ListReports", connect.NewUnaryHandler("/"+serviceName+"/ListReports", handler.ListReports, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteReport", connect.NewUnaryHandler("/"+serviceName+"/DeleteReport", handler.DeleteReport, codecOpt))
	mux.Handle("/"+serviceName+"/GetExpireMonitorStatus", connect.NewUnaryHandler("/"+serviceName+"/GetExpireMonitorStatus", handler.GetExpireMonitorStatus, codecOpt))
	mux.Handle("/"+serviceName+"/SetupExpireMonitor", connect.NewUnaryHandler("/"+serviceName+"/SetupExpireMonitor", handler.SetupExpireMonitor, codecOpt))
	mux.Handle("/"+serviceName+"/DisableExpireMonitor", connect.NewUnaryHandler("/"+serviceName+"/DisableExpireMonitor", handler.DisableExpireMonitor, codecOpt))
	mux.Handle("/"+serviceName+"/RemoveExpireMonitor", connect.NewUnaryHandler("/"+serviceName+"/RemoveExpireMonitor", handler.RemoveExpireMonitor, codecOpt))
	mux.Handle("/"+serviceName+"/ListTemplates", connect.NewUnaryHandler("/"+serviceName+"/ListTemplates", handler.ListTemplates, codecOpt))
	mux.Handle("/"+serviceName+"/GetTemplateSection", connect.NewUnaryHandler("/"+serviceName+"/GetTemplateSection", handler.GetTemplateSection, codecOpt))
	mux.Handle("/"+serviceName+"/RenderVouchers", connect.NewUnaryHandler("/"+serviceName+"/RenderVouchers", handler.RenderVouchers, codecOpt))

	return "/" + serviceName + "/", mux
}
