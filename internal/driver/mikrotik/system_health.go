package mikrotik

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// SystemHealth holds the parsed data from /system/health/print. Sensor
// availability depends on the board — fields that the router does not report
// are left empty.
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

// NewPrintSystemHealthCommand builds the command.Command for
// /system/health/print (one-shot).
func NewPrintSystemHealthCommand() command.Command {
	return command.Command{
		Raw:  "/system/health/print",
		Args: map[string]string{},
	}
}

// NewStreamSystemHealthCommand builds the command.Command for streaming
// /system/health/print interval=<n>, which makes RouterOS re-send the row
// periodically. interval defaults to "1s".
func NewStreamSystemHealthCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/health/print",
		Args: map[string]string{"interval": interval},
	}
}

// ParseSystemHealth converts a command.Result from /system/health/print
// into a SystemHealth struct.
func ParseSystemHealth(result command.Result) SystemHealth {
	if len(result.Rows) == 0 {
		return SystemHealth{}
	}
	row := result.Rows[0]
	return SystemHealth{
		Voltage:        row["voltage"],
		Temperature:    row["temperature"],
		CPUTemperature: row["cpu-temperature"],
		PSUVoltage:     row["psu-voltage"],
		PSUCurrent:     row["psu-current"],
		PSUTemperature: row["psu-temperature"],
		Fan1Speed:      row["fan1-speed"],
		Fan2Speed:      row["fan2-speed"],
	}
}
