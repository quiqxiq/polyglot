package system

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

func setIfNonEmpty(args map[string]string, key, value string) {
	if value != "" {
		args[key] = value
	}
}

// NewPrintResourceCommand builds the command.Command for /system/resource/print.
func NewPrintResourceCommand() command.Command {
	return command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{},
	}
}

// NewStreamResourceCommand builds the command.Command for streaming /system/resource/print.
func NewStreamResourceCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{"interval": interval},
	}
}

// NewPrintHealthCommand builds the command.Command for /system/health/print.
func NewPrintHealthCommand() command.Command {
	return command.Command{
		Raw:  "/system/health/print",
		Args: map[string]string{},
	}
}

// NewStreamHealthCommand builds the command.Command for streaming /system/health/print.
func NewStreamHealthCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/health/print",
		Args: map[string]string{"interval": interval},
	}
}

// NewPrintClockCommand builds the command.Command for /system/clock/print.
func NewPrintClockCommand() command.Command {
	return command.Command{
		Raw:  "/system/clock/print",
		Args: map[string]string{},
	}
}

// NewStreamClockCommand builds the command.Command for streaming /system/clock/print.
func NewStreamClockCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/clock/print",
		Args: map[string]string{"interval": interval},
	}
}

// NewPrintIdentityCommand builds the command.Command for /system/identity/print.
func NewPrintIdentityCommand() command.Command {
	return command.Command{
		Raw:  "/system/identity/print",
		Args: map[string]string{},
	}
}

// NewStreamIdentityCommand builds the command.Command for streaming /system/identity/print.
func NewStreamIdentityCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/identity/print",
		Args: map[string]string{"interval": interval},
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
		Raw:  "/system/routerboard/print",
		Args: map[string]string{},
	}
}

// NewStreamRouterboardCommand builds the command.Command for streaming /system/routerboard/print.
func NewStreamRouterboardCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/routerboard/print",
		Args: map[string]string{"interval": interval},
	}
}

// NewPrintLogsCommand builds the command.Command for /log/print.
func NewPrintLogsCommand(topicsFilter string) command.Command {
	args := map[string]string{}
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
	args := map[string]string{"follow": ""}
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
	args := map[string]string{}
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
	args := map[string]string{}
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
