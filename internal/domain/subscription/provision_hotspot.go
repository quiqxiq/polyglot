package subscription

// HotspotUserSpec mendefinisikan parameter pembuatan/pembaruan akun /ip/hotspot/user di MikroTik.
type HotspotUserSpec struct {
	Username    string
	Password    string
	Profile     string // nama user profile di router (mis. "WARKOP-10M" atau "isolir")
	Server      string // "all" atau nama server hotspot tertentu
	MacAddress  string // bind MAC address pelanggan jika ada
	IPAddress   string // static IP jika ada
	LimitUptime string // kuota waktu (kosong = unlimited bulanan)
	LimitBytes  string // kuota data (kosong = unlimited)
	Comment     string // format "polyglot:SUB-xxx"
	Disabled    bool   // true bila subscription SUSPENDED
}

// HotspotProfileSpec mendefinisikan parameter pembuatan/pemastian profil /ip/hotspot/user/profile di MikroTik.
type HotspotProfileSpec struct {
	Name         string // nama profil (sama dengan nama ServicePlan)
	RateLimit    string // "10M/5M" atau 8-segmen burst
	AddressPool  string // nama IP pool di router (ip_pool_name)
	SharedUsers  int    // batas perangkat bersamaan (default 1)
	ParentQueue  string // nama parent queue
	Price        string // harga voucher (format string desimal)
	SellingPrice string // harga modal/reseller
	Validity     string // masa aktif (mis. "30d")
	ExpireMode   string // mode expire Mikhmon ("ntf", "ntfc", "rem", "remc", "0")
	LockUser     bool   // kunci user ke MAC perangkat
	LockServer   bool   // kunci user ke server hotspot
	Comment      string
}

// HotspotProvisionSpec adalah paket spesifikasi lengkap untuk provisi layanan Hotspot ke router.
type HotspotProvisionSpec struct {
	User    HotspotUserSpec
	Profile HotspotProfileSpec
}
