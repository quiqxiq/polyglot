package dhcp

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestParseLeases(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":         "*1",
				"address":     "192.168.88.10",
				"mac-address": "00:11:22:33:44:55",
				"server":      "dhcp1",
				"status":      "bound",
				"blocked":     "true",
				"comment":     "SUSPENDED",
			},
			{
				// Missing .id should be skipped
				"address": "192.168.88.11",
			},
		},
	}

	leases := ParseLeases(res)
	if len(leases) != 1 {
		t.Fatalf("expected 1 lease, got %d", len(leases))
	}
	if leases[0].RosID != "*1" || !leases[0].Blocked || leases[0].MACAddress != "00:11:22:33:44:55" {
		t.Fatalf("unexpected lease parsed: %+v", leases[0])
	}
}

func TestParseServers(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":          "*A",
				"name":         "dhcp1",
				"interface":    "bridge1",
				"address-pool": "dhcp_pool1",
				"lease-time":   "10m",
				"disabled":     "false",
			},
		},
	}

	servers := ParseServers(res)
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0].Name != "dhcp1" || servers[0].Interface != "bridge1" || servers[0].Disabled {
		t.Fatalf("unexpected server parsed: %+v", servers[0])
	}
}
