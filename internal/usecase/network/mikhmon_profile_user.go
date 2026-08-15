package network

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/mikhmon"
	"github.com/quixiq/polyglot/internal/port"
)

// CreateProfile builds and executes the /ip/hotspot/user/profile/add command.
func (u *MikhmonUseCase) CreateProfile(ctx context.Context, driver port.DeviceDriver, p mikhmon.MikhmonProfileParams) (command.Result, error) {
	cmd := mikhmon.NewAddMikhmonProfileCommand(p)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetProfiles fetches all Hotspot User Profiles.
func (u *MikhmonUseCase) GetProfiles(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.HotspotUserProfile, error) {
	cmd := mikrotik.NewPrintHotspotUserProfilesCommand("")
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotUserProfiles(res), nil
}

// UpdateProfile updates an existing Hotspot User Profile by RouterOS .id.
func (u *MikhmonUseCase) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p mikhmon.MikhmonProfileParams) (command.Result, error) {
	cmd := mikhmon.NewSetMikhmonProfileCommand(rosID, p)
	return ExecuteCommand(ctx, driver, cmd)
}

// DeleteProfile removes a Hotspot User Profile by RouterOS .id.
func (u *MikhmonUseCase) DeleteProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewRemoveHotspotUserProfileCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetUsers lists hotspot users, optionally filtered by profile.
func (u *MikhmonUseCase) GetUsers(ctx context.Context, driver port.DeviceDriver, profileFilter string) ([]mikrotik.HotspotUser, error) {
	cmd := mikrotik.NewPrintHotspotUsersCommand(profileFilter)
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotUsers(res), nil
}

// GetUser fetches a single hotspot user by RouterOS .id.
func (u *MikhmonUseCase) GetUser(ctx context.Context, driver port.DeviceDriver, rosID string) (mikrotik.HotspotUser, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/user/print",
		Args: map[string]string{"?.id": rosID},
	}
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return mikrotik.HotspotUser{}, err
	}
	users := mikrotik.ParseHotspotUsers(res)
	if len(users) == 0 {
		return mikrotik.HotspotUser{}, fmt.Errorf("user %q not found", rosID)
	}
	return users[0], nil
}

// AddUser creates a new hotspot user (non-voucher type "up") with a pre-login Mikhmon comment.
func (u *MikhmonUseCase) AddUser(ctx context.Context, driver port.DeviceDriver, p mikrotik.HotspotUserParams) (command.Result, error) {
	if p.Comment != "" && !mikhmon.IsMikhmonComment(p.Comment) {
		code := mikhmon.GenerateVoucherCode(3, mikhmon.CharSetUpperNum)
		p.Comment = mikhmon.FormatPreLoginComment("up", code, p.Comment, time.Now())
	}
	cmd := mikrotik.NewAddHotspotUserCommand(p)
	return ExecuteCommand(ctx, driver, cmd)
}

// UpdateUser updates an existing hotspot user by RouterOS .id.
func (u *MikhmonUseCase) UpdateUser(ctx context.Context, driver port.DeviceDriver, rosID string, p mikrotik.HotspotUserParams) (command.Result, error) {
	cmd := mikrotik.NewSetHotspotUserCommand(rosID, p)
	return ExecuteCommand(ctx, driver, cmd)
}

// RemoveUser deletes a hotspot user by RouterOS .id.
func (u *MikhmonUseCase) RemoveUser(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewRemoveHotspotUserCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}

// ResetUserCounters resets byte/time counters for a hotspot user.
func (u *MikhmonUseCase) ResetUserCounters(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewResetHotspotUserCountersCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetUsersByTag retrieves hotspot users whose comment contains the given tag prefix.
func (u *MikhmonUseCase) GetUsersByTag(ctx context.Context, driver port.DeviceDriver, tag string) ([]mikrotik.HotspotUser, error) {
	all, err := u.GetUsers(ctx, driver, "")
	if err != nil {
		return nil, err
	}
	if tag == "" {
		return all, nil
	}
	filtered := make([]mikrotik.HotspotUser, 0)
	for _, user := range all {
		parsed, parseErr := mikhmon.ParseMikhmonComment(user.Comment)
		if parseErr != nil {
			continue
		}
		if strings.Contains(parsed.Tag, tag) {
			filtered = append(filtered, user)
		}
	}
	return filtered, nil
}

// GetActiveSessions fetches all active hotspot sessions.
func (u *MikhmonUseCase) GetActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.HotspotActiveSession, error) {
	cmd := mikrotik.NewPrintHotspotActiveCommand("")
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotActiveSessions(res), nil
}

// RemoveActiveSession kicks an active hotspot session.
func (u *MikhmonUseCase) RemoveActiveSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewDisconnectHotspotActiveCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}
