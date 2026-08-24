package ispadmin

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/importer"
)

func NewIspAdminServiceHandler(
	upsert *importer.UpsertUseCase,
	routerSrc *importer.RouterSource,
	reconciler *importer.Reconciler,
	exporter *importer.ExportUseCase,
	resolver func(ctx context.Context, deviceID string) (port.DeviceDriver, bool),
) (string, http.Handler) {
	handler := NewIspAdminConnectHandler(upsert, routerSrc, reconciler, exporter, resolver)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.IspAdminService"
	mux.Handle("/"+serviceName+"/ImportFile", connect.NewUnaryHandler("/"+serviceName+"/ImportFile", handler.ImportFile, codecOpt))
	mux.Handle("/"+serviceName+"/ImportRouter", connect.NewUnaryHandler("/"+serviceName+"/ImportRouter", handler.ImportRouter, codecOpt))
	mux.Handle("/"+serviceName+"/ExportCustomers", connect.NewUnaryHandler("/"+serviceName+"/ExportCustomers", handler.ExportCustomers, codecOpt))
	mux.Handle("/"+serviceName+"/Reconcile", connect.NewUnaryHandler("/"+serviceName+"/Reconcile", handler.Reconcile, codecOpt))

	return "/" + serviceName + "/", mux
}
