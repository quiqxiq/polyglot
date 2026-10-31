package setting

// ISPSettings holds business settings for ISP billing and automatic isolation.
type ISPSettings struct {
	BillingCycleStartDay int    // tanggal mulai periode tagihan (1-28)
	BillingDueDays       int    // hari kalender jatuh tempo setelah periode
	AutoIsolate          bool   // isolir otomatis aktif?
	IsolateGraceDays     int    // toleransi hari setelah jatuh tempo
	PPPoEIsolirProfile   string // profil isolir PPPoE di router
	HotspotIsolirProfile string // profil isolir hotspot di router
	PaymentRedirectURL   string // host:port halaman bayar tujuan dst-nat
	SuspendAfterDays     int    // ISOLATED → SUSPENDED otomatis (0 = off)
	IsolirAddressList    string // address-list penanda pelanggan terisolir

	// Billing Ops v3 (PLAN-BILLING-OPS-V3): mode siklus & konfigurasi keuangan.
	BillingMode                 string  // PRABAYAR | PASCABAYAR
	ClosingDate                 int     // tanggal tutup buku/siklus (1-28)
	UnpaidDaysDisplay           int     // ambang label "Belum Bayar H-N"
	PrepayInvoiceLeadDays       int     // mode PRABAYAR: hari sebelum siklus invoice terbit
	TaxPercent                  float64 // PPN default mitra bila plan.tax_percent = 0
	ReminderPreOffsets          []int   // offset pengingat sebelum jatuh tempo
	ReminderPostOffsets         []int   // offset pengingat setelah jatuh tempo
	GatewaySyncStaleMinutes     int     // transaksi PENDING lebih dari N menit disinkronkan
	DebtIncomeCategoryID        string  // kategori kas penerimaan hutang
	DebtReminderEnabled         bool    // pengingat hutang aktif
	DebtReminderDays            []int   // offset pengingat hutang
	GatewayFeeExpenseCategoryID string  // kategori kas beban biaya gateway
	MassPaymentEnabled          bool    // mass payment aktif

	// Blueprint isolir §5.4 (planning v2 D12).
	IsolirWebProxyEnabled bool     // aktifkan webproxy MikroTik untuk halaman isolir
	IsolirWebProxyPort    int      // port webproxy isolir (default 3125)
	IsolirBillingHosts    []string // host server billing yang di-whitelist
	IsolirErrorPageURL    string   // URL error.html untuk /tool fetch (mode self)
}
