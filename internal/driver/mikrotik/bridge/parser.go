package bridge

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseBridges converts command.Result from /interface/bridge/print into []Bridge.
func ParseBridges(result command.Result) []Bridge {
	bridges := make([]Bridge, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		bridges = append(bridges, Bridge{
			RosID:         id,
			Name:          name,
			MTU:           row["mtu"],
			ActualMTU:     row["actual-mtu"],
			MACAddress:    row["mac-address"],
			FastForward:   strings.EqualFold(row["fast-forward"], "true") || strings.EqualFold(row["fast-forward"], "yes"),
			VlanFiltering: strings.EqualFold(row["vlan-filtering"], "true") || strings.EqualFold(row["vlan-filtering"], "yes"),
			Disabled:      strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Running:       strings.EqualFold(row["running"], "true") || strings.EqualFold(row["running"], "yes"),
			Comment:       row["comment"],
		})
	}
	return bridges
}

// ParseBridgePorts converts command.Result from /interface/bridge/port/print into []BridgePort.
func ParseBridgePorts(result command.Result) []BridgePort {
	ports := make([]BridgePort, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		iface := row["interface"]
		if id == "" || iface == "" {
			continue
		}
		ports = append(ports, BridgePort{
			RosID:     id,
			Interface: iface,
			Bridge:    row["bridge"],
			Priority:  row["priority"],
			PathCost:  row["path-cost"],
			Horizon:   row["horizon"],
			Learn:     row["learn"],
			PVID:      row["pvid"],
			Disabled:  strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Inactive:  strings.EqualFold(row["inactive"], "true") || strings.EqualFold(row["inactive"], "yes"),
			Comment:   row["comment"],
		})
	}
	return ports
}

// ParseBridgeVLANs converts command.Result from /interface/bridge/vlan/print into []BridgeVLAN.
func ParseBridgeVLANs(result command.Result) []BridgeVLAN {
	vlans := make([]BridgeVLAN, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		bridge := row["bridge"]
		if id == "" || bridge == "" {
			continue
		}
		vlans = append(vlans, BridgeVLAN{
			RosID:           id,
			Bridge:          bridge,
			VlanIDs:         row["vlan-ids"],
			Tagged:          row["tagged"],
			Untagged:        row["untagged"],
			CurrentTagged:   row["current-tagged"],
			CurrentUntagged: row["current-untagged"],
			Disabled:        strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Comment:         row["comment"],
		})
	}
	return vlans
}

// ParseBridgeHosts converts command.Result from /interface/bridge/host/print into []BridgeHost.
func ParseBridgeHosts(result command.Result) []BridgeHost {
	hosts := make([]BridgeHost, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		mac := row["mac-address"]
		if id == "" || mac == "" {
			continue
		}
		hosts = append(hosts, BridgeHost{
			RosID:       id,
			Bridge:      row["bridge"],
			MACAddress:  mac,
			Interface:   row["interface"],
			OnInterface: row["on-interface"],
			Age:         row["age"],
			Local:       strings.EqualFold(row["local"], "true") || strings.EqualFold(row["local"], "yes"),
			Dynamic:     strings.EqualFold(row["dynamic"], "true") || strings.EqualFold(row["dynamic"], "yes"),
			Disabled:    strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
		})
	}
	return hosts
}
