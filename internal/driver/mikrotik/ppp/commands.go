package ppp

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

var setIfNonEmpty = rosutil.SetIfNonEmpty

// NewAddSecretCommand builds the command.Command for /ppp/secret/add.
func NewAddSecretCommand(p PPPoESecretParams) command.Command {
	profile := p.Profile
	if profile == "" {
		profile = "default"
	}
	service := p.Service
	if service == "" {
		service = "pppoe"
	}

	args := map[string]string{
		"name":     p.Name,
		"password": p.Password,
		"profile":  profile,
		"service":  service,
	}
	setIfNonEmpty(args, "local-address", p.LocalAddress)
	setIfNonEmpty(args, "remote-address", p.RemoteAddress)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}

	return command.Command{
		Raw:  "/ppp/secret/add",
		Args: args,
	}
}

// NewSetSecretCommand builds the command.Command for /ppp/secret/set.
func NewSetSecretCommand(rosID string, p PPPoESecretParams) command.Command {
	args := map[string]string{
		"numbers": rosID,
	}
	setIfNonEmpty(args, "password", p.Password)
	setIfNonEmpty(args, "profile", p.Profile)
	setIfNonEmpty(args, "service", p.Service)
	setIfNonEmpty(args, "local-address", p.LocalAddress)
	setIfNonEmpty(args, "remote-address", p.RemoteAddress)
	setIfNonEmpty(args, "comment", p.Comment)

	return command.Command{
		Raw:  "/ppp/secret/set",
		Args: args,
	}
}

// NewRemoveSecretCommand builds the command.Command for /ppp/secret/remove.
func NewRemoveSecretCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ppp/secret/remove",
		Args: map[string]string{"numbers": rosID},
	}
}

// NewPrintSecretsCommand builds the command.Command for /ppp/secret/print.
func NewPrintSecretsCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/secret/print",
		Args: args,
	}
}

// NewStreamSecretsCommand builds the command.Command for streaming /ppp/secret/print.
func NewStreamSecretsCommand(nameFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,name,service,profile,local-address,remote-address,comment,disabled,last-logged-out",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/secret/print",
		Args: args,
	}
}

// NewSetSecretDisabledCommand builds the command.Command for /ppp/secret/set disabled.
func NewSetSecretDisabledCommand(rosID string, disabled bool) command.Command {
	val := "no"
	if disabled {
		val = "yes"
	}
	return command.Command{
		Raw:  "/ppp/secret/set",
		Args: map[string]string{"numbers": rosID, "disabled": val},
	}
}

// NewPrintActiveCommand builds the command.Command for /ppp/active/print.
func NewPrintActiveCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/active/print",
		Args: args,
	}
}

// NewStreamActiveCommand builds the command.Command for /ppp/active/print follow.
func NewStreamActiveCommand(nameFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,name,service,caller-id,address,encoding,session-id,radius",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/active/print",
		Args: args,
	}
}

// NewStreamActiveStatsCommand builds the command.Command for /ppp/active/print stats.
func NewStreamActiveStatsCommand(interval string) command.Command {
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

// NewKickActiveCommand builds the command.Command for /ppp/active/remove.
func NewKickActiveCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ppp/active/remove",
		Args: map[string]string{".id": rosID},
	}
}

// NewPrintProfilesCommand builds the command.Command for /ppp/profile/print.
func NewPrintProfilesCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/profile/print",
		Args: args,
	}
}

// NewAddProfileCommand builds the command.Command for /ppp/profile/add.
func NewAddProfileCommand(p PPPProfileParams) command.Command {
	args := map[string]string{
		"name": p.Name,
	}
	setIfNonEmpty(args, "rate-limit", p.RateLimit)
	setIfNonEmpty(args, "local-address", p.LocalAddress)
	setIfNonEmpty(args, "remote-address", p.RemoteAddress)
	setIfNonEmpty(args, "dns-server", p.DNSServer)
	setIfNonEmpty(args, "parent-queue", p.ParentQueue)
	setIfNonEmpty(args, "address-list", p.AddressList)
	setIfNonEmpty(args, "session-timeout", p.SessionTimeout)
	setIfNonEmpty(args, "idle-timeout", p.IdleTimeout)
	setIfNonEmpty(args, "comment", p.Comment)
	setIfNonEmpty(args, "shared-users", p.SharedUsers)
	setIfNonEmpty(args, "only-one", p.OnlyOne)
	setIfNonEmpty(args, "use-mpls", p.UseMPLS)
	setIfNonEmpty(args, "use-compression", p.UseCompression)
	setIfNonEmpty(args, "use-encryption", p.UseEncryption)
	setIfNonEmpty(args, "change-tcp-mss", p.ChangeTCPMSS)
	setIfNonEmpty(args, "bridge-learning", p.BridgeLearning)
	setIfNonEmpty(args, "on-up", p.OnUp)
	setIfNonEmpty(args, "on-down", p.OnDown)
	return command.Command{Raw: "/ppp/profile/add", Args: args}
}

// NewSetProfileCommand builds the command.Command for /ppp/profile/set.
func NewSetProfileCommand(rosID string, p PPPProfileParams) command.Command {
	args := map[string]string{".id": rosID}
	setIfNonEmpty(args, "name", p.Name)
	setIfNonEmpty(args, "rate-limit", p.RateLimit)
	setIfNonEmpty(args, "local-address", p.LocalAddress)
	setIfNonEmpty(args, "remote-address", p.RemoteAddress)
	setIfNonEmpty(args, "dns-server", p.DNSServer)
	setIfNonEmpty(args, "parent-queue", p.ParentQueue)
	setIfNonEmpty(args, "address-list", p.AddressList)
	setIfNonEmpty(args, "session-timeout", p.SessionTimeout)
	setIfNonEmpty(args, "idle-timeout", p.IdleTimeout)
	setIfNonEmpty(args, "comment", p.Comment)
	setIfNonEmpty(args, "shared-users", p.SharedUsers)
	setIfNonEmpty(args, "only-one", p.OnlyOne)
	setIfNonEmpty(args, "use-mpls", p.UseMPLS)
	setIfNonEmpty(args, "use-compression", p.UseCompression)
	setIfNonEmpty(args, "use-encryption", p.UseEncryption)
	setIfNonEmpty(args, "change-tcp-mss", p.ChangeTCPMSS)
	setIfNonEmpty(args, "bridge-learning", p.BridgeLearning)
	setIfNonEmpty(args, "on-up", p.OnUp)
	setIfNonEmpty(args, "on-down", p.OnDown)
	return command.Command{Raw: "/ppp/profile/set", Args: args}
}

// NewRemoveProfileCommand builds the command.Command for /ppp/profile/remove.
func NewRemoveProfileCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ppp/profile/remove",
		Args: map[string]string{".id": rosID},
	}
}
