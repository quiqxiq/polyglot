package hotspot

// HotspotUser represents a hotspot user account on a network device.
type HotspotUser struct {
	RosID         string
	Name          string
	Password      string
	Profile       string
	Server        string
	MACAddress    string
	Address       string
	LimitUptime   string
	LimitBytesIn  string
	LimitBytesOut string
	Comment       string
	Disabled      bool
	Uptime        string
	BytesIn       string
	BytesOut      string
}

// HotspotUserProfile represents a hotspot user profile configuration.
type HotspotUserProfile struct {
	RosID          string
	Name           string
	RateLimit      string
	SessionTimeout string
	IdleTimeout    string
	SharedUsers    string
	ParentQueue    string
	AddressPool    string
	AddressList    string
	Comment        string
	OnLogin        string
}

// HotspotActiveSession represents an active hotspot session.
type HotspotActiveSession struct {
	RosID      string
	Server     string
	User       string
	Address    string
	MACAddress string
	LoginBy    string
}

// HotspotActiveStat represents real-time telemetry metrics of an active session.
type HotspotActiveStat struct {
	RosID           string
	Uptime          string
	SessionTimeLeft string
	IdleTime        string
	BytesIn         string
	BytesOut        string
	PacketsIn       string
	PacketsOut      string
}

// HotspotIPBinding represents an IP binding rule in the hotspot.
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

// HotspotCookie represents a login cookie session.
type HotspotCookie struct {
	RosID      string
	User       string
	MACAddress string
	ExpiresIn  string
	Domain     string
}

// GeneratedVoucher holds generated voucher credentials.
type GeneratedVoucher struct {
	Username string
	Password string
	Comment  string
}

// VoucherBatch holds a collection of generated vouchers.
type VoucherBatch struct {
	Vouchers []GeneratedVoucher
}

// VoucherStatusDetails aggregates all relevant status of a voucher.
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

// MikhmonTransaction represents a transaction record stored on the device.
type MikhmonTransaction struct {
	RosID    string
	Date     string
	Time     string
	Username string
	Price    string
	Address  string
	MAC      string
	Validity string
	Profile  string
	Comment  string
	RawName  string
}

// ExpireMonitorStatus describes the install/enabled state of the expire monitor.
type ExpireMonitorStatus struct {
	IsInstalled   bool
	IsEnabled     bool
	SchedulerID   string
	SchedulerName string
}

// MikhmonComment represents parsed metadata from a voucher comment.
type MikhmonComment struct {
	Type        string // "vc" or "up"
	Code        string
	CreatedDate string
	Tag         string

	IsActivated bool
	ExpireDate  string
	ExpireTime  string
	ExpireMode  string
	RawComment  string
}
