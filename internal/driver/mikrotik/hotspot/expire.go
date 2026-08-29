package hotspot

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/system"
	"github.com/quixiq/polyglot/internal/port"
)

// MikhmonExpireMonitorName is the standard RouterOS scheduler name used by Mikhmon.
const MikhmonExpireMonitorName = "Mikhmon-Expire-Monitor"

// mikhmonExpireScriptName and mikhmonExpireSchedulerName are the names used
// by the gateway's two-step expire-monitor setup (script + scheduler).
const (
	mikhmonExpireScriptName    = "mikhmon-expire-monitor"
	mikhmonExpireSchedulerName = "mikhmon-expire-scheduler"
)

// ExpireMonitorSchedulerNames lists the scheduler names recognised as the
// Mikhmon expire monitor — the legacy single-step form and the gateway
// two-step form (Fase 6 decision A).
var ExpireMonitorSchedulerNames = []string{
	MikhmonExpireMonitorName,
	mikhmonExpireSchedulerName,
}

// classifyExpireMonitorSchedulers maps raw scheduler rows to an
// ExpireMonitorStatus, preferring the legacy scheduler when both forms are
// present. Returns zero status (not installed) when no name matches.
func classifyExpireMonitorSchedulers(schedulers []system.SystemScheduler) port.ExpireMonitorStatus {
	for _, s := range schedulers {
		for _, name := range ExpireMonitorSchedulerNames {
			if s.Name == name {
				return port.ExpireMonitorStatus{
					IsInstalled:   true,
					IsEnabled:     !s.Disabled,
					SchedulerID:   s.RosID,
					SchedulerName: s.Name,
				}
			}
		}
	}
	return port.ExpireMonitorStatus{}
}

// BuildExpireMonitorScript generates the RouterOS script source for Mikhmon v4's
// automated expire monitor. When executed periodically (e.g. every 1 minute via scheduler),
// this script:
// 1. Converts currentDate and currentTime to integer numbers.
// 2. Finds all Hotspot users with a comment containing the current or previous year.
// 3. Parses the expiry date/time and mode ('N' or 'X') from the user's comment.
// 4. If expired:
//   - Mode 'N' (Notify): sets limit-uptime=1s and disconnects active session.
//   - Mode 'X' (Remove): deletes user completely and disconnects active session.
func BuildExpireMonitorScript() string {
	var sb strings.Builder
	sb.WriteString(":local dateint do={\n")
	sb.WriteString("    :local montharray ( \"jan\",\"feb\",\"mar\",\"apr\",\"may\",\"jun\",\"jul\",\"aug\",\"sep\",\"oct\",\"nov\",\"dec\" );\n")
	sb.WriteString("    :local days [ :pick $d 4 6 ];\n")
	sb.WriteString("    :local month [ :pick $d 0 3 ];\n")
	sb.WriteString("    :local year [ :pick $d 7 11 ];\n")
	sb.WriteString("    :local monthint ([ :find $montharray $month]);\n")
	sb.WriteString("    :local month ($monthint + 1);\n")
	sb.WriteString("    :if ( [len $month] = 1) do={\n")
	sb.WriteString("        :local zero (\"0\");\n")
	sb.WriteString("        :return [:tonum (\"$year$zero$month$days\")];\n")
	sb.WriteString("    } else={\n")
	sb.WriteString("        :return [:tonum (\"$year$month$days\")];\n")
	sb.WriteString("    }\n")
	sb.WriteString("};\n")

	sb.WriteString(":local timeint do={\n")
	sb.WriteString("    :local hours [ :pick $t 0 2 ];\n")
	sb.WriteString("    :local minutes [ :pick $t 3 5 ];\n")
	sb.WriteString("    :return ($hours * 60 + $minutes);\n")
	sb.WriteString("};\n")

	sb.WriteString(":local date [ /system clock get date ];\n")
	sb.WriteString(":local time [ /system clock get time ];\n")
	sb.WriteString(":local today [$dateint d=$date];\n")
	sb.WriteString(":local curtime [$timeint t=$time];\n")
	sb.WriteString(":local tyear [ :pick $date 7 11 ];\n")
	sb.WriteString(":local lyear ($tyear-1);\n")

	sb.WriteString(":foreach i in [ /ip hotspot user find where comment~\"/$tyear\" || comment~\"/$lyear\" ] do={\n")
	sb.WriteString("    :local comment [ /ip hotspot user get $i comment];\n")
	sb.WriteString("    :local limit [ /ip hotspot user get $i limit-uptime];\n")
	sb.WriteString("    :local name [ /ip hotspot user get $i name];\n")
	sb.WriteString("    :local gettime [:pic $comment 12 20];\n")
	sb.WriteString("    :if ([:pic $comment 3] = \"/\" and [:pic $comment 6] = \"/\") do={\n")
	sb.WriteString("        :local expd [$dateint d=$comment];\n")
	sb.WriteString("        :local expt [$timeint t=$gettime];\n")
	sb.WriteString("        :if (($expd < $today and $expt < $curtime) or ($expd < $today and $expt > $curtime) or ($expd = $today and $expt < $curtime) and $limit != \"00:00:01\") do={\n")
	sb.WriteString("            :if ([:pic $comment 21] = \"N\") do={\n")
	sb.WriteString("                [ /ip hotspot user set limit-uptime=1s $i ];\n")
	sb.WriteString("                [ /ip hotspot active remove [find where user=$name] ];\n")
	sb.WriteString("            } else={\n")
	sb.WriteString("                [ /ip hotspot user remove $i ];\n")
	sb.WriteString("                [ /ip hotspot active remove [find where user=$name] ];\n")
	sb.WriteString("            }\n")
	sb.WriteString("        }\n")
	sb.WriteString("    }\n")
	sb.WriteString("};\n")

	return sb.String()
}

// NewSetupMikhmonExpireMonitorCommand builds the command.Command for /system/scheduler/add
// installing the Mikhmon-Expire-Monitor scheduler on the router.
// Default interval is "00:01:00" (every 1 minute).
func NewSetupMikhmonExpireMonitorCommand(interval string) command.Command {
	if interval == "" {
		interval = "00:01:00"
	}
	script := BuildExpireMonitorScript()
	return system.NewAddSchedulerCommand(system.SystemSchedulerParams{
		Name:      MikhmonExpireMonitorName,
		StartTime: "00:00:00",
		Interval:  interval,
		OnEvent:   script,
		Comment:   "Mikhmon Expire Monitor",
		Disabled:  false,
	})
}

// NewUpdateMikhmonExpireMonitorCommand builds the command.Command for /system/scheduler/set
// updating an existing Mikhmon-Expire-Monitor scheduler on the router.
func NewUpdateMikhmonExpireMonitorCommand(rosID, interval string) command.Command {
	if interval == "" {
		interval = "00:01:00"
	}
	script := BuildExpireMonitorScript()
	return system.NewSetSchedulerCommand(rosID, system.SystemSchedulerParams{
		Name:     MikhmonExpireMonitorName,
		Interval: interval,
		OnEvent:  script,
		Disabled: false,
	})
}
