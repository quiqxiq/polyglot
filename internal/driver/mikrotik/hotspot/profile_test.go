package hotspot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOnLoginScript(t *testing.T) {
	t.Run("notify mode dengan mac lock dan recording", func(t *testing.T) {
		script := BuildOnLoginScript(MikhmonProfileParams{
			Name:            "1Day_10K",
			Price:           "10000",
			SellingPrice:    "8000",
			Validity:        "1d",
			ExpireMode:      ExpireModeNotifyRecord,
			LockUser:        true,
			LockServer:      true,
			EnableRecording: true,
		})

		// Verifikasi CSV metadata header
		assert.Contains(t, script, ",ntfc,10000,1d,8000,,Enable,Enable,")
		// Verifikasi mode 'N'
		assert.Contains(t, script, `:local mode "N";`)
		// Verifikasi scheduler addition
		assert.Contains(t, script, `/sys sch add name="$user" disable=no start-date=$date interval="1d";`)
		// Verifikasi MAC locking & Server locking
		assert.Contains(t, script, `/ip hotspot user set mac-address=$mac [find where name=$user]`)
		assert.Contains(t, script, `/ip hotspot user set server=$srv [find where name=$user]`)
		// Verifikasi transaction recording
		assert.Contains(t, script, `/system script add name="$date-|-$time-|-$user-|-10000-|-$address-|-$mac-|-1d-|-1Day_10K-|-$comment"`)
	})

	t.Run("no exp mode", func(t *testing.T) {
		script := BuildOnLoginScript(MikhmonProfileParams{
			Name:       "NoExp_50K",
			Price:      "50000",
			ExpireMode: ExpireModeNone,
			LockUser:   true,
		})
		assert.Contains(t, script, ",,50000,,0,noexp,Enable,Disable,")
		assert.Contains(t, script, `/ip hotspot user set mac-address=$mac`)
	})
}

func TestParseOnLoginScript(t *testing.T) {
	t.Run("format standar (remc)", func(t *testing.T) {
		script := `:put (",remc,3000,1h,3500,,Enable,Disable,");`
		meta, err := ParseOnLoginScript(script)
		require.NoError(t, err)
		assert.Equal(t, "remc", meta.ExpireMode)
		assert.Equal(t, 3000.0, meta.Price)
		assert.Equal(t, "1h", meta.Validity)
		assert.Equal(t, 3500.0, meta.SellingPrice)
		assert.Equal(t, "Enable", meta.LockUser)
		assert.Equal(t, "Disable", meta.LockServer)
	})

	t.Run("format noexp", func(t *testing.T) {
		script := `:put (",,50000,,0,noexp,Enable,Disable,");`
		meta, err := ParseOnLoginScript(script)
		require.NoError(t, err)
		assert.Equal(t, "0", meta.ExpireMode)
		assert.Equal(t, 50000.0, meta.Price)
		assert.Equal(t, "", meta.Validity)
		assert.Equal(t, 0.0, meta.SellingPrice)
		assert.Equal(t, "Enable", meta.LockUser)
	})

	t.Run("script kosong -> metadata kosong tanpa error", func(t *testing.T) {
		meta, err := ParseOnLoginScript("")
		require.NoError(t, err)
		assert.Equal(t, ProfileMeta{}, meta)
	})

	t.Run("bukan payload mikhmon -> error", func(t *testing.T) {
		_, err := ParseOnLoginScript("/system script add name=test")
		require.Error(t, err)
	})
}

func TestNormalizeProfileName(t *testing.T) {
	assert.Equal(t, "1Day_10K", NormalizeProfileName("1Day_10K"))
	assert.Equal(t, "Paket-1-Hari", NormalizeProfileName("Paket 1 Hari"))
	assert.Equal(t, "A-B", NormalizeProfileName("A   B"))
}

func TestNewAddMikhmonProfileCommand(t *testing.T) {
	cmd := NewAddMikhmonProfileCommand(MikhmonProfileParams{
		Name:        "1Day_10K",
		AddressPool: "hs-pool",
		RateLimit:   "5M/5M",
		Price:       "10000",
		Validity:    "1d",
	})

	assert.Equal(t, "/ip/hotspot/user/profile/add", cmd.Raw)
	assert.Equal(t, "1Day_10K", cmd.Args["name"])
	assert.Equal(t, "hs-pool", cmd.Args["address-pool"])
	assert.Equal(t, "5M/5M", cmd.Args["rate-limit"])
	assert.NotEmpty(t, cmd.Args["on-login"], "on-login script harus diisi")
	assert.Contains(t, cmd.Args["on-login"], ",ntf,10000,1d,")
}
