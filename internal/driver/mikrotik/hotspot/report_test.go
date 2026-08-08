package hotspot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestParseMikhmonTransactions(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{
			".id":     "*1",
			"name":    "dec/25/2024-|-10:30:15-|-budi-|-10000-|-192.168.1.100-|-AA:BB:CC:DD:EE:FF-|-1d-|-1Day_10K-|-Voucher 1 Hari",
			"owner":   "Dec2024",
			"comment": "mikhmon",
		},
		{
			".id":     "*2",
			"name":    "non-mikhmon-script",
			"owner":   "admin",
			"comment": "mikhmon",
		},
	}}

	records := ParseMikhmonTransactions(result)
	require.Len(t, records, 1, "hanya script ber-delimiter -|- yang di-parse")

	rec := records[0]
	assert.Equal(t, "*1", rec.RosID)
	assert.Equal(t, "dec/25/2024", rec.Date)
	assert.Equal(t, "10:30:15", rec.Time)
	assert.Equal(t, "budi", rec.Username)
	assert.Equal(t, "10000", rec.Price)
	assert.Equal(t, "192.168.1.100", rec.Address)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", rec.MAC)
	assert.Equal(t, "1d", rec.Validity)
	assert.Equal(t, "1Day_10K", rec.Profile)
	assert.Equal(t, "Voucher 1 Hari", rec.Comment)
}

func TestNewPrintMikhmonReportsCommand(t *testing.T) {
	cmd := NewPrintMikhmonReportsCommand()
	assert.Equal(t, "/system/script/print", cmd.Raw)
	assert.Equal(t, "mikhmon", cmd.Args["?comment"])
}
