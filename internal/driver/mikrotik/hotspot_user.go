package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// HotspotUserParams holds the parameters for creating or updating a RouterOS
// hotspot user (/ip/hotspot/user/add or /ip/hotspot/user/set).
//
// Field notes (from RouterOS /ip/hotspot/user reference):
//   - Name           : login username.
//   - Password       : login password.
//   - Profile        : name of an existing /ip/hotspot/user/profile entry.
//   - Server         : hotspot server name (e.g. "hotspot1"). Leave empty
//                      to apply to all servers ("all"). For voucher systems
//                      this is typically empty.
//   - MACAddress     : bind this user to a specific MAC address.
//   - Address        : assign a static IP to this user. Leave empty for
//                      dynamic assignment from the profile's pool.
//   - LimitUptime    : max cumulative online time (RouterOS duration string,
//                      e.g. "30d", "8h"). Empty = unlimited.
//   - LimitBytesIn   : max incoming bytes (numeric string). Empty = unlimited.
//   - LimitBytesOut  : max outgoing bytes (numeric string). Empty = unlimited.
//   - Comment        : free-text label. Convention: prefix "voucher" tag here
//                      for voucher-type users so the app can filter them.
//   - Disabled       : when true the user exists but cannot log in.
type HotspotUserParams struct {
	Name          string
	Password      string
	Profile       string
	Server        string
	MACAddress    string
	Address       string
	LimitUptime   string
	LimitBytesIn  string
	LimitBytesOut string
	Comment       string
	Disabled      bool
}

// HotspotUser represents one row returned by /ip/hotspot/user/print.
//
// Field notes:
//   - RosID  : internal RouterOS ID — required for set (via numbers=) and remove.
//   - Server : "all" when not server-specific.
type HotspotUser struct {
	RosID         string
	Name          string
	Password      string
	Profile       string
	Server        string
	MACAddress    string
	Address       string
	LimitUptime   string
	LimitBytesIn  string
	LimitBytesOut string
	Comment       string
	Disabled      bool
}

// NewPrintHotspotUsersCommand builds the command.Command for
// /ip/hotspot/user/print. Pass a non-empty nameFilter to look up one user.
func NewPrintHotspotUsersCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/user/print",
		Args: args,
	}
}

// NewStreamHotspotUsersCommand builds the command.Command for
// /ip/hotspot/user/print follow, streaming Hotspot user directory updates.
func NewStreamHotspotUsersCommand(profileFilter string) command.Command {
	args := map[string]string{
		"follow": "",
	}
	if profileFilter != "" {
		args["?profile"] = profileFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/user/print",
		Args: args,
	}
}

// NewAddHotspotUserCommand builds the command.Command for /ip/hotspot/user/add.
func NewAddHotspotUserCommand(p HotspotUserParams) command.Command {
	args := map[string]string{
		"name":     p.Name,
		"password": p.Password,
		"profile":  p.Profile,
	}
	setIfNonEmpty(args, "server", p.Server)
	setIfNonEmpty(args, "mac-address", p.MACAddress)
	setIfNonEmpty(args, "address", p.Address)
	setIfNonEmpty(args, "limit-uptime", p.LimitUptime)
	setIfNonEmpty(args, "limit-bytes-in", p.LimitBytesIn)
	setIfNonEmpty(args, "limit-bytes-out", p.LimitBytesOut)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/ip/hotspot/user/add", Args: args}
}

// NewSetHotspotUserCommand builds the command.Command for /ip/hotspot/user/set.
//
// ⚠️ RouterOS inconsistency: /ip/hotspot/user/set uses "numbers" (not ".id")
// to target the entry — this is unlike almost every other set endpoint.
// This is documented in MIKROTIK-COMMAND.md §4 and handled here so callers
// never need to remember the exception.
func NewSetHotspotUserCommand(rosID string, p HotspotUserParams) command.Command {
	args := map[string]string{
		"numbers": rosID, // intentional: hotspot/user/set uses numbers, not .id
	}
	setIfNonEmpty(args, "password", p.Password)
	setIfNonEmpty(args, "profile", p.Profile)
	setIfNonEmpty(args, "mac-address", p.MACAddress)
	setIfNonEmpty(args, "address", p.Address)
	setIfNonEmpty(args, "limit-uptime", p.LimitUptime)
	setIfNonEmpty(args, "limit-bytes-in", p.LimitBytesIn)
	setIfNonEmpty(args, "limit-bytes-out", p.LimitBytesOut)
	setIfNonEmpty(args, "comment", p.Comment)
	return command.Command{Raw: "/ip/hotspot/user/set", Args: args}
}

// NewRemoveHotspotUserCommand builds the command.Command for
// /ip/hotspot/user/remove. Note: unlike /set, remove uses ".id" (not "numbers").
func NewRemoveHotspotUserCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/hotspot/user/remove",
		Args: map[string]string{".id": rosID},
	}
}

// NewResetHotspotUserCountersCommand builds the command.Command for
// /ip/hotspot/user/reset-counters. rosID must come from a prior print result.
func NewResetHotspotUserCountersCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/hotspot/user/reset-counters",
		Args: map[string]string{".id": rosID},
	}
}


// ParseHotspotUsers converts command.Result rows from /ip/hotspot/user/print
// into typed HotspotUser values. Rows missing ".id" or "name" are skipped.
func ParseHotspotUsers(result command.Result) []HotspotUser {
	users := make([]HotspotUser, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		users = append(users, HotspotUser{
			RosID:         id,
			Name:          name,
			Password:      row["password"],
			Profile:       row["profile"],
			Server:        row["server"],
			MACAddress:    row["mac-address"],
			Address:       row["address"],
			LimitUptime:   row["limit-uptime"],
			LimitBytesIn:  row["limit-bytes-in"],
			LimitBytesOut: row["limit-bytes-out"],
			Comment:       row["comment"],
			Disabled:      strings.EqualFold(row["disabled"], "true"),
		})
	}
	return users
}
