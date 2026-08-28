package queue

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestParseSimpleQueues(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":      "*10",
				"name":     "sub-budi",
				"target":   "192.168.1.50/32",
				"max-limit": "10M/10M",
				"disabled": "false",
			},
		},
	}

	queues := ParseSimpleQueues(res)
	if len(queues) != 1 {
		t.Fatalf("expected 1 queue, got %d", len(queues))
	}
	if queues[0].Name != "sub-budi" || queues[0].MaxLimit != "10M/10M" || queues[0].Disabled {
		t.Fatalf("unexpected queue parsed: %+v", queues[0])
	}
}

