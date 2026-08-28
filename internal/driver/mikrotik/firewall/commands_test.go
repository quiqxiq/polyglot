package firewall

import (
	"testing"
)

func TestFirewallCommands(t *testing.T) {
	cmdFilter := NewAddFilterCommand(FirewallFilterParams{
		Chain:      "forward",
		Action:     "drop",
		SrcAddress: "192.168.1.50",
		Comment:    "SUSPENDED",
		Disabled:   false,
	})
	if cmdFilter.Raw != "/ip/firewall/filter/add" || cmdFilter.Args["chain"] != "forward" || cmdFilter.Args["src-address"] != "192.168.1.50" {
		t.Fatalf("unexpected filter command: %+v", cmdFilter)
	}

	cmdList := NewAddAddressListCommand(AddressListParams{
		List:    "ISOLIR_USERS",
		Address: "10.0.0.2",
		Comment: "user1",
	})
	if cmdList.Raw != "/ip/firewall/address-list/add" || cmdList.Args["list"] != "ISOLIR_USERS" || cmdList.Args["address"] != "10.0.0.2" {
		t.Fatalf("unexpected address list command: %+v", cmdList)
	}

	cmdNAT := NewPrintNATCommand("dstnat", "ISOLATION_REDIRECT", "ISOLIR_USERS")
	if cmdNAT.Raw != "/ip/firewall/nat/print" || cmdNAT.Args["?chain"] != "dstnat" {
		t.Fatalf("unexpected print nat command: %+v", cmdNAT)
	}
}

