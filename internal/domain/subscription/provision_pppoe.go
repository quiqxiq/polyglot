package subscription

// PPPoESecretSpec mendefinisikan parameter pembuatan/pembaruan akun /ppp/secret di MikroTik.
type PPPoESecretSpec struct {
	Username      string
	Password      string
	Profile       string // nama PPP profile di router (mis. "HOME-20M" atau "isolir")
	Service       string // default "pppoe"
	LocalAddress  string // static gateway jika ada
	RemoteAddress string // static IP pelanggan jika ada
	CallerID      string // bind MAC address pelanggan jika ada
	Routes        string // static routes jika ada
	Comment       string // format "polyglot:SUB-xxx"
	Disabled      bool   // true bila subscription SUSPENDED
}

// PPPoEProfileSpec mendefinisikan parameter pembuatan/pemastian profil /ppp/profile di MikroTik.
type PPPoEProfileSpec struct {
	Name              string // nama profil (sama dengan nama ServicePlan)
	RateLimit         string // "20M/10M" atau 8-segmen burst
	RemoteAddressPool string // nama IP pool di router untuk alokasi IP dinamis pelanggan
	LocalAddress      string // gateway IP pool (opsional)
	ParentQueue       string // nama parent queue di Simple Queue / Queue Tree
	AddressList       string // nama address-list firewall
	SessionTimeout    string // session-timeout di MikroTik (opsional)
	IdleTimeout       string // idle-timeout di MikroTik (opsional)
	Comment           string
}

// PPPoEProvisionSpec adalah paket spesifikasi lengkap untuk provisi layanan PPPoE ke router.
type PPPoEProvisionSpec struct {
	Secret  PPPoESecretSpec
	Profile PPPoEProfileSpec
}
