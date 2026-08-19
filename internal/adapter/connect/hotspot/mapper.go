package hotspot

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// ToProtoHotspotUsers converts mikrotik user list to proto list.
func ToProtoHotspotUsers(users []mikrotik.HotspotUser) []*devicepb.HotspotUser {
	pbUsers := make([]*devicepb.HotspotUser, len(users))
	for i, u := range users {
		pbUsers[i] = ToProtoHotspotUser(u)
	}
	return pbUsers
}

// ToProtoHotspotActiveSession converts a single active session to proto.
func ToProtoHotspotActiveSession(s port.HotspotActiveSession) *devicepb.HotspotActiveSession {
	return &devicepb.HotspotActiveSession{
		Id:         s.RosID,
		Server:     s.Server,
		User:       s.User,
		Address:    s.Address,
		MacAddress: s.MACAddress,
	}
}

// ToProtoActiveSessions converts mikrotik active sessions to proto list.
func ToProtoActiveSessions(sessions []mikrotik.HotspotActiveSession) []*devicepb.HotspotActiveSession {
	pbSessions := make([]*devicepb.HotspotActiveSession, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = ToProtoHotspotActiveSession(s)
	}
	return pbSessions
}

// ToProtoHotspotActiveStat converts a single active stat to proto.
func ToProtoHotspotActiveStat(s port.HotspotActiveStat) *devicepb.HotspotActiveStat {
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

// ToProtoActiveStats converts mikrotik active stats to proto list.
func ToProtoActiveStats(stats []mikrotik.HotspotActiveStat) []*devicepb.HotspotActiveStat {
	pbStats := make([]*devicepb.HotspotActiveStat, len(stats))
	for i, s := range stats {
		pbStats[i] = ToProtoHotspotActiveStat(s)
	}
	return pbStats
}

// ToProtoDHCPLeases converts mikrotik DHCP leases to proto list.
func ToProtoDHCPLeases(leases []mikrotik.DHCPLease) []*devicepb.DHCPLease {
	pbLeases := make([]*devicepb.DHCPLease, len(leases))
	for i, l := range leases {
		pbLeases[i] = &devicepb.DHCPLease{
			Id:         l.RosID,
			Address:    l.Address,
			MacAddress: l.MACAddress,
			HostName:   l.HostName,
			Status:     l.Status,
			Blocked:    l.Blocked,
			Comment:    l.Comment,
		}
	}
	return pbLeases
}
