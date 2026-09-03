package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// GetRoles retrieves the assigned RBAC roles for a user.
func (u *ManageUserUseCase) GetRoles(ctx context.Context, id uint) ([]string, error) {
	if u.roles == nil {
		return nil, nil
	}
	roles, err := u.roles.GetRolesForUser(fmt.Sprintf("%d", id))
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}
	return roles, nil
}

// GetPermissions retrieves the implicit permissions for a user based on their roles.
func (u *ManageUserUseCase) GetPermissions(ctx context.Context, id uint) ([]string, error) {
	if u.roles == nil {
		return nil, nil
	}
	perms, err := u.roles.GetImplicitPermissionsForUser(fmt.Sprintf("%d", id))
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for user: %w", err)
	}
	return perms, nil
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
		return 0, fmt.Errorf("failed to list users to count active owners: %w", err)
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
			return fmt.Errorf("failed to check device accessibility: %w", err)
		}
		if !accessible {
			return fmt.Errorf("%w: device %s", customer.ErrUnauthorizedDeviceAssign, devID)
		}
	}
	return nil
}

// AssignDevicesToUser assigns a list of device IDs to a user if permissions permit.
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
		return nil, fmt.Errorf("failed to assign devices: %w", err)
	}

	assigned, err := u.repo.GetAssignedDeviceIDs(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assigned device IDs: %w", err)
	}
	return assigned, nil
}

// ListUserAccessibleDevices returns device IDs accessible to the requested user.
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
	assigned, err := u.repo.GetAssignedDeviceIDs(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assigned device IDs: %w", err)
	}
	return assigned, nil
}
