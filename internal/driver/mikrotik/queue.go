package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// SimpleQueueParams holds the parameters for creating a RouterOS simple queue
// (/queue/simple/add). Simple queues are the primary bandwidth-shaping
// mechanism used in ISP billing — both for normal subscriber limits and
// for soft-suspend (throttle to near-zero speed instead of hard-block).
//
// Field notes (from RouterOS /queue/simple reference):
//   - Name        : queue name. Convention: use "sub-<username>" for subscriber
//                   queues and "suspended_<ip_underscored>" for suspend queues.
//   - Target      : IP address, subnet, or PPPoE interface targeted by this queue.
//                   For PPPoE subscribers this is typically the username
//                   (RouterOS matches the PPP interface automatically).
//   - MaxLimit    : maximum allowed rate in RouterOS format "rx/tx"
//                   (e.g. "10M/10M", "1k/1k" for suspend). This is the hard cap.
//   - LimitAt     : CIR (Committed Information Rate) — guaranteed minimum rate.
//                   Leave empty to match MaxLimit. Format same as MaxLimit.
//   - BurstLimit  : burst ceiling rate, e.g. "20M/20M". Empty = no burst.
//   - BurstThreshold: traffic level at which burst kicks in, e.g. "8M/8M".
//   - BurstTime   : burst duration window, e.g. "8s". RouterOS default is "8s".
//   - Priority    : 1–8 (1 = highest). Leave empty for default (8).
//   - Parent      : parent queue name for hierarchical (PCQ) shaping.
//   - Comment     : free-text label. Convention for suspend queues:
//                   "SUSPENDED - <reason> - <timestamp>".
//   - Disabled    : when true the queue entry exists but does not shape traffic.
type SimpleQueueParams struct {
	Name            string
	Target          string
	MaxLimit        string
	LimitAt         string
	BurstLimit      string
	BurstThreshold  string
	BurstTime       string
	Priority        string
	Parent          string
	Comment         string
	Disabled        bool
}

// SimpleQueue represents one row returned by /queue/simple/print.
//
// Counter fields (Bytes, Packets, Dropped, Rate, PacketRate) are runtime
// statistics — they reset on reboot and are useful for monitoring dashboards.
type SimpleQueue struct {
	RosID          string
	Name           string
	Target         string
	Parent         string
	MaxLimit       string
	LimitAt        string
	BurstLimit     string
	BurstThreshold string
	BurstTime      string
	Priority       string
	Queue          string // queue type: pfifo, fq_codel, etc.
	Bytes          string // cumulative bytes "rx/tx"
	Packets        string
	Dropped        string
	Rate           string // current rate bps "rx/tx"
	PacketRate     string
	Comment        string
	Disabled       bool
}

// SuspendQueueParams returns a pre-filled SimpleQueueParams for soft-suspension
// of a subscriber via a near-zero rate limit. The subscriber keeps network
// connectivity but at unusable speed — used instead of hard-block for
// subscribers on static IP or whose PPPoE profile cannot be changed.
//
// throttleSpeed is a RouterOS rate string for the limit (e.g. "1k" → 1 kbit/s).
// Defaults to "1k/1k" if empty — matches the billing app convention.
func SuspendQueueParams(targetIP, reason, throttleSpeed string) SimpleQueueParams {
	speed := throttleSpeed
	if speed == "" {
		speed = "1k/1k"
	}
	// IP underscored name convention from MIKROTIK-COMMAND.md §15
	name := "suspended_" + strings.ReplaceAll(targetIP, ".", "_")
	return SimpleQueueParams{
		Name:     name,
		Target:   targetIP,
		MaxLimit: speed,
		Comment:  "SUSPENDED - " + reason,
		Disabled: false,
	}
}

// NewPrintSimpleQueuesCommand builds the command.Command for
// /queue/simple/print. Pass a non-empty nameFilter to look up by queue name.
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
// with filter ?dynamic=false. Mikhmon uses non-dynamic simple queues as parent
// queue options for Hotspot user profile traffic shaping.
func NewPrintParentQueuesCommand() command.Command {
	return command.Command{
		Raw:  "/queue/simple/print",
		Args: map[string]string{"?dynamic": "false"},
	}
}

// QueueStreamParams defines filter and duration parameters for streaming queue statistics.
type QueueStreamParams struct {
	NameFilter   string // filter by queue name, e.g. "sub-budi"
	ParentFilter string // filter by parent queue name, e.g. "parent-total"
	ParentsOnly  bool   // filter non-dynamic static parent queues (?dynamic=false)
	Interval     string // RouterOS duration string (e.g. "1s", "500ms"). Defaults to "1s"
}

// NewStreamQueueStatsCommand builds the command.Command for
// /queue/simple/print stats interval=<N>, with optional filters by queue name,
// parent queue, or static parent queues only (?dynamic=false).
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
	setIfNonEmpty(args, "max-limit", p.MaxLimit)
	setIfNonEmpty(args, "limit-at", p.LimitAt)
	setIfNonEmpty(args, "burst-limit", p.BurstLimit)
	setIfNonEmpty(args, "burst-threshold", p.BurstThreshold)
	setIfNonEmpty(args, "burst-time", p.BurstTime)
	setIfNonEmpty(args, "priority", p.Priority)
	setIfNonEmpty(args, "parent", p.Parent)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{Raw: "/queue/simple/add", Args: args}
}

// NewRemoveSimpleQueueCommand builds the command.Command for
// /queue/simple/remove. Classified as ClassDestructive in commands.go.
func NewRemoveSimpleQueueCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/queue/simple/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParseSimpleQueues converts command.Result rows from /queue/simple/print
// into typed SimpleQueue values. Rows missing ".id" or "name" are skipped.
func ParseSimpleQueues(result command.Result) []SimpleQueue {
	queues := make([]SimpleQueue, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		queues = append(queues, SimpleQueue{
			RosID:          id,
			Name:           name,
			Target:         row["target"],
			Parent:         row["parent"],
			MaxLimit:       row["max-limit"],
			LimitAt:        row["limit-at"],
			BurstLimit:     row["burst-limit"],
			BurstThreshold: row["burst-threshold"],
			BurstTime:      row["burst-time"],
			Priority:       row["priority"],
			Queue:          row["queue"],
			Bytes:          row["bytes"],
			Packets:        row["packets"],
			Dropped:        row["dropped"],
			Rate:           row["rate"],
			PacketRate:     row["packet-rate"],
			Comment:        row["comment"],
			Disabled:       strings.EqualFold(row["disabled"], "true"),
		})
	}
	return queues
}
