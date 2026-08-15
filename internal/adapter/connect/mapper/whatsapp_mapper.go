package mapper

import (
	"fmt"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

// WASessionToProto maps a domain bot.WASession to Protobuf WASession.
func WASessionToProto(s bot.WASession) *devicepb.WASession {
	return &devicepb.WASession{
		Id:          fmt.Sprintf("%d", s.ID),
		Name:        s.DeviceName,
		PhoneNumber: s.PhoneNumber,
		Status:      string(s.Status),
		IsBotActive: s.IsBotEnabled,
		CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// WASessionListToProto maps a slice of bot.WASession to Protobuf WASession slice.
func WASessionListToProto(sessions []bot.WASession) []*devicepb.WASession {
	res := make([]*devicepb.WASession, len(sessions))
	for i, s := range sessions {
		res[i] = WASessionToProto(s)
	}
	return res
}
