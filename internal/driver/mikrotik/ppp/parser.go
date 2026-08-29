package ppp

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseSecrets converts command.Result rows from /ppp/secret/print into typed PPPoESecret values.
func ParseSecrets(result command.Result) []PPPoESecret {
	secrets := make([]PPPoESecret, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		secrets = append(secrets, PPPoESecret{
			RosID:         id,
			Name:          name,
			Profile:       row["profile"],
			Service:       row["service"],
			LocalAddress:  row["local-address"],
			RemoteAddress: row["remote-address"],
			Comment:       row["comment"],
			Disabled:      strings.EqualFold(row["disabled"], "true"),
			LastLoggedOut: row["last-logged-out"],
			CallerID:      row["caller-id"],
		})
	}
	return secrets
}

// ParseActiveSessions converts command.Result rows from /ppp/active/print into typed PPPActiveSession values.
func ParseActiveSessions(result command.Result) []PPPActiveSession {
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

// ParseActiveStats converts command.Result rows from /ppp/active/print stats into typed PPPActiveStat values.
func ParseActiveStats(result command.Result) []PPPActiveStat {
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

// ParseProfiles converts command.Result rows from /ppp/profile/print into typed PPPProfile values.
func ParseProfiles(result command.Result) []PPPProfile {
	profiles := make([]PPPProfile, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		profiles = append(profiles, PPPProfile{
			RosID:          id,
			Name:           name,
			RateLimit:      row["rate-limit"],
			LocalAddress:   row["local-address"],
			RemoteAddress:  row["remote-address"],
			DNSServer:      row["dns-server"],
			ParentQueue:    row["parent-queue"],
			AddressList:    row["address-list"],
			Comment:        row["comment"],
			SharedUsers:    row["shared-users"],
			OnlyOne:        row["only-one"],
			UseMPLS:        row["use-mpls"],
			UseCompression: row["use-compression"],
			UseEncryption:  row["use-encryption"],
			ChangeTCPMSS:   row["change-tcp-mss"],
			BridgeLearning: row["bridge-learning"],
		})
	}
	return profiles
}

// FilterInactiveSecrets returns the secrets of subscribers who are currently offline.
func FilterInactiveSecrets(secrets []PPPoESecret, active []PPPActiveSession) []PPPoESecret {
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

// EnrichActiveSessionsWithProfiles enriches PPP active sessions with profile from /ppp/secret.
func EnrichActiveSessionsWithProfiles(active []PPPActiveSession, secrets []PPPoESecret) []PPPActiveSession {
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
