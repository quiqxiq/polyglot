package auth

import (
	"fmt"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

// DomainUserToPb converts domain customer.User entity to proto User message.
func DomainUserToPb(u *customer.User, roles []string) *polyglotv1.User {
	if u == nil {
		return nil
	}

	primaryRole := u.Role
	if len(roles) > 0 {
		primaryRole = roles[0]
	}

	return &polyglotv1.User{
		Id:                uint64(u.ID),
		Username:          u.Username,
		Email:             u.Email,
		Role:              primaryRole,
		Roles:             roles,
		IsActive:          u.IsActive,
		FullName:          u.FullName,
		PhoneNumber:       u.PhoneNumber,
		Specialization:    u.Specialization,
		AssignedDeviceIds: u.AssignedDeviceIDs,
		CreatedAtUnix:     u.CreatedAt.Unix(),
		UpdatedAtUnix:     u.UpdatedAt.Unix(),
	}
}

// DomainUserToProfile converts domain customer.User entity to proto UserProfile message.
func DomainUserToProfile(u *customer.User, roles []string, perms []string) *polyglotv1.UserProfile {
	if u == nil {
		return nil
	}

	primaryRole := u.Role
	if len(roles) > 0 {
		primaryRole = roles[0]
	}

	return &polyglotv1.UserProfile{
		Id:             fmt.Sprintf("%d", u.ID),
		Username:       u.Username,
		Email:          u.Email,
		Role:           primaryRole,
		Roles:          roles,
		Permissions:    perms,
		FullName:       u.FullName,
		PhoneNumber:    u.PhoneNumber,
		Specialization: u.Specialization,
	}
}
