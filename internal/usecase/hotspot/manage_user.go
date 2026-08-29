package hotspot

import (
	"context"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// GetUsers lists hotspot users, optionally filtered (profile, comment batch
// tag, exact name, only-unused). Pass zero-value port.ListUsersFilter to
// list all users.
func (u *UseCase) GetUsers(ctx context.Context, driver port.DeviceDriver, f port.ListUsersFilter) ([]port.HotspotUser, error) {
	return u.gateway.ListUsers(ctx, driver, f)
}

// GetUser fetches a single hotspot user by RouterOS .id.
func (u *UseCase) GetUser(ctx context.Context, driver port.DeviceDriver, rosID string) (port.HotspotUser, error) {
	return u.gateway.GetUser(ctx, driver, rosID)
}

// AddUser creates a new hotspot user directly (non-batch).
func (u *UseCase) AddUser(ctx context.Context, driver port.DeviceDriver, p port.HotspotUserParams) (command.Result, error) {
	res, err := u.gateway.AddUser(ctx, driver, p)
	if err != nil {
		return res, err
	}
	if p.Comment != "" {
		_, _ = u.gateway.ParseUserComment(p.Comment) // validasi komentar best-effort, hasil tidak dipakai
	}
	return res, nil
}

// UpdateUser modifies an existing hotspot user by RouterOS .id.
func (u *UseCase) UpdateUser(ctx context.Context, driver port.DeviceDriver, rosID string, p port.HotspotUserParams) (command.Result, error) {
	return u.gateway.UpdateUser(ctx, driver, rosID, p)
}

// RemoveUser deletes a hotspot user by RouterOS .id.
func (u *UseCase) RemoveUser(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.RemoveUser(ctx, driver, rosID)
}

// ResetUserCounters resets byte/time counters for a hotspot user by RouterOS .id.
func (u *UseCase) ResetUserCounters(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.ResetUserCounters(ctx, driver, rosID)
}

// GetUsersByTag fetches all hotspot users whose comment contains tag.
func (u *UseCase) GetUsersByTag(ctx context.Context, driver port.DeviceDriver, tag string) ([]port.HotspotUser, error) {
	users, err := u.GetUsers(ctx, driver, port.ListUsersFilter{})
	if err != nil {
		return nil, err
	}

	filtered := make([]port.HotspotUser, 0)
	for _, usr := range users {
		if usr.Comment == "" {
			continue
		}
		parsed, parseErr := u.gateway.ParseUserComment(usr.Comment)
		if parseErr != nil {
			continue
		}
		if tag == "" || strings.EqualFold(parsed.Tag, tag) {
			filtered = append(filtered, usr)
		}
	}
	return filtered, nil
}

// DeleteUsersByFilter deletes hotspot users matching mode (by_profile, by_comment, expired).
func (u *UseCase) DeleteUsersByFilter(ctx context.Context, driver port.DeviceDriver, mode, value string) (int, error) {
	return u.gateway.DeleteUsersByFilter(ctx, driver, mode, value)
}
