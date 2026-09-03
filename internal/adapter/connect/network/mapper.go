package network

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
)

// ToProtoDHCPLeases maps a slice of domain DHCPLeases to protobuf DHCPLease messages.
func ToProtoDHCPLeases(leases []port.DHCPLease) []*devicepb.DHCPLease {
	out := make([]*devicepb.DHCPLease, len(leases))
	for i, l := range leases {
		out[i] = &devicepb.DHCPLease{
			Id:         l.RosID,
			Address:    l.Address,
			MacAddress: l.MACAddress,
			HostName:   l.HostName,
			Status:     l.Status,
			Blocked:    l.Blocked,
			Comment:    l.Comment,
		}
	}
	return out
}
