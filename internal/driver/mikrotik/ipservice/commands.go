package ipservice

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// NewPrintIPServicesCommand queries configured IP services with strict proplist.
func NewPrintIPServicesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "port", "disabled", "address", "certificate")
	return command.Command{
		Raw:  "/ip/service/print",
		Args: args,
	}
}

// NewSetIPServiceCommand updates an IP service (port, address restriction, certificate, state).
func NewSetIPServiceCommand(svc IPService) command.Command {
	args := map[string]string{
		".id": svc.ID,
	}
	if svc.Port > 0 {
		args["port"] = strconv.Itoa(svc.Port)
	}
	args["address"] = svc.Address
	args["certificate"] = svc.Certificate
	if svc.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/ip/service/set",
		Args: args,
	}
}

// NewPrintIPCloudCommand queries IP Cloud DDNS settings and resolved address.
func NewPrintIPCloudCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "ddns-enabled", "update-time", "public-address", "dns-name", "status")
	return command.Command{
		Raw:  "/ip/cloud/print",
		Args: args,
	}
}

// NewSetIPCloudCommand toggles IP Cloud dynamic DNS.
func NewSetIPCloudCommand(ddnsEnabled, updateTime bool) command.Command {
	ddns := "no"
	if ddnsEnabled {
		ddns = "yes"
	}
	ut := "no"
	if updateTime {
		ut = "yes"
	}
	return command.Command{
		Raw: "/ip/cloud/set",
		Args: map[string]string{
			"ddns-enabled": ddns,
			"update-time":  ut,
		},
	}
}

// NewPrintNeighborsCommand queries discovered neighbors (CDP/MNDP/LLDP) with strict proplist.
func NewPrintNeighborsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "interface", "mac-address", "identity", "platform", "version", "unpack", "age", "system-description", "system-caps")
	return command.Command{
		Raw:  "/ip/neighbor/print",
		Args: args,
	}
}

// NewPrintDiscoverySettingsCommand queries neighbor discovery interface settings.
func NewPrintDiscoverySettingsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "discover-interface-list")
	return command.Command{
		Raw:  "/ip/neighbor/discovery-settings/print",
		Args: args,
	}
}

// NewSetDiscoverySettingsCommand sets neighbor discovery allowed interface list.
func NewSetDiscoverySettingsCommand(interfaceList string) command.Command {
	return command.Command{
		Raw: "/ip/neighbor/discovery-settings/set",
		Args: map[string]string{
			"discover-interface-list": interfaceList,
		},
	}
}

// NewPrintARPCommand queries the ARP table with strict proplist.
func NewPrintARPCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "address", "mac-address", "interface", "status", "dynamic", "DHCP", "disabled", "comment")
	return command.Command{
		Raw:  "/ip/arp/print",
		Args: args,
	}
}

// NewStreamARPCommand builds a Mode A follow command for ARP table events.
func NewStreamARPCommand() command.Command {
	args := map[string]string{
		"follow": "",
	}
	rosutil.SetProplist(args, ".id", "address", "mac-address", "interface", "status", "dynamic", "DHCP", "disabled", "comment")
	return command.Command{
		Raw:  "/ip/arp/print",
		Args: args,
	}
}

// NewAddARPCommand adds a static ARP mapping.
func NewAddARPCommand(address, macAddress, iface, comment string) command.Command {
	args := map[string]string{
		"address":     address,
		"mac-address": macAddress,
		"interface":   iface,
	}
	if comment != "" {
		args["comment"] = comment
	}
	return command.Command{
		Raw:  "/ip/arp/add",
		Args: args,
	}
}

// NewSetARPCommand updates an ARP entry.
func NewSetARPCommand(id, address, macAddress, iface, comment string, disabled bool) command.Command {
	args := map[string]string{
		".id":         id,
		"address":     address,
		"mac-address": macAddress,
		"interface":   iface,
		"comment":     comment,
	}
	if disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/ip/arp/set",
		Args: args,
	}
}

// NewRemoveARPCommand removes an ARP entry by ID.
func NewRemoveARPCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/arp/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintSNMPCommand queries global SNMP settings.
func NewPrintSNMPCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "enabled", "contact", "location")
	return command.Command{
		Raw:  "/snmp/print",
		Args: args,
	}
}

// NewSetSNMPCommand updates global SNMP settings.
func NewSetSNMPCommand(enabled bool, contact, location string) command.Command {
	val := "no"
	if enabled {
		val = "yes"
	}
	return command.Command{
		Raw: "/snmp/set",
		Args: map[string]string{
			"enabled":  val,
			"contact":  contact,
			"location": location,
		},
	}
}

// NewPrintSNMPCommunitiesCommand queries SNMP community configs with strict proplist.
func NewPrintSNMPCommunitiesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "addresses", "security", "read-access", "write-access")
	return command.Command{
		Raw:  "/snmp/community/print",
		Args: args,
	}
}

// NewAddSNMPCommunityCommand creates an SNMP community.
func NewAddSNMPCommunityCommand(c SNMPCommunity) command.Command {
	args := map[string]string{
		"name": c.Name,
	}
	if c.Addresses != "" {
		args["addresses"] = c.Addresses
	}
	if c.Security != "" {
		args["security"] = c.Security
	}
	if c.ReadAccess {
		args["read-access"] = "yes"
	} else {
		args["read-access"] = "no"
	}
	if c.WriteAccess {
		args["write-access"] = "yes"
	} else {
		args["write-access"] = "no"
	}
	return command.Command{
		Raw:  "/snmp/community/add",
		Args: args,
	}
}

// NewSetSNMPCommunityCommand updates an SNMP community.
func NewSetSNMPCommunityCommand(c SNMPCommunity) command.Command {
	args := map[string]string{
		".id":       c.ID,
		"name":      c.Name,
		"addresses": c.Addresses,
		"security":  c.Security,
	}
	if c.ReadAccess {
		args["read-access"] = "yes"
	} else {
		args["read-access"] = "no"
	}
	if c.WriteAccess {
		args["write-access"] = "yes"
	} else {
		args["write-access"] = "no"
	}
	return command.Command{
		Raw:  "/snmp/community/set",
		Args: args,
	}
}

// NewRemoveSNMPCommunityCommand removes an SNMP community by ID.
func NewRemoveSNMPCommunityCommand(id string) command.Command {
	return command.Command{
		Raw: "/snmp/community/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}
