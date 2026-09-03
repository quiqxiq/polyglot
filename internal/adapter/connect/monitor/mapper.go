package monitor

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
)

func toProtoActiveSessions(sessions []port.HotspotActiveSession) []*devicepb.HotspotActiveSession {
	out := make([]*devicepb.HotspotActiveSession, len(sessions))
	for i, s := range sessions {
		out[i] = &devicepb.HotspotActiveSession{
			Id:         s.RosID,
			Server:     s.Server,
			User:       s.User,
			Address:    s.Address,
			MacAddress: s.MACAddress,
		}
	}
	return out
}

func toProtoActiveStats(stats []port.HotspotActiveStat) []*devicepb.HotspotActiveStat {
	out := make([]*devicepb.HotspotActiveStat, len(stats))
	for i, s := range stats {
		out[i] = &devicepb.HotspotActiveStat{
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
	return out
}

func toProtoHotspotUsers(users []port.HotspotUser) []*devicepb.HotspotUser {
	out := make([]*devicepb.HotspotUser, len(users))
	for i, u := range users {
		out[i] = &devicepb.HotspotUser{
			Id:          u.RosID,
			Name:        u.Name,
			Password:    u.Password,
			Profile:     u.Profile,
			LimitUptime: u.LimitUptime,
			LimitBytes:  u.LimitBytesIn,
			Uptime:      u.Uptime,
			BytesIn:     u.BytesIn,
			BytesOut:    u.BytesOut,
			Comment:     u.Comment,
			Disabled:    u.Disabled,
			Server:      u.Server,
		}
	}
	return out
}
