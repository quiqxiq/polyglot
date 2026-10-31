package netwatch

// NetwatchHost represents a host monitored by RouterOS Netwatch (/tool/netwatch).
type NetwatchHost struct {
	ID         string `json:"id"`
	Host       string `json:"host"`
	Interval   string `json:"interval"`
	Timeout    string `json:"timeout"`
	Status     string `json:"status"`
	Since      string `json:"since"`
	UpScript   string `json:"up_script"`
	DownScript string `json:"down_script"`
	Disabled   bool   `json:"disabled"`
	Comment    string `json:"comment"`
}
