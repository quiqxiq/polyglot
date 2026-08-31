package provisioner

import (
	"testing"

	"github.com/quixiq/polyglot/internal/port"
)

func TestDedicatedQueueName(t *testing.T) {
	if got := dedicatedQueueName("budi"); got != "dq-budi" {
		t.Errorf("dedicatedQueueName=%q want dq-budi", got)
	}
}

func TestDedicatedQueueFromAccount_WithBurst(t *testing.T) {
	acct := port.SubscriberAccount{
		Username:      "budi",
		RateLimit:     "10M/5M/20M/10M/4M/4M/8s/8s",
		BaseRateLimit: "10M/5M",
	}
	got := dedicatedQueueFromAccount(acct, "c1")
	want := port.DedicatedQueueParams{
		Name:           "dq-budi",
		Target:         "budi",
		MaxLimit:       "10M/5M/20M/10M/4M/4M/8s/8s",
		LimitAt:        "10M/5M",
		BurstLimit:     "20M/10M",
		BurstThreshold: "4M/4M",
		BurstTime:      "8s/8s",
		Comment:        "c1",
	}
	if got != want {
		t.Errorf("got=%+v want=%+v", got, want)
	}
}

func TestDedicatedQueueFromAccount_NoBurst(t *testing.T) {
	acct := port.SubscriberAccount{
		Username:      "ani",
		RateLimit:     "10M/5M",
		BaseRateLimit: "10M/5M",
	}
	got := dedicatedQueueFromAccount(acct, "")
	if got.BurstLimit != "" || got.BurstThreshold != "" || got.BurstTime != "" {
		t.Errorf("burst harus kosong: %+v", got)
	}
	if got.LimitAt != "10M/5M" || got.MaxLimit != "10M/5M" {
		t.Errorf("limit salah: %+v", got)
	}
}

func TestIsDedicated(t *testing.T) {
	cases := map[string]bool{
		"DEDICATED": true,
		"dedicated": true,
		"PPPOE":     false,
		"HOTSPOT":   false,
		"":          false,
	}
	for in, want := range cases {
		if got := isDedicated(in); got != want {
			t.Errorf("isDedicated(%q)=%v want %v", in, got, want)
		}
	}
}

func TestSplitBurstSegments(t *testing.T) {
	acct := port.SubscriberAccount{
		Username:      "x",
		RateLimit:     "10M/5M/20M/10M/4M/4M/8s/8s",
		BaseRateLimit: "10M/5M",
	}
	got := dedicatedQueueFromAccount(acct, "")
	if got.BurstLimit != "20M/10M" || got.BurstThreshold != "4M/4M" || got.BurstTime != "8s/8s" {
		t.Errorf("burst segmen salah: %+v", got)
	}
}
