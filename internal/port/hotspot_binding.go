package port

// HotspotIPBinding represents an entry from /ip/hotspot/ip-binding/print.
type HotspotIPBinding struct {
	RosID      string
	MACAddress string
	Address    string
	ToAddress  string
	Server     string
	Type       string // "bypassed" | "blocked" | "regular"
	Comment    string
	Disabled   bool
}

// HotspotIPBindingParams represents parameters for adding or setting an IP Binding.
type HotspotIPBindingParams struct {
	MACAddress string
	Address    string
	ToAddress  string
	Server     string
	Type       string
	Comment    string
	Disabled   bool
}

// HotspotCookie represents an active login cookie from /ip/hotspot/cookie/print.
type HotspotCookie struct {
	RosID      string
	User       string
	MACAddress string
	ExpiresIn  string
	Domain     string
}

// VoucherStatusDetails aggregates status data for quick voucher inspection.
type VoucherStatusDetails struct {
	Found         bool
	User          *HotspotUser
	Profile       *HotspotUserProfile
	IsOnline      bool
	ActiveSession *HotspotActiveSession
	HasCookie     bool
	Cookie        *HotspotCookie
	Status        string // "active" | "expired" | "unused" | "disabled" | "not_found"
	SisaWaktu     string
	SisaKuota     string
	ExpireDate    string
	MACLocked     string
	Message       string
}
