package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// SystemSchedulerParams holds parameters for /system/scheduler/add or /set.
type SystemSchedulerParams struct {
	Name      string
	StartTime string
	StartDate string
	Interval  string
	OnEvent   string
	Comment   string
	Disabled  bool
}

// SystemScheduler represents one row from /system/scheduler/print.
type SystemScheduler struct {
	RosID     string
	Name      string
	StartTime string
	StartDate string
	Interval  string
	OnEvent   string
	NextRun   string
	Comment   string
	Disabled  bool
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

// ParseSystemSchedulers converts command.Result rows from /system/scheduler/print.
func ParseSystemSchedulers(result command.Result) []SystemScheduler {
	schedulers := make([]SystemScheduler, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		schedulers = append(schedulers, SystemScheduler{
			RosID:     id,
			Name:      name,
			StartTime: row["start-time"],
			StartDate: row["start-date"],
			Interval:  row["interval"],
			OnEvent:   row["on-event"],
			NextRun:   row["next-run"],
			Comment:   row["comment"],
			Disabled:  strings.EqualFold(row["disabled"], "true"),
		})
	}
	return schedulers
}
