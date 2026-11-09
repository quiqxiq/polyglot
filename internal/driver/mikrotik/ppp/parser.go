package ppp

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	domainPPP "github.com/quixiq/polyglot/internal/domain/ppp"
)

// ParseSecrets converts command.Result rows from /ppp/secret/print into typed PPPoESecret values.
func ParseSecrets(result command.Result) []domainPPP.PPPoESecret {
	secrets := make([]domainPPP.PPPoESecret, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		callerID := row["caller-id"]
		if callerID == "" {
			callerID = row["last-caller-id"]
		}
		secrets = append(secrets, domainPPP.PPPoESecret{
			RosID:         id,
			Name:          name,
			Profile:       row["profile"],
			Service:       row["service"],
			LocalAddress:  row["local-address"],
			RemoteAddress: row["remote-address"],
			Comment:       row["comment"],
			Disabled:      strings.EqualFold(row["disabled"], "true"),
			LastLoggedOut: row["last-logged-out"],
			CallerID:      callerID,
		})
	}
	return secrets
}

// ParseActiveSessions converts command.Result rows from /ppp/active/print into typed ActiveSession values.
func ParseActiveSessions(result command.Result) []domainPPP.ActiveSession {
	sessions := make([]domainPPP.ActiveSession, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		sessions = append(sessions, domainPPP.ActiveSession{
			RosID:     id,
			Name:      name,
			Service:   row["service"],
			CallerID:  row["caller-id"],
			Address:   row["address"],
			Encoding:  row["encoding"],
			SessionID: row["session-id"],
			Radius:    strings.EqualFold(row["radius"], "true"),
			Profile:   row["profile"],
			Uptime:    row["uptime"],
		})
	}
	return sessions
}

// ParseProfiles converts command.Result rows from /ppp/profile/print into typed Profile values.
func ParseProfiles(result command.Result) []domainPPP.Profile {
	profiles := make([]domainPPP.Profile, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		profiles = append(profiles, domainPPP.Profile{
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
			OnUp:           row["on-up"],
			OnDown:         row["on-down"],
		})
	}
	return profiles
}

// FilterInactiveSecrets returns the secrets of subscribers who are currently offline.
func FilterInactiveSecrets(secrets []domainPPP.PPPoESecret, active []domainPPP.ActiveSession) []domainPPP.PPPoESecret {
	return domainPPP.FilterInactiveSecrets(secrets, active)
}

// EnrichActiveSessionsWithProfiles enriches PPP active sessions with profile from /ppp/secret.
func EnrichActiveSessionsWithProfiles(active []domainPPP.ActiveSession, secrets []domainPPP.PPPoESecret) []domainPPP.ActiveSession {
	return domainPPP.EnrichActiveSessionsWithProfiles(active, secrets)
}
