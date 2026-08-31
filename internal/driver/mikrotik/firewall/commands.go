package firewall

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

var setIfNonEmpty = rosutil.SetIfNonEmpty

// NewPrintFiltersCommand builds the command.Command for /ip/firewall/filter/print.
func NewPrintFiltersCommand(p FirewallFilterPrintParams) command.Command {
	args := map[string]string{}
	setIfNonEmpty(args, "?chain", p.Chain)
	setIfNonEmpty(args, "?src-address", p.SrcAddress)
	setIfNonEmpty(args, "?src-address-list", p.SrcAddressList)
	setIfNonEmpty(args, "?action", p.Action)
	return command.Command{Raw: "/ip/firewall/filter/print", Args: args}
}

// NewAddFilterCommand builds the command.Command for /ip/firewall/filter/add.
func NewAddFilterCommand(p FirewallFilterParams) command.Command {
	args := map[string]string{
		"chain":  p.Chain,
		"action": p.Action,
	}
	setIfNonEmpty(args, "src-address", p.SrcAddress)
	setIfNonEmpty(args, "src-address-list", p.SrcAddressList)
	setIfNonEmpty(args, "dst-address", p.DstAddress)
	setIfNonEmpty(args, "protocol", p.Protocol)
	setIfNonEmpty(args, "place-before", p.PlaceBefore)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/ip/firewall/filter/add", Args: args}
}

// NewSetFilterCommand builds the command.Command for /ip/firewall/filter/set.
func NewSetFilterCommand(rosID string, p FirewallFilterParams) command.Command {
	args := map[string]string{"numbers": rosID}
	setIfNonEmpty(args, "chain", p.Chain)
	setIfNonEmpty(args, "action", p.Action)
	setIfNonEmpty(args, "src-address", p.SrcAddress)
	setIfNonEmpty(args, "src-address-list", p.SrcAddressList)
	setIfNonEmpty(args, "dst-address", p.DstAddress)
	setIfNonEmpty(args, "protocol", p.Protocol)
	setIfNonEmpty(args, "comment", p.Comment)
	return command.Command{Raw: "/ip/firewall/filter/set", Args: args}
}

// NewRemoveFilterCommand builds the command.Command for /ip/firewall/filter/remove.
func NewRemoveFilterCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/firewall/filter/remove",
		Args: map[string]string{"numbers": rosID},
	}
}

// NewPrintAddressListCommand builds the command.Command for /ip/firewall/address-list/print.
func NewPrintAddressListCommand(p AddressListPrintParams) command.Command {
	args := map[string]string{}
	setIfNonEmpty(args, "?list", p.List)
	setIfNonEmpty(args, "?address", p.Address)
	return command.Command{Raw: "/ip/firewall/address-list/print", Args: args}
}

// NewAddAddressListCommand builds the command.Command for /ip/firewall/address-list/add.
func NewAddAddressListCommand(p AddressListParams) command.Command {
	args := map[string]string{
		"list":    p.List,
		"address": p.Address,
	}
	setIfNonEmpty(args, "timeout", p.Timeout)
	setIfNonEmpty(args, "comment", p.Comment)
	return command.Command{Raw: "/ip/firewall/address-list/add", Args: args}
}

// NewRemoveAddressListCommand builds the command.Command for /ip/firewall/address-list/remove.
func NewRemoveAddressListCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/firewall/address-list/remove",
		Args: map[string]string{".id": rosID},
	}
}

// NewPrintNATCommand builds the command.Command for /ip/firewall/nat/print.
func NewPrintNATCommand(chain, comment, srcAddressList string) command.Command {
	args := map[string]string{}
	setIfNonEmpty(args, "?chain", chain)
	setIfNonEmpty(args, "?comment", comment)
	setIfNonEmpty(args, "?src-address-list", srcAddressList)
	return command.Command{Raw: "/ip/firewall/nat/print", Args: args}
}

// NewAddNATCommand builds the command.Command for /ip/firewall/nat/add.
func NewAddNATCommand(p NATRuleParams) command.Command {
	args := map[string]string{
		"chain":  p.Chain,
		"action": p.Action,
	}
	setIfNonEmpty(args, "to-addresses", p.ToAddresses)
	setIfNonEmpty(args, "to-ports", p.ToPorts)
	setIfNonEmpty(args, "src-address", p.SrcAddress)
	setIfNonEmpty(args, "src-address-list", p.SrcAddressList)
	setIfNonEmpty(args, "dst-address", p.DstAddress)
	setIfNonEmpty(args, "protocol", p.Protocol)
	setIfNonEmpty(args, "dst-port", p.DstPort)
	setIfNonEmpty(args, "place-before", p.PlaceBefore)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/ip/firewall/nat/add", Args: args}
}

// NewSetNATCommand builds the command.Command for /ip/firewall/nat/set.
func NewSetNATCommand(rosID string, p NATRuleParams) command.Command {
	args := map[string]string{"numbers": rosID}
	setIfNonEmpty(args, "action", p.Action)
	setIfNonEmpty(args, "to-addresses", p.ToAddresses)
	setIfNonEmpty(args, "to-ports", p.ToPorts)
	setIfNonEmpty(args, "src-address-list", p.SrcAddressList)
	setIfNonEmpty(args, "protocol", p.Protocol)
	setIfNonEmpty(args, "dst-port", p.DstPort)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{Raw: "/ip/firewall/nat/set", Args: args}
}

// NewRemoveNATCommand builds the command.Command for /ip/firewall/nat/remove.
func NewRemoveNATCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/firewall/nat/remove",
		Args: map[string]string{"numbers": rosID},
	}
}

// NewPrintPoolsCommand builds the command.Command for /ip/pool/print.
func NewPrintPoolsCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{Raw: "/ip/pool/print", Args: args}
}

// NewPrintIPAddressesCommand builds the command.Command for /ip/address/print.
func NewPrintIPAddressesCommand(p IPAddressPrintParams) command.Command {
	args := map[string]string{}
	setIfNonEmpty(args, "?interface", p.Interface)
	setIfNonEmpty(args, "?address", p.Address)
	return command.Command{Raw: "/ip/address/print", Args: args}
}

// NewPrintIPRoutesCommand builds the command.Command for /ip/route/print.
func NewPrintIPRoutesCommand(dstFilter string) command.Command {
	args := map[string]string{}
	if dstFilter != "" {
		args["?dst-address"] = dstFilter
	}
	return command.Command{Raw: "/ip/route/print", Args: args}
}
