package firewall

// FirewallFilterParams holds the parameters for adding a RouterOS firewall filter rule.
type FirewallFilterParams struct {
	Chain          string
	Action         string
	SrcAddress     string
	SrcAddressList string
	DstAddress     string
	Protocol       string
	PlaceBefore    string
	Comment        string
	Disabled       bool
}

// FirewallFilter represents one row returned by /ip/firewall/filter/print.
type FirewallFilter struct {
	RosID          string
	Chain          string
	Action         string
	SrcAddress     string
	SrcAddressList string
	DstAddress     string
	Protocol       string
	Bytes          string
	Packets        string
	Comment        string
	Disabled       bool
}

// FirewallFilterPrintParams holds filter criteria for /ip/firewall/filter/print.
type FirewallFilterPrintParams struct {
	Chain          string
	SrcAddress     string
	SrcAddressList string
	Action         string
}

// NATRuleParams holds parameters for adding or modifying a RouterOS NAT rule.
type NATRuleParams struct {
	Chain          string
	Action         string
	ToAddresses    string
	ToPorts        string
	SrcAddress     string
	SrcAddressList string
	DstAddress     string
	Protocol       string
	DstPort        string
	PlaceBefore    string
	Comment        string
	Disabled       bool
}

// NATRule represents one row from /ip/firewall/nat/print.
type NATRule struct {
	RosID          string
	Chain          string
	Action         string
	ToAddresses    string
	ToPorts        string
	SrcAddress     string
	SrcAddressList string
	DstAddress     string
	Protocol       string
	DstPort        string
	Bytes          string
	Packets        string
	Comment        string
	Disabled       bool
}

// AddressListParams holds parameters for adding a member to an address-list.
type AddressListParams struct {
	List    string
	Address string
	Timeout string
	Comment string
}

// AddressListEntry represents one row returned by /ip/firewall/address-list/print.
type AddressListEntry struct {
	RosID    string
	List     string
	Address  string
	Timeout  string
	Comment  string
	Disabled bool
	Dynamic  bool
}

// AddressListPrintParams holds query filter criteria for /ip/firewall/address-list/print.
type AddressListPrintParams struct {
	List    string
	Address string
}

// IPPool represents one row from /ip/pool/print.
type IPPool struct {
	RosID    string
	Name     string
	Ranges   string
	NextPool string
	Comment  string
}

// IPAddress represents one row from /ip/address/print.
type IPAddress struct {
	RosID           string
	Address         string
	Network         string
	Interface       string
	ActualInterface string
	Disabled        bool
	Dynamic         bool
	Comment         string
}

// IPAddressPrintParams holds filter criteria for /ip/address/print.
type IPAddressPrintParams struct {
	Interface string
	Address   string
}

// IPRoute represents one row from /ip/route/print.
type IPRoute struct {
	RosID      string
	DstAddress string
	Gateway    string
	Distance   string
	Active     bool
	Dynamic    bool
	Static     bool
	Comment    string
}

// BlockIPFilterParams returns a pre-filled FirewallFilterParams for blocking a specific subscriber IP.
func BlockIPFilterParams(customerIP, reason string) FirewallFilterParams {
	return FirewallFilterParams{
		Chain:      "forward",
		Action:     "drop",
		SrcAddress: customerIP,
		Comment:    "SUSPENDED block_" + customerIP + " - " + reason,
	}
}

// BlockAddressListFilterParams returns a pre-filled FirewallFilterParams for a list-based block rule.
func BlockAddressListFilterParams(chain, listName string) FirewallFilterParams {
	var comment string
	if chain == "input" {
		comment = "Block suspended customers from accessing router (static IP)"
	} else {
		comment = "Block suspended customers (static IP)"
	}
	return FirewallFilterParams{
		Chain:          chain,
		Action:         "drop",
		SrcAddressList: listName,
		PlaceBefore:    "0",
		Comment:        comment,
	}
}
