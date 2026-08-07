package connectadapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/crypto/bcrypt"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
)

type AuthConnectHandler struct {
	pgStore    *postgres.Store
	jwtService *auth.JWTService
}

func NewAuthConnectHandler(pgStore *postgres.Store, jwtService *auth.JWTService) *AuthConnectHandler {
	return &AuthConnectHandler{
		pgStore:    pgStore,
		jwtService: jwtService,
	}
}

func (h *AuthConnectHandler) Login(ctx context.Context, req *connect.Request[devicepb.LoginRequest]) (*connect.Response[devicepb.LoginResponse], error) {
	if req.Msg.Username == "" || req.Msg.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username and password are required"))
	}

	if h.pgStore == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("database store unavailable"))
	}

	user, err := h.pgStore.FindUserByEmail(req.Msg.Username)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid username or password"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Msg.Password)); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid username or password"))
	}

	tokenStr, err := h.jwtService.GenerateToken(user.ID, user.Email, user.Role, user.TenantID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate auth token"))
	}

	exp := time.Now().Add(24 * time.Hour).Unix()

	return connect.NewResponse(&devicepb.LoginResponse{
		Token: tokenStr,
		User: &devicepb.UserProfile{
			Id:       fmt.Sprintf("%d", user.ID),
			Username: user.Email,
			Email:    user.Email,
			Role:     user.Role,
		},
		ExpiresAtUnix: exp,
	}), nil
}

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

func NewAuthServiceHandler(pgStore *postgres.Store, jwtService *auth.JWTService) (string, http.Handler) {
	handler := NewAuthConnectHandler(pgStore, jwtService)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(connectJSONCodec{})

	serviceName := "polyglot.v1.AuthService"
	mux.Handle("/"+serviceName+"/Login", connect.NewUnaryHandler(
		"/"+serviceName+"/Login",
		handler.Login,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetMe", connect.NewUnaryHandler(
		"/"+serviceName+"/GetMe",
		handler.GetMe,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/RefreshToken", connect.NewUnaryHandler(
		"/"+serviceName+"/RefreshToken",
		handler.RefreshToken,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
