package secscan

import (
	"time"
)

// Severity orders how much a failed check matters.
type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
	Low      Severity = "low"
	Info     Severity = "info"
)

// Severities in order, most severe first.
var Severities = []Severity{Critical, High, Medium, Low, Info}

// Status is a check's evaluation answer.
type Status string

const (
	Pass    Status = "pass"
	Fail    Status = "fail"
	Unknown Status = "unknown"
)

// Row represents one key-value map from RouterOS sentence reply.
type Row = map[string]string

// Inputs holds the raw row data per RouterOS path and tracks absent menus.
type Inputs struct {
	Rows   map[string][]Row
	Absent map[string]bool
	Now    time.Time
}

// Get returns the rows for a given path if present.
func (in Inputs) Get(path string) ([]Row, bool) {
	if in.Absent[path] {
		return nil, false
	}
	rows, ok := in.Rows[path]
	return rows, ok
}

// One returns the single settings row for a given path.
func (in Inputs) One(path string) (Row, bool) {
	rows, ok := in.Get(path)
	if !ok || len(rows) == 0 {
		return nil, false
	}
	return rows[0], true
}

// Check represents one audit rule in the security catalogue.
type Check struct {
	ID       string
	Category string
	Severity Severity
	Title    string
	Why      string
	Fix      string
	Link     string
	Menus    []string
	Eval     func(in Inputs) (Status, []string)
}

// Finding is one check's result.
type Finding struct {
	ID       string   `json:"id"`
	Category string   `json:"category"`
	Severity Severity `json:"severity"`
	Title    string   `json:"title"`
	Status   Status   `json:"status"`
	Detail   []string `json:"detail"`
	Why      string   `json:"why"`
	Fix      string   `json:"fix"`
	Link     string   `json:"link"`
}

// CategoryScore represents summary stats for a single category.
type CategoryScore struct {
	Name    string `json:"name"`
	Score   int    `json:"score"`
	Pass    int    `json:"pass"`
	Fail    int    `json:"fail"`
	Unknown int    `json:"unknown"`
}

// ServiceFact is one management service status.
type ServiceFact struct {
	Name       string `json:"name"`
	Port       string `json:"port"`
	Enabled    bool   `json:"enabled"`
	Restricted bool   `json:"restricted"`
	Plaintext  bool   `json:"plaintext"`
}

// Facts summarizes device attack-surface stats.
type Facts struct {
	Services        []ServiceFact `json:"services"`
	Installed       string        `json:"installed"`
	Latest          string        `json:"latest"`
	FirmwareCurrent string        `json:"firmware_current"`
	FirmwareUpgrade string        `json:"firmware_upgrade"`
	Users           int           `json:"users"`
	FullUsers       int           `json:"full_users"`
	MinPasswordLen  string        `json:"min_password_len"`
	AdminEnabled    bool          `json:"admin_enabled"`
}

// Report holds the full audit report.
type Report struct {
	Score      int             `json:"score"`
	Findings   []Finding       `json:"findings"`
	Categories []CategoryScore `json:"categories"`
	Failed     map[string]int  `json:"failed"`
	Passed     int             `json:"passed"`
	Unknown    int             `json:"unknown"`
	Facts      Facts           `json:"facts"`
}
