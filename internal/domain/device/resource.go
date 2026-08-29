package device

// SystemResource holds the parsed data from system resource metrics.
// It is the vendor-neutral representation of a router's system resource snapshot.
type SystemResource struct {
	CPULoad       int
	CPUCount      int
	CPUFrequency  string // MHz
	FreeMemory    string // bytes
	TotalMemory   string // bytes
	FreeHDDSpace  string // bytes
	TotalHDDSpace string // bytes
	Architecture  string
	Model         string
	SerialNumber  string
	FirmwareType  string
	Voltage       string // millivolts, may be empty
	Temperature   string // Celsius, may be empty
	BadBlocks     string
	Uptime        string
	Version       string
	BoardName     string
}
