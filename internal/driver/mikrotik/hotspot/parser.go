package hotspot

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseActiveSessions converts command.Result rows from /ip/hotspot/active/print into typed HotspotActiveSession values.
func ParseActiveSessions(result command.Result) []HotspotActiveSession {
	sessions := make([]HotspotActiveSession, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		user := row["user"]
		if id == "" || user == "" {
			continue
		}
		sessions = append(sessions, HotspotActiveSession{
			RosID:      id,
			Server:     row["server"],
			User:       user,
			Address:    row["address"],
			MACAddress: row["mac-address"],
			LoginBy:    row["login-by"],
		})
	}
	return sessions
}

// ParseActiveStats converts command.Result rows from /ip/hotspot/active/print stats into typed HotspotActiveStat values.
func ParseActiveStats(result command.Result) []HotspotActiveStat {
	stats := make([]HotspotActiveStat, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		stats = append(stats, HotspotActiveStat{
			RosID:           id,
			Uptime:          row["uptime"],
			SessionTimeLeft: row["session-time-left"],
			IdleTime:        row["idle-time"],
			BytesIn:         row["bytes-in"],
			BytesOut:        row["bytes-out"],
			PacketsIn:       row["packets-in"],
			PacketsOut:      row["packets-out"],
		})
	}
	return stats
}

// ParseServers converts command.Result rows from /ip/hotspot/print into typed HotspotServer values.
func ParseServers(result command.Result) []HotspotServer {
	servers := make([]HotspotServer, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		servers = append(servers, HotspotServer{
			RosID:       id,
			Name:        name,
			Interface:   row["interface"],
			Profile:     row["profile"],
			AddressPool: row["address"],
			Disabled:    strings.EqualFold(row["disabled"], "true"),
		})
	}
	return servers
}

// ParseUsers converts command.Result rows from /ip/hotspot/user/print into typed HotspotUser values.
func ParseUsers(result command.Result) []HotspotUser {
	users := make([]HotspotUser, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		users = append(users, HotspotUser{
			RosID:         id,
			Server:        row["server"],
			Name:          name,
			Profile:       row["profile"],
			Password:      row["password"],
			MACAddress:    row["mac-address"],
			Address:       row["address"],
			Comment:       row["comment"],
			Disabled:      strings.EqualFold(row["disabled"], "true"),
			BytesIn:       row["bytes-in"],
			BytesOut:      row["bytes-out"],
			LimitUptime:   row["limit-uptime"],
			LimitBytesIn:  row["limit-bytes-in"],
			LimitBytesOut: row["limit-bytes-out"],
			Uptime:        row["uptime"],
		})
	}
	return users
}

// ParseUserProfiles converts command.Result rows from /ip/hotspot/user/profile/print into typed HotspotUserProfile values.
func ParseUserProfiles(result command.Result) []HotspotUserProfile {
	profiles := make([]HotspotUserProfile, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		profiles = append(profiles, HotspotUserProfile{
			RosID:          id,
			Name:           name,
			SharedUsers:    row["shared-users"],
			RateLimit:      row["rate-limit"],
			SessionTimeout: row["session-timeout"],
			IdleTimeout:    row["idle-timeout"],
			ParentQueue:    row["parent-queue"],
			AddressPool:    row["address-pool"],
			Comment:        row["comment"],
			OnLogin:        row["on-login"],
		})
	}
	return profiles
}

// FilterInactiveUsers compares all registered Hotspot users against active sessions and returns inactive users.
func FilterInactiveUsers(users []HotspotUser, active []HotspotActiveSession) []HotspotUser {
	activeMap := make(map[string]bool, len(active))
	for _, s := range active {
		activeMap[s.User] = true
	}
	inactive := make([]HotspotUser, 0)
	for _, u := range users {
		if !activeMap[u.Name] {
			inactive = append(inactive, u)
		}
	}
	return inactive
}

