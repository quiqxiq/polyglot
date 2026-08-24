package app

import (
	"testing"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/internal/port"
)

func TestHotspotProfileParams(t *testing.T) {
	acct := port.SubscriberAccount{
		Profile: "HS-30D", RateLimit: "10M/5M",
		AddressPool: "pool-hs", SharedUsers: 2, ParentQueue: "pq",
		Price: "50000", SellingPrice: "40000", Validity: "30d",
		ExpireMode: "rem", LockUser: true, LockServer: true,
	}
	got := hotspotProfileParams(acct)
	want := port.MikhmonProfileParams{
		Name: "HS-30D", RateLimit: "10M/5M", AddressPool: "pool-hs",
		SharedUsers: "2", ParentQueue: "pq", Price: "50000",
		SellingPrice: "40000", Validity: "30d",
		ExpireMode: port.ExpireMode("rem"),
		LockUser:   true, LockServer: true, Comment: planProfileComment,
	}
	if got != want {
		t.Errorf("got=%+v want=%+v", got, want)
	}
}

func TestHotspotProfileParams_Defaults(t *testing.T) {
	// SharedUsers kosong → default 1; field opsional kosong tetap kosong.
	got := hotspotProfileParams(port.SubscriberAccount{Profile: "MIN"})
	if got.SharedUsers != "1" {
		t.Errorf("SharedUsers default=%q want 1", got.SharedUsers)
	}
	if got.AddressPool != "" || got.ParentQueue != "" || got.Price != "" {
		t.Errorf("field opsional harus kosong: %+v", got)
	}
}

func TestPPPProfileParams(t *testing.T) {
	acct := port.SubscriberAccount{
		Profile: "PPPOE-10M", RateLimit: "10M/5M/20M/10M/4M/4M/8s/8s",
		ParentQueue: "pq", AddressList: "paid",
	}
	got := pppProfileParams(acct)
	want := port.PPPProfileParams{
		Name:        "PPPOE-10M",
		RateLimit:   "10M/5M/20M/10M/4M/4M/8s/8s",
		ParentQueue: "pq",
		AddressList: "paid",
		Comment:     planProfileComment,
	}
	if got != want {
		t.Errorf("got=%+v want=%+v", got, want)
	}
}

func TestExpireModeMapping(t *testing.T) {
	// Konvensi domain (ntf/ntfc/rem/remc/0) lolos apa adanya; port layer
	// hanya meneruskan string ke Mikhmon expire monitor.
	for _, m := range []string{
		domainPlan.ExpireNotFiltered, domainPlan.ExpireNotFilteredCom,
		domainPlan.ExpireRemove, domainPlan.ExpireRemoveComment,
		domainPlan.ExpireNone,
	} {
		if string(hotspotExpireMode(m)) != m {
			t.Errorf("mode %q berubah saat dipetakan", m)
		}
	}
}

func TestIsolirAccount(t *testing.T) {
	acct := isolirAccount("ISOLIR", "0/0")
	if acct.Profile != "ISOLIR" || acct.RateLimit != "0/0" {
		t.Errorf("isolir account salah: %+v", acct)
	}
}
