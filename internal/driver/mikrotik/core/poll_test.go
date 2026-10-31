package core

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPollInterval(t *testing.T) {
	pi := NewPollInterval(2 * time.Second)
	if pi.Duration() != 2*time.Second {
		t.Fatalf("expected 2s, got %v", pi.Duration())
	}

	pi.Set(5 * time.Second)
	if pi.Duration() != 5*time.Second {
		t.Fatalf("expected 5s, got %v", pi.Duration())
	}

	pi.Set(0)
	if pi.Duration() != time.Second {
		t.Fatalf("expected default 1s on non-positive input, got %v", pi.Duration())
	}
}

func TestPollLoopExecutionAndRetime(t *testing.T) {
	var count atomic.Int32
	pi := NewPollInterval(10 * time.Millisecond)

	loop := NewPollLoop(func() {
		count.Add(1)
	}, pi.Duration)

	loop.Start()
	time.Sleep(35 * time.Millisecond)
	if c := count.Load(); c < 2 {
		t.Errorf("expected at least 2 ticks, got %d", c)
	}

	// Dynamic retime
	pi.Set(50 * time.Millisecond)
	loop.Retime()

	before := count.Load()
	time.Sleep(20 * time.Millisecond)
	// Shouldn't fire again immediately because of 50ms delay
	if c := count.Load(); c != before && c != before+1 {
		t.Errorf("unexpected tick count: %d", c)
	}

	loop.Stop()
	afterStop := count.Load()
	time.Sleep(50 * time.Millisecond)
	if count.Load() != afterStop {
		t.Errorf("tick after stop: before %d, now %d", afterStop, count.Load())
	}
}
