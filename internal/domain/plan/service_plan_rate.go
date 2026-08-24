package plan

import "strconv"

// rateSide mengonversi kbps ke notasi RouterOS ("10M" / "512k"),
// pembulatan ke Mbps terdekat untuk nilai >= 1000 kbps.
func rateSide(kbps int) string {
	switch {
	case kbps <= 0:
		return ""
	case kbps >= 1000:
		return strconv.Itoa((kbps+500)/1000) + "M"
	default:
		return strconv.Itoa(kbps) + "k"
	}
}

// RateLimit mengembalikan "rx/tx" tanpa burst; "" bila keduanya nol.
func (p ServicePlan) RateLimit() string {
	dl, ul := rateSide(p.BandwidthDownloadKbps), rateSide(p.BandwidthUploadKbps)
	if dl == "" && ul == "" {
		return ""
	}
	return dl + "/" + ul
}

// RateLimitWithBurst mengembalikan rate 8-segmen RouterOS bila semua
// komponen burst terisi; selain itu fallback ke RateLimit().
// Format: rx/tx/rx-burst/tx-burst/rx-thr/tx-thr/rx-time/tx-time.
func (p ServicePlan) RateLimitWithBurst() string {
	base := p.RateLimit()
	if base == "" ||
		p.BurstDownloadKbps <= 0 || p.BurstUploadKbps <= 0 ||
		p.BurstThresholdKbps <= 0 || p.BurstTimeSeconds <= 0 {
		return base
	}
	bd, bu := rateSide(p.BurstDownloadKbps), rateSide(p.BurstUploadKbps)
	th := rateSide(p.BurstThresholdKbps)
	bt := strconv.Itoa(p.BurstTimeSeconds) + "s"
	return base + "/" + bd + "/" + bu + "/" + th + "/" + th + "/" + bt + "/" + bt
}
