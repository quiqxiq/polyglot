package bridge

// Bridge represents a RouterOS /interface/bridge entry.
type Bridge struct {
	RosID         string `json:"ros_id"`
	Name          string `json:"name"`
	MTU           string `json:"mtu"`
	ActualMTU     string `json:"actual_mtu"`
	MACAddress    string `json:"mac_address"`
	FastForward   bool   `json:"fast_forward"`
	VlanFiltering bool   `json:"vlan_filtering"`
	Disabled      bool   `json:"disabled"`
	Running       bool   `json:"running"`
	Comment       string `json:"comment"`
}

// BridgePort represents a RouterOS /interface/bridge/port entry.
type BridgePort struct {
	RosID     string `json:"ros_id"`
	Interface string `json:"interface"`
	Bridge    string `json:"bridge"`
	Priority  string `json:"priority"`
	PathCost  string `json:"path_cost"`
	Horizon   string `json:"horizon"`
	Learn     string `json:"learn"`
	PVID      string `json:"pvid"`
	Disabled  bool   `json:"disabled"`
	Inactive  bool   `json:"inactive"`
	Comment   string `json:"comment"`
}

// BridgeVLAN represents a RouterOS /interface/bridge/vlan entry.
type BridgeVLAN struct {
	RosID           string `json:"ros_id"`
	Bridge          string `json:"bridge"`
	VlanIDs         string `json:"vlan_ids"`
	Tagged          string `json:"tagged"`
	Untagged        string `json:"untagged"`
	CurrentTagged   string `json:"current_tagged"`
	CurrentUntagged string `json:"current_untagged"`
	Disabled        bool   `json:"disabled"`
	Comment         string `json:"comment"`
}

// BridgeHost represents a learned MAC address on the bridge (/interface/bridge/host).
type BridgeHost struct {
	RosID       string `json:"ros_id"`
	Bridge      string `json:"bridge"`
	MACAddress  string `json:"mac_address"`
	Interface   string `json:"interface"`
	OnInterface string `json:"on_interface"`
	Age         string `json:"age"`
	Local       bool   `json:"local"`
	Dynamic     bool   `json:"dynamic"`
	Disabled    bool   `json:"disabled"`
}

// BridgeParams holds creation/edit params for a bridge.
type BridgeParams struct {
	Name          string
	VlanFiltering bool
	FastForward   bool
	Comment       string
	Disabled      bool
}

// BridgePortParams holds creation/edit params for a bridge port.
type BridgePortParams struct {
	Interface string
	Bridge    string
	PVID      string
	Comment   string
	Disabled  bool
}

// BridgeVLANParams holds creation/edit params for a bridge VLAN table entry.
type BridgeVLANParams struct {
	Bridge   string
	VlanIDs  string
	Tagged   string
	Untagged string
	Comment  string
	Disabled bool
}
