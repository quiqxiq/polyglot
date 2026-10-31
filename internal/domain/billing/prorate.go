package billing

import (
	"math"
	"time"
)

// ProrateMode adalah metode perhitungan prorata paket (planning v2 §3.1).
type ProrateMode string

const (
	// ProrateModeNone menagih harga penuh tanpa prorata.
	ProrateModeNone ProrateMode = "NONE"
	// ProrateModeCalendarDay memakai hari aktual kalender (planning v2 D2).
	ProrateModeCalendarDay ProrateMode = "CALENDAR_DAY"
	// ProrateModeFlat30 memakai basis 30 hari per siklus.
	ProrateModeFlat30 ProrateMode = "FLAT_30"
)

// ProrateParams is the input for calculating a prorated amount.
type ProrateParams struct {
	PlanPrice float64
	StartDate time.Time // hari aktif pertama (bagian tanggal saja)
	EndDate   time.Time // hari aktif terakhir periode, inklusif (bagian tanggal saja)
	Mode      ProrateMode
}

// ProrateResult is the output of CalculateProrate.
type ProrateResult struct {
	ActiveDays int
	PeriodDays int
	Amount     float64
	FullPeriod bool
}

// CalculateProrate menghitung nominal paket untuk sebagian periode:
// round(plan_price × sisa_hari_aktif / total_hari_siklus_berjalan).
// Denominator mengikuti hari pada bulan kalender dari EndDate (mode
// CALENDAR_DAY) atau basis 30 hari (mode FLAT_30).
func CalculateProrate(p ProrateParams) ProrateResult {
	start := dateOnly(p.StartDate)
	end := dateOnly(p.EndDate)
	res := ProrateResult{}
	if end.Before(start) {
		return res
	}
	active := int(end.Sub(start).Hours()/24) + 1
	periodDays := daysInMonth(end)
	if p.Mode == ProrateModeFlat30 {
		periodDays = 30
	}
	if periodDays <= 0 {
		return res
	}
	if active > periodDays {
		active = periodDays
	}
	res.ActiveDays = active
	res.PeriodDays = periodDays
	switch p.Mode {
	case ProrateModeNone:
		res.Amount = round2(p.PlanPrice)
		res.FullPeriod = true
	default:
		res.Amount = round2(p.PlanPrice * float64(active) / float64(periodDays))
		res.FullPeriod = active >= periodDays
	}
	return res
}

// dateOnly menormalkan time.Time ke tanggal UTC (tengah malam).
func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// daysInMonth mengembalikan jumlah hari pada bulan kalender t.
func daysInMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// round2 membulatkan nominal uang ke 2 desimal.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
