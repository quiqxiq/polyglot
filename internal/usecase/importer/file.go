package importer

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ParseCSV membaca baris data dari CSV (baris pertama = header).
func ParseCSV(r io.Reader) ([]Row, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1 // kolom fleksibel; divalidasi header
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}
	return parseRecords(records)
}

// ParseXLSX membaca sheet pertama (baris pertama = header).
func ParseXLSX(data []byte) ([]Row, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open XLSX: %w", err)
	}
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetList()
	if len(sheet) == 0 {
		return nil, errors.New("XLSX tanpa sheet")
	}
	rows, err := f.GetRows(sheet[0])
	if err != nil {
		return nil, fmt.Errorf("read XLSX: %w", err)
	}
	return parseRecords(rows)
}

// WriteCSV menghasilkan file CSV dari rows (format sama dengan parser).
func WriteCSV(rows []Row) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	header := []string{"id_pelanggan", "nama", "nomor_telepon", "email", "alamat",
		"latitude", "longitude", "tipe", "server", "username", "password",
		"paket", "harga", "rate_limit", "status", "local_address", "remote_address", "parent_queue"}
	if err := w.Write(header); err != nil {
		return nil, err
	}
	for _, r := range rows {
		rec := []string{
			r.CustomerCode, r.Name, r.Phone, r.Email, r.Address,
			fnum(r.Latitude), fnum(r.Longitude), r.ServiceType, r.DeviceName,
			r.Username, r.Password, r.PlanName, fmt.Sprintf("%.0f", r.Price),
			r.RateLimit, r.Status, r.LocalAddress, r.RemoteAddr, r.ParentQueue,
		}
		if err := w.Write(rec); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteXLSX menghasilkan file XLSX dari rows.
func WriteXLSX(rows []Row) ([]byte, error) {
	f := excelize.NewFile()
	const sheet = "Pelanggan"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}
	csvBytes, err := WriteCSV(rows)
	if err != nil {
		return nil, err
	}
	records, err := csv.NewReader(bytes.NewReader(csvBytes)).ReadAll()
	if err != nil {
		return nil, err
	}
	for i, rec := range records {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetSheetRow(sheet, cell, &rec); err != nil {
			return nil, err
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ─── shared ─────────────────────────────────────────────────────────────

func parseRecords(records [][]string) ([]Row, error) {
	if len(records) < 2 {
		return nil, errors.New("importer: file empty or header only")
	}
	m := mapHeaders(records[0])
	if len(m) == 0 {
		return nil, errors.New("importer: unrecognized header: download the export template as reference")
	}
	rows := make([]Row, 0, len(records)-1)
	for i, cells := range records[1:] {
		rowNo := i + 2 // +header, 1-based seperti spreadsheet
		if isBlankRow(cells) {
			continue
		}
		rows = append(rows, buildRow(m, cells, rowNo))
	}
	return rows, nil
}

func isBlankRow(cells []string) bool {
	for _, c := range cells {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func fnum(p *float64) string {
	if p == nil {
		return ""
	}
	return strconv.FormatFloat(*p, 'f', -1, 64)
}
