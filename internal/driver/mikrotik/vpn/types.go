package vpn

// WireGuardInterface represents a RouterOS /interface/wireguard entry.
type WireGuardInterface struct {
	RosID      string `json:"ros_id"`
	Name       string `json:"name"`
	ListenPort string `json:"listen_port"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	MTU        string `json:"mtu"`
	Running    bool   `json:"running"`
	Disabled   bool   `json:"disabled"`
	Comment    string `json:"comment"`
}

// WireGuardPeer represents a RouterOS /interface/wireguard/peers entry.
type WireGuardPeer struct {
	RosID           string `json:"ros_id"`
	Interface       string `json:"interface"`
	PublicKey       string `json:"public_key"`
	EndpointAddress string `json:"endpoint_address"`
	EndpointPort    string `json:"endpoint_port"`
	AllowedAddress  string `json:"allowed_address"`
	LatestHandshake string `json:"latest_handshake"`
	TransferRx      string `json:"transfer_rx"`
	TransferTx      string `json:"transfer_tx"`
	Disabled        bool   `json:"disabled"`
	Comment         string `json:"comment"`
}

// IPsecPeer represents a RouterOS /ip/ipsec/peer entry.
type IPsecPeer struct {
	RosID        string `json:"ros_id"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	Profile      string `json:"profile"`
	ExchangeMode string `json:"exchange_mode"`
	Disabled     bool   `json:"disabled"`
	Comment      string `json:"comment"`
}

// IPsecActivePeer represents an established IPsec active peer (/ip/ipsec/active-peers).
type IPsecActivePeer struct {
	RosID         string `json:"ros_id"`
	ID            string `json:"id"`
	State         string `json:"state"`
	Uptime        string `json:"uptime"`
	RemoteAddress string `json:"remote_address"`
	LocalAddress  string `json:"local_address"`
}
