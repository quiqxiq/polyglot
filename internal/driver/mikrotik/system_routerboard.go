package mikrotik

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// SystemRouterboard holds the parsed data from /system/routerboard/print.
// It is the vendor-neutral representation of the router's hardware/firmware
// identity, produced by the mikrotik driver and consumed by usecases.
type SystemRouterboard struct {
	BoardName       string
	Model           string
	SerialNumber    string
	FirmwareType    string
	FactoryFirmware string
	CurrentFirmware string
	UpgradeFirmware string
}

// NewPrintSystemRouterboardCommand builds the command.Command for
// /system/routerboard/print (one-shot).
func NewPrintSystemRouterboardCommand() command.Command {
	return command.Command{
		Raw:  "/system/routerboard/print",
		Args: map[string]string{},
	}
}

// NewStreamSystemRouterboardCommand builds the command.Command for streaming
// /system/routerboard/print interval=<n>, which makes RouterOS re-send the
// row periodically. interval defaults to "1s".
func NewStreamSystemRouterboardCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/routerboard/print",
		Args: map[string]string{"interval": interval},
	}
}

// ParseSystemRouterboard converts a command.Result from
// /system/routerboard/print into a SystemRouterboard struct.
func ParseSystemRouterboard(result command.Result) SystemRouterboard {
	if len(result.Rows) == 0 {
		return SystemRouterboard{}
	}
	row := result.Rows[0]
	return SystemRouterboard{
		BoardName:       row["board-name"],
		Model:           row["model"],
		SerialNumber:    row["serial-number"],
		FirmwareType:    row["firmware-type"],
		FactoryFirmware: row["factory-firmware"],
		CurrentFirmware: row["current-firmware"],
		UpgradeFirmware: row["upgrade-firmware"],
	}
}
