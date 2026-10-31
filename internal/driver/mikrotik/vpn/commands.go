package vpn

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// NewPrintWireGuardInterfacesCommand builds the command.Command for /interface/wireguard/print.
func NewPrintWireGuardInterfacesCommand(nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,listen-port,public-key,private-key,mtu,running,disabled,comment",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/interface/wireguard/print",
		Args: args,
	}
}

// NewPrintWireGuardPeersCommand builds the command.Command for /interface/wireguard/peers/print.
func NewPrintWireGuardPeersCommand(ifaceFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,interface,public-key,endpoint-address,endpoint-port,allowed-address,latest-handshake,rx,tx,disabled,comment",
	}
	if ifaceFilter != "" {
		args["?interface"] = ifaceFilter
	}
	return command.Command{
		Raw:  "/interface/wireguard/peers/print",
		Args: args,
	}
}

// NewStreamWireGuardPeersCommand builds the command.Command for streaming /interface/wireguard/peers/print follow.
func NewStreamWireGuardPeersCommand(ifaceFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,interface,public-key,endpoint-address,endpoint-port,allowed-address,latest-handshake,rx,tx,disabled,comment",
	}
	if ifaceFilter != "" {
		args["?interface"] = ifaceFilter
	}
	return command.Command{
		Raw:  "/interface/wireguard/peers/print",
		Args: args,
	}
}

// NewPrintIPsecPeersCommand builds the command.Command for /ip/ipsec/peer/print.
func NewPrintIPsecPeersCommand() command.Command {
	return command.Command{
		Raw: "/ip/ipsec/peer/print",
		Args: map[string]string{
			".proplist": ".id,name,address,profile,exchange-mode,disabled,comment",
		},
	}
}

// NewPrintIPsecActivePeersCommand builds the command.Command for /ip/ipsec/active-peers/print.
func NewPrintIPsecActivePeersCommand() command.Command {
	return command.Command{
		Raw: "/ip/ipsec/active-peers/print",
		Args: map[string]string{
			".proplist": ".id,id,state,uptime,remote-address,local-address",
		},
	}
}
