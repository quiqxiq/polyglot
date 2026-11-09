package report

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	domainReporting "github.com/quixiq/polyglot/internal/domain/reporting"
	"github.com/xuri/excelize/v2"
)

// ExportReport exports any requested report into CSV or XLSX format.
func (uc *ManageReportUseCase) ExportReport(ctx context.Context, reportType, format string, params map[string]string) ([]byte, string, string, error) {
	var headers []string
	var rows [][]string
	title := strings.ToLower(reportType)

	switch strings.ToUpper(reportType) {
	case "TAX":
		period := params["period"]
		rep, err := uc.TaxReport(ctx, period)
		if err != nil {
			return nil, "", "", err
		}
		title = fmt.Sprintf("tax-%s", rep.Period)
		headers = []string{"Periode", "Jumlah Faktur", "Subtotal", "Tarif PPN (%)", "Total PPN", "Total Tagihan", "PPN Terbayar"}
		rows = make([][]string, len(rep.Rows))
		for i, r := range rep.Rows {
			rows[i] = []string{
				r.Period,
				fmt.Sprintf("%d", r.InvoiceCount),
				fmt.Sprintf("%.2f", r.Subtotal),
				fmt.Sprintf("%.2f", r.TaxPercent),
				fmt.Sprintf("%.2f", r.TaxAmount),
				fmt.Sprintf("%.2f", r.Total),
				fmt.Sprintf("%.2f", r.PaidTax),
			}
		}

	case "SALES":
		period := params["period"]
		rep, err := uc.SalesReport(ctx, period)
		if err != nil {
			return nil, "", "", err
		}
		title = fmt.Sprintf("sales-%s", rep.Period)
		headers = []string{"Paket Layanan", "Tipe", "Jumlah Faktur", "Total Tagihan", "Total Terbayar"}
		rows = make([][]string, len(rep.Rows))
		for i, r := range rep.Rows {
			rows[i] = []string{
				r.PlanName,
				r.ServiceType,
				fmt.Sprintf("%d", r.InvoiceCount),
				fmt.Sprintf("%.2f", r.TotalBilled),
				fmt.Sprintf("%.2f", r.TotalPaid),
			}
		}

	case "CASHFLOW":
		from := params["from"]
		to := params["to"]
		rep, err := uc.CashflowReport(ctx, from, to)
		if err != nil {
			return nil, "", "", err
		}
		title = fmt.Sprintf("cashflow-%s-to-%s", rep.From.Format("20060102"), rep.To.Format("20060102"))
		headers = []string{"Kategori", "Arah Arus Kas", "Jumlah Transaksi", "Nominal"}
		rows = make([][]string, len(rep.Rows))
		for i, r := range rep.Rows {
			rows[i] = []string{
				r.CategoryName,
				r.Direction,
				fmt.Sprintf("%d", r.TrxCount),
				fmt.Sprintf("%.2f", r.TotalAmount),
			}
		}

	case "AGING_DETAIL":
		asOf := params["as_of"]
		rep, err := uc.AgingDetail(ctx, asOf, 1000)
		if err != nil {
			return nil, "", "", err
		}
		title = fmt.Sprintf("aging-detail-%s", rep.AsOf.Format("20060102"))
		headers = []string{"ID Pelanggan", "Nama Pelanggan", "No. Telepon", "Jumlah Faktur", "Total Piutang", "Jatuh Tempo Tertua", "Hari Menunggak"}
		rows = make([][]string, len(rep.Rows))
		for i, r := range rep.Rows {
			rows[i] = []string{
				r.CustomerID,
				r.CustomerName,
				r.Phone,
				fmt.Sprintf("%d", r.InvoiceCount),
				fmt.Sprintf("%.2f", r.Outstanding),
				r.OldestDue.Format("2006-01-02"),
				fmt.Sprintf("%d", r.DaysOverdue),
			}
		}

	default:
		return nil, "", "", fmt.Errorf("%w: unknown report type %q", domainReporting.ErrInvalidInput, reportType)
	}

	filename := fmt.Sprintf("%s-%s", title, time.Now().Format("20060102150405"))

	if strings.ToLower(format) == "xlsx" {
		f := excelize.NewFile()
		defer func() { _ = f.Close() }()
		sheet := "Report"
		f.SetSheetName("Sheet1", sheet)
		for colIdx, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
			_ = f.SetCellValue(sheet, cell, h)
		}
		for rowIdx, r := range rows {
			for colIdx, val := range r {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
				_ = f.SetCellValue(sheet, cell, val)
			}
		}
		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return nil, "", "", err
		}
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", filename + ".xlsx", nil
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(headers)
	_ = w.WriteAll(rows)
	w.Flush()
	return buf.Bytes(), "text/csv", filename + ".csv", nil
}
