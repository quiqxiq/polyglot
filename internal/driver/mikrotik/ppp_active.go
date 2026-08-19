package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// PPPActiveSession is the vendor-neutral PPP active session row.
// Canonical definition lives in internal/port (see port.PPPActiveSession docs).
type PPPActiveSession = port.PPPActiveSession

// NewPrintPPPActiveCommand builds the command.Command for /ppp/active/print.
// Pass a non-empty nameFilter to limit results to one specific username;
// pass empty string to list all active sessions.
//
// Typical use: check whether a subscriber is currently online before
// deciding to kick them (e.g. after profile change or suspension).
func NewPrintPPPActiveCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/active/print",
		Args: args,
	}
}

// PPPActiveStat is the vendor-neutral PPP active session telemetry stats row.
type PPPActiveStat = port.PPPActiveStat

// NewStreamPPPActiveCommand builds the command.Command for
// /ppp/active/print follow, which causes RouterOS to push a new row
// every time a PPPoE session changes state (login, logout, IP change).
// Use with Driver.Stream — isStreamingCommand returns true because
// "follow" is present. No polling needed: the driver notifies on change.
//
// Only static session identity fields are requested via .proplist to minimize overhead.
func NewStreamPPPActiveCommand(nameFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,name,service,caller-id,address,encoding,session-id,radius,profile",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/active/print",
		Args: args,
	}
}

// NewStreamPPPActiveStatsCommand builds the command.Command for
// /ppp/active/print stats interval=<interval>.
func NewStreamPPPActiveStatsCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw: "/ppp/active/print",
		Args: map[string]string{
			"stats":    "",
			"interval": interval,
		},
	}
}

// NewKickPPPActiveCommand builds the command.Command for /ppp/active/remove,
// which forcibly disconnects an active PPPoE session identified by its
// RouterOS session ID. The session ID must come from a prior /ppp/active/print
// result (PPPActiveSession.RosID).
//
// Classified as ClassDestructive — disconnecting a subscriber without notice
// is disruptive. Callers should prefer the lookup-then-kick pattern:
//  1. NewPrintPPPActiveCommand(username) → find session
//  2. NewKickPPPActiveCommand(session.RosID) → disconnect
func NewKickPPPActiveCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ppp/active/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParsePPPActiveSessions converts command.Result rows from a
// /ppp/active/print command into typed PPPActiveSession values.
// Rows missing ".id" or "name" are silently skipped.
func ParsePPPActiveSessions(result command.Result) []PPPActiveSession {
	sessions := make([]PPPActiveSession, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		sessions = append(sessions, PPPActiveSession{
			RosID:     id,
			Name:      name,
			Service:   row["service"],
			CallerID:  row["caller-id"],
			Address:   row["address"],
			Encoding:  row["encoding"],
			SessionID: row["session-id"],
			Radius:    strings.EqualFold(row["radius"], "true"),
			Profile:   row["profile"],
		})
	}
	return sessions
}

// ParsePPPActiveStats converts command.Result rows from a
// /ppp/active/print stats command into typed PPPActiveStat values.
func ParsePPPActiveStats(result command.Result) []PPPActiveStat {
	stats := make([]PPPActiveStat, 0, len(result.Rows))
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
			} else if len(parts) == 1 {
				bytesIn = parts[0]
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
			} else if len(parts) == 1 {
				packetsIn = parts[0]
			}
		}

		stats = append(stats, PPPActiveStat{
			RosID:         id,
			Uptime:        row["uptime"],
			LimitBytesIn:  row["limit-bytes-in"],
			LimitBytesOut: row["limit-bytes-out"],
			BytesIn:       bytesIn,
			BytesOut:      bytesOut,
			PacketsIn:     packetsIn,
			PacketsOut:    packetsOut,
		})
	}
	return stats
}

// FilterInactivePPPoESecrets compares all registered PPPoE secrets (from
// /ppp/secret/print) against currently active PPPoE sessions (from
// /ppp/active/print) and returns the secrets of subscribers who are currently
// offline (non-aktif).
func FilterInactivePPPoESecrets(secrets []PPPoESecret, active []PPPActiveSession) []PPPoESecret {
	activeMap := make(map[string]bool, len(active))
	for _, s := range active {
		activeMap[s.Name] = true
	}
	inactive := make([]PPPoESecret, 0)
	for _, s := range secrets {
		if !activeMap[s.Name] {
			inactive = append(inactive, s)
		}
	}
	return inactive
}
