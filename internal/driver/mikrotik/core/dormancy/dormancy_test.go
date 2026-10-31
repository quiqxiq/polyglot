package dormancy

import (
	"testing"
)

func TestDormancyTransitions(t *testing.T) {
	opts := Defaults()
	s := New(opts)

	if s.Dormant() {
		t.Fatal("expected new state to be active (not dormant)")
	}

	// 1. Sending empty observations up to threshold
	now := int64(1000)
	for i := 0; i < opts.EmptyThreshold-1; i++ {
		now += 1000
		v := s.Observe(Observation{TS: now, Empty: true}, now)
		if v != None {
			t.Fatalf("expected None before threshold, got %v", v)
		}
		if s.Dormant() {
			t.Fatalf("should not be dormant yet at streak %d", i+1)
		}
	}

	// 2. Crossing empty threshold puts it to sleep
	now += 1000
	v := s.Observe(Observation{TS: now, Empty: true}, now)
	if v != Sleep {
		t.Fatalf("expected Sleep at threshold, got %v", v)
	}
	if !s.Dormant() {
		t.Fatal("expected state to be dormant after crossing threshold")
	}

	// 3. Due when now >= wakeAt
	if s.DueForProbe(now) {
		t.Fatal("should not be due immediately after sleeping")
	}
	if !s.DueForProbe(s.WakeAt() + 1) {
		t.Fatal("should be due after wakeAt elapsed")
	}

	// 4. Probe begins probing
	s.MarkProbed(s.WakeAt() + 1)
	if !s.Probing() {
		t.Fatal("expected probing to be true")
	}

	// 5. Active observation wakes it up
	now = s.WakeAt() + 1000
	v = s.Observe(Observation{TS: now, Empty: false}, now)
	if v != Wake {
		t.Fatalf("expected Wake on non-empty payload, got %v", v)
	}
	if s.Dormant() {
		t.Fatal("expected state to be awake")
	}
}

func TestDormancyUnsupported(t *testing.T) {
	opts := Defaults()
	s := New(opts)

	now := int64(5000)
	// Unsupported (e.g. !trap menu missing) should immediately put to sleep
	v := s.Observe(Observation{TS: now, Unsupported: true}, now)
	if v != Sleep {
		t.Fatalf("expected Sleep immediately for unsupported menu, got %v", v)
	}
	if !s.Dormant() {
		t.Fatal("expected state to be dormant for unsupported menu")
	}
}
