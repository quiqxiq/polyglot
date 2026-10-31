package firewall

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// NewAddIPAddressCommand builds the command.Command for /ip/address/add.
func NewAddIPAddressCommand(p IPAddressParams) command.Command {
	args := map[string]string{
		"address":   p.Address,
		"interface": p.Interface,
	}
	if p.Network != "" {
		args["network"] = p.Network
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ip/address/add",
		Args: args,
	}
}

// NewSetIPAddressCommand builds the command.Command for /ip/address/set.
func NewSetIPAddressCommand(rosID string, p IPAddressParams) command.Command {
	args := map[string]string{
		".id": rosID,
	}
	if p.Address != "" {
		args["address"] = p.Address
	}
	if p.Interface != "" {
		args["interface"] = p.Interface
	}
	if p.Network != "" {
		args["network"] = p.Network
	}
	args["comment"] = p.Comment
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/ip/address/set",
		Args: args,
	}
}

// NewRemoveIPAddressCommand builds the command.Command for /ip/address/remove.
func NewRemoveIPAddressCommand(rosID string) command.Command {
	return command.Command{
		Raw: "/ip/address/remove",
		Args: map[string]string{
			".id": rosID,
		},
	}
}

// NewStreamIPAddressesCommand builds the command.Command for /ip/address/print with follow mode.
func NewStreamIPAddressesCommand(p IPAddressPrintParams) command.Command {
	args := map[string]string{
		".proplist": ".id,address,network,interface,actual-interface,invalid,dynamic,disabled,comment",
		"follow":    "",
	}
	if p.Interface != "" {
		args["?interface"] = p.Interface
	}
	if p.Address != "" {
		args["?address"] = p.Address
	}
	return command.Command{
		Raw:  "/ip/address/print",
		Args: args,
	}
}

// NewAddPoolCommand builds the command.Command for /ip/pool/add.
func NewAddPoolCommand(p IPPoolParams) command.Command {
	args := map[string]string{
		"name":   p.Name,
		"ranges": p.Ranges,
	}
	if p.NextPool != "" {
		args["next-pool"] = p.NextPool
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	return command.Command{
		Raw:  "/ip/pool/add",
		Args: args,
	}
}

// NewSetPoolCommand builds the command.Command for /ip/pool/set.
func NewSetPoolCommand(rosID string, p IPPoolParams) command.Command {
	args := map[string]string{
		".id": rosID,
	}
	if p.Name != "" {
		args["name"] = p.Name
	}
	if p.Ranges != "" {
		args["ranges"] = p.Ranges
	}
	if p.NextPool != "" {
		args["next-pool"] = p.NextPool
	}
	args["comment"] = p.Comment
	return command.Command{
		Raw:  "/ip/pool/set",
		Args: args,
	}
}

// NewRemovePoolCommand builds the command.Command for /ip/pool/remove.
func NewRemovePoolCommand(rosID string) command.Command {
	return command.Command{
		Raw: "/ip/pool/remove",
		Args: map[string]string{
			".id": rosID,
		},
	}
}
