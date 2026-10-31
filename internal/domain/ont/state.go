package ont

import (
	"strconv"
	"strings"
)

// DecideState menentukan state baru berdasarkan hasil poll OLT.
//   - linkUp true  → ONLINE (recovery langsung; strikes direset).
//   - linkUp false → OFFLINE setelah offlineStrikes berturut mencapai ambang.
//
// State REMOVED hanya ditetapkan oleh roster (tidak terlihat beberapa poll).
func DecideState(current State, linkUp bool, strikes, strikeThreshold int) State {
	if linkUp {
		return StateOnline
	}
	threshold := strikeThreshold
	if threshold <= 0 {
		threshold = 2
	}
	if strikes >= threshold {
		return StateOffline
	}
	if current == StateUnknown {
		return StateUnknown
	}
	return current
}

// ParseOpticalPower mengubah nilai SNMP/CLI menjadi dBm.
// Nilai "na", kosong, atau tidak valid dikembalikan sebagai (nil,false).
func ParseOpticalPower(raw string) (*float64, bool) {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, `"`)
	if s == "" || strings.EqualFold(s, "na") || strings.EqualFold(s, "n/a") {
		return nil, false
	}
	// Buang satuan bila ada (mis. "-21.40 dBm" atau "-21,4").
	s = strings.TrimSpace(strings.TrimSuffix(strings.ToLower(s), "dbm"))
	s = strings.ReplaceAll(s, ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, false
	}
	return &v, true
}

// RxClass klasifikasi kualitas Rx optik untuk badge UI/alert.
type RxClass string

const (
	RxUnavailable RxClass = "UNAVAILABLE"
	RxNormal      RxClass = "NORMAL"
	RxWarning     RxClass = "WARNING"
	RxCritical    RxClass = "CRITICAL"
)

// ClassifyRx mengklasifikasi Rx terhadap ambang warning/critical (dBm).
func ClassifyRx(rx *float64, warningDBM, criticalDBM float64) RxClass {
	if rx == nil {
		return RxUnavailable
	}
	switch {
	case *rx <= criticalDBM:
		return RxCritical
	case *rx <= warningDBM:
		return RxWarning
	default:
		return RxNormal
	}
}
