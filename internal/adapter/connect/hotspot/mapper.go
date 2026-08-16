package hotspot

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
)

// ToProtoHotspotProfiles converts mikrotik profile list to proto list.
func ToProtoHotspotProfiles(profiles []mikrotik.HotspotUserProfile) []*devicepb.HotspotProfile {
	pbProfiles := make([]*devicepb.HotspotProfile, len(profiles))
	for i, p := range profiles {
		pbProfiles[i] = &devicepb.HotspotProfile{
			Id:          p.RosID,
			Name:        p.Name,
			SharedUsers: p.SharedUsers,
			RateLimit:   p.RateLimit,
			ModeExpire:  p.OnLogin,
			ParentQueue: p.ParentQueue,
			Comment:     p.Comment,
		}
	}
	return pbProfiles
}

// ToProtoHotspotUsers converts mikrotik user list to proto list.
func ToProtoHotspotUsers(users []mikrotik.HotspotUser) []*devicepb.HotspotUser {
	pbUsers := make([]*devicepb.HotspotUser, len(users))
	for i, u := range users {
		pbUsers[i] = &devicepb.HotspotUser{
			Id:          u.RosID,
			Name:        u.Name,
			Password:    u.Password,
			Profile:     u.Profile,
			LimitUptime: u.LimitUptime,
			LimitBytes:  u.LimitBytesIn,
			Comment:     u.Comment,
			Disabled:    u.Disabled,
		}
	}
	return pbUsers
}

// ToProtoActiveSessions converts mikrotik active sessions to proto list.
func ToProtoActiveSessions(sessions []mikrotik.HotspotActiveSession) []*devicepb.HotspotActiveSession {
	pbSessions := make([]*devicepb.HotspotActiveSession, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = &devicepb.HotspotActiveSession{
			Id:         s.RosID,
			Server:     s.Server,
			User:       s.User,
			Address:    s.Address,
			MacAddress: s.MACAddress,
			Uptime:     s.Uptime,
			BytesIn:    s.BytesIn,
			BytesOut:   s.BytesOut,
		}
	}
	return pbSessions
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
