package auth

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	userUC "github.com/quixiq/polyglot/internal/usecase/user"
	"github.com/quixiq/polyglot/pkg/response"
)

type UserConnectHandler struct {
	userUC *userUC.ManageUserUseCase
}

func NewUserConnectHandler(uc *userUC.ManageUserUseCase) *UserConnectHandler {
	return &UserConnectHandler{
		userUC: uc,
	}
}

func (h *UserConnectHandler) ListUsers(ctx context.Context, req *connect.Request[devicepb.ListUsersRequest]) (*connect.Response[devicepb.ListUsersResponse], error) {
	page := int(req.Msg.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.Msg.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	users, total, err := h.userUC.ListUsers(ctx, page, pageSize, req.Msg.Search)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	pbUsers := make([]*devicepb.User, len(users))
	for i, u := range users {
		roles, _ := h.userUC.GetRoles(ctx, u.ID)
		pbUsers[i] = DomainUserToPb(u, roles)
	}

	return connect.NewResponse(&devicepb.ListUsersResponse{
		Users:    pbUsers,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}), nil
}

func (h *UserConnectHandler) CreateUser(ctx context.Context, req *connect.Request[devicepb.CreateUserRequest]) (*connect.Response[devicepb.CreateUserResponse], error) {
	u, err := h.userUC.CreateUser(ctx, req.Msg.Username, req.Msg.Email, req.Msg.Password, req.Msg.Role, req.Msg.FullName, req.Msg.PhoneNumber, req.Msg.Specialization)
	if err != nil {
		if errors.Is(err, userUC.ErrUserAlreadyExists) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		if errors.Is(err, userUC.ErrInvalidRole) || errors.Is(err, userUC.ErrUsernameRequired) || errors.Is(err, userUC.ErrPasswordTooShort) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, response.MapDomainError(err)
	}

	roles, _ := h.userUC.GetRoles(ctx, u.ID)
	return connect.NewResponse(&devicepb.CreateUserResponse{
		User: DomainUserToPb(u, roles),
	}), nil
}

func (h *UserConnectHandler) UpdateUser(ctx context.Context, req *connect.Request[devicepb.UpdateUserRequest]) (*connect.Response[devicepb.UpdateUserResponse], error) {
	if req.Msg.Id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user id is required"))
	}

	u, err := h.userUC.UpdateUser(ctx, uint(req.Msg.Id), req.Msg.Username, req.Msg.Email, req.Msg.Role, req.Msg.FullName, req.Msg.PhoneNumber, req.Msg.Specialization)
	if err != nil {
		if errors.Is(err, userUC.ErrInvalidRole) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, response.MapDomainError(err)
	}

	roles, _ := h.userUC.GetRoles(ctx, u.ID)
	return connect.NewResponse(&devicepb.UpdateUserResponse{
		User: DomainUserToPb(u, roles),
	}), nil
}

func (h *UserConnectHandler) ResetPassword(ctx context.Context, req *connect.Request[devicepb.ResetPasswordRequest]) (*connect.Response[devicepb.ResetPasswordResponse], error) {
	if req.Msg.Id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user id is required"))
	}

	if err := h.userUC.AdminResetPassword(ctx, uint(req.Msg.Id), req.Msg.NewPassword); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&devicepb.ResetPasswordResponse{
		Success: true,
	}), nil
}

func (h *UserConnectHandler) ToggleActive(ctx context.Context, req *connect.Request[devicepb.ToggleActiveRequest]) (*connect.Response[devicepb.ToggleActiveResponse], error) {
	if req.Msg.Id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user id is required"))
	}

	callerID, _, exists := iauth.IdentityFromContext(ctx)
	if exists && callerID == uint(req.Msg.Id) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot toggle your own active status"))
	}

	user, err := h.userUC.GetUser(ctx, uint(req.Msg.Id))
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	newActive := !user.IsActive
	if err := h.userUC.ToggleStatus(ctx, uint(req.Msg.Id), newActive); err != nil {
		return nil, response.MapDomainError(err)
	}

	updated, err := h.userUC.GetUser(ctx, uint(req.Msg.Id))
	if err != nil {
		updated = user
		updated.IsActive = newActive
	}

	roles, _ := h.userUC.GetRoles(ctx, updated.ID)
	return connect.NewResponse(&devicepb.ToggleActiveResponse{
		User: DomainUserToPb(updated, roles),
	}), nil
}

func (h *UserConnectHandler) DeleteUser(ctx context.Context, req *connect.Request[devicepb.DeleteUserRequest]) (*connect.Response[devicepb.DeleteUserResponse], error) {
	if req.Msg.Id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user id is required"))
	}

	callerID, _, exists := iauth.IdentityFromContext(ctx)
	if exists && callerID == uint(req.Msg.Id) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot delete your own account"))
	}

	if err := h.userUC.DeleteUser(ctx, uint(req.Msg.Id)); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteUserResponse{
		Success: true,
	}), nil
}
