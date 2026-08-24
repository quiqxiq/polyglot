package importer_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/usecase/importer"
)

const csvSample = `id_pelanggan,nama,nomor_telepon,alamat,tipe,server,username,password,paket,harga,rate_limit,status
01075,MATRAJI-KT,085606846141,"KATAPANG, SAMPANG",PPPOE,JAYA ABADI,MATRAJI-KT,rahasia,100-RB-100,100000,5M/5M,ACTIVE
01076,SITI,-6281234567890,SAMPANG KOTA,HOTSPOT,JAYA ABADI,siti99,siti123,HOME-20M,150000,,ISOLATED

01077,BUDI,08571112222,DUSUN MELATI,PPPOE,JAYA ABADI,budi77,budi456,100-RB-100,Rp. 125.000,,PENDING`

func TestParseCSV_Valid(t *testing.T) {
	rows, err := importer.ParseCSV(strings.NewReader(csvSample))
	require.NoError(t, err)
	require.Len(t, rows, 3)

	first := rows[0]
	assert.Equal(t, "01075", first.CustomerCode)
	assert.Equal(t, "MATRAJI-KT", first.Name)
	assert.Equal(t, "085606846141", first.Phone)
	assert.Equal(t, "JAYA ABADI", first.DeviceName)
	assert.Equal(t, "100-RB-100", first.PlanName)
	assert.InDelta(t, 100000, first.Price, 0.01)
	assert.Equal(t, "ACTIVE", first.Status)

	// Baris kosong dilewati: baris ke-3 kosong → BUDI = RowNumber 4.
	assert.Equal(t, 4, rows[2].RowNumber)
	assert.Equal(t, "PENDING", rows[2].Status)
}

func TestValidateRows_Errors(t *testing.T) {
	rows, err := importer.ParseCSV(strings.NewReader(csvSample))
	require.NoError(t, err)
	// Perbaiki nomor kedua agar lolos pola? ValidateRows hanya cek kehadiran.
	errs := importer.ValidateRows(rows)
	assert.Empty(t, errs)

	bad := []importer.Row{
		{RowNumber: 2, Name: "", Phone: "0812", Address: "x", Status: "ACTIVE"},
		{RowNumber: 3, Name: "SITI", Phone: "", Address: "x", Status: "ACTIVE"},
		{RowNumber: 4, Name: "BUDI", Phone: "0857", Address: "x", Status: "WEIRD"},
		{RowNumber: 5, Name: "EKO", Phone: "0858", Address: "x", Username: "eko", PlanName: ""},
	}
	errs = importer.ValidateRows(bad)
	require.Len(t, errs, 4)
}

func TestCSXXLSX_RoundTrip(t *testing.T) {
	src := []importer.Row{
		{CustomerCode: "C-1", Name: "A", Phone: "08111", Address: "AL1",
			ServiceType: "PPPOE", Username: "a1", Password: "p1",
			PlanName: "PLAN-A", Price: 100000, RateLimit: "5M/5M", Status: "ACTIVE"},
		{CustomerCode: "C-2", Name: "B", Phone: "08122", Address: "AL2",
			ServiceType: "HOTSPOT", Username: "b2", PlanName: "PLAN-B",
			Price: 200000, Status: "ISOLATED"},
	}

	csvBytes, err := importer.WriteCSV(src)
	require.NoError(t, err)
	back, err := importer.ParseCSV(strings.NewReader(string(csvBytes)))
	require.NoError(t, err)
	require.Len(t, back, 2)
	assert.Equal(t, "C-1", back[0].CustomerCode)
	assert.Equal(t, "ISOLATED", back[1].Status)
	assert.InDelta(t, 200000, back[1].Price, 0.01)

	xlsxBytes, err := importer.WriteXLSX(src)
	require.NoError(t, err)
	backX, err := importer.ParseXLSX(xlsxBytes)
	require.NoError(t, err)
	require.Len(t, backX, 2)
	assert.Equal(t, "B", backX[1].Name)
	assert.Equal(t, "PLAN-B", backX[1].PlanName)
}

func TestParseCSV_BadHeader(t *testing.T) {
	_, err := importer.ParseCSV(strings.NewReader("kolom_asal,kolom_ngasal\n1,2\n"))
	assert.ErrorContains(t, err, "header tidak dikenali")
}

func TestParseCSV_EmptyFile(t *testing.T) {
	_, err := importer.ParseCSV(strings.NewReader(""))
	assert.Error(t, err)
}
