package port

import "context"

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
}

// IsolationOptions membawa parameter isolation yang berasal dari
// system_settings (bukan konstanta kode).
type IsolationOptions struct {
	IsolirProfile string                   // profil tujuan saat isolir
	AddressList   string                   // penanda IP pelanggan terisolir
	Redirect      *IsolationRedirectConfig // nil = tanpa rule dst-nat global
}

// RouterAccountManager mengelola siklus hidup akun pelanggan di router BRAS
// dengan semantik ISP nyata:
//   - Provision : buat akun + pastikan profil paket ada (auto dari plan).
//   - Update    : ganti profil (ganti paket / isolir / restore).
//   - Suspend   : disabled=true — cuti/penghentian sementara.
//   - Terminate : hapus akun permanen.
//
// ISOLIR ≠ disable: akun dipindah ke profil isolir, sesi di-kick, dan IP
// ditandai address-list agar rule dst-nat redirect ke halaman bayar bekerja.
type RouterAccountManager interface {
	Provision(ctx context.Context, deviceID, serviceType string, acct SubscriberAccount) error
	UpdateAccount(ctx context.Context, deviceID, serviceType, username, newProfile string) error
	// EnsureProfile memastikan profil bernama profileName ada di router
	// dengan rate tertentu (auto-create bila belum ada). Dipakai sebelum
	// memindahkan akun ke profil tersebut (mis. saat ganti paket).
	EnsureProfile(ctx context.Context, deviceID, serviceType, profileName, rateLimit string) error
	Isolate(ctx context.Context, deviceID, serviceType, username string, opt IsolationOptions) error
	Restore(ctx context.Context, deviceID, serviceType, username, normalProfile, addressList string) error
	Suspend(ctx context.Context, deviceID, serviceType, username string) error
	Terminate(ctx context.Context, deviceID, serviceType, username string) error
}
