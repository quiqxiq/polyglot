package queue

import (
	"testing"
)

func TestNewPrintSimpleQueuesCommand(t *testing.T) {
	cmd := NewPrintSimpleQueuesCommand("")
	if cmd.Raw != "/queue/simple/print" {
		t.Fatalf("expected raw /queue/simple/print, got %s", cmd.Raw)
	}

	cmdNamed := NewPrintSimpleQueuesCommand("sub-budi")
	if cmdNamed.Args["?name"] != "sub-budi" {
		t.Fatalf("expected ?name filter, got %v", cmdNamed.Args)
	}
}

func TestSuspendQueueParams(t *testing.T) {
	p := SuspendQueueParams("192.168.1.100", "overdue", "")
	if p.Name != "suspended_192_168_1_100" || p.MaxLimit != "1k/1k" {
		t.Fatalf("unexpected suspend params: %+v", p)
	}
}

func TestNewAddSimpleQueueCommand(t *testing.T) {
	cmd := NewAddSimpleQueueCommand(SimpleQueueParams{
		Name:     "sub-test",
		Target:   "10.0.0.5",
		MaxLimit: "20M/20M",
		Disabled: false,
	})
	if cmd.Raw != "/queue/simple/add" || cmd.Args["name"] != "sub-test" || cmd.Args["disabled"] != "no" {
		t.Fatalf("unexpected add command: %+v", cmd)
	}
}
