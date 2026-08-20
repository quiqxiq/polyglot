package hotspot

import (
	"testing"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHotspotUserParamsFromProto(t *testing.T) {
	p := HotspotUserParamsFromProto(&devicepb.CreateHotspotUserRequest{
		Name:       "user1",
		Password:   "pass1",
		Profile:    "1Day_10K",
		Server:     "hotspot1",
		MacAddress: "00:11:22:33:44:55",
		TimeLimit:  "1d",
		DataLimit:  "1000M",
		Comment:    "vc-A3X-08.17.26",
	}, "vc-A3X-08.17.26")

	assert.Equal(t, "user1", p.Name)
	assert.Equal(t, "hotspot1", p.Server)
	assert.Equal(t, "1d", p.LimitUptime)
	// 1000M = 1000 * 1024 * 1024 bytes, dikonversi ke string byte untuk RouterOS.
	assert.Equal(t, "1048576000", p.LimitBytesOut)
	assert.Equal(t, "vc-A3X-08.17.26", p.Comment)
}

func TestHotspotUserParamsUpdateFromProto(t *testing.T) {
	p := HotspotUserParamsUpdateFromProto(&devicepb.UpdateHotspotUserRequest{
		RosId:      "*1",
		Password:   "newpass",
		Profile:    "7Day_25K",
		TimeLimit:  "7d",
		DataLimit:  "500",
		Comment:    "03/08/2026 catatan",
		ExpireDate: "03/08/2026",
	}, "03/08/2026 catatan")

	assert.Equal(t, "03/08/2026 catatan", p.Comment)
	assert.Equal(t, "500", p.LimitBytesOut) // tanpa unit -> byte polos
}

func TestProfileParamsFromProto(t *testing.T) {
	p := ProfileParamsFromProto(&devicepb.HotspotProfileParams{
		Name:            "1Day_10K",
		AddressPool:     "hs-pool",
		SharedUsers:     "2",
		RateLimit:       "5M/5M",
		ParentQueue:     "parent-queue",
		Price:           "10000",
		SellingPrice:    "8000",
		Validity:        "1d",
		ExpireMode:      "ntfc",
		LockUser:        true,
		LockServer:      false,
		EnableRecording: true,
		Comment:         "catatan",
	})

	assert.Equal(t, port.MikhmonProfileParams{
		Name:            "1Day_10K",
		AddressPool:     "hs-pool",
		SharedUsers:     "2",
		RateLimit:       "5M/5M",
		ParentQueue:     "parent-queue",
		Price:           "10000",
		SellingPrice:    "8000",
		Validity:        "1d",
		ExpireMode:      "ntfc",
		LockUser:        true,
		LockServer:      false,
		EnableRecording: true,
		Comment:         "catatan",
	}, p)
}

func TestToProtoHotspotHosts(t *testing.T) {
	hosts := ToProtoHotspotHosts([]map[string]string{
		{".id": "*1", "mac-address": "00:11:22:33:44:55", "address": "10.0.0.2",
			"to-address": "10.0.0.2", "server": "hotspot1", "bypassed": "true",
			"authorized": "false", "comment": "laptop"},
		{"mac-address": "aa:bb", "address": "10.0.0.3"}, // tanpa .id -> skip
	})
	require.Len(t, hosts, 1)
	assert.Equal(t, "*1", hosts[0].Id)
	assert.True(t, hosts[0].Bypassed)
	assert.False(t, hosts[0].Authorized)
	assert.Equal(t, "laptop", hosts[0].Comment)
}

func TestToProtoHotspotServers(t *testing.T) {
	servers := ToProtoHotspotServers([]map[string]string{
		{".id": "*2", "name": "hotspot1", "interface": "bridge1",
			"address-pool": "hs-pool", "disabled": "false", "comment": "utama"},
		{"name": "tanpa-id"}, // tanpa .id -> skip
	})
	require.Len(t, servers, 1)
	assert.Equal(t, "hotspot1", servers[0].Name)
	assert.Equal(t, "bridge1", servers[0].Interface)
	assert.False(t, servers[0].Disabled)
}

func TestToProtoHotspotProfile(t *testing.T) {
	p := ToProtoHotspotProfile(mikrotik.HotspotUserProfile{
		RosID:       "*1",
		Name:        "1Day_10K",
		SharedUsers: "1",
		RateLimit:   "5M/5M",
		ParentQueue: "parent-queue",
		AddressPool: "hs-pool",
		Comment:     "catatan",
		OnLogin:     `:put (",ntfc,10000,1d,8000,,Enable,Enable,");`,
	})
	assert.Equal(t, "ntfc", p.ModeExpire)
	assert.Equal(t, 10000.0, p.Price)
	assert.Equal(t, "1d", p.Validity)
	assert.Equal(t, 8000.0, p.SellingPrice)
	assert.Equal(t, "Enable", p.LockUser)
	assert.Equal(t, "Enable", p.LockServer)
	assert.Equal(t, "hs-pool", p.AddressPool)
}

func TestSanitizeBatchTag(t *testing.T) {
	assert.Equal(t, "Paket1Hari", sanitizeBatchTag("Paket 1 Hari"))
	assert.Equal(t, "A-B", sanitizeBatchTag("A - B"))
}
