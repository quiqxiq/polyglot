package port

import "context"

// IsolationRedirectConfig describes the global payment-page redirect rules
// applied to isolated subscribers (dst-nat by address-list).
type IsolationRedirectConfig struct {
	SrcAddressList string   // mis. "ISOLIR_USERS"
	PaymentHost    string   // host halaman bayar (tanpa scheme)
	PaymentPort    string   // port halaman bayar; kosong = 80
	Protocols      []string // default: tcp
	DstPorts       []string // default: ["80","443"]
	Disabled       bool     // true = rule dibuat nonaktif (E2E/dry-run)
}

// FirewallGateway abstracts firewall-level operations needed by the ISP
// isolation flow: NAT redirect rules (global, idempotent) dan
// address-list per pelanggan. Diimplementasikan oleh gateway MikroTik.
type FirewallGateway interface {
	// EnsureIsolationRedirect memastikan rule dst-nat redirect milik app
	// ada (dan parameternya sesuai). Idempoten: rule lama diperbarui, bukan
	// diduplikasi.
	EnsureIsolationRedirect(ctx context.Context, driver DeviceDriver, cfg IsolationRedirectConfig) error

	// DisableIsolationRedirect menonaktifkan (bukan menghapus) semua rule
	// redirect milik app pada router tersebut.
	DisableIsolationRedirect(ctx context.Context, driver DeviceDriver) error

	// AddToAddressList menandai IP pelanggan ke list (mis. ISOLIR_USERS).
	AddToAddressList(ctx context.Context, driver DeviceDriver, listName, address, comment string) error

	// RemoveFromAddressList menghapus penanda IP dari list.
	RemoveFromAddressList(ctx context.Context, driver DeviceDriver, listName, address string) error

	// RemoveFromAddressListByComment menghapus semua entri pada list yang
	// comment-nya memuat substring tertentu (mis. username pelanggan).
	RemoveFromAddressListByComment(ctx context.Context, driver DeviceDriver, listName, commentContains string) error
}
