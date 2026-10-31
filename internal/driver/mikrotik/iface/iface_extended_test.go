package iface

import (
	"testing"
)

func TestInterfaceListCommandsAndParsers(t *testing.T) {
	printCmd := NewPrintInterfaceListsCommand()
	if printCmd.Raw != "/interface/list/print" || printCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid list print: %v", printCmd)
	}

	memberCmd := NewPrintInterfaceListMembersCommand()
	if memberCmd.Raw != "/interface/list/member/print" {
		t.Fatalf("invalid member print: %v", memberCmd)
	}

	lists := ParseInterfaceLists([]map[string]string{
		{".id": "*1", "name": "LAN", "builtin": "true"},
	})
	if len(lists) != 1 || lists[0].Name != "LAN" || !lists[0].Builtin {
		t.Errorf("unexpected lists parse: %+v", lists)
	}

	members := ParseInterfaceListMembers([]map[string]string{
		{".id": "*1", "interface": "ether2", "list": "LAN", "disabled": "false"},
	})
	if len(members) != 1 || members[0].Interface != "ether2" || members[0].List != "LAN" {
		t.Errorf("unexpected members parse: %+v", members)
	}
}

func TestVLANCommandsAndStreaming(t *testing.T) {
	printCmd := NewPrintVLANInterfacesCommand()
	if printCmd.Raw != "/interface/vlan/print" || printCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid vlan print: %v", printCmd)
	}

	streamCmd := NewStreamVLANInterfacesCommand()
	if streamCmd.Raw != "/interface/vlan/print" || streamCmd.Args["follow"] != "" {
		t.Fatalf("invalid vlan stream: %v", streamCmd)
	}

	vlans := ParseVLANInterfaces([]map[string]string{
		{".id": "*1", "name": "vlan10", "vlan-id": "10", "interface": "bridge", "mtu": "1500"},
	})
	if len(vlans) != 1 || vlans[0].VLANID != 10 || vlans[0].Name != "vlan10" {
		t.Errorf("unexpected vlans parse: %+v", vlans)
	}
}

func TestEthernetAndCableTest(t *testing.T) {
	ethCmd := NewPrintEthernetCommand()
	if ethCmd.Raw != "/interface/ethernet/print" || ethCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid eth print: %v", ethCmd)
	}

	cableCmd := NewCableTestCommand("ether1")
	if cableCmd.Raw != "/interface/ethernet/cable-test" || cableCmd.Args["numbers"] != "ether1" {
		t.Fatalf("invalid cable test cmd: %v", cableCmd)
	}

	eths := ParseEthernetInterfaces([]map[string]string{
		{".id": "*1", "name": "ether1", "mac-address": "00:11:22:33:44:55", "speed": "1Gbps", "running": "true"},
	})
	if len(eths) != 1 || eths[0].Speed != "1Gbps" || !eths[0].Running {
		t.Errorf("unexpected eths parse: %+v", eths)
	}

	ct := ParseCableTestResult([]map[string]string{
		{"name": "ether1", "status": "ok", "cable-pairs": "open (1), open (2)"},
	})
	if ct == nil || ct.Status != "ok" {
		t.Errorf("unexpected cable test parse: %+v", ct)
	}
}

func TestVRRPAndDetectInternet(t *testing.T) {
	vrrpCmd := NewPrintVRRPCommand()
	if vrrpCmd.Raw != "/interface/vrrp/print" || vrrpCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid vrrp print: %v", vrrpCmd)
	}

	vrrps := ParseVRRPInterfaces([]map[string]string{
		{".id": "*1", "name": "vrrp1", "interface": "ether1", "vrid": "10", "priority": "100", "running": "true"},
	})
	if len(vrrps) != 1 || vrrps[0].VRID != 10 || !vrrps[0].Running {
		t.Errorf("unexpected vrrps parse: %+v", vrrps)
	}

	di := ParseDetectInternetState([]map[string]string{
		{"interface": "ether1", "state": "internet", "internet": "yes"},
	})
	if len(di) != 1 || di[0].State != "internet" {
		t.Errorf("unexpected detect internet parse: %+v", di)
	}
}
