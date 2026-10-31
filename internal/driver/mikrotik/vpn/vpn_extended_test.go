package vpn

import (
	"testing"
)

func TestIPsecCommandsAndParsers(t *testing.T) {
	identCmd := NewPrintIPsecIdentitiesCommand()
	if identCmd.Raw != "/ip/ipsec/identity/print" || identCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid ident print: %v", identCmd)
	}

	policyCmd := NewPrintIPsecPoliciesCommand()
	if policyCmd.Raw != "/ip/ipsec/policy/print" {
		t.Fatalf("invalid policy print: %v", policyCmd)
	}

	moveCmd := NewMoveIPsecPolicyCommand("*3", "*1")
	if moveCmd.Raw != "/ip/ipsec/policy/move" || moveCmd.Args["destination"] != "*1" {
		t.Errorf("unexpected move policy: %v", moveCmd)
	}

	saCmd := NewPrintIPsecInstalledSAsCommand()
	if saCmd.Raw != "/ip/ipsec/installed-sa/print" {
		t.Fatalf("invalid sa print: %v", saCmd)
	}

	policies := ParseIPsecPolicies([]map[string]string{
		{".id": "*1", "peer": "peer1", "tunnel": "yes", "src-address": "192.168.1.0/24", "dst-address": "10.0.0.0/24", "action": "encrypt"},
	})
	if len(policies) != 1 || !policies[0].Tunnel || policies[0].Action != "encrypt" {
		t.Errorf("unexpected policies parse: %+v", policies)
	}

	sas := ParseIPsecInstalledSAs([]map[string]string{
		{".id": "*1", "src-address": "1.2.3.4", "dst-address": "5.6.7.8", "spi": "0x123456", "state": "mature"},
	})
	if len(sas) != 1 || sas[0].SPI != "0x123456" {
		t.Errorf("unexpected sa parse: %+v", sas)
	}
}

func TestOVPNCommandsAndParsers(t *testing.T) {
	serverCmd := NewPrintOVPNServerCommand()
	if serverCmd.Raw != "/interface/ovpn-server/server/print" || serverCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid ovpn server print: %v", serverCmd)
	}

	clientCmd := NewPrintOVPNClientCommand()
	if clientCmd.Raw != "/interface/ovpn-client/print" {
		t.Fatalf("invalid ovpn client print: %v", clientCmd)
	}

	srv := ParseOVPNServer([]map[string]string{
		{"enabled": "yes", "port": "1194", "mode": "ip", "protocol": "udp", "auth": "sha1"},
	})
	if srv == nil || !srv.Enabled || srv.Port != 1194 {
		t.Errorf("unexpected ovpn server parse: %+v", srv)
	}

	clients := ParseOVPNClients([]map[string]string{
		{".id": "*1", "name": "ovpn-out1", "connect-to": "vpn.example.com", "port": "1194", "user": "test"},
	})
	if len(clients) != 1 || clients[0].ConnectTo != "vpn.example.com" {
		t.Errorf("unexpected ovpn client parse: %+v", clients)
	}
}

func TestPPPoECommandsAndParsers(t *testing.T) {
	srvCmd := NewPrintPPPoEServerCommand()
	if srvCmd.Raw != "/interface/pppoe-server/server/print" {
		t.Fatalf("invalid pppoe server print: %v", srvCmd)
	}

	clientCmd := NewPrintPPPoEClientCommand()
	if clientCmd.Raw != "/interface/pppoe-client/print" {
		t.Fatalf("invalid pppoe client print: %v", clientCmd)
	}

	servers := ParsePPPoEServers([]map[string]string{
		{".id": "*1", "service-name": "pppoe-lan", "interface": "bridge", "one-session-per-host": "yes"},
	})
	if len(servers) != 1 || !servers[0].OneSessionPerHost || servers[0].ServiceName != "pppoe-lan" {
		t.Errorf("unexpected pppoe servers parse: %+v", servers)
	}

	clients := ParsePPPoEClients([]map[string]string{
		{".id": "*1", "name": "pppoe-out1", "interface": "ether1", "user": "dsl01", "add-default-route": "yes"},
	})
	if len(clients) != 1 || !clients[0].AddDefaultRoute || clients[0].User != "dsl01" {
		t.Errorf("unexpected pppoe clients parse: %+v", clients)
	}
}

func TestWireGuardClientConfig(t *testing.T) {
	showCmd := NewShowWireGuardClientConfigCommand("*1", true)
	if showCmd.Raw != "/interface/wireguard/peers/show-client-config" || showCmd.Args[".id"] != "*1" {
		t.Fatalf("invalid show wg config cmd: %v", showCmd)
	}

	conf := ParseWireGuardClientConfig([]map[string]string{
		{"ret": "[Interface]\nPrivateKey = ..."},
	})
	if conf != "[Interface]\nPrivateKey = ..." {
		t.Errorf("unexpected wg conf parse: %s", conf)
	}
}

func TestIPsecExtendedFeatures(t *testing.T) {
	// Peer mutations
	addPeer := NewAddIPsecPeerCommand(IPsecPeerParams{
		Name:    "peer-branch",
		Address: "198.51.100.2",
		Profile: "default",
	})
	if addPeer.Raw != "/ip/ipsec/peer/add" || addPeer.Args["address"] != "198.51.100.2" {
		t.Errorf("invalid add ipsec peer: %v", addPeer)
	}

	setPeer := NewSetIPsecPeerCommand("*1", IPsecPeerParams{
		Address: "198.51.100.3",
	})
	if setPeer.Raw != "/ip/ipsec/peer/set" || setPeer.Args[".id"] != "*1" {
		t.Errorf("invalid set ipsec peer: %v", setPeer)
	}

	remPeer := NewRemoveIPsecPeerCommand("*1")
	if remPeer.Raw != "/ip/ipsec/peer/remove" || remPeer.Args[".id"] != "*1" {
		t.Errorf("invalid rem ipsec peer: %v", remPeer)
	}

	// Profile commands & parsers
	profPrint := NewPrintIPsecProfilesCommand()
	if profPrint.Raw != "/ip/ipsec/profile/print" {
		t.Errorf("invalid prof print: %v", profPrint)
	}

	addProf := NewAddIPsecProfileCommand(IPsecProfile{
		Name:          "ike2-prof",
		HashAlgorithm: "sha256",
		EncAlgorithm:  "aes-256",
		DHGroup:       "modp2048",
		NatTraversal:  true,
	})
	if addProf.Raw != "/ip/ipsec/profile/add" || addProf.Args["nat-traversal"] != "yes" {
		t.Errorf("invalid add prof: %v", addProf)
	}

	profs := ParseIPsecProfiles([]map[string]string{
		{".id": "*1", "name": "ike2-prof", "hash-algorithm": "sha256", "nat-traversal": "yes"},
	})
	if len(profs) != 1 || !profs[0].NatTraversal || profs[0].HashAlgorithm != "sha256" {
		t.Errorf("unexpected profs parse: %+v", profs)
	}

	// Proposal commands & parsers
	addProp := NewAddIPsecProposalCommand(IPsecProposal{
		Name:           "prop1",
		AuthAlgorithms: "sha256",
		EncAlgorithms:  "aes-256-gcm",
		PFSGroup:       "modp2048",
	})
	if addProp.Raw != "/ip/ipsec/proposal/add" || addProp.Args["enc-algorithms"] != "aes-256-gcm" {
		t.Errorf("invalid add prop: %v", addProp)
	}

	props := ParseIPsecProposals([]map[string]string{
		{".id": "*1", "name": "prop1", "enc-algorithms": "aes-256-gcm", "pfs-group": "modp2048"},
	})
	if len(props) != 1 || props[0].PFSGroup != "modp2048" {
		t.Errorf("unexpected props parse: %+v", props)
	}

	// Mode config commands & parsers
	addMC := NewAddIPsecModeConfigCommand(IPsecModeConfig{
		Name:        "cfg1",
		Responder:   true,
		AddressPool: "vpn-pool",
		SystemDNS:   true,
	})
	if addMC.Raw != "/ip/ipsec/mode-config/add" || addMC.Args["responder"] != "yes" || addMC.Args["system-dns"] != "yes" {
		t.Errorf("invalid add mode-config: %v", addMC)
	}

	mcs := ParseIPsecModeConfigs([]map[string]string{
		{".id": "*1", "name": "cfg1", "responder": "yes", "system-dns": "yes"},
	})
	if len(mcs) != 1 || !mcs[0].Responder || !mcs[0].SystemDNS {
		t.Errorf("unexpected mode-configs parse: %+v", mcs)
	}
}
