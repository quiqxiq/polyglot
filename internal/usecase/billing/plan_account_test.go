package billing

import (
	"strconv"
	"testing"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
)

func TestSubscriberAccountFromPlan_Hotspot(t *testing.T) {
	pl := domainPlan.ServicePlan{
		Name: "HS-30D-10M", ServiceType: domainPlan.TypeHotspot,
		BandwidthDownloadKbps: 10240, BandwidthUploadKbps: 5120,
		SharedUsers: 2, IPPoolName: "pool-hs", ParentQueue: "pq-all",
		AddressList: "paid",
		Price:       40000, SellingPrice: 50000, Validity: "30d",
		ExpireMode: domainPlan.ExpireRemove, LockUser: true,
		LockServer: true,
	}
	sub := domainSubscription.Subscription{
		RemoteUsername: "0812xxx", RemotePassword: "secret",
	}

	acct := subscriberAccountFromPlan(sub, pl)

	if acct.Profile != "HS-30D-10M" || acct.Username != "0812xxx" || acct.Password != "secret" {
		t.Fatalf("identitas salah: %+v", acct)
	}
	if acct.RateLimit != "10M/5M" {
		t.Errorf("RateLimit=%q want 10M/5M", acct.RateLimit)
	}
	if acct.SharedUsers != 2 || acct.AddressPool != "pool-hs" ||
		acct.ParentQueue != "pq-all" || acct.AddressList != "paid" ||
		acct.Price != "50000" || acct.SellingPrice != "40000" ||
		acct.Validity != "30d" || acct.ExpireMode != "rem" ||
		!acct.LockUser || !acct.LockServer {
		t.Errorf("parameter mikhmon tidak lengkap: %+v", acct)
	}
}

func TestSubscriberAccountFromPlan_Burst(t *testing.T) {
	pl := domainPlan.ServicePlan{
		Name: "PPPOE-BURST",
		BandwidthDownloadKbps: 10240, BandwidthUploadKbps: 5120,
		BurstDownloadKbps: 20480, BurstUploadKbps: 10240,
		BurstThresholdKbps: 4096, BurstTimeSeconds: 8,
	}
	acct := subscriberAccountFromPlan(domainSubscription.Subscription{}, pl)
	if acct.RateLimit != "10M/5M/20M/10M/4M/4M/8s/8s" {
		t.Errorf("burst rate=%q", acct.RateLimit)
	}
}

func TestSubscriberAccountFromPlan_CustomRateOverride(t *testing.T) {
	pl := domainPlan.ServicePlan{Name: "P1",
		BandwidthDownloadKbps: 10240, BandwidthUploadKbps: 5120,
	}
	sub := domainSubscription.Subscription{RateLimit: "2M/1M"}

	acct := subscriberAccountFromPlan(sub, pl)
	if acct.RateLimit != "2M/1M" {
		t.Errorf("custom override=%q want 2M/1M", acct.RateLimit)
	}
}

func TestFormatMoney(t *testing.T) {
	cases := []struct {
		selling, base, want string
	}{
		{"50000", "40000", "50000"},
		{"0", "40000", "40000"},
		{"0", "0", ""},
	}
	for _, c := range cases {
		s, _ := strconv.ParseFloat(c.selling, 64)
		b, _ := strconv.ParseFloat(c.base, 64)
		jual, _ := formatMoney(s, b)
		if jual != c.want {
			t.Errorf("formatMoney(%v,%v)=%q want %q", c.selling, c.base, jual, c.want)
		}
	}
}
