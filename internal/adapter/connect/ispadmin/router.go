package ispadmin

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/importer"
)

// NewISPAdminServiceHandler mounts the IspAdminService ConnectRPC routes.
func NewISPAdminServiceHandler(
	upsert *importer.UpsertUseCase,
	routerSrc *importer.RouterSource,
	reconciler *importer.Reconciler,
	exporter *importer.ExportUseCase,
	resolver func(ctx context.Context, deviceID string) (port.DeviceDriver, bool),
) (string, http.Handler) {
	handler := NewISPAdminConnectHandler(upsert, routerSrc, reconciler, exporter, resolver)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.IspAdminService"
	mux.Handle("/"+serviceName+"/ImportFile", connect.NewUnaryHandler("/"+serviceName+"/ImportFile", handler.ImportFile, opts...))
	mux.Handle("/"+serviceName+"/ImportRouter", connect.NewUnaryHandler("/"+serviceName+"/ImportRouter", handler.ImportRouter, opts...))
	mux.Handle("/"+serviceName+"/ExportCustomers", connect.NewUnaryHandler("/"+serviceName+"/ExportCustomers", handler.ExportCustomers, opts...))
	mux.Handle("/"+serviceName+"/Reconcile", connect.NewUnaryHandler("/"+serviceName+"/Reconcile", handler.Reconcile, opts...))

	return "/" + serviceName + "/", mux
}

// NewIspAdminServiceHandler preserves backward compatibility.
var NewIspAdminServiceHandler = NewISPAdminServiceHandler
