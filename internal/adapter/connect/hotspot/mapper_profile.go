package hotspot

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/port"
)

// ProfileParamsFromProto maps a HotspotProfileParams proto into
// port.MikhmonProfileParams ready for the on-login script builder.
func ProfileParamsFromProto(p *devicepb.HotspotProfileParams) port.MikhmonProfileParams {
	return port.MikhmonProfileParams{
		Name:            p.Name,
		AddressPool:     p.AddressPool,
		SharedUsers:     p.SharedUsers,
		RateLimit:       p.RateLimit,
		ParentQueue:     p.ParentQueue,
		Price:           p.Price,
		SellingPrice:    p.SellingPrice,
		Validity:        p.Validity,
		ExpireMode:      port.ExpireMode(p.ExpireMode),
		LockUser:        p.LockUser,
		LockServer:      p.LockServer,
		EnableRecording: p.EnableRecording,
		Comment:         p.Comment,
	}
}

// ToProtoHotspotProfile converts a single mikrotik profile to proto, parsing
// the Mikhmon metadata out of the on-login script (mode/price/validity/locks).
func ToProtoHotspotProfile(p mikrotik.HotspotUserProfile) *devicepb.HotspotProfile {
	meta, _ := hotspot.ParseOnLoginScript(p.OnLogin)
	return &devicepb.HotspotProfile{
		Id:           p.RosID,
		Name:         p.Name,
		SharedUsers:  p.SharedUsers,
		RateLimit:    p.RateLimit,
		ModeExpire:   meta.ExpireMode,
		Validity:     meta.Validity,
		Price:        meta.Price,
		SellingPrice: meta.SellingPrice,
		LockUser:     meta.LockUser,
		LockServer:   meta.LockServer,
		ParentQueue:  p.ParentQueue,
		AddressPool:  p.AddressPool,
		Comment:      p.Comment,
	}
}

// ToProtoHotspotProfiles converts mikrotik profile list to proto list.
func ToProtoHotspotProfiles(profiles []mikrotik.HotspotUserProfile) []*devicepb.HotspotProfile {
	pbProfiles := make([]*devicepb.HotspotProfile, len(profiles))
	for i, p := range profiles {
		pbProfiles[i] = ToProtoHotspotProfile(p)
	}
	return pbProfiles
}
