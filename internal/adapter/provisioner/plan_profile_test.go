package provisioner

import (
	"testing"

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
		ParentQueue: "pq", AddressList: "paid", RemoteAddressPool: "pool-pppoe",
	}
	got := pppProfileParams(acct)
	want := port.PPPProfileParams{
		Name:          "PPPOE-10M",
		RateLimit:     "10M/5M/20M/10M/4M/4M/8s/8s",
		ParentQueue:   "pq",
		AddressList:   "paid",
		RemoteAddress: "pool-pppoe",
		Comment:       planProfileComment,
	}
	if got != want {
		t.Errorf("got=%+v want=%+v", got, want)
	}
}

func TestExpireModeMapping(t *testing.T) {
	// Konvensi (ntf/ntfc/rem/remc/0) lolos apa adanya; port layer
	// hanya meneruskan string ke Mikhmon expire monitor.
	for _, m := range []string{
		"ntf", "ntfc", "rem", "remc", "0",
	} {
		if string(hotspotExpireMode(m)) != m {
			t.Errorf("mode %q berubah saat dipetakan", m)
		}
	}
}

func TestIsolirAccount(t *testing.T) {
	acct := isolirAccount("isolir", isolirRateLimit, "ISOLIR_USERS")
	if acct.Profile != "isolir" || acct.RateLimit != isolirRateLimit || acct.AddressList != "ISOLIR_USERS" {
		t.Errorf("isolir account salah: %+v", acct)
	}
	if isolirRateLimit == "0/0" || isolirRateLimit == "" {
		t.Errorf("rate isolir tidak boleh unlimited/kosong: %q", isolirRateLimit)
	}
}

func TestPlanProfileDiffers(t *testing.T) {
	acct := port.SubscriberAccount{
		Profile: "PLAN", RateLimit: "10M/10M", ParentQueue: "pq",
		AddressList: "paid", RemoteAddressPool: "pool-pppoe",
	}
	same := profileSnapshot{rate: "10M/10M", parentQueue: "pq", addressList: "paid", addressPool: "pool-pppoe"}
	if planProfileDiffers(same, acct) {
		t.Fatalf("profil identik tidak boleh dianggap berbeda")
	}

	for name, snap := range map[string]profileSnapshot{
		"rate":        {rate: "5M/5M", parentQueue: "pq", addressList: "paid", addressPool: "pool-pppoe"},
		"parentQueue": {rate: "10M/10M", parentQueue: "lain", addressList: "paid", addressPool: "pool-pppoe"},
		"addressList": {rate: "10M/10M", parentQueue: "pq", addressList: "isolir", addressPool: "pool-pppoe"},
		"addressPool": {rate: "10M/10M", parentQueue: "pq", addressList: "paid", addressPool: "pool-lain"},
	} {
		if !planProfileDiffers(snap, acct) {
			t.Fatalf("%s berbeda harus terdeteksi", name)
		}
	}

	// Field kosong pada akun tidak dianggap perbedaan (tidak menimpa profil
	// dashboard dengan nilai kosong).
	sparse := port.SubscriberAccount{Profile: "PLAN", RateLimit: "10M/10M"}
	if planProfileDiffers(profileSnapshot{rate: "10M/10M", parentQueue: "pq"}, sparse) {
		t.Fatalf("field kosong tidak boleh memicu update")
	}
}
