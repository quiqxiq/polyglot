package firewall

import (
	"testing"
)

func TestIPv6FirewallRules(t *testing.T) {
	// Filter
	printFilter := NewPrintIPv6FilterRulesCommand("input")
	if printFilter.Raw != "/ipv6/firewall/filter/print" || printFilter.Args["?chain"] != "input" {
		t.Errorf("invalid print ipv6 filter: %v", printFilter)
	}

	addFilter := NewAddIPv6FilterRuleCommand(IPv6FirewallRule{
		Chain:       "input",
		Action:      "accept",
		Protocol:    "icmpv6",
		PlaceBefore: "*1",
		Comment:     "Accept ICMPv6",
	})
	if addFilter.Raw != "/ipv6/firewall/filter/add" || addFilter.Args["protocol"] != "icmpv6" {
		t.Errorf("invalid add ipv6 filter: %v", addFilter)
	}

	setFilter := NewSetIPv6FilterRuleCommand("*2", IPv6FirewallRule{
		Chain:    "input",
		Action:   "drop",
		Disabled: true,
	})
	if setFilter.Raw != "/ipv6/firewall/filter/set" || setFilter.Args[".id"] != "*2" || setFilter.Args["disabled"] != "yes" {
		t.Errorf("invalid set ipv6 filter: %v", setFilter)
	}

	moveFilter := NewMoveIPv6FilterRuleCommand("*2", "*1")
	if moveFilter.Raw != "/ipv6/firewall/filter/move" || moveFilter.Args[".id"] != "*2" || moveFilter.Args["destination"] != "*1" {
		t.Errorf("invalid move ipv6 filter: %v", moveFilter)
	}

	streamFilter := NewStreamIPv6FilterRulesCommand("forward")
	if streamFilter.Raw != "/ipv6/firewall/filter/print" || streamFilter.Args["follow"] != "" || streamFilter.Args["?chain"] != "forward" {
		t.Errorf("invalid stream ipv6 filter: %v", streamFilter)
	}

	remFilter := NewRemoveIPv6FilterRuleCommand("*2")
	if remFilter.Raw != "/ipv6/firewall/filter/remove" || remFilter.Args[".id"] != "*2" {
		t.Errorf("invalid rem ipv6 filter: %v", remFilter)
	}

	// NAT, Mangle, Raw
	printNAT := NewPrintIPv6NATRulesCommand("")
	if printNAT.Raw != "/ipv6/firewall/nat/print" {
		t.Errorf("invalid print nat: %v", printNAT)
	}
	addNAT := NewAddIPv6NATRuleCommand(IPv6FirewallRule{Chain: "srcnat", Action: "masquerade"})
	if addNAT.Raw != "/ipv6/firewall/nat/add" {
		t.Errorf("invalid add nat: %v", addNAT)
	}

	printMangle := NewPrintIPv6MangleRulesCommand("")
	if printMangle.Raw != "/ipv6/firewall/mangle/print" {
		t.Errorf("invalid print mangle: %v", printMangle)
	}
	addMangle := NewAddIPv6MangleRuleCommand(IPv6FirewallRule{Chain: "prerouting", Action: "mark-connection"})
	if addMangle.Raw != "/ipv6/firewall/mangle/add" {
		t.Errorf("invalid add mangle: %v", addMangle)
	}

	printRaw := NewPrintIPv6RawRulesCommand("")
	if printRaw.Raw != "/ipv6/firewall/raw/print" {
		t.Errorf("invalid print raw: %v", printRaw)
	}
	addRaw := NewAddIPv6RawRuleCommand(IPv6FirewallRule{Chain: "prerouting", Action: "drop"})
	if addRaw.Raw != "/ipv6/firewall/raw/add" {
		t.Errorf("invalid add raw: %v", addRaw)
	}

	// Parsing
	rules := ParseIPv6FirewallRules([]map[string]string{
		{".id": "*1", "chain": "input", "action": "accept", "protocol": "icmpv6", "bytes": "1024", "disabled": "false"},
	})
	if len(rules) != 1 || rules[0].Protocol != "icmpv6" || rules[0].Bytes != "1024" || rules[0].Disabled {
		t.Errorf("unexpected ipv6 rules parse: %+v", rules)
	}
}
