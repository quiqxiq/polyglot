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
