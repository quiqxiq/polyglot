package system

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// LicenseInfo represents RouterOS license level details.
type LicenseInfo struct {
	SystemID   string `json:"system_id"`
	SoftwareID string `json:"software_id"`
	Level      string `json:"level"`
	NoLevel    bool   `json:"no_level"`
}

// DeviceMode holds RouterOS device mode and permitted capabilities.
type DeviceMode struct {
	Mode             string `json:"mode"`
	ContainerAllowed bool   `json:"container_allowed"`
	SnifferAllowed   bool   `json:"sniffer_allowed"`
}

// ResetConfigParams holds parameters for /system/reset-configuration.
type ResetConfigParams struct {
	NoDefaults    bool
	SkipBackup    bool
	KeepUsers     bool
	RunAfterReset string
}

// NewRebootCommand builds a command to reboot the router.
func NewRebootCommand() command.Command {
	return command.Command{
		Raw:  "/system/reboot",
		Args: map[string]string{},
	}
}

// NewResetConfigurationCommand builds a command to factory reset router configuration.
func NewResetConfigurationCommand(params ResetConfigParams) command.Command {
	args := make(map[string]string)
	if params.NoDefaults {
		args["no-defaults"] = "yes"
	}
	if params.SkipBackup {
		args["skip-backup"] = "yes"
	}
	if params.KeepUsers {
		args["keep-users"] = "yes"
	}
	if params.RunAfterReset != "" {
		args["run-after-reset"] = params.RunAfterReset
	}
	return command.Command{
		Raw:  "/system/reset-configuration",
		Args: args,
	}
}

// NewPrintLicenseCommand queries RouterOS system license details.
func NewPrintLicenseCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "system-id", "software-id", "level", "nolevel")
	return command.Command{
		Raw:  "/system/license/print",
		Args: args,
	}
}

// NewPrintDeviceModeCommand queries current device mode restrictions.
func NewPrintDeviceModeCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "mode", "container", "sniffer")
	return command.Command{
		Raw:  "/system/device-mode/print",
		Args: args,
	}
}

// ParseLicenseInfo converts reply rows into LicenseInfo.
func ParseLicenseInfo(rows []map[string]string) *LicenseInfo {
	if len(rows) == 0 {
		return nil
	}
	s := rows[0]
	return &LicenseInfo{
		SystemID:   s["system-id"],
		SoftwareID: s["software-id"],
		Level:      s["level"],
		NoLevel:    s["nolevel"] == "true" || s["nolevel"] == "yes",
	}
}

// ParseDeviceMode converts reply rows into DeviceMode.
func ParseDeviceMode(rows []map[string]string) *DeviceMode {
	if len(rows) == 0 {
		return nil
	}
	s := rows[0]
	return &DeviceMode{
		Mode:             s["mode"],
		ContainerAllowed: s["container"] != "false" && s["container"] != "no",
		SnifferAllowed:   s["sniffer"] != "false" && s["sniffer"] != "no",
	}
}
