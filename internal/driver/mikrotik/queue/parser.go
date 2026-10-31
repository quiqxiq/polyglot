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

// ParseQueueTrees converts command.Result rows from /queue/tree/print into typed QueueTree values.
func ParseQueueTrees(result command.Result) []QueueTree {
	trees := make([]QueueTree, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		trees = append(trees, QueueTree{
			RosID:      id,
			Name:       name,
			Parent:     row["parent"],
			PacketMark: row["packet-mark"],
			LimitAt:    row["limit-at"],
			MaxLimit:   row["max-limit"],
			Priority:   row["priority"],
			Bytes:      row["bytes"],
			Packets:    row["packets"],
			Dropped:    row["dropped"],
			Rate:       row["rate"],
			PacketRate: row["packet-rate"],
			Disabled:   strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
		})
	}
	return trees
}
