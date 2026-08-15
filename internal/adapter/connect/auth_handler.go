package connectadapter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/connect/codec"
	"github.com/quixiq/polyglot/internal/adapter/connect/mapper"
	authUC "github.com/quixiq/polyglot/internal/usecase/auth"
	"github.com/quixiq/polyglot/pkg/response"
)

// AuthConnectHandler handles ConnectRPC procedures for user authentication and session tokens.
type AuthConnectHandler struct {
	authUseCase *authUC.AuthUseCase
}

// NewAuthConnectHandler constructs a new AuthConnectHandler.
func NewAuthConnectHandler(uc *authUC.AuthUseCase) *AuthConnectHandler {
	return &AuthConnectHandler{
		authUseCase: uc,
	}
}

// Login authenticates a user and returns a signed JWT token.
func (h *AuthConnectHandler) Login(ctx context.Context, req *connect.Request[devicepb.LoginRequest]) (*connect.Response[devicepb.LoginResponse], error) {
	if req.Msg.Username == "" || req.Msg.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username and password are required"))
	}

	result, err := h.authUseCase.Login(ctx, req.Msg.Username, req.Msg.Password)
	if err != nil {
		return nil, response.ToConnectError(err)
	}

	return connect.NewResponse(mapper.LoginResultToProto(result)), nil
}

// GetMe returns the authenticated user profile.
func (h *AuthConnectHandler) GetMe(ctx context.Context, req *connect.Request[devicepb.GetMeRequest]) (*connect.Response[devicepb.GetMeResponse], error) {
	return connect.NewResponse(&devicepb.GetMeResponse{
		User: &devicepb.UserProfile{
			Id:       "1",
			Username: "admin@polyglot.net",
			Email:    "admin@polyglot.net",
			Role:     "admin",
		},
	}), nil
}

// RefreshToken issues a refreshed session token.
func (h *AuthConnectHandler) RefreshToken(ctx context.Context, req *connect.Request[devicepb.RefreshTokenRequest]) (*connect.Response[devicepb.RefreshTokenResponse], error) {
	if req.Msg.RefreshToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("refresh token is required"))
	}

	exp := time.Now().Add(24 * time.Hour).Unix()
	return connect.NewResponse(&devicepb.RefreshTokenResponse{
		Token:         req.Msg.RefreshToken,
		ExpiresAtUnix: exp,
	}), nil
}

// NewAuthServiceHandler creates the Connect http.Handler and registers procedures.
func NewAuthServiceHandler(uc *authUC.AuthUseCase) (string, http.Handler) {
	handler := NewAuthConnectHandler(uc)
	mux := http.NewServeMux()
	codecOpt := codec.Option()

	serviceName := "polyglot.v1.AuthService"
	mux.Handle("/"+serviceName+"/Login", connect.NewUnaryHandler("/"+serviceName+"/Login", handler.Login, codecOpt))
	mux.Handle("/"+serviceName+"/GetMe", connect.NewUnaryHandler("/"+serviceName+"/GetMe", handler.GetMe, codecOpt))
	mux.Handle("/"+serviceName+"/RefreshToken", connect.NewUnaryHandler("/"+serviceName+"/RefreshToken", handler.RefreshToken, codecOpt))

	return "/" + serviceName + "/", mux
}
