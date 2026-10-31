package ipservice

import (
	"testing"
)

func TestIPServiceCommandsAndParsers(t *testing.T) {
	printCmd := NewPrintIPServicesCommand()
	if printCmd.Raw != "/ip/service/print" || printCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid print cmd: %v", printCmd)
	}

	setCmd := NewSetIPServiceCommand(IPService{
		ID:       "*1",
		Port:     8728,
		Disabled: false,
		Address:  "192.168.1.0/24",
	})
	if setCmd.Raw != "/ip/service/set" || setCmd.Args["port"] != "8728" || setCmd.Args["disabled"] != "no" {
		t.Errorf("unexpected set cmd: %v", setCmd)
	}

	services := ParseIPServices([]map[string]string{
		{".id": "*1", "name": "api", "port": "8728", "disabled": "false", "address": "192.168.1.0/24"},
	})
	if len(services) != 1 || services[0].Port != 8728 || services[0].Name != "api" {
		t.Errorf("unexpected services parse: %+v", services)
	}
}

func TestIPCloudCommandsAndParsers(t *testing.T) {
	printCmd := NewPrintIPCloudCommand()
	if printCmd.Raw != "/ip/cloud/print" {
		t.Fatalf("invalid cloud print: %v", printCmd)
	}

	setCmd := NewSetIPCloudCommand(true, true)
	if setCmd.Args["ddns-enabled"] != "yes" || setCmd.Args["update-time"] != "yes" {
		t.Errorf("unexpected cloud set: %v", setCmd)
	}

	cloud := ParseIPCloud([]map[string]string{
		{"ddns-enabled": "yes", "update-time": "yes", "public-address": "203.0.113.1", "dns-name": "xyz.sn.mynetname.net", "status": "updated"},
	})
	if cloud == nil || !cloud.DDNSEnabled || cloud.PublicAddress != "203.0.113.1" {
		t.Errorf("unexpected cloud parse: %+v", cloud)
	}
}

func TestNeighborAndDiscovery(t *testing.T) {
	printCmd := NewPrintNeighborsCommand()
	if printCmd.Raw != "/ip/neighbor/print" {
		t.Errorf("invalid neighbor cmd: %v", printCmd)
	}

	neighbors := ParseNeighbors([]map[string]string{
		{".id": "*1", "interface": "ether1", "identity": "Core-SW", "mac-address": "00:11:22:33:44:55", "platform": "MikroTik"},
	})
	if len(neighbors) != 1 || neighbors[0].Identity != "Core-SW" {
		t.Errorf("unexpected neighbor parse: %+v", neighbors)
	}

	disc := ParseDiscoverySettings([]map[string]string{
		{"discover-interface-list": "LAN"},
	})
	if disc == nil || disc.DiscoverInterfaceList != "LAN" {
		t.Errorf("unexpected discovery settings: %+v", disc)
	}
}

func TestARPCommandsAndStreaming(t *testing.T) {
	printCmd := NewPrintARPCommand()
	if printCmd.Raw != "/ip/arp/print" {
		t.Errorf("invalid arp print: %v", printCmd)
	}

	streamCmd := NewStreamARPCommand()
	if streamCmd.Raw != "/ip/arp/print" || streamCmd.Args["follow"] != "" {
		t.Errorf("invalid stream arp: %v", streamCmd)
	}

	entries := ParseARPEntries([]map[string]string{
		{".id": "*1", "address": "192.168.1.50", "mac-address": "AA:BB:CC:DD:EE:FF", "interface": "bridge", "dynamic": "true"},
	})
	if len(entries) != 1 || !entries[0].Dynamic || entries[0].Address != "192.168.1.50" {
		t.Errorf("unexpected arp parse: %+v", entries)
	}
}

func TestSNMPCommandsAndParsers(t *testing.T) {
	snmpCmd := NewPrintSNMPCommand()
	if snmpCmd.Raw != "/snmp/print" {
		t.Errorf("invalid snmp print: %v", snmpCmd)
	}

	snmp := ParseSNMPSettings([]map[string]string{
		{"enabled": "yes", "contact": "noc@example.com", "location": "DC-1"},
	})
	if snmp == nil || !snmp.Enabled || snmp.Contact != "noc@example.com" {
		t.Errorf("unexpected snmp parse: %+v", snmp)
	}

	comm := ParseSNMPCommunities([]map[string]string{
		{".id": "*1", "name": "public", "read-access": "yes", "write-access": "no"},
	})
	if len(comm) != 1 || !comm[0].ReadAccess || comm[0].WriteAccess {
		t.Errorf("unexpected comm parse: %+v", comm)
	}
}
