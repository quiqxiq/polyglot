package dns

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

var setIfNonEmpty = rosutil.SetIfNonEmpty

// NewPrintDNSSettingsCommand builds the command.Command for /ip/dns/print.
func NewPrintDNSSettingsCommand() command.Command {
	return command.Command{
		Raw: "/ip/dns/print",
		Args: map[string]string{
			".proplist": "servers,dynamic-servers,allow-remote-requests,cache-size,cache-used",
		},
	}
}

// NewPrintDNSStaticCommand builds the command.Command for /ip/dns/static/print.
func NewPrintDNSStaticCommand(nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,address,ttl,comment,disabled,dynamic",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ip/dns/static/print",
		Args: args,
	}
}

// NewStreamDNSStaticCommand builds the command.Command for streaming /ip/dns/static/print follow.
func NewStreamDNSStaticCommand(nameFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,name,address,ttl,comment,disabled,dynamic",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ip/dns/static/print",
		Args: args,
	}
}

// NewAddDNSStaticCommand builds the command.Command for /ip/dns/static/add.
func NewAddDNSStaticCommand(p DNSStaticParams) command.Command {
	args := map[string]string{
		"name":    p.Name,
		"address": p.Address,
	}
	setIfNonEmpty(args, "ttl", p.TTL)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/ip/dns/static/add", Args: args}
}

// NewSetDNSStaticCommand builds the command.Command for /ip/dns/static/set.
func NewSetDNSStaticCommand(rosID string, p DNSStaticParams) command.Command {
	args := map[string]string{"numbers": rosID}
	setIfNonEmpty(args, "name", p.Name)
	setIfNonEmpty(args, "address", p.Address)
	setIfNonEmpty(args, "ttl", p.TTL)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{Raw: "/ip/dns/static/set", Args: args}
}

// NewRemoveDNSStaticCommand builds the command.Command for /ip/dns/static/remove.
func NewRemoveDNSStaticCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/dns/static/remove",
		Args: map[string]string{"numbers": rosID},
	}
}

// NewPrintDNSCacheCommand builds the command.Command for /ip/dns/cache/print.
func NewPrintDNSCacheCommand(nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,type,data,ttl,static",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ip/dns/cache/print",
		Args: args,
	}
}
