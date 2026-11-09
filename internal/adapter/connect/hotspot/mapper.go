package hotspot

import (
	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
)

// toProtoHotspotUsers converts hotspot user list to proto list.
func toProtoHotspotUsers(users []port.HotspotUser) []*polyglotv1.HotspotUser {
	pbUsers := make([]*polyglotv1.HotspotUser, len(users))
	for i, u := range users {
		pbUsers[i] = toProtoHotspotUser(u)
	}
	return pbUsers
}

// toProtoHotspotActiveSession converts a single active session to proto.
func toProtoHotspotActiveSession(s port.HotspotActiveSession) *polyglotv1.HotspotActiveSession {
	return &polyglotv1.HotspotActiveSession{
		Id:         s.RosID,
		Server:     s.Server,
		User:       s.User,
		Address:    s.Address,
		MacAddress: s.MACAddress,
	}
}

// toProtoActiveSessions converts hotspot active sessions to proto list.
func toProtoActiveSessions(sessions []port.HotspotActiveSession) []*polyglotv1.HotspotActiveSession {
	pbSessions := make([]*polyglotv1.HotspotActiveSession, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = toProtoHotspotActiveSession(s)
	}
	return pbSessions
}

// toProtoDHCPLease converts a single port.DHCPLease to proto.
func toProtoDHCPLease(l port.DHCPLease) *polyglotv1.DHCPLease {
	return &polyglotv1.DHCPLease{
		Id:         l.RosID,
		Address:    l.Address,
		MacAddress: l.MACAddress,
		HostName:   l.HostName,
		Status:     l.Status,
		Blocked:    l.Blocked,
		Comment:    l.Comment,
	}
}

// toProtoDHCPLeases converts a slice of port.DHCPLease to proto slice.
func toProtoDHCPLeases(leases []port.DHCPLease) []*polyglotv1.DHCPLease {
	pbLeases := make([]*polyglotv1.DHCPLease, len(leases))
	for i, l := range leases {
		pbLeases[i] = toProtoDHCPLease(l)
	}
	return pbLeases
}
