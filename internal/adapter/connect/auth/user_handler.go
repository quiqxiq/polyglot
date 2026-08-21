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
	callerID, callerRoles, _ := iauth.IdentityFromContext(ctx)

	u, err := h.userUC.CreateUser(
		ctx,
		callerID,
		callerRoles,
		req.Msg.Username,
		req.Msg.Email,
		req.Msg.Password,
		req.Msg.Role,
		req.Msg.FullName,
		req.Msg.PhoneNumber,
		req.Msg.Specialization,
	)
	if err != nil {
		if errors.Is(err, userUC.ErrUserAlreadyExists) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		if errors.Is(err, userUC.ErrAdminCannotCreateAdminOrOwner) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
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

	callerID, callerRoles, _ := iauth.IdentityFromContext(ctx)

	u, err := h.userUC.UpdateUser(
		ctx,
		callerID,
		callerRoles,
		uint(req.Msg.Id),
		req.Msg.Username,
		req.Msg.Email,
		req.Msg.Role,
		req.Msg.FullName,
		req.Msg.PhoneNumber,
		req.Msg.Specialization,
	)
	if err != nil {
		if errors.Is(err, userUC.ErrCannotModifyOwner) ||
			errors.Is(err, userUC.ErrCannotModifyAdmin) ||
			errors.Is(err, userUC.ErrAdminCannotAssignAdminOrOwner) ||
			errors.Is(err, userUC.ErrCannotAssignOwnerRole) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		if errors.Is(err, userUC.ErrInvalidRole) || errors.Is(err, userUC.ErrLastOwnerDemotion) {
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

	callerID, callerRoles, _ := iauth.IdentityFromContext(ctx)

	if err := h.userUC.AdminResetPassword(ctx, callerID, callerRoles, uint(req.Msg.Id), req.Msg.NewPassword); err != nil {
		if errors.Is(err, userUC.ErrCannotModifyOwner) || errors.Is(err, userUC.ErrCannotModifyAdmin) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		if errors.Is(err, userUC.ErrPasswordTooShort) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ResetPasswordResponse{
		Success: true,
	}), nil
}

func (h *UserConnectHandler) ToggleActive(ctx context.Context, req *connect.Request[devicepb.ToggleActiveRequest]) (*connect.Response[devicepb.ToggleActiveResponse], error) {
	if req.Msg.Id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user id is required"))
	}

	callerID, callerRoles, exists := iauth.IdentityFromContext(ctx)
	if exists && callerID == uint(req.Msg.Id) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot toggle your own active status"))
	}

	user, err := h.userUC.GetUser(ctx, uint(req.Msg.Id))
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	newActive := !user.IsActive
	if err := h.userUC.ToggleStatus(ctx, callerID, callerRoles, uint(req.Msg.Id), newActive); err != nil {
		if errors.Is(err, userUC.ErrCannotModifyOwner) || errors.Is(err, userUC.ErrCannotModifyAdmin) || errors.Is(err, userUC.ErrSelfOperation) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
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

	callerID, callerRoles, exists := iauth.IdentityFromContext(ctx)
	if exists && callerID == uint(req.Msg.Id) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot delete your own account"))
	}

	if err := h.userUC.DeleteUser(ctx, callerID, callerRoles, uint(req.Msg.Id)); err != nil {
		if errors.Is(err, userUC.ErrCannotModifyOwner) || errors.Is(err, userUC.ErrCannotModifyAdmin) || errors.Is(err, userUC.ErrSelfOperation) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteUserResponse{
		Success: true,
	}), nil
}
