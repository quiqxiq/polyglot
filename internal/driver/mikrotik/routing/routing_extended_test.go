package routing

import (
	"testing"
)

func TestOSPFCommandsAndParsers(t *testing.T) {
	instCmd := NewPrintOSPFInstancesCommand()
	if instCmd.Raw != "/routing/ospf/instance/print" || instCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid inst print: %v", instCmd)
	}

	instances := ParseOSPFInstances([]map[string]string{
		{".id": "*1", "name": "default", "version": "2", "router-id": "10.0.0.1", "disabled": "false"},
	})
	if len(instances) != 1 || instances[0].RouterID != "10.0.0.1" || instances[0].Version != 2 {
		t.Errorf("unexpected instances parse: %+v", instances)
	}

	areaCmd := NewPrintOSPFAreasCommand()
	if areaCmd.Raw != "/routing/ospf/area/print" {
		t.Fatalf("invalid area print: %v", areaCmd)
	}

	areas := ParseOSPFAreas([]map[string]string{
		{".id": "*1", "name": "backbone", "instance": "default", "area-id": "0.0.0.0"},
	})
	if len(areas) != 1 || areas[0].AreaID != "0.0.0.0" {
		t.Errorf("unexpected areas parse: %+v", areas)
	}

	tmpls := ParseOSPFTemplates([]map[string]string{
		{".id": "*1", "area": "backbone", "interfaces": "ether1", "cost": "10", "priority": "1"},
	})
	if len(tmpls) != 1 || tmpls[0].Cost != 10 {
		t.Errorf("unexpected tmpls parse: %+v", tmpls)
	}
}

func TestRoutingRuleCommandsAndParsers(t *testing.T) {
	ruleCmd := NewPrintRoutingRulesCommand()
	if ruleCmd.Raw != "/routing/rule/print" || ruleCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid rule print: %v", ruleCmd)
	}

	addCmd := NewAddRoutingRuleCommand(RoutingRule{
		Action:     "lookup",
		DstAddress: "192.168.0.0/16",
		Table:      "custom-table",
	})
	if addCmd.Raw != "/routing/rule/add" || addCmd.Args["table"] != "custom-table" {
		t.Errorf("unexpected add rule: %v", addCmd)
	}

	moveCmd := NewMoveRoutingRuleCommand("*2", "*1")
	if moveCmd.Raw != "/routing/rule/move" || moveCmd.Args["destination"] != "*1" {
		t.Errorf("unexpected move rule: %v", moveCmd)
	}

	rules := ParseRoutingRules([]map[string]string{
		{".id": "*1", "action": "lookup", "dst-address": "192.168.0.0/16", "table": "custom-table"},
	})
	if len(rules) != 1 || rules[0].Action != "lookup" {
		t.Errorf("unexpected rules parse: %+v", rules)
	}
}

func TestIPv6CommandsAndParsers(t *testing.T) {
	routeCmd := NewPrintIPv6RoutesCommand()
	if routeCmd.Raw != "/ipv6/route/print" || routeCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid route6 print: %v", routeCmd)
	}

	streamCmd := NewStreamIPv6RoutesCommand()
	if streamCmd.Raw != "/ipv6/route/print" || streamCmd.Args["follow"] != "" {
		t.Fatalf("invalid stream route6: %v", streamCmd)
	}

	routes := ParseIPv6Routes([]map[string]string{
		{".id": "*1", "dst-address": "2001:db8::/32", "gateway": "fe80::1%ether1", "distance": "1", "active": "true"},
	})
	if len(routes) != 1 || routes[0].Distance != 1 || !routes[0].Active {
		t.Errorf("unexpected route6 parse: %+v", routes)
	}

	addrs := ParseIPv6Addresses([]map[string]string{
		{".id": "*1", "address": "2001:db8::1/64", "interface": "ether1", "disabled": "false"},
	})
	if len(addrs) != 1 || addrs[0].Address != "2001:db8::1/64" {
		t.Errorf("unexpected addr6 parse: %+v", addrs)
	}

	set := ParseIPv6Settings([]map[string]string{
		{"disable-ipv6": "yes"},
	})
	if set == nil || !set.DisableIPv6 {
		t.Errorf("unexpected ipv6 settings parse: %+v", set)
	}
}
