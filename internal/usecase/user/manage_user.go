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
	"github.com/quixiq/polyglot/pkg/phone"
)

var (
	ErrUserNotFound                   = errors.New("user not found")
	ErrUsernameRequired               = errors.New("username is required")
	ErrPasswordTooShort               = errors.New("password must be at least 8 characters")
	ErrInvalidRole                    = errors.New("invalid role")
	ErrUserAlreadyExists              = errors.New("username or email already taken")
	ErrSelfOperation                  = errors.New("cannot perform this operation on your own account")
	ErrCannotModifyOwner              = errors.New("owner account can only be modified by itself")
	ErrCannotModifyAdmin              = errors.New("admin account can only be modified by itself or owner")
	ErrAdminCannotCreateAdminOrOwner  = errors.New("admin can only create accounts with role agent or teknisi")
	ErrAdminCannotAssignAdminOrOwner  = errors.New("admin can only assign role agent or teknisi")
	ErrCannotAssignOwnerRole          = errors.New("only owner can assign owner role")
	ErrLastOwnerDemotion              = errors.New("system requires at least one active owner account")
)

var KnownRoles = map[string]bool{
	"owner":      true,
	"admin":      true,
	"agent":      true,
	"teknisi":    true,
	"technician": true,
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

func isOwnerRole(roles []string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, "owner") {
			return true
		}
	}
	return false
}

func (u *ManageUserUseCase) countActiveOwners(ctx context.Context) (int, error) {
	users, _, err := u.repo.List(ctx, 1, 1000, "")
	if err != nil {
		return 0, err
	}
	count := 0
	for _, usr := range users {
		if strings.EqualFold(usr.Role, "owner") && usr.IsActive {
			count++
		}
	}
	return count, nil
}

func (u *ManageUserUseCase) CreateUser(
	ctx context.Context,
	actorID uint,
	actorRoles []string,
	username, email, password, role, fullName, phoneNumber, specialization string,
) (*customer.User, error) {
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

	// Hierarchy check: Non-owner (admin) can only create agent or teknisi
	if !isOwnerRole(actorRoles) {
		if role == "owner" || role == "admin" {
			return nil, ErrAdminCannotCreateAdminOrOwner
		}
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
		Username:       username,
		Email:          email,
		PasswordHash:   string(hash),
		Role:           role,
		FullName:       strings.TrimSpace(fullName),
		PhoneNumber:    phone.Normalize(phoneNumber),
		Specialization: strings.TrimSpace(specialization),
		IsActive:       true,
	}

	if err := u.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	if u.roles != nil {
		_, _ = u.roles.AddRoleForUser(fmt.Sprintf("%d", newUser.ID), role)
	}

	logger.WithComponent("UserUseCase").WithFields(map[string]any{
		"username": username,
		"role":     role,
		"actor_id": actorID,
	}).Info("user created successfully")
	return newUser, nil
}

func (u *ManageUserUseCase) UpdateUser(
	ctx context.Context,
	actorID uint,
	actorRoles []string,
	targetID uint,
	username, email, role, fullName, phoneNumber, specialization string,
) (*customer.User, error) {
	targetUser, err := u.repo.FindByID(ctx, targetID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	isSelf := actorID > 0 && actorID == targetID
	actorIsOwner := isOwnerRole(actorRoles)

	// Rule 1: Owner account protection - ONLY the owner account itself can modify its data.
	if strings.EqualFold(targetUser.Role, "owner") && !isSelf {
		return nil, ErrCannotModifyOwner
	}

	// Rule 2: Admin account protection - Admin can modify self; Owner can modify admin;
	// Another admin (or lower role) CANNOT modify an admin.
	if strings.EqualFold(targetUser.Role, "admin") && !isSelf && !actorIsOwner {
		return nil, ErrCannotModifyAdmin
	}

	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	role = strings.ToLower(strings.TrimSpace(role))

	// Rule 3: Role assignment restrictions
	if role != "" {
		if !KnownRoles[role] {
			return nil, ErrInvalidRole
		}

		if !actorIsOwner {
			// Non-owner (admin) cannot assign "owner" role to anyone (including self)
			if role == "owner" {
				return nil, ErrCannotAssignOwnerRole
			}
			// When modifying other users (agent/teknisi), admin cannot promote them to "admin" or "owner"
			if !isSelf && role == "admin" {
				return nil, ErrAdminCannotAssignAdminOrOwner
			}
		}

		// Rule 4: If an owner is demoting themselves from owner to non-owner,
		// ensure there is at least one other active owner in the system.
		if isSelf && strings.EqualFold(targetUser.Role, "owner") && !strings.EqualFold(role, "owner") {
			ownersCount, err := u.countActiveOwners(ctx)
			if err == nil && ownersCount <= 1 {
				return nil, ErrLastOwnerDemotion
			}
		}

		targetUser.Role = role
	}

	if username != "" {
		targetUser.Username = username
	}
	if email != "" {
		targetUser.Email = email
	}
	if fullName != "" {
		targetUser.FullName = strings.TrimSpace(fullName)
	}
	if phoneNumber != "" {
		targetUser.PhoneNumber = phone.Normalize(phoneNumber)
	}
	if specialization != "" {
		targetUser.Specialization = strings.TrimSpace(specialization)
	}

	if err := u.repo.Update(ctx, targetUser); err != nil {
		return nil, err
	}

	if role != "" && u.roles != nil {
		_, _ = u.roles.DeleteRolesForUser(fmt.Sprintf("%d", targetID))
		_, _ = u.roles.AddRoleForUser(fmt.Sprintf("%d", targetID), role)
	}

	logger.WithComponent("UserUseCase").WithFields(map[string]any{
		"target_id": targetID,
		"actor_id":  actorID,
	}).Info("user updated successfully")
	return targetUser, nil
}

func (u *ManageUserUseCase) DeleteUser(ctx context.Context, actorID uint, actorRoles []string, targetID uint) error {
	if actorID > 0 && actorID == targetID {
		return ErrSelfOperation
	}

	targetUser, err := u.repo.FindByID(ctx, targetID)
	if err != nil {
		return ErrUserNotFound
	}

	// Owner account cannot be deleted by anyone
	if strings.EqualFold(targetUser.Role, "owner") {
		return ErrCannotModifyOwner
	}

	// Admin account can only be deleted by owner
	if strings.EqualFold(targetUser.Role, "admin") && !isOwnerRole(actorRoles) {
		return ErrCannotModifyAdmin
	}

	if err := u.repo.Delete(ctx, targetID); err != nil {
		return err
	}
	if u.roles != nil {
		_, _ = u.roles.DeleteRolesForUser(fmt.Sprintf("%d", targetID))
	}
	logger.WithComponent("UserUseCase").WithFields(map[string]any{
		"target_id": targetID,
		"actor_id":  actorID,
	}).Info("user deleted successfully")
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

func (u *ManageUserUseCase) AdminResetPassword(ctx context.Context, actorID uint, actorRoles []string, targetID uint, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	targetUser, err := u.repo.FindByID(ctx, targetID)
	if err != nil {
		return ErrUserNotFound
	}

	isSelf := actorID > 0 && actorID == targetID

	// Owner account password can only be reset by self
	if strings.EqualFold(targetUser.Role, "owner") && !isSelf {
		return ErrCannotModifyOwner
	}

	// Admin account password can only be reset by self or owner
	if strings.EqualFold(targetUser.Role, "admin") && !isSelf && !isOwnerRole(actorRoles) {
		return ErrCannotModifyAdmin
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return u.repo.UpdatePassword(ctx, targetID, string(hash))
}

func (u *ManageUserUseCase) ToggleStatus(ctx context.Context, actorID uint, actorRoles []string, targetID uint, isActive bool) error {
	if actorID > 0 && actorID == targetID {
		return ErrSelfOperation
	}

	targetUser, err := u.repo.FindByID(ctx, targetID)
	if err != nil {
		return ErrUserNotFound
	}

	// Owner account cannot be deactivated
	if strings.EqualFold(targetUser.Role, "owner") {
		return ErrCannotModifyOwner
	}

	// Admin account can only be toggled by owner
	if strings.EqualFold(targetUser.Role, "admin") && !isOwnerRole(actorRoles) {
		return ErrCannotModifyAdmin
	}

	return u.repo.UpdateStatus(ctx, targetID, isActive)
}
