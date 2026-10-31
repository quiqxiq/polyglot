package postgres

import (
	"context"
	"fmt"
	"time"

	domainReporting "github.com/quixiq/polyglot/internal/domain/reporting"
	"github.com/quixiq/polyglot/internal/port"
)

var _ port.AdvancedReportRepository = (*ReportingRepository)(nil)

// ARAging menghitung umur piutang dari invoices yang masih outstanding.
func (r *ReportingRepository) ARAging(ctx context.Context, tenantID string, asOf time.Time) (domainReporting.ARAgingReport, error) {
	asOf = time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 0, 0, 0, 0, time.UTC)
	report := domainReporting.ARAgingReport{AsOf: asOf}
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT CASE
			WHEN due_date >= ? THEN 'CURRENT'
			WHEN due_date >= ? - INTERVAL '30 days' THEN 'D1_30'
			WHEN due_date >= ? - INTERVAL '60 days' THEN 'D31_60'
			WHEN due_date >= ? - INTERVAL '90 days' THEN 'D61_90'
			ELSE 'D90_PLUS' END AS bucket,
		COUNT(*) AS invoice_count,
		COALESCE(SUM(total - paid_amount), 0) AS outstanding
		FROM invoices
		WHERE tenant_id = ? AND deleted_at IS NULL
		  AND status IN ('UNPAID', 'PARTIAL', 'OVERDUE')
		GROUP BY bucket`,
		asOf, asOf, asOf, asOf, tenantID).Rows()
	if err != nil {
		return report, fmt.Errorf("ar aging query: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var row domainReporting.ARAgingRow
		if err := rows.Scan(&row.Bucket, &row.InvoiceCount, &row.Outstanding); err != nil {
			return report, fmt.Errorf("scan ar aging: %w", err)
		}
		report.Rows = append(report.Rows, row)
		report.TotalCount += row.InvoiceCount
		report.Total += row.Outstanding
	}
	return report, rows.Err()
}

// GatewayReconciliation membandingkan transaksi gateway settled dengan
// payment yang tercatat di buku kas.
func (r *ReportingRepository) GatewayReconciliation(ctx context.Context, tenantID string, from, to time.Time) (domainReporting.GatewayReconciliationReport, error) {
	report := domainReporting.GatewayReconciliationReport{From: from, To: to}
	var rows []struct {
		Gateway          string
		TransactionCount int
		GatewayAmount    float64
		FeeAmount        float64
		RecordedAmount   float64
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT g.gateway,
			COUNT(*) AS transaction_count,
			COALESCE(SUM(g.amount), 0) AS gateway_amount,
			COALESCE(SUM(g.fee_amount), 0) AS fee_amount,
			COALESCE(SUM(CASE WHEN p.voided_at IS NULL THEN p.amount ELSE 0 END), 0) AS recorded_amount
		FROM gateway_transactions g
		LEFT JOIN payments p ON p.id = g.payment_id
		WHERE g.tenant_id = ? AND g.status = 'SETTLED'
		  AND g.paid_at >= ? AND g.paid_at < ?
		GROUP BY g.gateway
		ORDER BY g.gateway`,
		tenantID, from, to).Scan(&rows).Error
	if err != nil {
		return report, fmt.Errorf("gateway reconciliation query: %w", err)
	}
	for _, row := range rows {
		report.Rows = append(report.Rows, domainReporting.GatewayReconRow{
			Gateway:          row.Gateway,
			TransactionCount: row.TransactionCount,
			GatewayAmount:    row.GatewayAmount,
			FeeAmount:        row.FeeAmount,
			RecordedAmount:   row.RecordedAmount,
		})
		report.TotalGateway += row.GatewayAmount
		report.TotalRecorded += row.RecordedAmount
		report.TotalFee += row.FeeAmount
	}
	return report, nil
}

// PSBStats menghitung statistik efisiensi penyelesaian PSB pada rentang tanggal.
func (r *ReportingRepository) PSBStats(ctx context.Context, tenantID string, from, to time.Time) (domainReporting.PSBStatsReport, error) {
	report := domainReporting.PSBStatsReport{From: from, To: to}
	var agg struct {
		Total           int
		Converted       int
		Closed          int
		AvgSurveyHours  float64
		AvgInstallHours float64
		AvgLeadHours    float64
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'ACTIVE') AS converted,
			COUNT(*) FILTER (WHERE status IN ('REJECTED','CANCELLED')) AS closed,
			COALESCE(AVG(EXTRACT(EPOCH FROM (survey_date::timestamp - created_at)) / 3600)
				FILTER (WHERE survey_date IS NOT NULL), 0) AS avg_survey_hours,
			COALESCE(AVG(EXTRACT(EPOCH FROM (installed_at - created_at)) / 3600)
				FILTER (WHERE installed_at IS NOT NULL), 0) AS avg_install_hours,
			COALESCE(AVG(EXTRACT(EPOCH FROM (installed_at - created_at)) / 3600)
				FILTER (WHERE installed_at IS NOT NULL), 0) AS avg_lead_hours
		FROM registrations
		WHERE tenant_id = ? AND created_at >= ? AND created_at < ?`,
		tenantID, from, to).Scan(&agg).Error
	if err != nil {
		return report, fmt.Errorf("psb stats query: %w", err)
	}
	report.Total = agg.Total
	report.Converted = agg.Converted
	report.Closed = agg.Closed
	report.AvgSurveyHours = roundHours(agg.AvgSurveyHours)
	report.AvgInstallHours = roundHours(agg.AvgInstallHours)
	report.AvgLeadHours = roundHours(agg.AvgLeadHours)

	var statusRows []struct {
		Status string
		Count  int
	}
	if serr := r.db.WithContext(ctx).Raw(`
		SELECT status, COUNT(*) AS count FROM registrations
		WHERE tenant_id = ? AND created_at >= ? AND created_at < ?
		GROUP BY status ORDER BY status`, tenantID, from, to).Scan(&statusRows).Error; serr != nil {
		return report, fmt.Errorf("psb status query: %w", serr)
	}
	for _, row := range statusRows {
		report.ByStatus = append(report.ByStatus, domainReporting.PSBStatusCount{Status: row.Status, Count: row.Count})
	}
	return report, nil
}

// TaxReport menghitung rekap PPN per periode tagihan.
func (r *ReportingRepository) TaxReport(ctx context.Context, tenantID, period string) (domainReporting.TaxReport, error) {
	report := domainReporting.TaxReport{Period: period}
	var rows []struct {
		Period       string
		InvoiceCount int
		Subtotal     float64
		TaxPercent   float64
		TaxAmount    float64
		Total        float64
		PaidTax      float64
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT period,
			COUNT(*) AS invoice_count,
			COALESCE(SUM(subtotal), 0) AS subtotal,
			COALESCE(AVG(tax_percent), 0) AS tax_percent,
			COALESCE(SUM(tax_amount), 0) AS tax_amount,
			COALESCE(SUM(total), 0) AS total,
			COALESCE(SUM(CASE WHEN status = 'PAID' THEN tax_amount ELSE 0 END), 0) AS paid_tax
		FROM invoices
		WHERE tenant_id = ? AND period = ? AND deleted_at IS NULL
		GROUP BY period`, tenantID, period).Scan(&rows).Error
	if err != nil {
		return report, fmt.Errorf("tax report query: %w", err)
	}
	for _, row := range rows {
		report.Rows = append(report.Rows, domainReporting.TaxRow{
			Period:       row.Period,
			InvoiceCount: row.InvoiceCount,
			Subtotal:     row.Subtotal,
			TaxPercent:   row.TaxPercent,
			TaxAmount:    row.TaxAmount,
			Total:        row.Total,
			PaidTax:      row.PaidTax,
		})
		report.TotalSubtotal += row.Subtotal
		report.TotalTax += row.TaxAmount
		report.TotalBilled += row.Total
		report.TotalPaidTax += row.PaidTax
	}
	return report, nil
}

// SalesReport menghitung rekap penjualan per paket layanan pada periode tertentu.
func (r *ReportingRepository) SalesReport(ctx context.Context, tenantID, period string) (domainReporting.SalesReport, error) {
	report := domainReporting.SalesReport{Period: period}
	var rows []struct {
		PlanName     string
		ServiceType  string
		InvoiceCount int
		TotalBilled  float64
		TotalPaid    float64
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(sp.name, 'Lainnya') AS plan_name,
			COALESCE(sp.service_type, 'PPPOE') AS service_type,
			COUNT(*) AS invoice_count,
			COALESCE(SUM(i.total), 0) AS total_billed,
			COALESCE(SUM(i.paid_amount), 0) AS total_paid
		FROM invoices i
		LEFT JOIN subscriptions s ON s.id = i.subscription_id
		LEFT JOIN service_plans sp ON sp.id = s.plan_id
		WHERE i.tenant_id = ? AND i.period = ? AND i.deleted_at IS NULL
		GROUP BY sp.name, sp.service_type
		ORDER BY total_billed DESC`, tenantID, period).Scan(&rows).Error
	if err != nil {
		return report, fmt.Errorf("sales report query: %w", err)
	}
	for _, row := range rows {
		report.Rows = append(report.Rows, domainReporting.SalesRow{
			PlanName:     row.PlanName,
			ServiceType:  row.ServiceType,
			InvoiceCount: row.InvoiceCount,
			TotalBilled:  row.TotalBilled,
			TotalPaid:    row.TotalPaid,
		})
		report.TotalBilled += row.TotalBilled
		report.TotalPaid += row.TotalPaid
	}

	var meta struct {
		PaidCount     int
		CustomerCount int
	}
	_ = r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT customer_id) AS customer_count,
			COUNT(*) FILTER (WHERE status = 'PAID') AS paid_count
		FROM invoices
		WHERE tenant_id = ? AND period = ? AND deleted_at IS NULL`, tenantID, period).Scan(&meta)
	report.PaidCount = meta.PaidCount
	report.CustomerCount = meta.CustomerCount
	return report, nil
}

// CollectionReport menghitung rasio pengumpulan pembayaran.
func (r *ReportingRepository) CollectionReport(ctx context.Context, tenantID, period string) (domainReporting.CollectionReport, error) {
	report := domainReporting.CollectionReport{Period: period}
	var agg struct {
		TotalInvoices    int
		PaidInvoices     int
		OverdueInvoices  int
		TotalBilled      float64
		TotalCollected   float64
		TotalOutstanding float64
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) AS total_invoices,
			COUNT(*) FILTER (WHERE status = 'PAID') AS paid_invoices,
			COUNT(*) FILTER (WHERE status IN ('UNPAID','PARTIAL','OVERDUE') AND due_date < CURRENT_DATE) AS overdue_invoices,
			COALESCE(SUM(total), 0) AS total_billed,
			COALESCE(SUM(paid_amount), 0) AS total_collected,
			COALESCE(SUM(total - paid_amount), 0) AS total_outstanding
		FROM invoices
		WHERE tenant_id = ? AND period = ? AND deleted_at IS NULL`, tenantID, period).Scan(&agg).Error
	if err != nil {
		return report, fmt.Errorf("collection report query: %w", err)
	}
	report.TotalInvoices = agg.TotalInvoices
	report.PaidInvoices = agg.PaidInvoices
	report.OverdueInvoices = agg.OverdueInvoices
	report.TotalBilled = agg.TotalBilled
	report.TotalCollected = agg.TotalCollected
	report.TotalOutstanding = agg.TotalOutstanding
	if report.TotalBilled > 0 {
		report.CollectionRate = (report.TotalCollected / report.TotalBilled) * 100
	}
	return report, nil
}

// CashflowReport merangkum mutasi kas masuk dan keluar per kategori.
func (r *ReportingRepository) CashflowReport(ctx context.Context, tenantID string, from, to time.Time) (domainReporting.CashflowReport, error) {
	report := domainReporting.CashflowReport{From: from, To: to}
	var rows []struct {
		CategoryName string
		Direction    string
		TrxCount     int
		TotalAmount  float64
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(cc.name, 'Lainnya') AS category_name,
			ct.direction,
			COUNT(*) AS trx_count,
			COALESCE(SUM(ct.amount), 0) AS total_amount
		FROM cash_transactions ct
		LEFT JOIN cash_categories cc ON cc.id = ct.category_id
		WHERE ct.tenant_id = ? AND ct.trx_date >= ? AND ct.trx_date < ?
		GROUP BY cc.name, ct.direction
		ORDER BY total_amount DESC`, tenantID, from, to).Scan(&rows).Error
	if err != nil {
		return report, fmt.Errorf("cashflow report query: %w", err)
	}
	for _, row := range rows {
		report.Rows = append(report.Rows, domainReporting.CashflowRow{
			CategoryName: row.CategoryName,
			Direction:    row.Direction,
			TrxCount:     row.TrxCount,
			TotalAmount:  row.TotalAmount,
		})
		if row.Direction == "IN" {
			report.TotalIn += row.TotalAmount
		} else {
			report.TotalOut += row.TotalAmount
		}
	}
	report.NetCashflow = report.TotalIn - report.TotalOut
	return report, nil
}

// AgingDetail menampilkan rincian piutang per pelanggan.
func (r *ReportingRepository) AgingDetail(ctx context.Context, tenantID string, asOf time.Time, limit int) (domainReporting.AgingDetailReport, error) {
	report := domainReporting.AgingDetailReport{AsOf: asOf}
	if limit <= 0 {
		limit = 100
	}
	var rows []struct {
		CustomerID   string
		CustomerName string
		Phone        string
		InvoiceCount int
		Outstanding  float64
		OldestDue    time.Time
		DaysOverdue  int
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT c.id AS customer_id,
			c.name AS customer_name,
			c.phone,
			COUNT(*) AS invoice_count,
			COALESCE(SUM(i.total - i.paid_amount), 0) AS outstanding,
			MIN(i.due_date) AS oldest_due,
			COALESCE(EXTRACT(DAY FROM (?::date - MIN(i.due_date))), 0)::int AS days_overdue
		FROM invoices i
		JOIN customers c ON c.id = i.customer_id
		WHERE i.tenant_id = ? AND i.deleted_at IS NULL
		  AND i.status IN ('UNPAID', 'PARTIAL', 'OVERDUE')
		  AND i.due_date < ?
		GROUP BY c.id, c.name, c.phone
		ORDER BY outstanding DESC
		LIMIT ?`, asOf, tenantID, asOf, limit).Scan(&rows).Error
	if err != nil {
		return report, fmt.Errorf("aging detail query: %w", err)
	}
	for _, row := range rows {
		report.Rows = append(report.Rows, domainReporting.AgingDetailRow{
			CustomerID:   row.CustomerID,
			CustomerName: row.CustomerName,
			Phone:        row.Phone,
			InvoiceCount: row.InvoiceCount,
			Outstanding:  row.Outstanding,
			OldestDue:    row.OldestDue,
			DaysOverdue:  row.DaysOverdue,
		})
		report.TotalCustomers++
		report.TotalOutstanding += row.Outstanding
	}
	return report, nil
}

// ClosingReport merangkum metrik operasional & keuangan pada tanggal tutup buku.
func (r *ReportingRepository) ClosingReport(ctx context.Context, tenantID, period string, closingDate time.Time) (domainReporting.ClosingReport, error) {
	report := domainReporting.ClosingReport{Period: period, ClosingDate: closingDate}
	var invAgg struct {
		TotalInvoices    int
		TotalBilled      float64
		TotalCollected   float64
		TotalOutstanding float64
	}
	_ = r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) AS total_invoices,
			COALESCE(SUM(total), 0) AS total_billed,
			COALESCE(SUM(paid_amount), 0) AS total_collected,
			COALESCE(SUM(total - paid_amount), 0) AS total_outstanding
		FROM invoices
		WHERE tenant_id = ? AND period = ? AND deleted_at IS NULL`, tenantID, period).Scan(&invAgg)
	report.TotalInvoices = invAgg.TotalInvoices
	report.TotalBilled = invAgg.TotalBilled
	report.TotalCollected = invAgg.TotalCollected
	report.TotalOutstanding = invAgg.TotalOutstanding

	var activeCust int64
	_ = r.db.WithContext(ctx).Table("customers").Where("tenant_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL", tenantID).Count(&activeCust)
	report.ActiveCustomers = int(activeCust)

	var newSubs int64
	_ = r.db.WithContext(ctx).Table("subscriptions").Where("tenant_id = ? AND start_date >= ? AND start_date <= ?", tenantID, closingDate.AddDate(0, -1, 0), closingDate).Count(&newSubs)
	report.NewSubscriptions = int(newSubs)

	var debtOut float64
	_ = r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(remaining_amount), 0) FROM customer_debts
		WHERE tenant_id = ? AND status IN ('OPEN','PARTIAL')`, tenantID).Scan(&debtOut)
	report.DebtsOutstanding = debtOut

	return report, nil
}

func roundHours(v float64) float64 {
	return float64(int64(v*10+0.5)) / 10
}
