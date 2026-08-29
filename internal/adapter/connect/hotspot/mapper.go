package hotspot

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
)

// toProtoHotspotUsers converts hotspot user list to proto list.
func toProtoHotspotUsers(users []port.HotspotUser) []*devicepb.HotspotUser {
	pbUsers := make([]*devicepb.HotspotUser, len(users))
	for i, u := range users {
		pbUsers[i] = toProtoHotspotUser(u)
	}
	return pbUsers
}

// toProtoHotspotActiveSession converts a single active session to proto.
func toProtoHotspotActiveSession(s port.HotspotActiveSession) *devicepb.HotspotActiveSession {
	return &devicepb.HotspotActiveSession{
		Id:         s.RosID,
		Server:     s.Server,
		User:       s.User,
		Address:    s.Address,
		MacAddress: s.MACAddress,
	}
}

// toProtoActiveSessions converts hotspot active sessions to proto list.
func toProtoActiveSessions(sessions []port.HotspotActiveSession) []*devicepb.HotspotActiveSession {
	pbSessions := make([]*devicepb.HotspotActiveSession, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = toProtoHotspotActiveSession(s)
	}
	return pbSessions
}

// toProtoHotspotActiveStat converts a single active stat to proto.
func toProtoHotspotActiveStat(s port.HotspotActiveStat) *devicepb.HotspotActiveStat {
	return &devicepb.HotspotActiveStat{
		Id:              s.RosID,
		Uptime:          s.Uptime,
		SessionTimeLeft: s.SessionTimeLeft,
		IdleTime:        s.IdleTime,
		BytesIn:         s.BytesIn,
		BytesOut:        s.BytesOut,
		PacketsIn:       s.PacketsIn,
		PacketsOut:      s.PacketsOut,
	}
}

// toProtoActiveStats converts hotspot active stats to proto list.
func toProtoActiveStats(stats []port.HotspotActiveStat) []*devicepb.HotspotActiveStat {
	pbStats := make([]*devicepb.HotspotActiveStat, len(stats))
	for i, s := range stats {
		pbStats[i] = toProtoHotspotActiveStat(s)
	}
	return pbStats
}

// toProtoDHCPLease converts a single port.DHCPLease to proto.
func toProtoDHCPLease(l port.DHCPLease) *devicepb.DHCPLease {
	return &devicepb.DHCPLease{
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
func toProtoDHCPLeases(leases []port.DHCPLease) []*devicepb.DHCPLease {
	pbLeases := make([]*devicepb.DHCPLease, len(leases))
	for i, l := range leases {
		pbLeases[i] = toProtoDHCPLease(l)
	}
	return pbLeases
}
