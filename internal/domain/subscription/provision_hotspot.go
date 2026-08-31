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
	Name           string // nama profil (sama dengan nama ServicePlan)
	RateLimit      string // "10M/5M" atau 8-segmen burst
	AddressPool    string // nama IP pool di router (ip_pool_name)
	AddressList    string // firewall address-list (mis. "ISOLIR_USERS", "VIP")
	SharedUsers    int    // batas perangkat bersamaan (default 1)
	ParentQueue    string // nama parent queue
	SessionTimeout string // session-timeout di MikroTik (opsional)
	IdleTimeout    string // idle-timeout di MikroTik (opsional)
	Comment        string
	OnLogin        string
	OnLogout       string
}

// HotspotProvisionSpec adalah paket spesifikasi lengkap untuk provisi layanan Hotspot ke router.
type HotspotProvisionSpec struct {
	User    HotspotUserSpec
	Profile HotspotProfileSpec
}
