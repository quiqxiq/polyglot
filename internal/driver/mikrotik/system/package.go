package system

import (
	"fmt"
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// PackageEntry represents one installed RouterOS package.
type PackageEntry struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	Scheduled string `json:"scheduled"`
	Size      int64  `json:"size"`
	Disabled  bool   `json:"disabled"`
}

// PackageUpdate represents the latest RouterOS version status.
type PackageUpdate struct {
	InstalledVersion string `json:"installed_version"`
	LatestVersion    string `json:"latest_version"`
	Status           string `json:"status"`
	Channel          string `json:"channel"`
}

// RouterBoardSettings holds RouterBOARD boot and firmware settings.
type RouterBoardSettings struct {
	AutoUpgrade  bool   `json:"auto_upgrade"`
	BootDevice   string `json:"boot_device"`
	CPUFrequency string `json:"cpu_frequency"`
}

// NewPrintPackagesCommand builds a command to query installed packages with strict proplist.
func NewPrintPackagesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "version", "build-time", "scheduled", "size", "disabled")
	return command.Command{
		Raw:  "/system/package/print",
		Args: args,
	}
}

// NewPrintPackageUpdateCommand queries current update availability status.
func NewPrintPackageUpdateCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "installed-version", "latest-version", "status", "channel")
	return command.Command{
		Raw:  "/system/package/update/print",
		Args: args,
	}
}

// NewCheckPackageUpdatesCommand requests the router to check MikroTik upgrade servers.
func NewCheckPackageUpdatesCommand() command.Command {
	return command.Command{
		Raw:  "/system/package/update/check-for-updates",
		Args: map[string]string{},
	}
}

// NewInstallPackageUpdateCommand triggers download and install of available updates.
func NewInstallPackageUpdateCommand() command.Command {
	return command.Command{
		Raw:  "/system/package/update/install",
		Args: map[string]string{},
	}
}

// NewSetPackageUpdateChannelCommand changes update channel (stable, testing, development, long-term).
func NewSetPackageUpdateChannelCommand(channel string) command.Command {
	return command.Command{
		Raw: "/system/package/update/set",
		Args: map[string]string{
			"channel": channel,
		},
	}
}

// NewPackageScheduleCommand schedules enable, disable, uninstall, or unschedule for a package.
func NewPackageScheduleCommand(action, idOrNumbers string) (command.Command, error) {
	var raw string
	switch action {
	case "enable":
		raw = "/system/package/enable"
	case "disable":
		raw = "/system/package/disable"
	case "uninstall":
		raw = "/system/package/uninstall"
	case "unschedule":
		raw = "/system/package/unschedule"
	default:
		return command.Command{}, fmt.Errorf("unknown package action: %s", action)
	}

	return command.Command{
		Raw: raw,
		Args: map[string]string{
			"numbers": idOrNumbers,
		},
	}, nil
}

// NewUpgradeRouterBoardCommand builds command to upgrade RouterBOARD bootloader firmware.
func NewUpgradeRouterBoardCommand() command.Command {
	return command.Command{
		Raw:  "/system/routerboard/upgrade",
		Args: map[string]string{},
	}
}

// NewPrintRouterBoardSettingsCommand queries RouterBOARD hardware settings.
func NewPrintRouterBoardSettingsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "auto-upgrade", "boot-device", "cpu-frequency")
	return command.Command{
		Raw:  "/system/routerboard/settings/print",
		Args: args,
	}
}

// NewSetRouterBoardSettingsCommand updates RouterBOARD auto-upgrade toggle.
func NewSetRouterBoardSettingsCommand(autoUpgrade bool) command.Command {
	val := "no"
	if autoUpgrade {
		val = "yes"
	}
	return command.Command{
		Raw: "/system/routerboard/settings/set",
		Args: map[string]string{
			"auto-upgrade": val,
		},
	}
}

// ParsePackages converts RouterOS reply rows into PackageEntry slice.
func ParsePackages(rows []map[string]string) []PackageEntry {
	out := make([]PackageEntry, 0, len(rows))
	for _, r := range rows {
		sz, _ := strconv.ParseInt(r["size"], 10, 64)
		out = append(out, PackageEntry{
			ID:        r[".id"],
			Name:      r["name"],
			Version:   r["version"],
			BuildTime: r["build-time"],
			Scheduled: r["scheduled"],
			Size:      sz,
			Disabled:  r["disabled"] == "true",
		})
	}
	return out
}

// ParsePackageUpdate converts RouterOS reply rows into PackageUpdate struct.
func ParsePackageUpdate(rows []map[string]string) *PackageUpdate {
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &PackageUpdate{
		InstalledVersion: r["installed-version"],
		LatestVersion:    r["latest-version"],
		Status:           r["status"],
		Channel:          r["channel"],
	}
}

// ParseRouterBoardSettings converts reply rows into RouterBoardSettings.
func ParseRouterBoardSettings(rows []map[string]string) *RouterBoardSettings {
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &RouterBoardSettings{
		AutoUpgrade:  r["auto-upgrade"] == "yes" || r["auto-upgrade"] == "true",
		BootDevice:   r["boot-device"],
		CPUFrequency: r["cpu-frequency"],
	}
}
