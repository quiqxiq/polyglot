package mikrotik

import "github.com/quixiq/polyglot/internal/domain/command"

// NewPrintSystemResourceCommand builds the command.Command for /system/resource/print.
func NewPrintSystemResourceCommand() command.Command {
	return command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{},
	}
}

// NewStreamSystemResourceCommand builds the command.Command for /system/resource/print with interval.
func NewStreamSystemResourceCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{"interval": interval},
	}
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

// NewPrintSystemClockCommand builds the command.Command for /system/clock/print.
func NewPrintSystemClockCommand() command.Command {
	return command.Command{
		Raw:  "/system/clock/print",
		Args: map[string]string{},
	}
}

// NewPrintLogCommand builds the command.Command for /log/print.
func NewPrintLogCommand(topicsFilter string) command.Command {
	args := map[string]string{}
	if topicsFilter != "" {
		args["?topics~"+topicsFilter] = ""
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

// NewPrintSystemSchedulersCommand builds the command.Command for /system/scheduler/print.
func NewPrintSystemSchedulersCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/system/scheduler/print",
		Args: args,
	}
}

// NewAddSystemSchedulerCommand builds the command.Command for /system/scheduler/add.
func NewAddSystemSchedulerCommand(p SystemSchedulerParams) command.Command {
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

// NewSetSystemSchedulerCommand builds the command.Command for /system/scheduler/set.
func NewSetSystemSchedulerCommand(rosID string, p SystemSchedulerParams) command.Command {
	args := map[string]string{".id": rosID}
	setIfNonEmpty(args, "name", p.Name)
	setIfNonEmpty(args, "start-time", p.StartTime)
	setIfNonEmpty(args, "start-date", p.StartDate)
	setIfNonEmpty(args, "interval", p.Interval)
	setIfNonEmpty(args, "on-event", p.OnEvent)
	setIfNonEmpty(args, "comment", p.Comment)
	return command.Command{Raw: "/system/scheduler/set", Args: args}
}

// NewPrintSystemScriptsCommand builds the command.Command for /system/script/print.
func NewPrintSystemScriptsCommand(ownerFilter, commentFilter string) command.Command {
	args := map[string]string{}
	if ownerFilter != "" {
		args["?owner"] = ownerFilter
	}
	if commentFilter != "" {
		args["?comment"] = commentFilter
	}
	return command.Command{
		Raw:  "/system/script/print",
		Args: args,
	}
}

// NewRemoveSystemScriptCommand builds the command.Command for /system/script/remove.
func NewRemoveSystemScriptCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/system/script/remove",
		Args: map[string]string{".id": rosID},
	}
}
