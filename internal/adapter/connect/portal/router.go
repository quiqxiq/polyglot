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
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.PortalService"
	mux.Handle("/"+serviceName+"/RequestOTP", connect.NewUnaryHandler("/"+serviceName+"/RequestOTP", handler.RequestOTP, codecOpt))
	mux.Handle("/"+serviceName+"/Login", connect.NewUnaryHandler("/"+serviceName+"/Login", handler.Login, codecOpt))
	mux.Handle("/"+serviceName+"/Overview", connect.NewUnaryHandler("/"+serviceName+"/Overview", handler.Overview, codecOpt))
	mux.Handle("/"+serviceName+"/MyInvoices", connect.NewUnaryHandler("/"+serviceName+"/MyInvoices", handler.MyInvoices, codecOpt))
	mux.Handle("/"+serviceName+"/MyPayments", connect.NewUnaryHandler("/"+serviceName+"/MyPayments", handler.MyPayments, codecOpt))
	mux.Handle("/"+serviceName+"/Logout", connect.NewUnaryHandler("/"+serviceName+"/Logout", handler.Logout, codecOpt))

	return "/" + serviceName + "/", mux
}
