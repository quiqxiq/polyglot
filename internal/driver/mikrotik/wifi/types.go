package wifi

// WiFiInterface represents a RouterOS wireless or wifiwave2 interface.
type WiFiInterface struct {
	RosID       string `json:"ros_id"`
	Name        string `json:"name"`
	MasterIface string `json:"master_interface"`
	SSID        string `json:"ssid"`
	Band        string `json:"band"`
	Channel     string `json:"channel"`
	Frequency   string `json:"frequency"`
	Width       string `json:"width"`
	Security    string `json:"security"`
	Running     bool   `json:"running"`
	Disabled    bool   `json:"disabled"`
	Comment     string `json:"comment"`
}

// WiFiRegistration represents an associated wireless client (/interface/wireless/registration-table or /interface/wifi/registration-table).
type WiFiRegistration struct {
	RosID      string `json:"ros_id"`
	Interface  string `json:"interface"`
	MACAddress string `json:"mac_address"`
	SSID       string `json:"ssid"`
	Signal     string `json:"signal"`
	TxRate     string `json:"tx_rate"`
	RxRate     string `json:"rx_rate"`
	Uptime     string `json:"uptime"`
	BytesTx    string `json:"bytes_tx"`
	BytesRx    string `json:"bytes_rx"`
	PacketsTx  string `json:"packets_tx"`
	PacketsRx  string `json:"packets_rx"`
	Comment    string `json:"comment"`
}
