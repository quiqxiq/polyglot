package ipservice

// IPService represents a router management service (api, winbox, ssh, www, etc).
type IPService struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Port        int    `json:"port"`
	Disabled    bool   `json:"disabled"`
	Address     string `json:"address"`
	Certificate string `json:"certificate"`
}

// IPCloud represents MikroTik IP Cloud dynamic DNS and status.
type IPCloud struct {
	DDNSEnabled   bool   `json:"ddns_enabled"`
	UpdateTime    bool   `json:"update_time"`
	PublicAddress string `json:"public_address"`
	DNSName       string `json:"dns_name"`
	Status        string `json:"status"`
}

// NeighborEntry represents a discovered network neighbor via MNDP, CDP, or LLDP.
type NeighborEntry struct {
	ID                string `json:"id"`
	Interface         string `json:"interface"`
	MacAddress        string `json:"mac_address"`
	Identity          string `json:"identity"`
	Platform          string `json:"platform"`
	Version           string `json:"version"`
	Unpack            string `json:"unpack"`
	Age               string `json:"age"`
	SystemDescription string `json:"system_description"`
	SystemCaps        string `json:"system_caps"`
}

// DiscoverySettings represents neighbor discovery interface configuration.
type DiscoverySettings struct {
	DiscoverInterfaceList string `json:"discover_interface_list"`
}

// ARPEntry represents one ARP cache table mapping.
type ARPEntry struct {
	ID         string `json:"id"`
	Address    string `json:"address"`
	MacAddress string `json:"mac_address"`
	Interface  string `json:"interface"`
	Status     string `json:"status"`
	Dynamic    bool   `json:"dynamic"`
	DHCP       bool   `json:"dhcp"`
	Disabled   bool   `json:"disabled"`
	Comment    string `json:"comment"`
}

// SNMPSettings represents global SNMP agent parameters.
type SNMPSettings struct {
	Enabled  bool   `json:"enabled"`
	Contact  string `json:"contact"`
	Location string `json:"location"`
}

// SNMPCommunity represents an SNMP v1/v2c community configuration.
type SNMPCommunity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Addresses   string `json:"addresses"`
	Security    string `json:"security"`
	ReadAccess  bool   `json:"read_access"`
	WriteAccess bool   `json:"write_access"`
}
