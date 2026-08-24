package mikrotik

import "testing"

func TestDedicatedQueueParamsToROS(t *testing.T) {
	got := dedicatedQueueToROS(DedicatedQueueParams{
		Name: "dq-budi", Target: "budi",
		MaxLimit: "20M/10M", LimitAt: "10M/5M",
		BurstLimit: "20M/10M", BurstThreshold: "4M", BurstTime: "8s",
		Comment: "x",
	})
	if got.Raw != "/queue/simple/add" {
		t.Fatalf("raw=%q", got.Raw)
	}
	if got.Args["name"] != "dq-budi" || got.Args["target"] != "budi" ||
		got.Args["max-limit"] != "20M/10M" || got.Args["limit-at"] != "10M/5M" ||
		got.Args["burst-limit"] != "20M/10M" ||
		got.Args["burst-threshold"] != "4M" || got.Args["burst-time"] != "8s" {
		t.Fatalf("args salah: %+v", got.Args)
	}
}

func TestDedicatedQueueParamsNoBurst(t *testing.T) {
	got := dedicatedQueueToROS(DedicatedQueueParams{
		Name: "d", Target: "d", MaxLimit: "10M/5M", LimitAt: "10M/5M",
	})
	if _, ok := got.Args["burst-limit"]; ok {
		t.Fatal("burst-limit tidak boleh muncul tanpa burst")
	}
	if _, ok := got.Args["burst-threshold"]; ok {
		t.Fatal("burst-threshold tidak boleh muncul tanpa burst")
	}
}
