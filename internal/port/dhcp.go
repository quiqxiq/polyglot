package port

// DHCPLease represents one row returned by /ip/dhcp-server/lease/print.
//
// DHCP leases are read-only or set-only (for block/unblock) in this
// application — there is no add or remove for leases, since RouterOS manages
// them dynamically based on client requests.
//
// Field notes (from RouterOS /ip/dhcp-server/lease reference):
//   - RosID         : internal ID — required for /set.
//   - Address       : IP address given to the client.
//   - MACAddress    : hardware address of the client.
//   - ClientID      : DHCP option 61 client identifier (may be empty).
//   - Server        : DHCP server that issued this lease.
//   - Status        : "bound", "waiting", "offered", "expired", etc.
//   - ActiveAddress : actual active IP (may differ from Address during
//     transition states).
//   - HostName      : client's self-reported hostname (DHCP option 12).
//   - Blocked       : true when the lease is blocked via /set blocked=yes,
//     which is the MAC-based suspension mechanism.
//   - Comment       : free-text label. Convention for blocked leases:
//     "SUSPENDED - <reason> - <timestamp>".
type DHCPLease struct {
	RosID         string
	Address       string
	MACAddress    string
	ClientID      string
	Server        string
	Status        string
	ActiveAddress string
	HostName      string
	Blocked       bool
	Comment       string
}

// DHCPLeaseBlockParams holds the parameters for blocking or unblocking a
// DHCP lease via /ip/dhcp-server/lease/set. This is the MAC-based
// suspension method for subscribers on DHCP.
type DHCPLeaseBlockParams struct {
	Blocked bool
	Comment string // should include reason and timestamp for audit trail
}
