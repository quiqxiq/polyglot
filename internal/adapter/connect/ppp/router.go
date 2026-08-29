package ppp

import (
	"net/http"

	"connectrpc.com/connect"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	pppUC "github.com/quixiq/polyglot/internal/usecase/ppp"
)

// NewPPPServiceHandler creates the Connect http.Handler for PPPService.
func NewPPPServiceHandler(uc *pppUC.UseCase, provider ConnectDriverProvider) (string, http.Handler) {
	handler := NewPPPConnectHandler(uc, provider)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.PPPService"

	// Secret Operations
	mux.Handle("/"+serviceName+"/ListSecrets", connect.NewUnaryHandler("/"+serviceName+"/ListSecrets", handler.ListSecrets, opts...))
	mux.Handle("/"+serviceName+"/GetSecret", connect.NewUnaryHandler("/"+serviceName+"/GetSecret", handler.GetSecret, opts...))
	mux.Handle("/"+serviceName+"/CreateSecret", connect.NewUnaryHandler("/"+serviceName+"/CreateSecret", handler.CreateSecret, opts...))
	mux.Handle("/"+serviceName+"/UpdateSecret", connect.NewUnaryHandler("/"+serviceName+"/UpdateSecret", handler.UpdateSecret, opts...))
	mux.Handle("/"+serviceName+"/DeleteSecret", connect.NewUnaryHandler("/"+serviceName+"/DeleteSecret", handler.DeleteSecret, opts...))
	mux.Handle("/"+serviceName+"/SetSecretDisabled", connect.NewUnaryHandler("/"+serviceName+"/SetSecretDisabled", handler.SetSecretDisabled, opts...))

	// Profile Operations
	mux.Handle("/"+serviceName+"/ListProfiles", connect.NewUnaryHandler("/"+serviceName+"/ListProfiles", handler.ListProfiles, opts...))
	mux.Handle("/"+serviceName+"/GetProfile", connect.NewUnaryHandler("/"+serviceName+"/GetProfile", handler.GetProfile, opts...))
	mux.Handle("/"+serviceName+"/CreateProfile", connect.NewUnaryHandler("/"+serviceName+"/CreateProfile", handler.CreateProfile, opts...))
	mux.Handle("/"+serviceName+"/UpdateProfile", connect.NewUnaryHandler("/"+serviceName+"/UpdateProfile", handler.UpdateProfile, opts...))
	mux.Handle("/"+serviceName+"/DeleteProfile", connect.NewUnaryHandler("/"+serviceName+"/DeleteProfile", handler.DeleteProfile, opts...))

	// Active / Inactive Operations
	mux.Handle("/"+serviceName+"/ListActiveSessions", connect.NewUnaryHandler("/"+serviceName+"/ListActiveSessions", handler.ListActiveSessions, opts...))
	mux.Handle("/"+serviceName+"/KickActiveSession", connect.NewUnaryHandler("/"+serviceName+"/KickActiveSession", handler.KickActiveSession, opts...))
	mux.Handle("/"+serviceName+"/KickActiveSessions", connect.NewUnaryHandler("/"+serviceName+"/KickActiveSessions", handler.KickActiveSessions, opts...))
	mux.Handle("/"+serviceName+"/ListInactiveSecrets", connect.NewUnaryHandler("/"+serviceName+"/ListInactiveSecrets", handler.ListInactiveSecrets, opts...))

	// Streaming Live Updates
	mux.Handle("/"+serviceName+"/StreamActiveSessions", connect.NewServerStreamHandler("/"+serviceName+"/StreamActiveSessions", handler.StreamActiveSessions, opts...))
	mux.Handle("/"+serviceName+"/StreamActiveStats", connect.NewServerStreamHandler("/"+serviceName+"/StreamActiveStats", handler.StreamActiveStats, opts...))
	mux.Handle("/"+serviceName+"/StreamInactiveSecrets", connect.NewServerStreamHandler("/"+serviceName+"/StreamInactiveSecrets", handler.StreamInactiveSecrets, opts...))

	return "/" + serviceName + "/", mux
}
