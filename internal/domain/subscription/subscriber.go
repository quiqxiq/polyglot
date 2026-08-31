package subscription

// SubscriberAccount adalah data akun jaringan yang diprovisikan ke router.
type SubscriberAccount struct {
	Username  string
	Password  string
	Profile   string // nama profil paket di router (default: nama paket)
	RateLimit string // untuk auto-buat profil paket, mis. "5M/5M" atau 8-segmen burst
	Comment   string

	// Parameter profil lengkap dari ServicePlan (opsional; kosong = abaikan).
	AddressPool  string // hotspot: IP pool profil
	ParentQueue  string // parent queue (hotspot & ppp profile)
	AddressList  string // ppp profile: address-list
	SharedUsers  int    // hotspot: batasi login bersamaan
	Price        string // hotspot: harga jual (desimal tanpa simbol)
	SellingPrice string // hotspot: harga modal
	Validity     string // hotspot: masa aktif ("30d")
	ExpireMode   string // hotspot: ntf|ntfc|rem|remc|0
	LockUser     bool   // hotspot: kunci user ke MAC
	LockServer   bool   // hotspot: kunci user ke server

	BaseRateLimit     string // rate tanpa burst ("10M/5M") — CIR untuk queue DEDICATED
	RemoteAddressPool string // ppp profile: pool IP sumber alamat pelanggan (remote-address)
	LocalAddress      string // ppp profile: IP gateway router (local-address)
	DNSServer         string // ppp profile: DNS server
}
