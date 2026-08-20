package hotspot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateVoucherCode(t *testing.T) {
	t.Run("panjang dan karakter set numeric", func(t *testing.T) {
		code := GenerateVoucherCode(6, CharSetNumeric)
		assert.Len(t, code, 6)
		for _, c := range code {
			assert.Contains(t, "1234567890", string(c))
		}
	})

	t.Run("panjang dan karakter set upper", func(t *testing.T) {
		code := GenerateVoucherCode(4, CharSetUpper)
		assert.Len(t, code, 4)
		for _, c := range code {
			assert.Contains(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", string(c))
		}
	})
}

func TestNewAddMikhmonVoucherCommand(t *testing.T) {
	cmd := NewAddMikhmonVoucherCommand(VoucherGenerateParams{
		Profile:     "1Day_10K",
		Server:      "hotspot1",
		LimitUptime: "1d",
		CommentTag:  "VoucherMember",
	}, "user123", "pass123")

	assert.Equal(t, "/ip/hotspot/user/add", cmd.Raw)
	assert.Equal(t, "user123", cmd.Args["name"])
	assert.Equal(t, "pass123", cmd.Args["password"])
	assert.Equal(t, "1Day_10K", cmd.Args["profile"])
	assert.Equal(t, "hotspot1", cmd.Args["server"])
	assert.Equal(t, "1d", cmd.Args["limit-uptime"])
	assert.Contains(t, cmd.Args["comment"], "vc-")
	assert.Contains(t, cmd.Args["comment"], "VoucherMember")
}

func TestNewGenerateVoucherBatchCommands(t *testing.T) {
	batch := NewGenerateVoucherBatchCommands(VoucherGenerateParams{
		Profile:    "1Day_10K",
		Prefix:     "m_",
		UserLength: 5,
		CharSet:    CharSetLowerNum,
	}, 3)

	require.Len(t, batch.Vouchers, 3)
	require.Len(t, batch.Commands, 3)

	for i := 0; i < 3; i++ {
		v := batch.Vouchers[i]
		cmd := batch.Commands[i]
		assert.Equal(t, v.Username, cmd.Args["name"])
		assert.Contains(t, v.Username, "m_")
		assert.Equal(t, "/ip/hotspot/user/add", cmd.Raw)
	}
}
