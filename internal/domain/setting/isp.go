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
}
