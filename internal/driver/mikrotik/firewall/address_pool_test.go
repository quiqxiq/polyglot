package firewall

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestAddressPoolCommandsAndParsers(t *testing.T) {
	// IP Address commands
	addCmd := NewAddIPAddressCommand(IPAddressParams{
		Address:   "192.168.88.1/24",
		Interface: "bridge",
		Comment:   "LAN Gateway",
	})
	if addCmd.Raw != "/ip/address/add" || addCmd.Args["address"] != "192.168.88.1/24" {
		t.Errorf("invalid add address cmd: %v", addCmd)
	}

	setCmd := NewSetIPAddressCommand("*1", IPAddressParams{
		Address:   "192.168.88.254/24",
		Interface: "bridge",
		Disabled:  true,
	})
	if setCmd.Raw != "/ip/address/set" || setCmd.Args[".id"] != "*1" || setCmd.Args["disabled"] != "yes" {
		t.Errorf("invalid set address cmd: %v", setCmd)
	}

	remCmd := NewRemoveIPAddressCommand("*1")
	if remCmd.Raw != "/ip/address/remove" || remCmd.Args[".id"] != "*1" {
		t.Errorf("invalid rem address cmd: %v", remCmd)
	}

	streamCmd := NewStreamIPAddressesCommand(IPAddressPrintParams{
		Interface: "bridge",
	})
	if streamCmd.Raw != "/ip/address/print" || streamCmd.Args["?interface"] != "bridge" || streamCmd.Args[".proplist"] == "" {
		t.Errorf("invalid stream address cmd: %v", streamCmd)
	}

	// Address parsing
	addrs := ParseIPAddresses(command.Result{
		Rows: []map[string]string{
			{".id": "*1", "address": "192.168.88.1/24", "network": "192.168.88.0", "interface": "bridge", "disabled": "false"},
		},
	})
	if len(addrs) != 1 || addrs[0].Address != "192.168.88.1/24" {
		t.Errorf("unexpected addresses parse: %+v", addrs)
	}

	// Pool commands
	addPool := NewAddPoolCommand(IPPoolParams{
		Name:     "dhcp-pool",
		Ranges:   "192.168.88.10-192.168.88.100",
		NextPool: "none",
	})
	if addPool.Raw != "/ip/pool/add" || addPool.Args["ranges"] != "192.168.88.10-192.168.88.100" {
		t.Errorf("invalid add pool cmd: %v", addPool)
	}

	setPool := NewSetPoolCommand("*2", IPPoolParams{
		Ranges: "192.168.88.50-192.168.88.200",
	})
	if setPool.Raw != "/ip/pool/set" || setPool.Args[".id"] != "*2" {
		t.Errorf("invalid set pool cmd: %v", setPool)
	}

	remPool := NewRemovePoolCommand("*2")
	if remPool.Raw != "/ip/pool/remove" || remPool.Args[".id"] != "*2" {
		t.Errorf("invalid rem pool cmd: %v", remPool)
	}

	pools := ParsePools(command.Result{
		Rows: []map[string]string{
			{".id": "*2", "name": "dhcp-pool", "ranges": "192.168.88.10-192.168.88.100"},
		},
	})
	if len(pools) != 1 || pools[0].Ranges != "192.168.88.10-192.168.88.100" {
		t.Errorf("unexpected pools parse: %+v", pools)
	}
}
