package vpn

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestVPNCommands(t *testing.T) {
	cmdWG := NewPrintWireGuardInterfacesCommand("wg0")
	if cmdWG.Raw != "/interface/wireguard/print" || cmdWG.Args["?name"] != "wg0" || cmdWG.Args[".proplist"] == "" {
		t.Fatalf("unexpected wg iface command: %+v", cmdWG)
	}

	cmdPeers := NewPrintWireGuardPeersCommand("wg0")
	if cmdPeers.Raw != "/interface/wireguard/peers/print" || cmdPeers.Args["?interface"] != "wg0" || cmdPeers.Args[".proplist"] == "" {
		t.Fatalf("unexpected wg peers command: %+v", cmdPeers)
	}

	cmdIPsec := NewPrintIPsecPeersCommand()
	if cmdIPsec.Raw != "/ip/ipsec/peer/print" || cmdIPsec.Args[".proplist"] == "" {
		t.Fatalf("unexpected ipsec peers command: %+v", cmdIPsec)
	}
}

func TestParseWireGuardInterfaces(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":         "*1",
				"name":        "wg0",
				"listen-port": "13231",
				"public-key":  "abcdef12345=",
				"mtu":         "1420",
				"running":     "true",
				"disabled":    "false",
			},
		},
	}
	ifaces := ParseWireGuardInterfaces(res)
	if len(ifaces) != 1 || ifaces[0].Name != "wg0" || ifaces[0].ListenPort != "13231" || !ifaces[0].Running {
		t.Fatalf("unexpected wg ifaces parsed: %+v", ifaces)
	}
}

func TestParseWireGuardPeers(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":              "*10",
				"interface":        "wg0",
				"public-key":       "peerkey123=",
				"endpoint-address": "198.51.100.1",
				"endpoint-port":    "51820",
				"allowed-address":  "10.10.0.2/32",
				"latest-handshake": "1m20s",
				"rx":               "1048576",
				"tx":               "2097152",
				"disabled":         "no",
			},
		},
	}
	peers := ParseWireGuardPeers(res)
	if len(peers) != 1 || peers[0].Interface != "wg0" || peers[0].TransferRx != "1048576" {
		t.Fatalf("unexpected wg peers parsed: %+v", peers)
	}
}
