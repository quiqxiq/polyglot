package vpn

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseWireGuardInterfaces converts command.Result from /interface/wireguard/print into []WireGuardInterface.
func ParseWireGuardInterfaces(result command.Result) []WireGuardInterface {
	ifaces := make([]WireGuardInterface, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		ifaces = append(ifaces, WireGuardInterface{
			RosID:      id,
			Name:       name,
			ListenPort: row["listen-port"],
			PublicKey:  row["public-key"],
			PrivateKey: row["private-key"],
			MTU:        row["mtu"],
			Running:    strings.EqualFold(row["running"], "true") || strings.EqualFold(row["running"], "yes"),
			Disabled:   strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Comment:    row["comment"],
		})
	}
	return ifaces
}

// ParseWireGuardPeers converts command.Result from /interface/wireguard/peers/print into []WireGuardPeer.
func ParseWireGuardPeers(result command.Result) []WireGuardPeer {
	peers := make([]WireGuardPeer, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		peers = append(peers, WireGuardPeer{
			RosID:           id,
			Interface:       row["interface"],
			PublicKey:       row["public-key"],
			EndpointAddress: row["endpoint-address"],
			EndpointPort:    row["endpoint-port"],
			AllowedAddress:  row["allowed-address"],
			LatestHandshake: row["latest-handshake"],
			TransferRx:      row["rx"],
			TransferTx:      row["tx"],
			Disabled:        strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Comment:         row["comment"],
		})
	}
	return peers
}

// ParseIPsecPeers converts command.Result from /ip/ipsec/peer/print into []IPsecPeer.
func ParseIPsecPeers(result command.Result) []IPsecPeer {
	peers := make([]IPsecPeer, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		peers = append(peers, IPsecPeer{
			RosID:        id,
			Name:         name,
			Address:      row["address"],
			Profile:      row["profile"],
			ExchangeMode: row["exchange-mode"],
			Disabled:     strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Comment:      row["comment"],
		})
	}
	return peers
}
