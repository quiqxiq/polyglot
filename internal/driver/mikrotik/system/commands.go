package system

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

var setIfNonEmpty = rosutil.SetIfNonEmpty

// NewPrintResourceCommand builds the command.Command for /system/resource/print.
func NewPrintResourceCommand() command.Command {
	return command.Command{
		Raw: "/system/resource/print",
		Args: map[string]string{
			".proplist": "cpu-load,cpu-count,cpu-frequency,free-memory,total-memory,free-hdd-space,total-hdd-space,architecture-name,board-name,version,uptime,model,serial-number,current-firmware,voltage,board-voltage,temperature,bad-blocks",
		},
	}
}

// NewStreamResourceCommand builds the command.Command for streaming /system/resource/print.
func NewStreamResourceCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw: "/system/resource/print",
		Args: map[string]string{
			"interval":  interval,
			".proplist": "cpu-load,cpu-count,cpu-frequency,free-memory,total-memory,free-hdd-space,total-hdd-space,architecture-name,board-name,version,uptime,model,serial-number,current-firmware,voltage,board-voltage,temperature,bad-blocks",
		},
	}
}

// NewPrintHealthCommand builds the command.Command for /system/health/print.
func NewPrintHealthCommand() command.Command {
	return command.Command{
		Raw: "/system/health/print",
		Args: map[string]string{
			".proplist": ".id,name,value,type,voltage,temperature,cpu-temperature,psu-voltage,psu-current,psu-temperature,fan1-speed,fan2-speed",
		},
	}
}

// NewStreamHealthCommand builds the command.Command for streaming /system/health/print.
func NewStreamHealthCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw: "/system/health/print",
		Args: map[string]string{
			"interval":  interval,
			".proplist": ".id,name,value,type,voltage,temperature,cpu-temperature,psu-voltage,psu-current,psu-temperature,fan1-speed,fan2-speed",
		},
	}
}

// NewPrintClockCommand builds the command.Command for /system/clock/print.
func NewPrintClockCommand() command.Command {
	return command.Command{
		Raw: "/system/clock/print",
		Args: map[string]string{
			".proplist": "time,date,time-zone-name,gmt-offset,dst-active",
		},
	}
}

// NewStreamClockCommand builds the command.Command for streaming /system/clock/print.
func NewStreamClockCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw: "/system/clock/print",
		Args: map[string]string{
			"interval":  interval,
			".proplist": "time,date,time-zone-name,gmt-offset,dst-active",
		},
	}
}

// NewPrintIdentityCommand builds the command.Command for /system/identity/print.
func NewPrintIdentityCommand() command.Command {
	return command.Command{
		Raw: "/system/identity/print",
		Args: map[string]string{
			".proplist": "name",
		},
	}
}

// NewStreamIdentityCommand builds the command.Command for streaming /system/identity/print.
func NewStreamIdentityCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw: "/system/identity/print",
		Args: map[string]string{
			"interval":  interval,
			".proplist": "name",
		},
	}
}

// NewSetIdentityCommand builds the command.Command for /system/identity/set.
func NewSetIdentityCommand(name string) command.Command {
	return command.Command{
		Raw:  "/system/identity/set",
		Args: map[string]string{"name": name},
	}
}

// NewPrintRouterboardCommand builds the command.Command for /system/routerboard/print.
func NewPrintRouterboardCommand() command.Command {
	return command.Command{
		Raw: "/system/routerboard/print",
		Args: map[string]string{
			".proplist": "routerboard,model,serial-number,current-firmware,upgrade-firmware",
		},
	}
}

// NewStreamRouterboardCommand builds the command.Command for streaming /system/routerboard/print.
func NewStreamRouterboardCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw: "/system/routerboard/print",
		Args: map[string]string{
			"interval":  interval,
			".proplist": "routerboard,model,serial-number,current-firmware,upgrade-firmware",
		},
	}
}

// NewPrintLogsCommand builds the command.Command for /log/print.
func NewPrintLogsCommand(topicsFilter string) command.Command {
	args := map[string]string{
		".proplist": "time,topics,message",
	}
	if topicsFilter != "" {
		args["?topics~"] = topicsFilter
	}
	return command.Command{
		Raw:  "/log/print",
		Args: args,
	}
}

// NewStreamLogsCommand builds the command.Command for streaming /log/print follow.
func NewStreamLogsCommand(topicsFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": "time,topics,message",
	}
	if topicsFilter != "" {
		args["?topics~"] = topicsFilter
	}
	return command.Command{
		Raw:  "/log/print",
		Args: args,
	}
}

// NewPingCommand builds the command.Command for /ping.
func NewPingCommand(host, count string) command.Command {
	if count == "" {
		count = "4"
	}
	return command.Command{
		Raw: "/ping",
		Args: map[string]string{
			"address": host,
			"count":   count,
		},
	}
}

// NewPingStreamCommand builds the command.Command for continuous /ping streaming.
func NewPingStreamCommand(host string) command.Command {
	return command.Command{
		Raw: "/ping",
		Args: map[string]string{
			"address":  host,
			"interval": "1s",
		},
	}
}

// NewPrintSchedulerCommand builds the command.Command for /system/scheduler/print.
func NewPrintSchedulerCommand(nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,on-event,start-time,start-date,interval,comment,disabled,run-count,next-run",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/system/scheduler/print",
		Args: args,
	}
}

// NewAddSchedulerCommand builds the command.Command for /system/scheduler/add.
func NewAddSchedulerCommand(p SystemSchedulerParams) command.Command {
	args := map[string]string{
		"name":     p.Name,
		"on-event": p.OnEvent,
	}
	setIfNonEmpty(args, "start-time", p.StartTime)
	setIfNonEmpty(args, "start-date", p.StartDate)
	setIfNonEmpty(args, "interval", p.Interval)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{Raw: "/system/scheduler/add", Args: args}
}

// NewSetSchedulerCommand builds the command.Command for /system/scheduler/set.
func NewSetSchedulerCommand(rosID string, p SystemSchedulerParams) command.Command {
	args := map[string]string{".id": rosID}
	setIfNonEmpty(args, "name", p.Name)
	setIfNonEmpty(args, "start-time", p.StartTime)
	setIfNonEmpty(args, "start-date", p.StartDate)
	setIfNonEmpty(args, "interval", p.Interval)
	setIfNonEmpty(args, "on-event", p.OnEvent)
	setIfNonEmpty(args, "comment", p.Comment)
	return command.Command{Raw: "/system/scheduler/set", Args: args}
}

// NewRemoveSchedulerCommand builds the command.Command for /system/scheduler/remove.
func NewRemoveSchedulerCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/system/scheduler/remove",
		Args: map[string]string{".id": rosID},
	}
}

// NewPrintScriptCommand builds the command.Command for /system/script/print.
func NewPrintScriptCommand(nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,source,comment,dont-require-permissions,run-count,last-started",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/system/script/print",
		Args: args,
	}
}

// NewAddScriptCommand builds the command.Command for /system/script/add.
func NewAddScriptCommand(p SystemScriptParams) command.Command {
	args := map[string]string{
		"name":   p.Name,
		"source": p.Source,
	}
	setIfNonEmpty(args, "comment", p.Comment)
	if p.DontReqPwr {
		args["dont-require-permissions"] = "yes"
	}
	return command.Command{Raw: "/system/script/add", Args: args}
}

// NewSetScriptCommand builds the command.Command for /system/script/set.
func NewSetScriptCommand(rosID string, p SystemScriptParams) command.Command {
	args := map[string]string{".id": rosID}
	setIfNonEmpty(args, "name", p.Name)
	setIfNonEmpty(args, "source", p.Source)
	setIfNonEmpty(args, "comment", p.Comment)
	return command.Command{Raw: "/system/script/set", Args: args}
}

// NewRemoveScriptCommand builds the command.Command for /system/script/remove.
func NewRemoveScriptCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/system/script/remove",
		Args: map[string]string{".id": rosID},
	}
}

// NewRunScriptCommand builds the command.Command for /system/script/run.
func NewRunScriptCommand(scriptName string) command.Command {
	return command.Command{
		Raw:  "/system/script/run",
		Args: map[string]string{"number": scriptName},
	}
}
