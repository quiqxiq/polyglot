package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameRequired  = errors.New("username is required")
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrInvalidRole       = errors.New("invalid role")
	ErrUserAlreadyExists = errors.New("username or email already taken")
	ErrSelfOperation     = errors.New("cannot perform this operation on your own account")
)

var KnownRoles = map[string]bool{
	"owner":   true,
	"admin":   true,
	"agent":   true,
	"teknisi": true,
}

type ManageUserUseCase struct {
	repo  port.UserRepository
	roles port.RoleAuthorizer
}

func NewManageUserUseCase(repo port.UserRepository, roles port.RoleAuthorizer) *ManageUserUseCase {
	return &ManageUserUseCase{
		repo:  repo,
		roles: roles,
	}
}

func (u *ManageUserUseCase) ListUsers(ctx context.Context, page, pageSize int, search string) ([]*customer.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return u.repo.List(ctx, page, pageSize, search)
}

func (u *ManageUserUseCase) GetUser(ctx context.Context, id uint) (*customer.User, error) {
	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (u *ManageUserUseCase) GetRoles(ctx context.Context, id uint) ([]string, error) {
	if u.roles == nil {
		return nil, nil
	}
	return u.roles.GetRolesForUser(fmt.Sprintf("%d", id))
}

func (u *ManageUserUseCase) GetPermissions(ctx context.Context, id uint) ([]string, error) {
	if u.roles == nil {
		return nil, nil
	}
	return u.roles.GetImplicitPermissionsForUser(fmt.Sprintf("%d", id))
}

func (u *ManageUserUseCase) CreateUser(ctx context.Context, username, email, password, role string) (*customer.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	role = strings.ToLower(strings.TrimSpace(role))

	if username == "" {
		return nil, ErrUsernameRequired
	}
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}
	if !KnownRoles[role] {
		return nil, ErrInvalidRole
	}

	if existing, _ := u.repo.FindByUsername(ctx, username); existing != nil {
		return nil, ErrUserAlreadyExists
	}
	if email != "" {
		if existing, _ := u.repo.FindByEmail(ctx, email); existing != nil {
			return nil, ErrUserAlreadyExists
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &customer.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		IsActive:     true,
	}

	if err := u.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	if u.roles != nil {
		_, _ = u.roles.AddRoleForUser(fmt.Sprintf("%d", newUser.ID), role)
	}

	logger.WithComponent("UserUseCase").WithField("username", username).Info("user created successfully")
	return newUser, nil
}

func (u *ManageUserUseCase) UpdateUser(ctx context.Context, id uint, username, email, role string) (*customer.User, error) {
	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	role = strings.ToLower(strings.TrimSpace(role))

	if username != "" {
		user.Username = username
	}
	if email != "" {
		user.Email = email
	}
	if role != "" {
		if !KnownRoles[role] {
			return nil, ErrInvalidRole
		}
		user.Role = role
	}

	if err := u.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	if role != "" && u.roles != nil {
		_, _ = u.roles.DeleteRolesForUser(fmt.Sprintf("%d", id))
		_, _ = u.roles.AddRoleForUser(fmt.Sprintf("%d", id), role)
	}

	logger.WithComponent("UserUseCase").WithField("user_id", id).Info("user updated successfully")
	return user, nil
}

func (u *ManageUserUseCase) DeleteUser(ctx context.Context, id uint) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	if u.roles != nil {
		_, _ = u.roles.DeleteRolesForUser(fmt.Sprintf("%d", id))
	}
	logger.WithComponent("UserUseCase").WithField("user_id", id).Info("user deleted successfully")
	return nil
}

func (u *ManageUserUseCase) ChangePassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("current password does not match")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	return u.repo.UpdatePassword(ctx, id, string(hash))
}

func (u *ManageUserUseCase) AdminResetPassword(ctx context.Context, id uint, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return u.repo.UpdatePassword(ctx, id, string(hash))
}

func (u *ManageUserUseCase) ToggleStatus(ctx context.Context, id uint, isActive bool) error {
	return u.repo.UpdateStatus(ctx, id, isActive)
}
