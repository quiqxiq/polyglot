package port

// ListUsersFilter holds optional filters for /ip/hotspot/user/print.
// Mirrors the legacy Mikhmon list/query filters: profile, comment (batch tag)
// and only_unused (uptime=0s — vouchers that have never logged in).
type ListUsersFilter struct {
	Name       string // ?name=   — exact username lookup
	Profile    string // ?profile= — profile filter
	Comment    string // ?comment= — batch/tag filter (e.g. "MyTag")
	OnlyUnused bool   // ?uptime=0s — never logged in
}

// HotspotUserParams holds the parameters for creating or updating a RouterOS
// hotspot user (/ip/hotspot/user/add or /ip/hotspot/user/set).
//
// Field notes (from RouterOS /ip/hotspot/user reference):
//   - Name           : login username.
//   - Password       : login password.
//   - Profile        : name of an existing /ip/hotspot/user/profile entry.
//   - Server         : hotspot server name (e.g. "hotspot1"). Leave empty
//     to apply to all servers ("all"). For voucher systems
//     this is typically empty.
//   - MACAddress     : bind this user to a specific MAC address.
//   - Address        : assign a static IP to this user. Leave empty for
//     dynamic assignment from the profile's pool.
//   - LimitUptime    : max cumulative online time (RouterOS duration string,
//     e.g. "30d", "8h"). Empty = unlimited.
//   - LimitBytesIn   : max incoming bytes (numeric string). Empty = unlimited.
//   - LimitBytesOut  : max outgoing bytes (numeric string). Empty = unlimited.
//   - Comment        : free-text label. Convention: prefix "voucher" tag here
//     for voucher-type users so the app can filter them.
//   - Disabled       : when true the user exists but cannot log in.
type HotspotUserParams struct {
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
}

// HotspotUser represents one row returned by /ip/hotspot/user/print.
//
// Field notes:
//   - RosID  : internal RouterOS ID — required for set (via numbers=) and remove.
//   - Server : "all" when not server-specific.
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
