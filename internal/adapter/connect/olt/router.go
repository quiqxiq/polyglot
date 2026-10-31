package olt

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	oltUC "github.com/quixiq/polyglot/internal/usecase/olt"
)

// NewOltServiceHandler mounts OltService Connect handlers (di balik JWT+RBAC).
func NewOltServiceHandler(svc *oltUC.Service, settings port.SettingReader) (string, http.Handler) {
	handler := NewConnectHandler(svc, settings)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()
	serviceName := "polyglot.v1.OltService"

	rpcs := map[string]http.Handler{
		"ListOlts":             connect.NewUnaryHandler("/"+serviceName+"/ListOlts", handler.ListOlts, opts...),
		"GetOlt":               connect.NewUnaryHandler("/"+serviceName+"/GetOlt", handler.GetOlt, opts...),
		"CreateOlt":            connect.NewUnaryHandler("/"+serviceName+"/CreateOlt", handler.CreateOlt, opts...),
		"UpdateOlt":            connect.NewUnaryHandler("/"+serviceName+"/UpdateOlt", handler.UpdateOlt, opts...),
		"DeleteOlt":            connect.NewUnaryHandler("/"+serviceName+"/DeleteOlt", handler.DeleteOlt, opts...),
		"TestOltConnection":    connect.NewUnaryHandler("/"+serviceName+"/TestOltConnection", handler.TestOltConnection, opts...),
		"ListOnus":             connect.NewUnaryHandler("/"+serviceName+"/ListOnus", handler.ListOnus, opts...),
		"GetOnu":               connect.NewUnaryHandler("/"+serviceName+"/GetOnu", handler.GetOnu, opts...),
		"ListOnuEvents":        connect.NewUnaryHandler("/"+serviceName+"/ListOnuEvents", handler.ListOnuEvents, opts...),
		"GetOnuOpticalHistory": connect.NewUnaryHandler("/"+serviceName+"/GetOnuOpticalHistory", handler.GetOnuOpticalHistory, opts...),
		"SyncOlt":              connect.NewUnaryHandler("/"+serviceName+"/SyncOlt", handler.SyncOlt, opts...),
		"SaveOltConfig":        connect.NewUnaryHandler("/"+serviceName+"/SaveOltConfig", handler.SaveOltConfig, opts...),
		"RenameOnu":            connect.NewUnaryHandler("/"+serviceName+"/RenameOnu", handler.RenameOnu, opts...),
		"RebootOnu":            connect.NewUnaryHandler("/"+serviceName+"/RebootOnu", handler.RebootOnu, opts...),
		"SetOnuAdminState":     connect.NewUnaryHandler("/"+serviceName+"/SetOnuAdminState", handler.SetOnuAdminState, opts...),
		"DeleteOnu":            connect.NewUnaryHandler("/"+serviceName+"/DeleteOnu", handler.DeleteOnu, opts...),
	}
	for name, h := range rpcs {
		mux.Handle("/"+serviceName+"/"+name, h)
	}
	return "/" + serviceName + "/", mux
}
