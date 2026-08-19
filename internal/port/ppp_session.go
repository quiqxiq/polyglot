package port

// PPPActiveSession represents one row returned by /ppp/active/print.
// This is a read-only monitoring resource — the only mutation is
// removing (kicking/disconnecting) an active session via Remove.
//
// Field notes (from RouterOS /ppp/active/print output):
//   - RosID          : internal session ID — required for /ppp/active/remove.
//   - Name           : PPPoE username (matches /ppp/secret name).
//   - Service        : pppoe, pptp, l2tp, etc.
//   - CallerID       : MAC address of the connected CPE.
//   - Address        : IP address assigned to the client for this session.
//   - Uptime         : connection duration as RouterOS time string (e.g. "3h25m10s").
//   - Encoding       : encryption/compression info (may be empty).
//   - SessionID      : protocol-level session identifier.
//   - LimitBytesIn   : byte limit for incoming traffic (empty = unlimited).
//   - LimitBytesOut  : byte limit for outgoing traffic (empty = unlimited).
//   - Radius         : "true" if the session is authenticated via RADIUS.
type PPPActiveSession struct {
	RosID         string
	Name          string
	Service       string
	CallerID      string
	Address       string
	Uptime        string
	Encoding      string
	SessionID     string
	LimitBytesIn  string
	LimitBytesOut string
	Radius        bool
	Profile       string
}

// PPPActiveStat represents real-time telemetry metrics returned by
// /ppp/active/print stats interval=.
type PPPActiveStat struct {
	RosID         string
	Uptime        string
	LimitBytesIn  string
	LimitBytesOut string
	BytesIn       string
	BytesOut      string
	PacketsIn     string
	PacketsOut    string
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

// PPPoESecretParams holds parameters for creating or updating a PPPoE secret.
type PPPoESecretParams struct {
	Name          string
	Password      string
	Profile       string
	Service       string
	LocalAddress  string
	RemoteAddress string
	Comment       string
	Disabled      bool
	CallerID      string
}

// PPPProfile represents one row returned by /ppp/profile/print.
type PPPProfile struct {
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

// PPPProfileParams holds parameters for creating or updating a PPP profile.
type PPPProfileParams struct {
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

