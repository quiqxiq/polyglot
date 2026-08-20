package auth

import (
	"net/http"

	"connectrpc.com/connect"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	authUC "github.com/quixiq/polyglot/internal/usecase/auth"
	userUC "github.com/quixiq/polyglot/internal/usecase/user"
)

// NewAuthServiceHandler creates the Connect http.Handler for AuthService.
func NewAuthServiceHandler(
	authUC *authUC.AuthUseCase,
	refreshUC *authUC.RefreshTokenUseCase,
	userUC *userUC.ManageUserUseCase,
	secure bool,
) (string, http.Handler) {
	handler := NewAuthConnectHandler(authUC, refreshUC, userUC, secure)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.AuthService"
	mux.Handle("/"+serviceName+"/Login", connect.NewUnaryHandler("/"+serviceName+"/Login", handler.Login, codecOpt))
	mux.Handle("/"+serviceName+"/GetMe", connect.NewUnaryHandler("/"+serviceName+"/GetMe", handler.GetMe, codecOpt))
	mux.Handle("/"+serviceName+"/RefreshToken", connect.NewUnaryHandler("/"+serviceName+"/RefreshToken", handler.RefreshToken, codecOpt))
	mux.Handle("/"+serviceName+"/Logout", connect.NewUnaryHandler("/"+serviceName+"/Logout", handler.Logout, codecOpt))

	return "/" + serviceName + "/", mux
}

// NewUserServiceHandler creates the Connect http.Handler for UserService.
func NewUserServiceHandler(uc *userUC.ManageUserUseCase) (string, http.Handler) {
	handler := NewUserConnectHandler(uc)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.UserService"
	mux.Handle("/"+serviceName+"/ListUsers", connect.NewUnaryHandler("/"+serviceName+"/ListUsers", handler.ListUsers, codecOpt))
	mux.Handle("/"+serviceName+"/CreateUser", connect.NewUnaryHandler("/"+serviceName+"/CreateUser", handler.CreateUser, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateUser", connect.NewUnaryHandler("/"+serviceName+"/UpdateUser", handler.UpdateUser, codecOpt))
	mux.Handle("/"+serviceName+"/ResetPassword", connect.NewUnaryHandler("/"+serviceName+"/ResetPassword", handler.ResetPassword, codecOpt))
	mux.Handle("/"+serviceName+"/ToggleActive", connect.NewUnaryHandler("/"+serviceName+"/ToggleActive", handler.ToggleActive, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteUser", connect.NewUnaryHandler("/"+serviceName+"/DeleteUser", handler.DeleteUser, codecOpt))

	return "/" + serviceName + "/", mux
}
