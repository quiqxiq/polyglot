package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// HotspotUserProfileParams holds the parameters for creating or updating a
// RouterOS hotspot user profile (/ip/hotspot/user/profile/add or /set).
//
// Field notes (from RouterOS /ip/hotspot/user/profile reference):
//   - Name           : unique profile name.
//   - RateLimit      : bandwidth limit in RouterOS format ("rx/tx", e.g. "5M/5M").
//     Send empty string to clear (unlimited).
//   - SessionTimeout : max single-session duration (RouterOS duration, e.g. "1h").
//     Send empty to clear.
//   - IdleTimeout    : disconnect after idle for this long. Send empty to clear.
//   - SharedUsers    : max simultaneous logins with this profile.
//   - ParentQueue    : parent queue name for shaping hierarchy.
//   - AddressPool    : IP pool name for clients — this is the correct field
//     for /ip/hotspot/user/profile (unlike /ppp/profile which
//     uses remote-address). See MIKROTIK-COMMAND.md §6 note.
//   - Comment        : free-text label.
//
// ⚠️ Fields local-address, remote-address, dns-server, address-list are NOT
// valid properties of /ip/hotspot/user/profile (they belong to /ppp/profile).
// They are deliberately excluded here to avoid sending invalid attributes.
type HotspotUserProfileParams struct {
	Name           string
	RateLimit      string
	SessionTimeout string
	IdleTimeout    string
	SharedUsers    string
	ParentQueue    string
	AddressPool    string
	Comment        string
	OnLogin        string
}

// HotspotUserProfile is the vendor-neutral hotspot user profile row.
// Canonical definition lives in internal/port (see port.HotspotUserProfile docs).
type HotspotUserProfile = port.HotspotUserProfile

// NewPrintHotspotUserProfilesCommand builds the command.Command for
// /ip/hotspot/user/profile/print. Pass a non-empty nameFilter to look up
// one profile by name.
func NewPrintHotspotUserProfilesCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/user/profile/print",
		Args: args,
	}
}

// NewAddHotspotUserProfileCommand builds the command.Command for
// /ip/hotspot/user/profile/add.
func NewAddHotspotUserProfileCommand(p HotspotUserProfileParams) command.Command {
	args := map[string]string{"name": p.Name}
	setIfNonEmpty(args, "rate-limit", p.RateLimit)
	setIfNonEmpty(args, "session-timeout", p.SessionTimeout)
	setIfNonEmpty(args, "idle-timeout", p.IdleTimeout)
	setIfNonEmpty(args, "shared-users", p.SharedUsers)
	setIfNonEmpty(args, "parent-queue", p.ParentQueue)
	setIfNonEmpty(args, "address-pool", p.AddressPool)
	setIfNonEmpty(args, "comment", p.Comment)
	setIfNonEmpty(args, "on-login", p.OnLogin)
	return command.Command{Raw: "/ip/hotspot/user/profile/add", Args: args}
}

// NewSetHotspotUserProfileCommand builds the command.Command for
// /ip/hotspot/user/profile/set. rosID must come from a prior print result.
// Non-empty string fields are updated; empty fields in params are not sent
// (RouterOS preserves the existing value). To explicitly clear a field
// (e.g. remove rate-limit), include the empty string — the router accepts
// an empty value to reset to no-limit.
func NewSetHotspotUserProfileCommand(rosID string, p HotspotUserProfileParams) command.Command {
	args := map[string]string{".id": rosID}
	setIfNonEmpty(args, "name", p.Name)
	setIfNonEmpty(args, "rate-limit", p.RateLimit)
	setIfNonEmpty(args, "session-timeout", p.SessionTimeout)
	setIfNonEmpty(args, "idle-timeout", p.IdleTimeout)
	setIfNonEmpty(args, "shared-users", p.SharedUsers)
	setIfNonEmpty(args, "parent-queue", p.ParentQueue)
	setIfNonEmpty(args, "address-pool", p.AddressPool)
	setIfNonEmpty(args, "comment", p.Comment)
	setIfNonEmpty(args, "on-login", p.OnLogin)
	return command.Command{Raw: "/ip/hotspot/user/profile/set", Args: args}
}

// NewRemoveHotspotUserProfileCommand builds the command.Command for
// /ip/hotspot/user/profile/remove.
func NewRemoveHotspotUserProfileCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/hotspot/user/profile/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParseHotspotUserProfiles converts command.Result rows from
// /ip/hotspot/user/profile/print into typed HotspotUserProfile values.
// Rows missing ".id" or "name" are skipped.
func ParseHotspotUserProfiles(result command.Result) []HotspotUserProfile {
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
			RateLimit:      row["rate-limit"],
			SessionTimeout: row["session-timeout"],
			IdleTimeout:    row["idle-timeout"],
			SharedUsers:    row["shared-users"],
			ParentQueue:    row["parent-queue"],
			AddressPool:    row["address-pool"],
			Comment:        row["comment"],
			OnLogin:        row["on-login"],
		})
	}
	return profiles
}

// FindHotspotUserProfileRosID looks up the RouterOS internal ID for a hotspot
// user profile by name. Returns ("", ErrProfileNotFound) when not found.
func FindHotspotUserProfileRosID(result command.Result, name string) (string, error) {
	for _, p := range ParseHotspotUserProfiles(result) {
		if strings.EqualFold(p.Name, name) {
			return p.RosID, nil
		}
	}
	return "", ErrProfileNotFound
}
