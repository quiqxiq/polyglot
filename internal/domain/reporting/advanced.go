package reporting

import "time"

// ARAgingRow adalah satu bucket umur piutang tagihan pelanggan.
type ARAgingRow struct {
	Bucket       string  `json:"bucket"` // CURRENT | D1_30 | D31_60 | D61_90 | D90_PLUS
	InvoiceCount int     `json:"invoice_count"`
	Outstanding  float64 `json:"outstanding"`
}

// ARAgingReport merangkum AR Aging per bucket.
type ARAgingReport struct {
	AsOf       time.Time    `json:"as_of"`
	Rows       []ARAgingRow `json:"rows"`
	TotalCount int          `json:"total_count"`
	Total      float64      `json:"total_outstanding"`
}

// GatewayReconRow membandingkan transaksi gateway settled dengan kas tercatat.
type GatewayReconRow struct {
	Gateway          string  `json:"gateway"`
	TransactionCount int     `json:"transaction_count"`
	GatewayAmount    float64 `json:"gateway_amount"`
	FeeAmount        float64 `json:"fee_amount"`
	RecordedAmount   float64 `json:"recorded_amount"`
}

// GatewayReconciliationReport merangkum rekonsiliasi per gateway.
type GatewayReconciliationReport struct {
	From          time.Time         `json:"from"`
	To            time.Time         `json:"to"`
	Rows          []GatewayReconRow `json:"rows"`
	TotalGateway  float64           `json:"total_gateway"`
	TotalRecorded float64           `json:"total_recorded"`
	TotalFee      float64           `json:"total_fee"`
}

// PSBStatusCount adalah jumlah pendaftaran per status.
type PSBStatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// PSBStatsReport merangkum efisiensi & durasi penyelesaian PSB.
type PSBStatsReport struct {
	From            time.Time        `json:"from"`
	To              time.Time        `json:"to"`
	Total           int              `json:"total"`
	Converted       int              `json:"converted"`
	Closed          int              `json:"closed"`
	ByStatus        []PSBStatusCount `json:"by_status"`
	AvgSurveyHours  float64          `json:"avg_survey_hours"`
	AvgInstallHours float64          `json:"avg_install_hours"`
	AvgLeadHours    float64          `json:"avg_lead_hours"`
}

// TaxRow adalah rekap PPN per periode / paket.
type TaxRow struct {
	Period       string  `json:"period"`
	InvoiceCount int     `json:"invoice_count"`
	Subtotal     float64 `json:"subtotal"`
	TaxPercent   float64 `json:"tax_percent"`
	TaxAmount    float64 `json:"tax_amount"`
	Total        float64 `json:"total"`
	PaidTax      float64 `json:"paid_tax"`
}

// TaxReport merangkum laporan pajak PPN.
type TaxReport struct {
	Period        string   `json:"period"`
	Rows          []TaxRow `json:"rows"`
	TotalSubtotal float64  `json:"total_subtotal"`
	TotalTax      float64  `json:"total_tax"`
	TotalBilled   float64  `json:"total_billed"`
	TotalPaidTax  float64  `json:"total_paid_tax"`
}

// SalesRow adalah rekap penjualan per paket / tipe layanan.
type SalesRow struct {
	PlanName     string  `json:"plan_name"`
	ServiceType  string  `json:"service_type"`
	InvoiceCount int     `json:"invoice_count"`
	TotalBilled  float64 `json:"total_billed"`
	TotalPaid    float64 `json:"total_paid"`
}

// SalesReport merangkum laporan pendapatan per paket/layanan.
type SalesReport struct {
	Period        string     `json:"period"`
	Rows          []SalesRow `json:"rows"`
	TotalBilled   float64    `json:"total_billed"`
	TotalPaid     float64    `json:"total_paid"`
	PaidCount     int        `json:"paid_count"`
	CustomerCount int        `json:"customer_count"`
}

// CollectionReport merangkum rasio penagihan & keterlambatan.
type CollectionReport struct {
	Period           string  `json:"period"`
	TotalInvoices    int     `json:"total_invoices"`
	PaidInvoices     int     `json:"paid_invoices"`
	OverdueInvoices  int     `json:"overdue_invoices"`
	TotalBilled      float64 `json:"total_billed"`
	TotalCollected   float64 `json:"total_collected"`
	TotalOutstanding float64 `json:"total_outstanding"`
	CollectionRate   float64 `json:"collection_rate"` // (TotalCollected / TotalBilled) * 100
}

// CashflowRow mutasi kas per kategori.
type CashflowRow struct {
	CategoryName string  `json:"category_name"`
	Direction    string  `json:"direction"`
	TrxCount     int     `json:"trx_count"`
	TotalAmount  float64 `json:"total_amount"`
}

// CashflowReport merangkum arus kas masuk dan keluar.
type CashflowReport struct {
	From        time.Time     `json:"from"`
	To          time.Time     `json:"to"`
	Rows        []CashflowRow `json:"rows"`
	TotalIn     float64       `json:"total_in"`
	TotalOut    float64       `json:"total_out"`
	NetCashflow float64       `json:"net_cashflow"`
}

// AgingDetailRow rincian piutang per pelanggan.
type AgingDetailRow struct {
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Phone        string    `json:"phone"`
	InvoiceCount int       `json:"invoice_count"`
	Outstanding  float64   `json:"outstanding"`
	OldestDue    time.Time `json:"oldest_due"`
	DaysOverdue  int       `json:"days_overdue"`
}

// AgingDetailReport laporan rincian piutang per pelanggan.
type AgingDetailReport struct {
	AsOf             time.Time        `json:"as_of"`
	Rows             []AgingDetailRow `json:"rows"`
	TotalCustomers   int              `json:"total_customers"`
	TotalOutstanding float64          `json:"total_outstanding"`
}

// ClosingReport ringkasan operasional dan keuangan pada tanggal tutup buku.
type ClosingReport struct {
	Period           string    `json:"period"`
	ClosingDate      time.Time `json:"closing_date"`
	TotalInvoices    int       `json:"total_invoices"`
	TotalBilled      float64   `json:"total_billed"`
	TotalCollected   float64   `json:"total_collected"`
	TotalOutstanding float64   `json:"total_outstanding"`
	ActiveCustomers  int       `json:"active_customers"`
	NewSubscriptions int       `json:"new_subscriptions"`
	DebtsOutstanding float64   `json:"debts_outstanding"`
}
