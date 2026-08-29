package portal

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	uc "github.com/quixiq/polyglot/internal/usecase/portal"
)

// NewPortalServiceHandler mounts PortalService Connect handlers to http.ServeMux
// secara PUBLIK (di rootMux tanpa JWT staff; autentikasi memakai session token
// portal internal).
func NewPortalServiceHandler(u *uc.UseCase) (string, http.Handler) {
	handler := NewPortalConnectHandler(u)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.PortalService"
	mux.Handle("/"+serviceName+"/RequestOTP", connect.NewUnaryHandler("/"+serviceName+"/RequestOTP", handler.RequestOTP, opts...))
	mux.Handle("/"+serviceName+"/Login", connect.NewUnaryHandler("/"+serviceName+"/Login", handler.Login, opts...))
	mux.Handle("/"+serviceName+"/Overview", connect.NewUnaryHandler("/"+serviceName+"/Overview", handler.Overview, opts...))
	mux.Handle("/"+serviceName+"/MyInvoices", connect.NewUnaryHandler("/"+serviceName+"/MyInvoices", handler.MyInvoices, opts...))
	mux.Handle("/"+serviceName+"/MyPayments", connect.NewUnaryHandler("/"+serviceName+"/MyPayments", handler.MyPayments, opts...))
	mux.Handle("/"+serviceName+"/Logout", connect.NewUnaryHandler("/"+serviceName+"/Logout", handler.Logout, opts...))

	return "/" + serviceName + "/", mux
}
