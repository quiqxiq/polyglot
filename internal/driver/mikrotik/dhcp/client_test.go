package dhcp

import (
	"testing"
)

func TestDHCPClientCommandsAndParsers(t *testing.T) {
	printCmd := NewPrintDHCPClientsCommand()
	if printCmd.Raw != "/ip/dhcp-client/print" || printCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid print cmd: %v", printCmd)
	}

	addCmd := NewAddDHCPClientCommand("ether1", "WAN client", true)
	if addCmd.Raw != "/ip/dhcp-client/add" || addCmd.Args["add-default-route"] != "yes" {
		t.Errorf("unexpected add cmd: %v", addCmd)
	}

	renewCmd := NewRenewDHCPClientCommand("*1")
	if renewCmd.Raw != "/ip/dhcp-client/renew" || renewCmd.Args["numbers"] != "*1" {
		t.Errorf("unexpected renew cmd: %v", renewCmd)
	}

	clients := ParseDHCPClients([]map[string]string{
		{
			".id":       "*1",
			"interface": "ether1",
			"status":    "bound",
			"address":   "192.168.100.20/24",
			"gateway":   "192.168.100.1",
			"disabled":  "false",
		},
	})
	if len(clients) != 1 || clients[0].Status != "bound" || clients[0].Address != "192.168.100.20/24" {
		t.Errorf("unexpected parse result: %+v", clients)
	}
}
