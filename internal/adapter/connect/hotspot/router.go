package hotspot

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
)

// NewHotspotServiceHandler mounts HotspotService Connect handlers.
func NewHotspotServiceHandler(
	uc *hotspotUC.UseCase,
	provider ConnectDriverProvider,
) (string, http.Handler) {
	handler := NewHotspotConnectHandler(uc, provider)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.HotspotService"

	// Profiles
	mux.Handle("/"+serviceName+"/ListProfiles", connect.NewUnaryHandler("/"+serviceName+"/ListProfiles", handler.ListProfiles, opts...))
	mux.Handle("/"+serviceName+"/CreateProfile", connect.NewUnaryHandler("/"+serviceName+"/CreateProfile", handler.CreateProfile, opts...))
	mux.Handle("/"+serviceName+"/UpdateProfile", connect.NewUnaryHandler("/"+serviceName+"/UpdateProfile", handler.UpdateProfile, opts...))
	mux.Handle("/"+serviceName+"/DeleteProfile", connect.NewUnaryHandler("/"+serviceName+"/DeleteProfile", handler.DeleteProfile, opts...))

	// Users
	mux.Handle("/"+serviceName+"/ListUsers", connect.NewUnaryHandler("/"+serviceName+"/ListUsers", handler.ListUsers, opts...))
	mux.Handle("/"+serviceName+"/GetUser", connect.NewUnaryHandler("/"+serviceName+"/GetUser", handler.GetUser, opts...))
	mux.Handle("/"+serviceName+"/CreateUser", connect.NewUnaryHandler("/"+serviceName+"/CreateUser", handler.CreateUser, opts...))
	mux.Handle("/"+serviceName+"/UpdateUser", connect.NewUnaryHandler("/"+serviceName+"/UpdateUser", handler.UpdateUser, opts...))
	mux.Handle("/"+serviceName+"/ResetUserCounters", connect.NewUnaryHandler("/"+serviceName+"/ResetUserCounters", handler.ResetUserCounters, opts...))
	mux.Handle("/"+serviceName+"/DeleteUser", connect.NewUnaryHandler("/"+serviceName+"/DeleteUser", handler.DeleteUser, opts...))
	mux.Handle("/"+serviceName+"/DeleteHotspotUsers", connect.NewUnaryHandler("/"+serviceName+"/DeleteHotspotUsers", handler.DeleteHotspotUsers, opts...))

	// Active Sessions
	mux.Handle("/"+serviceName+"/ListActiveSessions", connect.NewUnaryHandler("/"+serviceName+"/ListActiveSessions", handler.ListActiveSessions, opts...))
	mux.Handle("/"+serviceName+"/KickActiveSession", connect.NewUnaryHandler("/"+serviceName+"/KickActiveSession", handler.KickActiveSession, opts...))

	// Vouchers
	mux.Handle("/"+serviceName+"/GenerateVouchers", connect.NewUnaryHandler("/"+serviceName+"/GenerateVouchers", handler.GenerateVouchers, opts...))
	mux.Handle("/"+serviceName+"/GetVoucherBatch", connect.NewUnaryHandler("/"+serviceName+"/GetVoucherBatch", handler.GetVoucherBatch, opts...))
	mux.Handle("/"+serviceName+"/CheckVoucherStatus", connect.NewUnaryHandler("/"+serviceName+"/CheckVoucherStatus", handler.CheckVoucherStatus, opts...))

	// Hosts & Servers
	mux.Handle("/"+serviceName+"/ListHosts", connect.NewUnaryHandler("/"+serviceName+"/ListHosts", handler.ListHosts, opts...))
	mux.Handle("/"+serviceName+"/RemoveHost", connect.NewUnaryHandler("/"+serviceName+"/RemoveHost", handler.RemoveHost, opts...))
	mux.Handle("/"+serviceName+"/ListHotspotServers", connect.NewUnaryHandler("/"+serviceName+"/ListHotspotServers", handler.ListHotspotServers, opts...))

	// IP Bindings
	mux.Handle("/"+serviceName+"/ListHotspotIPBindings", connect.NewUnaryHandler("/"+serviceName+"/ListHotspotIPBindings", handler.ListHotspotIPBindings, opts...))
	mux.Handle("/"+serviceName+"/CreateHotspotIPBinding", connect.NewUnaryHandler("/"+serviceName+"/CreateHotspotIPBinding", handler.CreateHotspotIPBinding, opts...))
	mux.Handle("/"+serviceName+"/UpdateHotspotIPBinding", connect.NewUnaryHandler("/"+serviceName+"/UpdateHotspotIPBinding", handler.UpdateHotspotIPBinding, opts...))
	mux.Handle("/"+serviceName+"/DeleteHotspotIPBinding", connect.NewUnaryHandler("/"+serviceName+"/DeleteHotspotIPBinding", handler.DeleteHotspotIPBinding, opts...))

	// Cookies
	mux.Handle("/"+serviceName+"/ListHotspotCookies", connect.NewUnaryHandler("/"+serviceName+"/ListHotspotCookies", handler.ListHotspotCookies, opts...))
	mux.Handle("/"+serviceName+"/DeleteHotspotCookie", connect.NewUnaryHandler("/"+serviceName+"/DeleteHotspotCookie", handler.DeleteHotspotCookie, opts...))

	// Reports
	mux.Handle("/"+serviceName+"/ListReports", connect.NewUnaryHandler("/"+serviceName+"/ListReports", handler.ListReports, opts...))
	mux.Handle("/"+serviceName+"/DeleteReport", connect.NewUnaryHandler("/"+serviceName+"/DeleteReport", handler.DeleteReport, opts...))

	// Expire Monitor
	mux.Handle("/"+serviceName+"/GetExpireMonitorStatus", connect.NewUnaryHandler("/"+serviceName+"/GetExpireMonitorStatus", handler.GetExpireMonitorStatus, opts...))
	mux.Handle("/"+serviceName+"/SetupExpireMonitor", connect.NewUnaryHandler("/"+serviceName+"/SetupExpireMonitor", handler.SetupExpireMonitor, opts...))
	mux.Handle("/"+serviceName+"/DisableExpireMonitor", connect.NewUnaryHandler("/"+serviceName+"/DisableExpireMonitor", handler.DisableExpireMonitor, opts...))
	mux.Handle("/"+serviceName+"/RemoveExpireMonitor", connect.NewUnaryHandler("/"+serviceName+"/RemoveExpireMonitor", handler.RemoveExpireMonitor, opts...))

	// Templates
	mux.Handle("/"+serviceName+"/ListTemplates", connect.NewUnaryHandler("/"+serviceName+"/ListTemplates", handler.ListTemplates, opts...))
	mux.Handle("/"+serviceName+"/GetTemplateSection", connect.NewUnaryHandler("/"+serviceName+"/GetTemplateSection", handler.GetTemplateSection, opts...))
	mux.Handle("/"+serviceName+"/RenderVouchers", connect.NewUnaryHandler("/"+serviceName+"/RenderVouchers", handler.RenderVouchers, opts...))

	return "/" + serviceName + "/", mux
}
