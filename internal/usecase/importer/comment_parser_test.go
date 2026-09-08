package importer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quixiq/polyglot/internal/usecase/importer"
)

func TestGuessCustomerName(t *testing.T) {
	tests := []struct {
		name     string
		username string
		comment  string
		want     string
	}{
		{
			name:     "empty comment falls back to username",
			username: "user01",
			comment:  "",
			want:     "user01",
		},
		{
			name:     "explicit prefix with dash",
			username: "user02",
			comment:  "Nama: Budi Santoso - 08123456789",
			want:     "Budi Santoso",
		},
		{
			name:     "delimiter pattern name phone address",
			username: "user03",
			comment:  "Ahmad Dahlan - 081987654321 - Jl. Melati No. 5",
			want:     "Ahmad Dahlan",
		},
		{
			name:     "pipe delimiter",
			username: "user04",
			comment:  "Siti Aminah | 082112345678 | Rp 150.000",
			want:     "Siti Aminah",
		},
		{
			name:     "mikhmon voucher comment falls back to username",
			username: "vc1029",
			comment:  "vc-1029 exp: 2026-09-08 12:00:00",
			want:     "vc1029",
		},
		{
			name:     "numeric comment falls back to username",
			username: "cust99",
			comment:  "192.168.88.50",
			want:     "cust99",
		},
		{
			name:     "empty comment with underscore falls back to formatted username",
			username: "susanto_madek",
			comment:  "",
			want:     "Susanto Madek",
		},
		{
			name:     "empty comment with hyphen falls back to formatted username",
			username: "susanto-madek",
			comment:  "",
			want:     "Susanto Madek",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := importer.GuessCustomerName(tt.username, tt.comment)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatAutoName(t *testing.T) {
	assert.Equal(t, "Susanto Madek", importer.FormatAutoName("susanto_madek"))
	assert.Equal(t, "Susanto Madek", importer.FormatAutoName("susanto-madek"))
	assert.Equal(t, "Budi Santoso Rt01", importer.FormatAutoName("budi_santoso_rt01"))
	assert.Equal(t, "Paket 10mbps", importer.FormatAutoName("paket_10mbps"))
	assert.Equal(t, "susanto", importer.FormatAutoName("susanto"))
	assert.Equal(t, "USER123", importer.FormatAutoName("USER123"))
	assert.Equal(t, "", importer.FormatAutoName(""))
	assert.Equal(t, "", importer.FormatAutoName("   "))
}

func TestGuessPhone(t *testing.T) {
	assert.Equal(t, "08123456789", importer.GuessPhone("Budi Santoso 08123456789"))
	assert.Equal(t, "+6281234567890", importer.GuessPhone("Ahmad (+6281234567890)"))
	assert.Equal(t, "", importer.GuessPhone("No phone here 1234"))
}

func TestExtractPriceFromComment(t *testing.T) {
	assert.Equal(t, float64(150000), importer.ExtractPriceFromComment("Paket 10M Rp 150.000"))
	assert.Equal(t, float64(250000), importer.ExtractPriceFromComment("Budi rp.250.000"))
	assert.Equal(t, float64(0), importer.ExtractPriceFromComment("Tanpa harga"))
}
