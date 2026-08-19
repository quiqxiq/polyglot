package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// HotspotActiveSession is the vendor-neutral hotspot active session row.
// Canonical definition lives in internal/port (see port.HotspotActiveSession docs).
type HotspotActiveSession = port.HotspotActiveSession

// HotspotServer represents one row returned by /ip/hotspot/print.
// This is a read-only resource — the application only reads hotspot server
// configuration, never creates/modifies/deletes servers.
//
// Field notes:
//   - AddressPool : shown as "address" in the raw RouterOS output — the pool
//     of IPs assigned to hotspot clients.
type HotspotServer struct {
	RosID       string
	Name        string
	Interface   string
	Profile     string
	AddressPool string
	Disabled    bool
}

// NewPrintHotspotActiveCommand builds the command.Command for
// /ip/hotspot/active/print. Pass a non-empty userFilter to limit results
// to sessions belonging to one specific username.
//
// Note: the filter field is "user" (not "name") — specific to active sessions.
func NewPrintHotspotActiveCommand(userFilter string) command.Command {
	args := map[string]string{}
	if userFilter != "" {
		args["?user"] = userFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/active/print",
		Args: args,
	}
}

// HotspotActiveStat is the vendor-neutral hotspot active session telemetry stats row.
type HotspotActiveStat = port.HotspotActiveStat

// NewStreamHotspotActiveCommand builds the command.Command for
// /ip/hotspot/active/print follow, which causes RouterOS to push a new row
// every time a hotspot session changes state (login, logout, expiry).
// Use with Driver.Stream — isStreamingCommand returns true because
// "follow" is present.
//
// Only static session identity fields are requested via .proplist to minimize overhead.
// userFilter: limit to one username (empty = all sessions).
func NewStreamHotspotActiveCommand(userFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,server,user,domain,address,mac-address,login-by",
	}
	if userFilter != "" {
		args["?user"] = userFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/active/print",
		Args: args,
	}
}

// NewStreamHotspotActiveStatsCommand builds the command.Command for
// /ip/hotspot/active/print stats interval=<interval>.
func NewStreamHotspotActiveStatsCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw: "/ip/hotspot/active/print",
		Args: map[string]string{
			"stats":    "",
			"interval": interval,
		},
	}
}

// NewDisconnectHotspotActiveCommand builds the command.Command for
// /ip/hotspot/active/remove, which forcibly disconnects one active hotspot
// session. The session ID must come from a prior /ip/hotspot/active/print
// result (HotspotActiveSession.RosID).
func NewDisconnectHotspotActiveCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/hotspot/active/remove",
		Args: map[string]string{".id": rosID},
	}
}

// NewPrintHotspotServersCommand builds the command.Command for
// /ip/hotspot/print (list configured hotspot servers). No filter — callers
// always get the full list, which is typically short (1–3 servers per router).
func NewPrintHotspotServersCommand() command.Command {
	return command.Command{
		Raw:  "/ip/hotspot/print",
		Args: map[string]string{},
	}
}

// ParseHotspotActiveSessions converts command.Result rows from
// /ip/hotspot/active/print into typed HotspotActiveSession values.
// Rows missing ".id" or "user" are skipped.
func ParseHotspotActiveSessions(result command.Result) []HotspotActiveSession {
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

// ParseHotspotActiveStats converts command.Result rows from
// /ip/hotspot/active/print stats into typed HotspotActiveStat values.
func ParseHotspotActiveStats(result command.Result) []HotspotActiveStat {
	stats := make([]HotspotActiveStat, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}

		bytesIn := row["bytes-in"]
		if bytesIn == "" {
			bytesIn = row["rx-byte"]
		}
		if bytesIn == "" {
			bytesIn = row["rx-bytes"]
		}

		bytesOut := row["bytes-out"]
		if bytesOut == "" {
			bytesOut = row["tx-byte"]
		}
		if bytesOut == "" {
			bytesOut = row["tx-bytes"]
		}

		if bytesIn == "" && bytesOut == "" && row["bytes"] != "" {
			parts := strings.Split(row["bytes"], "/")
			if len(parts) == 2 {
				bytesIn = parts[0]
				bytesOut = parts[1]
			}
		}

		packetsIn := row["packets-in"]
		if packetsIn == "" {
			packetsIn = row["rx-packet"]
		}
		if packetsIn == "" {
			packetsIn = row["rx-packets"]
		}

		packetsOut := row["packets-out"]
		if packetsOut == "" {
			packetsOut = row["tx-packet"]
		}
		if packetsOut == "" {
			packetsOut = row["tx-packets"]
		}

		if packetsIn == "" && packetsOut == "" && row["packets"] != "" {
			parts := strings.Split(row["packets"], "/")
			if len(parts) == 2 {
				packetsIn = parts[0]
				packetsOut = parts[1]
			}
		}

		stats = append(stats, HotspotActiveStat{
			RosID:           id,
			Uptime:          row["uptime"],
			SessionTimeLeft: row["session-time-left"],
			IdleTime:        row["idle-time"],
			BytesIn:         bytesIn,
			BytesOut:        bytesOut,
			PacketsIn:       packetsIn,
			PacketsOut:      packetsOut,
		})
	}
	return stats
}

// ParseHotspotServers converts command.Result rows from /ip/hotspot/print
// into typed HotspotServer values. Rows missing ".id" or "name" are skipped.
func ParseHotspotServers(result command.Result) []HotspotServer {
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
			AddressPool: row["address-pool"],
			Disabled:    strings.EqualFold(row["disabled"], "true"),
		})
	}
	return servers
}

// FilterInactiveHotspotUsers compares all registered Hotspot users (from
// /ip/hotspot/user/print) against currently active Hotspot sessions (from
// /ip/hotspot/active/print) and returns the users who are currently offline
// (non-aktif).
func FilterInactiveHotspotUsers(users []HotspotUser, active []HotspotActiveSession) []HotspotUser {
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
