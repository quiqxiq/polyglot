package port

// HotspotActiveSession represents one row returned by
// /ip/hotspot/active/print. This is a read-only monitoring resource —
// the only write operation is Remove (disconnect a session).
//
// Field notes (from RouterOS /ip/hotspot/active reference):
//   - RosID          : session ID — required for /ip/hotspot/active/remove.
//   - Server         : hotspot server name.
//   - User           : username (matches /ip/hotspot/user name).
//   - Address        : IP address of the connected client.
//   - MACAddress     : MAC address of the client.
//   - LoginBy        : authentication method (http-pap, cookie, mac, etc.).
//   - Uptime         : total session duration (RouterOS time string).
//   - SessionTimeLeft: remaining allowed session time (empty if unlimited).
//   - IdleTime       : current idle duration.
//   - BytesIn        : bytes received from client in this session.
//   - BytesOut       : bytes sent to client in this session.
//   - PacketsIn      : packets received from client.
//   - PacketsOut     : packets sent to client.
type HotspotActiveSession struct {
	RosID      string
	Server     string
	User       string
	Address    string
	MACAddress string
	LoginBy    string
}

// HotspotActiveStat represents dynamic real-time telemetry metrics returned by
// /ip/hotspot/active/print stats interval=.
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
