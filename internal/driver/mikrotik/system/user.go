package system

import (
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// UserPolicies is the exhaustive list of 17 RouterOS user policies.
var UserPolicies = []string{
	"local", "telnet", "ssh", "ftp", "reboot", "read", "write", "policy", "test",
	"winbox", "password", "web", "sniff", "sensitive", "api", "romon", "rest-api",
}

// RosUser represents one RouterOS API or administrative user.
type RosUser struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Group        string `json:"group"`
	Address      string `json:"address"`
	Disabled     bool   `json:"disabled"`
	Comment      string `json:"comment"`
	LastLoggedIn string `json:"last_logged_in"`
}

// RosGroup represents a RouterOS user group with policy permissions.
type RosGroup struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Policy  []string `json:"policy"`
	Skin    string   `json:"skin"`
	Comment string   `json:"comment"`
}

// RosActiveSession represents an ongoing login session on the router.
type RosActiveSession struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	When    string `json:"when"`
	Address string `json:"address"`
	Via     string `json:"via"`
	Group   string `json:"group"`
}

// UserSettings represents global user security settings.
type UserSettings struct {
	MinPasswordLength int `json:"min_password_length"`
}

// NewPrintUsersCommand builds a query for /user/print with strict proplist.
func NewPrintUsersCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "group", "address", "disabled", "comment", "last-logged-in")
	return command.Command{
		Raw:  "/user/print",
		Args: args,
	}
}

// NewAddUserCommand builds a command to create a new RouterOS user.
func NewAddUserCommand(user RosUser, password string) command.Command {
	args := map[string]string{
		"name":  user.Name,
		"group": user.Group,
	}
	if password != "" {
		args["password"] = password
	}
	if user.Address != "" {
		args["address"] = user.Address
	}
	if user.Comment != "" {
		args["comment"] = user.Comment
	}
	if user.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/user/add",
		Args: args,
	}
}

// NewSetUserCommand builds a command to update an existing RouterOS user.
func NewSetUserCommand(user RosUser, password string) command.Command {
	args := map[string]string{
		".id": user.ID,
	}
	if user.Name != "" {
		args["name"] = user.Name
	}
	if user.Group != "" {
		args["group"] = user.Group
	}
	if password != "" {
		args["password"] = password
	}
	args["address"] = user.Address
	args["comment"] = user.Comment
	if user.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/user/set",
		Args: args,
	}
}

// NewRemoveUserCommand builds a command to delete a user by ID.
func NewRemoveUserCommand(id string) command.Command {
	return command.Command{
		Raw: "/user/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintGroupsCommand builds a query for /user/group/print with strict proplist.
func NewPrintGroupsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "policy", "skin", "comment")
	return command.Command{
		Raw:  "/user/group/print",
		Args: args,
	}
}

// NewAddGroupCommand builds a command to create a user group.
func NewAddGroupCommand(group RosGroup) command.Command {
	args := map[string]string{
		"name": group.Name,
	}
	if len(group.Policy) > 0 {
		args["policy"] = strings.Join(group.Policy, ",")
	}
	if group.Comment != "" {
		args["comment"] = group.Comment
	}
	return command.Command{
		Raw:  "/user/group/add",
		Args: args,
	}
}

// NewSetGroupCommand builds a command to update a user group.
// RouterOS /user/group/set requires explicit !negations for unselected policies.
func NewSetGroupCommand(group RosGroup) command.Command {
	selected := make(map[string]bool, len(group.Policy))
	for _, p := range group.Policy {
		selected[p] = true
	}

	var policies []string
	for _, p := range UserPolicies {
		if selected[p] {
			policies = append(policies, p)
		} else {
			policies = append(policies, "!"+p)
		}
	}

	args := map[string]string{
		".id":    group.ID,
		"name":   group.Name,
		"policy": strings.Join(policies, ","),
	}
	args["comment"] = group.Comment
	return command.Command{
		Raw:  "/user/group/set",
		Args: args,
	}
}

// NewRemoveGroupCommand builds a command to delete a group by ID.
func NewRemoveGroupCommand(id string) command.Command {
	return command.Command{
		Raw: "/user/group/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintActiveUsersCommand builds a query for active logged-in sessions.
func NewPrintActiveUsersCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "when", "address", "via", "group")
	return command.Command{
		Raw:  "/user/active/print",
		Args: args,
	}
}

// NewStreamActiveUsersCommand builds an event streaming command for active sessions (Mode A).
func NewStreamActiveUsersCommand() command.Command {
	args := map[string]string{
		"follow": "",
	}
	rosutil.SetProplist(args, ".id", "name", "when", "address", "via", "group")
	return command.Command{
		Raw:  "/user/active/print",
		Args: args,
	}
}

// NewRemoveActiveUserCommand builds a command to terminate an active session.
func NewRemoveActiveUserCommand(id string) command.Command {
	return command.Command{
		Raw: "/user/active/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintUserSettingsCommand builds a query for user security settings.
func NewPrintUserSettingsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "minimum-password-length")
	return command.Command{
		Raw:  "/user/settings/print",
		Args: args,
	}
}

// ParseUsers converts RouterOS reply rows into RosUser slice.
func ParseUsers(rows []map[string]string) []RosUser {
	out := make([]RosUser, 0, len(rows))
	for _, r := range rows {
		out = append(out, RosUser{
			ID:           r[".id"],
			Name:         r["name"],
			Group:        r["group"],
			Address:      r["address"],
			Disabled:     r["disabled"] == "true",
			Comment:      r["comment"],
			LastLoggedIn: r["last-logged-in"],
		})
	}
	return out
}

// ParseGroups converts reply rows into RosGroup slice.
func ParseGroups(rows []map[string]string) []RosGroup {
	out := make([]RosGroup, 0, len(rows))
	for _, r := range rows {
		rawPol := r["policy"]
		var policies []string
		if rawPol != "" {
			for _, p := range strings.Split(rawPol, ",") {
				if p = strings.TrimSpace(p); p != "" {
					policies = append(policies, p)
				}
			}
		}
		out = append(out, RosGroup{
			ID:      r[".id"],
			Name:    r["name"],
			Policy:  policies,
			Skin:    r["skin"],
			Comment: r["comment"],
		})
	}
	return out
}

// ParseActiveSessions converts reply rows into RosActiveSession slice.
func ParseActiveSessions(rows []map[string]string) []RosActiveSession {
	out := make([]RosActiveSession, 0, len(rows))
	for _, r := range rows {
		out = append(out, RosActiveSession{
			ID:      r[".id"],
			Name:    r["name"],
			When:    r["when"],
			Address: r["address"],
			Via:     r["via"],
			Group:   r["group"],
		})
	}
	return out
}

// ParseUserSettings converts reply rows into UserSettings.
func ParseUserSettings(rows []map[string]string) *UserSettings {
	if len(rows) == 0 {
		return nil
	}
	minLen, _ := strconv.Atoi(rows[0]["minimum-password-length"])
	return &UserSettings{
		MinPasswordLength: minLen,
	}
}
