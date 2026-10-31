package ont

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	ontUC "github.com/quixiq/polyglot/internal/usecase/ont"
)

// NewOntServiceHandler mounts OntService Connect handlers (di balik JWT+RBAC).
func NewOntServiceHandler(svc *ontUC.Service) (string, http.Handler) {
	handler := NewConnectHandler(svc)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()
	serviceName := "polyglot.v1.OntService"

	rpcs := map[string]http.Handler{
		"GetAcsConfig":        connect.NewUnaryHandler("/"+serviceName+"/GetAcsConfig", handler.GetAcsConfig, opts...),
		"SyncAcsDevices":      connect.NewUnaryHandler("/"+serviceName+"/SyncAcsDevices", handler.SyncAcsDevices, opts...),
		"ListAcsDevices":      connect.NewUnaryHandler("/"+serviceName+"/ListAcsDevices", handler.ListAcsDevices, opts...),
		"RefreshOntTelemetry": connect.NewUnaryHandler("/"+serviceName+"/RefreshOntTelemetry", handler.RefreshOntTelemetry, opts...),
		"RebootOnt":           connect.NewUnaryHandler("/"+serviceName+"/RebootOnt", handler.RebootOnt, opts...),
		"SetOntWifi":          connect.NewUnaryHandler("/"+serviceName+"/SetOntWifi", handler.SetOntWifi, opts...),
		"PushOntPppoe":        connect.NewUnaryHandler("/"+serviceName+"/PushOntPppoe", handler.PushOntPppoe, opts...),
		"LinkOnt":             connect.NewUnaryHandler("/"+serviceName+"/LinkOnt", handler.LinkOnt, opts...),
		"UnlinkOnt":           connect.NewUnaryHandler("/"+serviceName+"/UnlinkOnt", handler.UnlinkOnt, opts...),
		"PinAcsDevice":        connect.NewUnaryHandler("/"+serviceName+"/PinAcsDevice", handler.PinAcsDevice, opts...),
		"UnpinAcsDevice":      connect.NewUnaryHandler("/"+serviceName+"/UnpinAcsDevice", handler.UnpinAcsDevice, opts...),
		"EnsureAcsPreset":     connect.NewUnaryHandler("/"+serviceName+"/EnsureAcsPreset", handler.EnsureAcsPreset, opts...),
	}
	for name, h := range rpcs {
		mux.Handle("/"+serviceName+"/"+name, h)
	}
	return "/" + serviceName + "/", mux
}
