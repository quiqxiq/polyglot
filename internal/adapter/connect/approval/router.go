package approval

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
)

// NewServiceHandler mounts ApprovalService Connect handlers.
func NewServiceHandler(h *ConnectHandler) (string, http.Handler) {
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()
	sn := "polyglot.v1.ApprovalService"
	mux.Handle("/"+sn+"/ListApprovals", connect.NewUnaryHandler("/"+sn+"/ListApprovals", h.ListApprovals, opts...))
	mux.Handle("/"+sn+"/GetApproval", connect.NewUnaryHandler("/"+sn+"/GetApproval", h.GetApproval, opts...))
	mux.Handle("/"+sn+"/DecideApproval", connect.NewUnaryHandler("/"+sn+"/DecideApproval", h.DecideApproval, opts...))
	mux.Handle("/"+sn+"/CancelApproval", connect.NewUnaryHandler("/"+sn+"/CancelApproval", h.CancelApproval, opts...))
	return "/" + sn + "/", mux
}
