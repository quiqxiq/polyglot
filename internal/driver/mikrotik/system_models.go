package mikrotik

// SystemResource holds the parsed data from /system/resource/print.
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

// SystemIdentity holds the router's configured identity (name).
type SystemIdentity struct {
	Name string
}

// SystemClock holds the current clock state from /system/clock/print.
type SystemClock struct {
	Time         string
	Date         string
	TimeZoneName string
	GMTOffset    string
}

// LogEntry represents one row returned by /log/print.
type LogEntry struct {
	RosID   string
	Time    string
	Topics  string // comma-separated categories (e.g. "ppp,info")
	Message string
}

// PingResult holds one "!re" row from a /ping command.
type PingResult struct {
	Seq        int
	Host       string
	Sent       int
	Received   int
	PacketLoss int
	MinRTT     string
	AvgRTT     string
	MaxRTT     string
	TTL        string
	Time       string
}

// SystemSchedulerParams holds parameters for /system/scheduler/add or /set.
type SystemSchedulerParams struct {
	Name      string
	StartTime string
	StartDate string
	Interval  string
	OnEvent   string
	Comment   string
	Disabled  bool
}

// SystemScheduler represents one row from /system/scheduler/print.
type SystemScheduler struct {
	RosID     string
	Name      string
	StartTime string
	StartDate string
	Interval  string
	OnEvent   string
	NextRun   string
	Comment   string
	Disabled  bool
}

// SystemScript represents one row from /system/script/print.
type SystemScript struct {
	RosID   string
	Name    string
	Owner   string
	Source  string
	Comment string
}
