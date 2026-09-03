package device

// Interface represents one row returned by /interface/print.
//
// Field notes (from RouterOS /interface reference):
//   - RosID      : internal ID — required for enable/disable.
//   - Name       : interface name (e.g. "ether1", "pppoe-out1", "bridge").
//   - Type       : interface type ("ether", "vlan", "bridge", "pppoe-in", etc.).
//   - MTU        : configured MTU.
//   - ActualMTU  : effective MTU (may differ for PPP interfaces).
//   - L2MTU      : Layer-2 MTU.
//   - MACAddress : hardware address.
//   - Running    : true when the physical link is up.
//   - Disabled   : true when administratively disabled.
//   - RxByte     : cumulative received bytes (since last reboot/reset).
//   - TxByte     : cumulative transmitted bytes.
//   - RxPacket   : cumulative received packets.
//   - TxPacket   : cumulative transmitted packets.
//   - Comment    : free-text label.
type Interface struct {
	RosID      string
	Name       string
	Type       string
	MTU        string
	ActualMTU  string
	L2MTU      string
	MACAddress string
	Running    bool
	Disabled   bool
	RxByte     string
	TxByte     string
	RxPacket   string
	TxPacket   string
	Comment    string
}

// InterfaceTrafficStats holds the real-time traffic statistics returned by
// /interface/monitor-traffic with the "once" flag. Values are instantaneous
// rates at the moment of the query — not cumulative.
//
// All rate fields are in bits per second (bps) as strings (RouterOS format).
type InterfaceTrafficStats struct {
	RxBitsPerSecond    string
	TxBitsPerSecond    string
	RxPacketsPerSecond string
	TxPacketsPerSecond string
	RxDropsPerSecond   string
	TxDropsPerSecond   string
	RxErrorsPerSecond  string
	TxErrorsPerSecond  string
}
