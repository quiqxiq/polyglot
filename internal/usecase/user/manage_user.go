package user

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/phone"
)

// Error sentinels for this use case live in internal/domain/customer
// (errors.go) per DEVELOPMENT-GUIDELINES.md §6.

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
		return nil, customer.ErrUserNotFound
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

func (u *ManageUserUseCase) validateDeviceAssignments(ctx context.Context, actorID uint, actorRoles []string, deviceIDs []string) error {
	if len(deviceIDs) == 0 || isOwnerRole(actorRoles) {
		return nil
	}
	for _, devID := range deviceIDs {
		devID = strings.TrimSpace(devID)
		if devID == "" {
			continue
		}
		accessible, err := u.repo.IsDeviceAccessibleByUser(ctx, actorID, devID)
		if err != nil {
			return err
		}
		if !accessible {
			return fmt.Errorf("%w: device %s", customer.ErrUnauthorizedDeviceAssign, devID)
		}
	}
	return nil
}

func (u *ManageUserUseCase) CreateUser(
	ctx context.Context,
	actorID uint,
	actorRoles []string,
	username, email, password, role, fullName, phoneNumber, specialization string,
	assignedDeviceIDs []string,
) (*customer.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	role = strings.ToLower(strings.TrimSpace(role))

	if username == "" {
		return nil, customer.ErrUsernameRequired
	}
	if len(password) < 8 {
		return nil, customer.ErrPasswordTooShort
	}
	if !KnownRoles[role] {
		return nil, customer.ErrInvalidRole
	}

	// Hierarchy check: Non-owner (admin) can only create agent or teknisi
	if !isOwnerRole(actorRoles) {
		if role == "owner" || role == "admin" {
			return nil, customer.ErrAdminCannotCreateAdmin
		}
	}

	if err := u.validateDeviceAssignments(ctx, actorID, actorRoles, assignedDeviceIDs); err != nil {
		return nil, err
	}

	if existing, _ := u.repo.FindByUsername(ctx, username); existing != nil {
		return nil, customer.ErrUserAlreadyExists
	}
	if email != "" {
		if existing, _ := u.repo.FindByEmail(ctx, email); existing != nil {
			return nil, customer.ErrUserAlreadyExists
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

	if len(assignedDeviceIDs) > 0 {
		var assignedBy *uint
		if actorID > 0 {
			assignedBy = &actorID
		}
		if err := u.repo.AssignDevices(ctx, newUser.ID, assignedDeviceIDs, assignedBy); err != nil {
			logger.WithComponent("UserUseCase").WithError(err).WithField("user_id", newUser.ID).Error("failed to assign devices")
		} else {
			newUser.AssignedDeviceIDs = assignedDeviceIDs
		}
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
	assignedDeviceIDs []string,
) (*customer.User, error) {
	targetUser, err := u.repo.FindByID(ctx, targetID)
	if err != nil {
		return nil, customer.ErrUserNotFound
	}

	isSelf := actorID > 0 && actorID == targetID
	actorIsOwner := isOwnerRole(actorRoles)

	// Rule 1: Owner account protection - ONLY the owner account itself can modify its data.
	if strings.EqualFold(targetUser.Role, "owner") && !isSelf {
		return nil, customer.ErrCannotModifyOwner
	}

	// Rule 2: Admin account protection - Admin can modify self; Owner can modify admin;
	// Another admin (or lower role) CANNOT modify an admin.
	if strings.EqualFold(targetUser.Role, "admin") && !isSelf && !actorIsOwner {
		return nil, customer.ErrCannotModifyAdmin
	}

	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	role = strings.ToLower(strings.TrimSpace(role))

	// Rule 3: Role assignment restrictions
	if role != "" {
		if !KnownRoles[role] {
			return nil, customer.ErrInvalidRole
		}

		if !actorIsOwner {
			// Non-owner (admin) cannot assign "owner" role to anyone (including self)
			if role == "owner" {
				return nil, customer.ErrCannotAssignOwnerRole
			}
			// When modifying other users (agent/teknisi), admin cannot promote them to "admin" or "owner"
			if !isSelf && role == "admin" {
				return nil, customer.ErrAdminCannotAssignElevated
			}
		}

		// Rule 4: If an owner is demoting themselves from owner to non-owner,
		// ensure there is at least one other active owner in the system.
		if isSelf && strings.EqualFold(targetUser.Role, "owner") && !strings.EqualFold(role, "owner") {
			ownersCount, err := u.countActiveOwners(ctx)
			if err == nil && ownersCount <= 1 {
				return nil, customer.ErrLastOwnerDemotion
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

	if len(assignedDeviceIDs) > 0 || actorIsOwner {
		if err := u.validateDeviceAssignments(ctx, actorID, actorRoles, assignedDeviceIDs); err != nil {
			return nil, err
		}
		var assignedBy *uint
		if actorID > 0 {
			assignedBy = &actorID
		}
		if err := u.repo.AssignDevices(ctx, targetID, assignedDeviceIDs, assignedBy); err != nil {
			logger.WithComponent("UserUseCase").WithError(err).WithField("user_id", targetID).Error("failed to update device assignments")
		} else {
			targetUser.AssignedDeviceIDs = assignedDeviceIDs
		}
	}

	logger.WithComponent("UserUseCase").WithFields(map[string]any{
		"target_id": targetID,
		"actor_id":  actorID,
	}).Info("user updated successfully")
	return targetUser, nil
}

func (u *ManageUserUseCase) AssignDevicesToUser(
	ctx context.Context,
	actorID uint,
	actorRoles []string,
	targetUserID uint,
	deviceIDs []string,
) ([]string, error) {
	if targetUserID == 0 {
		return nil, customer.ErrUserNotFound
	}
	targetUser, err := u.repo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, customer.ErrUserNotFound
	}

	actorIsOwner := isOwnerRole(actorRoles)
	if strings.EqualFold(targetUser.Role, "owner") && !actorIsOwner {
		return nil, customer.ErrCannotModifyOwner
	}
	if strings.EqualFold(targetUser.Role, "admin") && !actorIsOwner && actorID != targetUserID {
		return nil, customer.ErrCannotModifyAdmin
	}

	if err := u.validateDeviceAssignments(ctx, actorID, actorRoles, deviceIDs); err != nil {
		return nil, err
	}

	var assignedBy *uint
	if actorID > 0 {
		assignedBy = &actorID
	}
	if err := u.repo.AssignDevices(ctx, targetUserID, deviceIDs, assignedBy); err != nil {
		return nil, err
	}

	return u.repo.GetAssignedDeviceIDs(ctx, targetUserID)
}

func (u *ManageUserUseCase) ListUserAccessibleDevices(
	ctx context.Context,
	actorID uint,
	actorRoles []string,
	targetUserID uint,
) ([]string, error) {
	if targetUserID == 0 {
		targetUserID = actorID
	}
	targetUser, err := u.repo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, customer.ErrUserNotFound
	}
	if strings.EqualFold(targetUser.Role, "owner") {
		return []string{"*"}, nil
	}
	return u.repo.GetAssignedDeviceIDs(ctx, targetUserID)
}

func (u *ManageUserUseCase) DeleteUser(ctx context.Context, actorID uint, actorRoles []string, targetID uint) error {
	if actorID > 0 && actorID == targetID {
		return customer.ErrSelfOperation
	}

	targetUser, err := u.repo.FindByID(ctx, targetID)
	if err != nil {
		return customer.ErrUserNotFound
	}

	// Owner account cannot be deleted by anyone
	if strings.EqualFold(targetUser.Role, "owner") {
		return customer.ErrCannotModifyOwner
	}

	// Admin account can only be deleted by owner
	if strings.EqualFold(targetUser.Role, "admin") && !isOwnerRole(actorRoles) {
		return customer.ErrCannotModifyAdmin
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
		return customer.ErrPasswordTooShort
	}

	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return customer.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return customer.ErrWrongPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	return u.repo.UpdatePassword(ctx, id, string(hash))
}

func (u *ManageUserUseCase) AdminResetPassword(ctx context.Context, actorID uint, actorRoles []string, targetID uint, newPassword string) error {
	if len(newPassword) < 8 {
		return customer.ErrPasswordTooShort
	}

	targetUser, err := u.repo.FindByID(ctx, targetID)
	if err != nil {
		return customer.ErrUserNotFound
	}

	isSelf := actorID > 0 && actorID == targetID

	// Owner account password can only be reset by self
	if strings.EqualFold(targetUser.Role, "owner") && !isSelf {
		return customer.ErrCannotModifyOwner
	}

	// Admin account password can only be reset by self or owner
	if strings.EqualFold(targetUser.Role, "admin") && !isSelf && !isOwnerRole(actorRoles) {
		return customer.ErrCannotModifyAdmin
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return u.repo.UpdatePassword(ctx, targetID, string(hash))
}

func (u *ManageUserUseCase) ToggleStatus(ctx context.Context, actorID uint, actorRoles []string, targetID uint, isActive bool) error {
	if actorID > 0 && actorID == targetID {
		return customer.ErrSelfOperation
	}

	targetUser, err := u.repo.FindByID(ctx, targetID)
	if err != nil {
		return customer.ErrUserNotFound
	}

	// Owner account cannot be deactivated
	if strings.EqualFold(targetUser.Role, "owner") {
		return customer.ErrCannotModifyOwner
	}

	// Admin account can only be toggled by owner
	if strings.EqualFold(targetUser.Role, "admin") && !isOwnerRole(actorRoles) {
		return customer.ErrCannotModifyAdmin
	}

	return u.repo.UpdateStatus(ctx, targetID, isActive)
}
