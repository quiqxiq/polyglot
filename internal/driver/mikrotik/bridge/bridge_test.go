package bridge

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestBridgeCommands(t *testing.T) {
	cmdBridges := NewPrintBridgesCommand("bridge1")
	if cmdBridges.Raw != "/interface/bridge/print" || cmdBridges.Args["?name"] != "bridge1" || cmdBridges.Args[".proplist"] == "" {
		t.Fatalf("unexpected bridges command: %+v", cmdBridges)
	}

	cmdPorts := NewPrintBridgePortsCommand("bridge1")
	if cmdPorts.Raw != "/interface/bridge/port/print" || cmdPorts.Args["?bridge"] != "bridge1" || cmdPorts.Args[".proplist"] == "" {
		t.Fatalf("unexpected bridge ports command: %+v", cmdPorts)
	}

	cmdVlans := NewPrintBridgeVLANsCommand("bridge1")
	if cmdVlans.Raw != "/interface/bridge/vlan/print" || cmdVlans.Args["?bridge"] != "bridge1" || cmdVlans.Args[".proplist"] == "" {
		t.Fatalf("unexpected bridge vlans command: %+v", cmdVlans)
	}

	cmdHosts := NewPrintBridgeHostsCommand("bridge1", "00:11:22:33:44:55")
	if cmdHosts.Raw != "/interface/bridge/host/print" || cmdHosts.Args[".proplist"] == "" {
		t.Fatalf("unexpected bridge hosts command: %+v", cmdHosts)
	}
}

func TestParseBridges(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":            "*1",
				"name":           "bridge-lan",
				"mtu":            "1500",
				"actual-mtu":     "1500",
				"mac-address":    "48:8F:5A:11:22:33",
				"fast-forward":   "true",
				"vlan-filtering": "no",
				"running":        "yes",
				"disabled":       "false",
			},
		},
	}
	bridges := ParseBridges(res)
	if len(bridges) != 1 || bridges[0].Name != "bridge-lan" || !bridges[0].Running || !bridges[0].FastForward {
		t.Fatalf("unexpected bridges parsed: %+v", bridges)
	}
}

func TestParseBridgePorts(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":       "*2",
				"interface": "ether2",
				"bridge":    "bridge-lan",
				"pvid":      "10",
				"disabled":  "false",
				"inactive":  "false",
			},
		},
	}
	ports := ParseBridgePorts(res)
	if len(ports) != 1 || ports[0].Interface != "ether2" || ports[0].PVID != "10" {
		t.Fatalf("unexpected bridge ports parsed: %+v", ports)
	}
}

func TestBridgeVLANMutationsAndStreaming(t *testing.T) {
	addVlan := NewAddBridgeVLANCommand(BridgeVLANParams{
		Bridge:   "bridge",
		VlanIDs:  "10,20",
		Tagged:   "ether1",
		Untagged: "ether2",
		Comment:  "Office VLANs",
	})
	if addVlan.Raw != "/interface/bridge/vlan/add" || addVlan.Args["vlan-ids"] != "10,20" {
		t.Errorf("invalid add bridge vlan: %v", addVlan)
	}

	setVlan := NewSetBridgeVLANCommand("*1", BridgeVLANParams{
		VlanIDs:  "10,20,30",
		Tagged:   "ether1,ether3",
		Disabled: false,
	})
	if setVlan.Raw != "/interface/bridge/vlan/set" || setVlan.Args[".id"] != "*1" {
		t.Errorf("invalid set bridge vlan: %v", setVlan)
	}

	remVlan := NewRemoveBridgeVLANCommand("*1")
	if remVlan.Raw != "/interface/bridge/vlan/remove" || remVlan.Args[".id"] != "*1" {
		t.Errorf("invalid remove bridge vlan: %v", remVlan)
	}

	streamHosts := NewStreamBridgeHostsCommand("bridge")
	if streamHosts.Raw != "/interface/bridge/host/print" || streamHosts.Args["follow"] != "" || streamHosts.Args["?bridge"] != "bridge" {
		t.Errorf("invalid stream bridge hosts: %v", streamHosts)
	}

	vlans := ParseBridgeVLANs(command.Result{
		Rows: []map[string]string{
			{".id": "*1", "bridge": "bridge", "vlan-ids": "10,20", "tagged": "ether1", "untagged": "ether2"},
		},
	})
	if len(vlans) != 1 || vlans[0].VlanIDs != "10,20" {
		t.Errorf("unexpected vlans parse: %+v", vlans)
	}

	hosts := ParseBridgeHosts(command.Result{
		Rows: []map[string]string{
			{".id": "*2", "bridge": "bridge", "mac-address": "00:11:22:33:44:55", "interface": "ether2", "local": "false", "dynamic": "true"},
		},
	})
	if len(hosts) != 1 || hosts[0].MACAddress != "00:11:22:33:44:55" || !hosts[0].Dynamic {
		t.Errorf("unexpected hosts parse: %+v", hosts)
	}
}
