package routerjob

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	jobuc "github.com/quixiq/polyglot/internal/usecase/routerjob"
)

// NewRouterJobServiceHandler mounts RouterJobService Connect handlers
// (di balik JWT+RBAC).
func NewRouterJobServiceHandler(jobs port.RouterJobRepository, exporter *jobuc.Exporter) (string, http.Handler) {
	handler := NewConnectHandler(jobs, exporter)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()
	serviceName := "polyglot.v1.RouterJobService"

	mux.Handle("/"+serviceName+"/ListRouterJobs", connect.NewUnaryHandler("/"+serviceName+"/ListRouterJobs", handler.ListRouterJobs, opts...))
	mux.Handle("/"+serviceName+"/RetryRouterJob", connect.NewUnaryHandler("/"+serviceName+"/RetryRouterJob", handler.RetryRouterJob, opts...))
	mux.Handle("/"+serviceName+"/ExportRouterJobScript", connect.NewUnaryHandler("/"+serviceName+"/ExportRouterJobScript", handler.ExportRouterJobScript, opts...))

	return "/" + serviceName + "/", mux
}
