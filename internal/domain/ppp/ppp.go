package ppp

// ActiveSession represents one row returned by /ppp/active/print.
// This is a read-only monitoring resource — the only mutation is
// removing (kicking/disconnecting) an active session.
type ActiveSession struct {
	RosID     string
	Name      string
	Service   string
	CallerID  string
	Address   string
	Encoding  string
	SessionID string
	Radius    bool
	Profile   string
}

// ActiveStat represents real-time telemetry metrics returned by
// /ppp/active/print stats interval=.
type ActiveStat struct {
	RosID         string
	Uptime        string
	LimitBytesIn  string
	LimitBytesOut string
}

// PPPoESecret represents one row returned by /ppp/secret/print. Only fields
// that are stable across RouterOS versions and genuinely useful to the
// application are exposed.
type PPPoESecret struct {
	RosID         string // RouterOS internal .id, needed for set/remove
	Name          string
	Profile       string
	Service       string
	LocalAddress  string
	RemoteAddress string
	Comment       string
	Disabled      bool
	LastLoggedOut string
	CallerID      string
}

// Profile represents one row returned by /ppp/profile/print.
type Profile struct {
	RosID          string
	Name           string
	RateLimit      string
	LocalAddress   string
	RemoteAddress  string
	DNSServer      string
	ParentQueue    string
	AddressList    string
	Comment        string
	SharedUsers    string
	OnlyOne        string
	UseMPLS        string
	UseCompression string
	UseEncryption  string
	ChangeTCPMSS   string
	BridgeLearning string
}
