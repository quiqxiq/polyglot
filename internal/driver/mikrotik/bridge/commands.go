package bridge

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

var setIfNonEmpty = rosutil.SetIfNonEmpty

// NewPrintBridgesCommand builds the command.Command for /interface/bridge/print.
func NewPrintBridgesCommand(nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,mtu,actual-mtu,mac-address,fast-forward,vlan-filtering,running,disabled,comment",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/interface/bridge/print",
		Args: args,
	}
}

// NewStreamBridgesCommand builds the command.Command for streaming /interface/bridge/print follow.
func NewStreamBridgesCommand(nameFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,name,mtu,actual-mtu,mac-address,fast-forward,vlan-filtering,running,disabled,comment",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/interface/bridge/print",
		Args: args,
	}
}

// NewAddBridgeCommand builds the command.Command for /interface/bridge/add.
func NewAddBridgeCommand(p BridgeParams) command.Command {
	args := map[string]string{
		"name": p.Name,
	}
	if p.VlanFiltering {
		args["vlan-filtering"] = "yes"
	}
	if p.FastForward {
		args["fast-forward"] = "yes"
	} else {
		args["fast-forward"] = "no"
	}
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/interface/bridge/add", Args: args}
}

// NewPrintBridgePortsCommand builds the command.Command for /interface/bridge/port/print.
func NewPrintBridgePortsCommand(bridgeFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,interface,bridge,priority,path-cost,horizon,learn,pvid,disabled,inactive,comment",
	}
	if bridgeFilter != "" {
		args["?bridge"] = bridgeFilter
	}
	return command.Command{
		Raw:  "/interface/bridge/port/print",
		Args: args,
	}
}

// NewAddBridgePortCommand builds the command.Command for /interface/bridge/port/add.
func NewAddBridgePortCommand(p BridgePortParams) command.Command {
	args := map[string]string{
		"interface": p.Interface,
		"bridge":    p.Bridge,
	}
	setIfNonEmpty(args, "pvid", p.PVID)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/interface/bridge/port/add", Args: args}
}

// NewPrintBridgeVLANsCommand builds the command.Command for /interface/bridge/vlan/print.
func NewPrintBridgeVLANsCommand(bridgeFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,bridge,vlan-ids,tagged,untagged,current-tagged,current-untagged,disabled,comment",
	}
	if bridgeFilter != "" {
		args["?bridge"] = bridgeFilter
	}
	return command.Command{
		Raw:  "/interface/bridge/vlan/print",
		Args: args,
	}
}

// NewAddBridgeVLANCommand builds the command.Command for /interface/bridge/vlan/add.
func NewAddBridgeVLANCommand(p BridgeVLANParams) command.Command {
	args := map[string]string{
		"bridge":   p.Bridge,
		"vlan-ids": p.VlanIDs,
	}
	setIfNonEmpty(args, "tagged", p.Tagged)
	setIfNonEmpty(args, "untagged", p.Untagged)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/interface/bridge/vlan/add",
		Args: args,
	}
}

// NewSetBridgeVLANCommand builds the command.Command for /interface/bridge/vlan/set.
func NewSetBridgeVLANCommand(rosID string, p BridgeVLANParams) command.Command {
	args := map[string]string{
		".id": rosID,
	}
	setIfNonEmpty(args, "bridge", p.Bridge)
	setIfNonEmpty(args, "vlan-ids", p.VlanIDs)
	setIfNonEmpty(args, "tagged", p.Tagged)
	setIfNonEmpty(args, "untagged", p.Untagged)
	args["comment"] = p.Comment
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/interface/bridge/vlan/set",
		Args: args,
	}
}

// NewRemoveBridgeVLANCommand builds the command.Command for /interface/bridge/vlan/remove.
func NewRemoveBridgeVLANCommand(rosID string) command.Command {
	return command.Command{
		Raw: "/interface/bridge/vlan/remove",
		Args: map[string]string{
			".id": rosID,
		},
	}
}

// NewPrintBridgeHostsCommand builds the command.Command for /interface/bridge/host/print.
func NewPrintBridgeHostsCommand(bridgeFilter, macFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,bridge,mac-address,interface,on-interface,age,local,dynamic,disabled",
	}
	if bridgeFilter != "" {
		args["?bridge"] = bridgeFilter
	}
	if macFilter != "" {
		args["?mac-address"] = macFilter
	}
	return command.Command{
		Raw:  "/interface/bridge/host/print",
		Args: args,
	}
}

// NewStreamBridgeHostsCommand builds a Mode A follow command for /interface/bridge/host/print.
func NewStreamBridgeHostsCommand(bridgeFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,bridge,mac-address,interface,on-interface,age,local,dynamic,disabled",
	}
	if bridgeFilter != "" {
		args["?bridge"] = bridgeFilter
	}
	return command.Command{
		Raw:  "/interface/bridge/host/print",
		Args: args,
	}
}
