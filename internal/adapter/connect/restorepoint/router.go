package restorepoint

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
)

// NewServiceHandler mounts RestorePointService Connect handlers.
func NewServiceHandler(h *ConnectHandler) (string, http.Handler) {
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()
	sn := "polyglot.v1.RestorePointService"
	mux.Handle("/"+sn+"/ListRestorePoints", connect.NewUnaryHandler("/"+sn+"/ListRestorePoints", h.ListRestorePoints, opts...))
	mux.Handle("/"+sn+"/GetRestorePoint", connect.NewUnaryHandler("/"+sn+"/GetRestorePoint", h.GetRestorePoint, opts...))
	mux.Handle("/"+sn+"/CreateRestorePoint", connect.NewUnaryHandler("/"+sn+"/CreateRestorePoint", h.CreateRestorePoint, opts...))
	mux.Handle("/"+sn+"/RestoreFromPoint", connect.NewUnaryHandler("/"+sn+"/RestoreFromPoint", h.RestoreFromPoint, opts...))
	mux.Handle("/"+sn+"/DownloadRestorePointArtifact", connect.NewUnaryHandler("/"+sn+"/DownloadRestorePointArtifact", h.DownloadRestorePointArtifact, opts...))
	mux.Handle("/"+sn+"/DeleteRestorePoint", connect.NewUnaryHandler("/"+sn+"/DeleteRestorePoint", h.DeleteRestorePoint, opts...))
	return "/" + sn + "/", mux
}
