package queue

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// SuspendQueueParams returns a pre-filled SimpleQueueParams for soft-suspension
// of a subscriber via a near-zero rate limit.
func SuspendQueueParams(targetIP, reason, throttleSpeed string) SimpleQueueParams {
	speed := throttleSpeed
	if speed == "" {
		speed = "1k/1k"
	}
	name := "suspended_" + strings.ReplaceAll(targetIP, ".", "_")
	return SimpleQueueParams{
		Name:     name,
		Target:   targetIP,
		MaxLimit: speed,
		Comment:  "SUSPENDED - " + reason,
		Disabled: false,
	}
}

// NewPrintSimpleQueuesCommand builds the command.Command for /queue/simple/print.
func NewPrintSimpleQueuesCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/queue/simple/print",
		Args: args,
	}
}

// NewPrintParentQueuesCommand builds the command.Command for /queue/simple/print
// with filter ?dynamic=false.
func NewPrintParentQueuesCommand() command.Command {
	return command.Command{
		Raw:  "/queue/simple/print",
		Args: map[string]string{"?dynamic": "false"},
	}
}

// NewStreamQueueStatsCommand builds the command.Command for streaming /queue/simple/print.
func NewStreamQueueStatsCommand(p QueueStreamParams) command.Command {
	interval := p.Interval
	if interval == "" {
		interval = "1s"
	}
	args := map[string]string{
		"stats":    "",
		"interval": interval,
	}
	if p.NameFilter != "" {
		args["?name"] = p.NameFilter
	}
	if p.ParentFilter != "" {
		args["?parent"] = p.ParentFilter
	}
	if p.ParentsOnly {
		args["?dynamic"] = "false"
	}
	return command.Command{
		Raw:  "/queue/simple/print",
		Args: args,
	}
}

// NewAddSimpleQueueCommand builds the command.Command for /queue/simple/add.
func NewAddSimpleQueueCommand(p SimpleQueueParams) command.Command {
	args := map[string]string{
		"name":   p.Name,
		"target": p.Target,
	}
	if p.MaxLimit != "" {
		args["max-limit"] = p.MaxLimit
	}
	if p.LimitAt != "" {
		args["limit-at"] = p.LimitAt
	}
	if p.BurstLimit != "" {
		args["burst-limit"] = p.BurstLimit
	}
	if p.BurstThreshold != "" {
		args["burst-threshold"] = p.BurstThreshold
	}
	if p.BurstTime != "" {
		args["burst-time"] = p.BurstTime
	}
	if p.Priority != "" {
		args["priority"] = p.Priority
	}
	if p.Parent != "" {
		args["parent"] = p.Parent
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{Raw: "/queue/simple/add", Args: args}
}

// DedicatedQueueToROS builds /queue/simple/add from typed params.
func DedicatedQueueToROS(p DedicatedQueueParams) command.Command {
	args := map[string]string{
		"name":      p.Name,
		"target":    p.Target,
		"max-limit": p.MaxLimit,
		"limit-at":  p.LimitAt,
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	if p.BurstLimit != "" {
		args["burst-limit"] = p.BurstLimit
	}
	if p.BurstThreshold != "" {
		args["burst-threshold"] = p.BurstThreshold
	}
	if p.BurstTime != "" {
		args["burst-time"] = p.BurstTime
	}
	return command.Command{Raw: "/queue/simple/add", Args: args}
}

// NewRemoveSimpleQueueCommand builds the command.Command for /queue/simple/remove.
func NewRemoveSimpleQueueCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/queue/simple/remove",
		Args: map[string]string{".id": rosID},
	}
}

// NewSetSimpleQueueEnabledCommand builds the command.Command for /queue/simple/set enabled/disabled.
func NewSetSimpleQueueEnabledCommand(rosID string, enabled bool) command.Command {
	val := "no"
	if !enabled {
		val = "yes"
	}
	return command.Command{
		Raw:  "/queue/simple/set",
		Args: map[string]string{"numbers": rosID, "disabled": val},
	}
}
