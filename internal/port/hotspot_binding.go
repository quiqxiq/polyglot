package port

// HotspotIPBindingParams represents parameters for adding or setting an IP Binding.
type HotspotIPBindingParams struct {
	MACAddress string
	Address    string
	ToAddress  string
	Server     string
	Type       string
	Comment    string
	Disabled   bool
}
