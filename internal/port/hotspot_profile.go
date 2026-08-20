package port

// HotspotUserProfile represents one row returned by
// /ip/hotspot/user/profile/print.
type HotspotUserProfile struct {
	RosID          string
	Name           string
	RateLimit      string
	SessionTimeout string
	IdleTimeout    string
	SharedUsers    string
	ParentQueue    string
	AddressPool    string
	Comment        string
	OnLogin        string
}

// ExpireMode defines how expired vouchers are handled by Mikhmon's expire monitor script.
type ExpireMode string

// MikhmonProfileParams holds the parameters required to create or update a
// Hotspot User Profile with full Mikhmon v4 metadata and auto-expiry logic.
type MikhmonProfileParams struct {
	Name            string     // Profile name (e.g. "1Day_10K")
	AddressPool     string     // IP pool name
	SharedUsers     string     // Concurrent logins (default "1")
	RateLimit       string     // rx/tx rate limit (e.g. "5M/5M")
	ParentQueue     string     // Parent queue name
	Price           string     // Selling price (e.g. "10000")
	SellingPrice    string     // Cost price (e.g. "8000")
	Validity        string     // Validity duration (e.g. "1d", "7d", "30d")
	ExpireMode      ExpireMode // ExpireModeNotify ("ntf") or ExpireModeRemove ("rem")
	LockUser        bool       // Lock user to MAC address on first login
	LockServer      bool       // Lock user to Hotspot server
	EnableRecording bool       // Save transaction log in /system script
	Comment         string     // Profile comment
}
