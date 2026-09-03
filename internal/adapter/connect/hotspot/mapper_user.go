package hotspot

import (
	"fmt"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
	"github.com/quixiq/polyglot/internal/port"
)

// toProtoHotspotUser converts a single mikrotik user to proto.
func toProtoHotspotUser(u port.HotspotUser) *devicepb.HotspotUser {
	limitBytes := u.LimitBytesOut
	if limitBytes == "" {
		limitBytes = u.LimitBytesIn
	}
	return &devicepb.HotspotUser{
		Id:          u.RosID,
		Name:        u.Name,
		Password:    u.Password,
		Profile:     u.Profile,
		Server:      u.Server,
		LimitUptime: u.LimitUptime,
		LimitBytes:  limitBytes,
		Comment:     u.Comment,
		Disabled:    u.Disabled,
		Uptime:      u.Uptime,
		BytesIn:     u.BytesIn,
		BytesOut:    u.BytesOut,
	}
}

// HotspotUserParamsFromProto maps a CreateHotspotUserRequest into
// port.HotspotUserParams. comment must already be resolved by the handler
// (auto vc-/up- prefix via hotspot.BuildCreateUserComment). data_limit
// ("1000M") is converted to a RouterOS byte string via hotspot.ParseDataLimit.
func HotspotUserParamsFromProto(req *devicepb.CreateHotspotUserRequest, comment string) port.HotspotUserParams {
	return port.HotspotUserParams{
		Name:          req.Name,
		Password:      req.Password,
		Profile:       req.Profile,
		Server:        req.Server,
		MACAddress:    req.MacAddress,
		LimitUptime:   req.TimeLimit,
		LimitBytesOut: dataLimitBytes(req.DataLimit),
		Comment:       comment,
	}
}

// HotspotUserParamsUpdateFromProto maps an UpdateHotspotUserRequest into
// port.HotspotUserParams. comment must already be resolved by the handler
// (legacy rebuild rules via hotspot.BuildUpdatedComment).
func HotspotUserParamsUpdateFromProto(req *devicepb.UpdateHotspotUserRequest, comment string) port.HotspotUserParams {
	return port.HotspotUserParams{
		Name:          req.Name,
		Password:      req.Password,
		Profile:       req.Profile,
		Server:        req.Server,
		MACAddress:    req.MacAddress,
		LimitUptime:   req.TimeLimit,
		LimitBytesOut: dataLimitBytes(req.DataLimit),
		Comment:       comment,
	}
}

// dataLimitBytes converts a human-readable data limit ("1000M", "1g") to a
// RouterOS byte string ("1048576000"); empty when unparseable or zero.
func dataLimitBytes(s string) string {
	if b := domainHotspot.ParseDataLimit(s); b > 0 {
		return fmt.Sprintf("%d", b)
	}
	return ""
}
