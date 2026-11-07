package monitor

import (
	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
)

func toProtoActiveSessions(sessions []port.HotspotActiveSession) []*polyglotv1.HotspotActiveSession {
	out := make([]*polyglotv1.HotspotActiveSession, len(sessions))
	for i, s := range sessions {
		out[i] = &polyglotv1.HotspotActiveSession{
			Id:         s.RosID,
			Server:     s.Server,
			User:       s.User,
			Address:    s.Address,
			MacAddress: s.MACAddress,
		}
	}
	return out
}

func toProtoActiveStats(stats []port.HotspotActiveStat) []*polyglotv1.HotspotActiveStat {
	out := make([]*polyglotv1.HotspotActiveStat, len(stats))
	for i, s := range stats {
		out[i] = &polyglotv1.HotspotActiveStat{
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

func toProtoHotspotUsers(users []port.HotspotUser) []*polyglotv1.HotspotUser {
	out := make([]*polyglotv1.HotspotUser, len(users))
	for i, u := range users {
		out[i] = &polyglotv1.HotspotUser{
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
