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
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.PPPService"

	// Secret Operations
	mux.Handle("/"+serviceName+"/ListSecrets", connect.NewUnaryHandler("/"+serviceName+"/ListSecrets", handler.ListSecrets, codecOpt))
	mux.Handle("/"+serviceName+"/GetSecret", connect.NewUnaryHandler("/"+serviceName+"/GetSecret", handler.GetSecret, codecOpt))
	mux.Handle("/"+serviceName+"/CreateSecret", connect.NewUnaryHandler("/"+serviceName+"/CreateSecret", handler.CreateSecret, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateSecret", connect.NewUnaryHandler("/"+serviceName+"/UpdateSecret", handler.UpdateSecret, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteSecret", connect.NewUnaryHandler("/"+serviceName+"/DeleteSecret", handler.DeleteSecret, codecOpt))
	mux.Handle("/"+serviceName+"/SetSecretDisabled", connect.NewUnaryHandler("/"+serviceName+"/SetSecretDisabled", handler.SetSecretDisabled, codecOpt))

	// Profile Operations
	mux.Handle("/"+serviceName+"/ListProfiles", connect.NewUnaryHandler("/"+serviceName+"/ListProfiles", handler.ListProfiles, codecOpt))
	mux.Handle("/"+serviceName+"/GetProfile", connect.NewUnaryHandler("/"+serviceName+"/GetProfile", handler.GetProfile, codecOpt))
	mux.Handle("/"+serviceName+"/CreateProfile", connect.NewUnaryHandler("/"+serviceName+"/CreateProfile", handler.CreateProfile, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateProfile", connect.NewUnaryHandler("/"+serviceName+"/UpdateProfile", handler.UpdateProfile, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteProfile", connect.NewUnaryHandler("/"+serviceName+"/DeleteProfile", handler.DeleteProfile, codecOpt))

	// Active / Inactive Operations
	mux.Handle("/"+serviceName+"/ListActiveSessions", connect.NewUnaryHandler("/"+serviceName+"/ListActiveSessions", handler.ListActiveSessions, codecOpt))
	mux.Handle("/"+serviceName+"/KickActiveSession", connect.NewUnaryHandler("/"+serviceName+"/KickActiveSession", handler.KickActiveSession, codecOpt))
	mux.Handle("/"+serviceName+"/KickActiveSessions", connect.NewUnaryHandler("/"+serviceName+"/KickActiveSessions", handler.KickActiveSessions, codecOpt))
	mux.Handle("/"+serviceName+"/ListInactiveSecrets", connect.NewUnaryHandler("/"+serviceName+"/ListInactiveSecrets", handler.ListInactiveSecrets, codecOpt))

	// Streaming Live Updates
	mux.Handle("/"+serviceName+"/StreamActiveSessions", connect.NewServerStreamHandler("/"+serviceName+"/StreamActiveSessions", handler.StreamActiveSessions, codecOpt))
	mux.Handle("/"+serviceName+"/StreamActiveStats", connect.NewServerStreamHandler("/"+serviceName+"/StreamActiveStats", handler.StreamActiveStats, codecOpt))
	mux.Handle("/"+serviceName+"/StreamInactiveSecrets", connect.NewServerStreamHandler("/"+serviceName+"/StreamInactiveSecrets", handler.StreamInactiveSecrets, codecOpt))

	return "/" + serviceName + "/", mux
}
