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

// SystemIdentity holds the router's configured identity name.
type SystemIdentity struct {
	Name string
}

// LogEntry represents one log event from the device.
type LogEntry struct {
	RosID   string
	Time    string
	Topics  string
	Message string
}

// SystemHealth holds the parsed sensor data from system health.
type SystemHealth struct {
	Voltage        string
	Temperature    string
	CPUTemperature string
	PSUVoltage     string
	PSUCurrent     string
	PSUTemperature string
	Fan1Speed      string
	Fan2Speed      string
}

// SystemClock holds the parsed clock data.
type SystemClock struct {
	Time         string
	Date         string
	TimeZoneName string
	GMTOffset    string
	DSTActive    bool
}

// SystemRouterboard holds the parsed routerboard hardware details.
type SystemRouterboard struct {
	BoardName       string
	Model           string
	SerialNumber    string
	FirmwareType    string
	FactoryFirmware string
	CurrentFirmware string
	UpgradeFirmware string
}
