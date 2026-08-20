package hotspot

import (
	"context"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/port"
)

// DeleteUsersByFilter deletes hotspot users matching mode ("profile", "comment", "expired").
func (g *Gateway) DeleteUsersByFilter(ctx context.Context, driver port.DeviceDriver, mode, value string) (int, error) {
	var usersToDelete []port.HotspotUser
	switch strings.ToLower(mode) {
	case "profile", "by_profile":
		if value == "" {
			return 0, fmt.Errorf("profile name is required for profile-based deletion")
		}
		users, err := g.ListUsers(ctx, driver, port.ListUsersFilter{Profile: value})
		if err != nil {
			return 0, fmt.Errorf("list users by profile %q: %w", value, err)
		}
		usersToDelete = users

	case "comment", "by_comment":
		if value == "" {
			return 0, fmt.Errorf("comment batch tag is required for comment-based deletion")
		}
		users, err := g.ListUsers(ctx, driver, port.ListUsersFilter{Comment: value})
		if err != nil {
			return 0, fmt.Errorf("list users by comment %q: %w", value, err)
		}
		usersToDelete = users

	case "expired":
		allUsers, err := g.ListUsers(ctx, driver, port.ListUsersFilter{})
		if err != nil {
			return 0, fmt.Errorf("list all users: %w", err)
		}
		for _, u := range allUsers {
			if isUserExpired(u) {
				usersToDelete = append(usersToDelete, u)
			}
		}

	default:
		return 0, fmt.Errorf("unsupported delete mode %q (allowed: profile, comment, expired)", mode)
	}

	if len(usersToDelete) == 0 {
		return 0, nil
	}

	deletedCount := 0
	for _, u := range usersToDelete {
		if u.RosID == "" {
			continue
		}
		if _, err := g.RemoveUser(ctx, driver, u.RosID); err == nil {
			deletedCount++
		}
	}

	return deletedCount, nil
}

func isUserExpired(u port.HotspotUser) bool {
	if u.LimitUptime == "1s" {
		return true
	}
	if strings.Contains(strings.ToLower(u.Comment), "expired") {
		return true
	}
	return false
}
