package queue

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

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

