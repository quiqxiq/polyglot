package mikrotik

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

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

// NewPrintSystemIdentityCommand builds the command.Command for /system/identity/print.
func NewPrintSystemIdentityCommand() command.Command {
	return command.Command{
		Raw:  "/system/identity/print",
		Args: map[string]string{},
	}
}

// NewSetSystemIdentityCommand builds the command.Command for /system/identity/set.
func NewSetSystemIdentityCommand(name string) command.Command {
	return command.Command{
		Raw:  "/system/identity/set",
		Args: map[string]string{"name": name},
	}
}

// ParseSystemIdentity converts a command.Result from /system/identity/print into a SystemIdentity struct.
func ParseSystemIdentity(result command.Result) SystemIdentity {
	if len(result.Rows) == 0 {
		return SystemIdentity{}
	}
	return SystemIdentity{Name: result.Rows[0]["name"]}
}

// NewPrintSystemClockCommand builds the command.Command for /system/clock/print.
func NewPrintSystemClockCommand() command.Command {
	return command.Command{
		Raw:  "/system/clock/print",
		Args: map[string]string{},
	}
}

// ParseSystemClock converts a command.Result from /system/clock/print into a SystemClock struct.
func ParseSystemClock(result command.Result) SystemClock {
	if len(result.Rows) == 0 {
		return SystemClock{}
	}
	row := result.Rows[0]
	return SystemClock{
		Time:         row["time"],
		Date:         row["date"],
		TimeZoneName: row["time-zone-name"],
		GMTOffset:    row["gmt-offset"],
	}
}
