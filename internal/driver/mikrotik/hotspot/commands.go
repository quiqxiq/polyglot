package hotspot

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

var setIfNonEmpty = rosutil.SetIfNonEmpty

// NewPrintActiveCommand builds the command.Command for /ip/hotspot/active/print.
func NewPrintActiveCommand(userFilter string) command.Command {
	args := map[string]string{}
	if userFilter != "" {
		args["?user"] = userFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/active/print",
		Args: args,
	}
}

// NewStreamActiveCommand builds the command.Command for /ip/hotspot/active/print follow.
func NewStreamActiveCommand(userFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,server,user,address,mac-address,login-by",
	}
	if userFilter != "" {
		args["?user"] = userFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/active/print",
		Args: args,
	}
}

// NewStreamActiveStatsCommand builds the command.Command for /ip/hotspot/active/print stats.
func NewStreamActiveStatsCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw: "/ip/hotspot/active/print",
		Args: map[string]string{
			"stats":     "",
			"interval":  interval,
			".proplist": ".id,uptime,bytes-in,bytes-out",
		},
	}
}

// NewKickActiveCommand builds the command.Command for /ip/hotspot/active/remove.
func NewKickActiveCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/hotspot/active/remove",
		Args: map[string]string{".id": rosID},
	}
}

// NewPrintServersCommand builds the command.Command for /ip/hotspot/print.
func NewPrintServersCommand() command.Command {
	return command.Command{
		Raw:  "/ip/hotspot/print",
		Args: map[string]string{},
	}
}

// NewPrintUsersCommand builds the command.Command for /ip/hotspot/user/print.
func NewPrintUsersCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/user/print",
		Args: args,
	}
}

// NewStreamUsersCommand builds the command.Command for streaming /ip/hotspot/user/print.
func NewStreamUsersCommand(nameFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,server,name,profile,comment,disabled,uptime,bytes-in,bytes-out,limit-uptime,limit-bytes-total",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/user/print",
		Args: args,
	}
}

// NewAddUserCommand builds the command.Command for /ip/hotspot/user/add.
func NewAddUserCommand(p HotspotUserParams) command.Command {
	args := map[string]string{
		"name": p.Name,
	}
	setIfNonEmpty(args, "server", p.Server)
	setIfNonEmpty(args, "password", p.Password)
	setIfNonEmpty(args, "profile", p.Profile)
	setIfNonEmpty(args, "comment", p.Comment)
	setIfNonEmpty(args, "limit-uptime", p.LimitUptime)
	setIfNonEmpty(args, "limit-bytes-total", p.LimitBytes)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/ip/hotspot/user/add", Args: args}
}

// NewSetUserCommand builds the command.Command for /ip/hotspot/user/set.
func NewSetUserCommand(rosID string, p HotspotUserParams) command.Command {
	args := map[string]string{".id": rosID}
	setIfNonEmpty(args, "name", p.Name)
	setIfNonEmpty(args, "server", p.Server)
	setIfNonEmpty(args, "password", p.Password)
	setIfNonEmpty(args, "profile", p.Profile)
	setIfNonEmpty(args, "comment", p.Comment)
	setIfNonEmpty(args, "limit-uptime", p.LimitUptime)
	setIfNonEmpty(args, "limit-bytes-total", p.LimitBytes)
	return command.Command{Raw: "/ip/hotspot/user/set", Args: args}
}

// NewRemoveUserCommand builds the command.Command for /ip/hotspot/user/remove.
func NewRemoveUserCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/hotspot/user/remove",
		Args: map[string]string{".id": rosID},
	}
}

// NewPrintUserProfilesCommand builds the command.Command for /ip/hotspot/user/profile/print.
func NewPrintUserProfilesCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ip/hotspot/user/profile/print",
		Args: args,
	}
}

// NewAddUserProfileCommand builds the command.Command for /ip/hotspot/user/profile/add.
func NewAddUserProfileCommand(p HotspotProfileParams) command.Command {
	args := map[string]string{
		"name": p.Name,
	}
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

// NewSetUserProfileCommand builds the command.Command for /ip/hotspot/user/profile/set.
func NewSetUserProfileCommand(rosID string, p HotspotProfileParams) command.Command {
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

// NewRemoveUserProfileCommand builds the command.Command for /ip/hotspot/user/profile/remove.
func NewRemoveUserProfileCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/hotspot/user/profile/remove",
		Args: map[string]string{".id": rosID},
	}
}
