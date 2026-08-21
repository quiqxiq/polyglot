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

// NewStreamPPPActiveCommand builds the command.Command for /ppp/active/print
// follow, filtering only identity/state fields to minimize RouterOS CPU overhead.
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
// It requests only dynamic counter fields via .proplist to minimize RouterOS CPU.
func NewStreamPPPActiveStatsCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw: "/ppp/active/print",
		Args: map[string]string{
			"stats":     "",
			"interval":  interval,
			".proplist": ".id,uptime,limit-bytes-in,limit-bytes-out",
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
		stats = append(stats, PPPActiveStat{
			RosID:         id,
			Uptime:        row["uptime"],
			LimitBytesIn:  row["limit-bytes-in"],
			LimitBytesOut: row["limit-bytes-out"],
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

// EnrichPPPActiveSessionsWithProfiles enriches PPP active sessions with the subscriber's
// profile from /ppp/secret because RouterOS /ppp/active/print does not include a profile property.
func EnrichPPPActiveSessionsWithProfiles(active []PPPActiveSession, secrets []PPPoESecret) []PPPActiveSession {
	secretMap := make(map[string]string, len(secrets))
	for _, s := range secrets {
		if s.Profile != "" {
			secretMap[s.Name] = s.Profile
		}
	}
	for i := range active {
		if active[i].Profile == "" {
			if prof, ok := secretMap[active[i].Name]; ok {
				active[i].Profile = prof
			}
		}
	}
	return active
}
