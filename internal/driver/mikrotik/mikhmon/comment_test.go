package mikhmon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatPreLoginComment(t *testing.T) {
	tm := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)

	t.Run("dengan tag", func(t *testing.T) {
		formatted := FormatPreLoginComment("vc", "A3X", "Paket1Hari", tm)
		assert.Equal(t, "vc-A3X-08.03.26-Paket1Hari", formatted)
	})

	t.Run("tanpa tag", func(t *testing.T) {
		formatted := FormatPreLoginComment("vc", "B7Y", "", tm)
		assert.Equal(t, "vc-B7Y-08.03.26", formatted)
	})

	t.Run("type default vc", func(t *testing.T) {
		formatted := FormatPreLoginComment("", "C9Z", "Tag", tm)
		assert.Equal(t, "vc-C9Z-08.03.26-Tag", formatted)
	})
}

func TestParseMikhmonComment(t *testing.T) {
	t.Run("parse pre-login comment", func(t *testing.T) {
		comment := "vc-A3X-08.03.26-Voucher_1_Hari"
		parsed, err := ParseMikhmonComment(comment)
		require.NoError(t, err)
		assert.Equal(t, "vc", parsed.Type)
		assert.Equal(t, "A3X", parsed.Code)
		assert.Equal(t, "08.03.26", parsed.CreatedDate)
		assert.Equal(t, "Voucher_1_Hari", parsed.Tag)
		assert.False(t, parsed.IsActivated)
	})

	t.Run("parse post-login comment activated", func(t *testing.T) {
		comment := "03/08/2026 15:30:00 N vc-A3X-08.03.26-Voucher_1_Hari"
		parsed, err := ParseMikhmonComment(comment)
		require.NoError(t, err)
		assert.True(t, parsed.IsActivated)
		assert.Equal(t, "03/08/2026", parsed.ExpireDate)
		assert.Equal(t, "15:30:00", parsed.ExpireTime)
		assert.Equal(t, "N", parsed.ExpireMode)
		assert.Equal(t, "vc", parsed.Type)
		assert.Equal(t, "A3X", parsed.Code)
		assert.Equal(t, "Voucher_1_Hari", parsed.Tag)
	})

	t.Run("invalid comment format", func(t *testing.T) {
		_, err := ParseMikhmonComment("invalid-comment")
		assert.ErrorIs(t, err, ErrInvalidMikhmonComment)
	})
}

func TestIsExpired(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

	t.Run("belum aktif — false", func(t *testing.T) {
		assert.False(t, IsExpired("vc-A3X-08.03.26-Tag", now))
	})

	t.Run("sudah expired — true", func(t *testing.T) {
		// Expiry 03/08/2026 10:00:00 (sebelum now 12:00:00)
		comment := "03/08/2026 10:00:00 N vc-A3X-08.03.26-Tag"
		assert.True(t, IsExpired(comment, now))
	})

	t.Run("belum expired — false", func(t *testing.T) {
		// Expiry 03/08/2026 15:00:00 (setelah now 12:00:00)
		comment := "03/08/2026 15:00:00 N vc-A3X-08.03.26-Tag"
		assert.False(t, IsExpired(comment, now))
	})
}
