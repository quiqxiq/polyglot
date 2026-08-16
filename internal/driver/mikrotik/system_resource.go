package mikrotik

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
)

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

// NewPrintSystemResourceCommand builds the command.Command for /system/resource/print.
func NewPrintSystemResourceCommand() command.Command {
	return command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{},
	}
}

// NewStreamSystemResourceCommand builds the command.Command for streaming /system/resource/print.
func NewStreamSystemResourceCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{"interval": interval},
	}
}

// ParseSystemResource converts a command.Result from /system/resource/print into a SystemResource struct.
func ParseSystemResource(result command.Result) SystemResource {
	if len(result.Rows) == 0 {
		return SystemResource{}
	}
	row := result.Rows[0]
	cpuLoad, _ := strconv.Atoi(row["cpu-load"])
	cpuCount, _ := strconv.Atoi(row["cpu-count"])

	voltage := row["voltage"]
	if voltage == "" {
		voltage = row["board-voltage"]
	}

	return SystemResource{
		CPULoad:       cpuLoad,
		CPUCount:      cpuCount,
		CPUFrequency:  row["cpu-frequency"],
		FreeMemory:    row["free-memory"],
		TotalMemory:   row["total-memory"],
		FreeHDDSpace:  row["free-hdd-space"],
		TotalHDDSpace: row["total-hdd-space"],
		Architecture:  row["architecture-name"],
		Model:         row["model"],
		SerialNumber:  row["serial-number"],
		FirmwareType:  row["current-firmware"],
		Voltage:       voltage,
		Temperature:   row["temperature"],
		BadBlocks:     row["bad-blocks"],
		Uptime:        row["uptime"],
		Version:       row["version"],
		BoardName:     row["board-name"],
	}
}
