package billing

import (
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
		Price:       40000,
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
		acct.ParentQueue != "pq-all" || acct.AddressList != "paid" {
		t.Errorf("parameter hotspot tidak lengkap: %+v", acct)
	}
}

func TestSubscriberAccountFromPlan_Burst(t *testing.T) {
	pl := domainPlan.ServicePlan{
		Name:                  "PPPOE-BURST",
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

func TestBuildHotspotProvisionSpec(t *testing.T) {
	pl := domainPlan.ServicePlan{
		Name:                  "HS-10M",
		ServiceType:           domainPlan.TypeHotspot,
		BandwidthDownloadKbps: 10240,
		BandwidthUploadKbps:   5120,
		SharedUsers:           3,
		IPPoolName:            "hs-pool",
		AddressList:           "VIP",
	}
	sub := domainSubscription.Subscription{
		ID:             "sub-hs-1",
		RemoteUsername: "user1",
		RemotePassword: "pwd",
	}
	spec := BuildHotspotProvisionSpec(sub, pl)
	if spec.User.Username != "user1" || spec.User.Password != "pwd" || spec.User.Profile != "HS-10M" {
		t.Fatalf("spec user salah: %+v", spec.User)
	}
	if spec.Profile.Name != "HS-10M" || spec.Profile.RateLimit != "10M/5M" || spec.Profile.SharedUsers != 3 || spec.Profile.AddressPool != "hs-pool" || spec.Profile.AddressList != "VIP" {
		t.Fatalf("spec profile salah: %+v", spec.Profile)
	}
}
