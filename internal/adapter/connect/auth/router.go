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
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.AuthService"
	mux.Handle("/"+serviceName+"/Login", connect.NewUnaryHandler("/"+serviceName+"/Login", handler.Login, opts...))
	mux.Handle("/"+serviceName+"/GetMe", connect.NewUnaryHandler("/"+serviceName+"/GetMe", handler.GetMe, opts...))
	mux.Handle("/"+serviceName+"/UpdateMe", connect.NewUnaryHandler("/"+serviceName+"/UpdateMe", handler.UpdateMe, opts...))
	mux.Handle("/"+serviceName+"/ChangePassword", connect.NewUnaryHandler("/"+serviceName+"/ChangePassword", handler.ChangePassword, opts...))
	mux.Handle("/"+serviceName+"/RefreshToken", connect.NewUnaryHandler("/"+serviceName+"/RefreshToken", handler.RefreshToken, opts...))
	mux.Handle("/"+serviceName+"/Logout", connect.NewUnaryHandler("/"+serviceName+"/Logout", handler.Logout, opts...))

	return "/" + serviceName + "/", mux
}

// NewUserServiceHandler creates the Connect http.Handler for UserService.
func NewUserServiceHandler(uc *userUC.ManageUserUseCase) (string, http.Handler) {
	handler := NewUserConnectHandler(uc)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.UserService"
	mux.Handle("/"+serviceName+"/ListUsers", connect.NewUnaryHandler("/"+serviceName+"/ListUsers", handler.ListUsers, opts...))
	mux.Handle("/"+serviceName+"/CreateUser", connect.NewUnaryHandler("/"+serviceName+"/CreateUser", handler.CreateUser, opts...))
	mux.Handle("/"+serviceName+"/UpdateUser", connect.NewUnaryHandler("/"+serviceName+"/UpdateUser", handler.UpdateUser, opts...))
	mux.Handle("/"+serviceName+"/ResetPassword", connect.NewUnaryHandler("/"+serviceName+"/ResetPassword", handler.ResetPassword, opts...))
	mux.Handle("/"+serviceName+"/ToggleActive", connect.NewUnaryHandler("/"+serviceName+"/ToggleActive", handler.ToggleActive, opts...))
	mux.Handle("/"+serviceName+"/DeleteUser", connect.NewUnaryHandler("/"+serviceName+"/DeleteUser", handler.DeleteUser, opts...))
	mux.Handle("/"+serviceName+"/AssignDevicesToUser", connect.NewUnaryHandler("/"+serviceName+"/AssignDevicesToUser", handler.AssignDevicesToUser, opts...))
	mux.Handle("/"+serviceName+"/ListUserAccessibleDevices", connect.NewUnaryHandler("/"+serviceName+"/ListUserAccessibleDevices", handler.ListUserAccessibleDevices, opts...))

	return "/" + serviceName + "/", mux
}
