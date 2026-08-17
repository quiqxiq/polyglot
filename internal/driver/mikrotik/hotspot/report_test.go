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

func TestParseMikhmonTransactions_BrokenRecords(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{
			".id":     "*1",
			"name":    "aug/17/2026-|-14:20:00-|-VIP123", // partial — hanya 3 segmen
			"comment": "mikhmon",
		},
		{
			".id":     "*2",
			"name":    "bukan-laporan", // tanpa delimiter -|-
			"comment": "mikhmon",
		},
		{
			".id":     "*3",
			"name":    "aug/17/2026-|-14:25:00-|-VIP456-|-3000-|-192.168.88.51-|-AA:BB:CC:00:11:22-|-1h-|-1Jam-3K-|-vc-A1B-08.17.26",
			"comment": "mikhmon",
		},
	}}

	records := ParseMikhmonTransactions(result)
	require.Len(t, records, 2, "record partial tetap di-parse; non-delimiter dilewati")

	assert.Equal(t, "VIP123", records[0].Username, "username dari record partial harus terbaca")
	assert.Empty(t, records[0].Price, "price kosong pada record partial")

	assert.Equal(t, "3000", records[1].Price)
	assert.Equal(t, "1Jam-3K", records[1].Profile)
}
