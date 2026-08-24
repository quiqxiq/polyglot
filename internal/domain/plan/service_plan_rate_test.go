package plan

import "testing"

func TestRateLimit(t *testing.T) {
	cases := []struct {
		name string
		p    ServicePlan
		want string
	}{
		{"mbps", ServicePlan{BandwidthDownloadKbps: 10240, BandwidthUploadKbps: 5120}, "10M/5M"},
		{"kbps", ServicePlan{BandwidthDownloadKbps: 512, BandwidthUploadKbps: 256}, "512k/256k"},
		{"rounding", ServicePlan{BandwidthDownloadKbps: 1536, BandwidthUploadKbps: 512}, "2M/512k"},
		{"zero", ServicePlan{}, ""},
	}
	for _, c := range cases {
		if got := c.p.RateLimit(); got != c.want {
			t.Errorf("%s: RateLimit()=%q want %q", c.name, got, c.want)
		}
	}
}

func TestRateLimitWithBurst(t *testing.T) {
	p := ServicePlan{
		BandwidthDownloadKbps: 10240, BandwidthUploadKbps: 5120,
		BurstDownloadKbps: 20480, BurstUploadKbps: 10240,
		BurstThresholdKbps: 4096, BurstTimeSeconds: 8,
	}
	// RouterOS: rx/tx/rx-burst/tx-burst/rx-threshold/tx-threshold/rx-burst-time/tx-burst-time
	want := "10M/5M/20M/10M/4M/4M/8s/8s"
	if got := p.RateLimitWithBurst(); got != want {
		t.Errorf("RateLimitWithBurst()=%q want %q", got, want)
	}
	// Tanpa burst → fallback ke rate polos.
	if got := (ServicePlan{BandwidthDownloadKbps: 10240, BandwidthUploadKbps: 5120}).RateLimitWithBurst(); got != "10M/5M" {
		t.Errorf("fallback=%q", got)
	}
	// Burst parsial (threshold hilang) → fallback juga.
	partial := p
	partial.BurstThresholdKbps = 0
	if got := partial.RateLimitWithBurst(); got != "10M/5M" {
		t.Errorf("partial fallback=%q", got)
	}
}
