package queue

import (
	"testing"
)

func TestNewPrintSimpleQueuesCommand(t *testing.T) {
	cmd := NewPrintSimpleQueuesCommand("")
	if cmd.Raw != "/queue/simple/print" {
		t.Fatalf("expected raw /queue/simple/print, got %s", cmd.Raw)
	}
	if cmd.Args[".proplist"] == "" {
		t.Fatal("expected .proplist to be populated")
	}

	cmdNamed := NewPrintSimpleQueuesCommand("sub-budi")
	if cmdNamed.Args["?name"] != "sub-budi" {
		t.Fatalf("expected ?name filter, got %v", cmdNamed.Args)
	}
}

func TestNewStreamQueueStatsCommand(t *testing.T) {
	cmd := NewStreamQueueStatsCommand(StreamParams{Interval: "1s"})
	if cmd.Raw != "/queue/simple/print" || cmd.Args["stats"] != "" || cmd.Args["interval"] != "1s" {
		t.Fatalf("unexpected stream stats command: %+v", cmd)
	}
	if cmd.Args[".proplist"] == "" {
		t.Fatal("expected .proplist to be populated")
	}
}

func TestNewQueueTreeCommands(t *testing.T) {
	cmdPrint := NewPrintQueueTreeCommand("tree1")
	if cmdPrint.Raw != "/queue/tree/print" || cmdPrint.Args["?name"] != "tree1" {
		t.Fatalf("unexpected print queue tree command: %+v", cmdPrint)
	}
	if cmdPrint.Args[".proplist"] == "" {
		t.Fatal("expected .proplist to be populated")
	}

	cmdStream := NewStreamQueueTreeStatsCommand("2s")
	if cmdStream.Raw != "/queue/tree/print" || cmdStream.Args["stats"] != "" || cmdStream.Args["interval"] != "2s" {
		t.Fatalf("unexpected stream queue tree stats command: %+v", cmdStream)
	}
	if cmdStream.Args[".proplist"] == "" {
		t.Fatal("expected .proplist to be populated")
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
