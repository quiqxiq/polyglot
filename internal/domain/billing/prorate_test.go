package billing_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestCalculateProrate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		price      float64
		start      time.Time
		end        time.Time
		mode       billing.ProrateMode
		wantActive int
		wantPeriod int
		wantAmount float64
		wantFull   bool
	}{
		{
			name:       "Februari pendek 28 hari",
			price:      100000,
			start:      date(2026, time.February, 10),
			end:        date(2026, time.February, 28),
			mode:       billing.ProrateModeCalendarDay,
			wantActive: 19, wantPeriod: 28, wantAmount: 67857.14,
		},
		{
			name:       "April 30 hari prorata dari tanggal 1",
			price:      100000,
			start:      date(2026, time.April, 1),
			end:        date(2026, time.April, 30),
			mode:       billing.ProrateModeCalendarDay,
			wantActive: 30, wantPeriod: 30, wantAmount: 100000, wantFull: true,
		},
		{
			name:       "Tahun kabisat Februari 29 hari",
			price:      100000,
			start:      date(2028, time.February, 1),
			end:        date(2028, time.February, 29),
			mode:       billing.ProrateModeCalendarDay,
			wantActive: 29, wantPeriod: 29, wantAmount: 100000, wantFull: true,
		},
		{
			name:       "Aktivasi pertengahan Agustus 12/31",
			price:      100000,
			start:      date(2026, time.August, 20),
			end:        date(2026, time.August, 31),
			mode:       billing.ProrateModeCalendarDay,
			wantActive: 12, wantPeriod: 31, wantAmount: 38709.68,
		},
		{
			name:       "Mode NONE selalu harga penuh",
			price:      100000,
			start:      date(2026, time.August, 20),
			end:        date(2026, time.August, 31),
			mode:       billing.ProrateModeNone,
			wantActive: 12, wantPeriod: 31, wantAmount: 100000, wantFull: true,
		},
		{
			name:       "Mode FLAT_30 basis 30 hari",
			price:      100000,
			start:      date(2026, time.February, 10),
			end:        date(2026, time.February, 28),
			mode:       billing.ProrateModeFlat30,
			wantActive: 19, wantPeriod: 30, wantAmount: 63333.33,
		},
		{
			name:       "Mode FLAT_30 bulan 31 hari dijepit 30",
			price:      100000,
			start:      date(2026, time.January, 1),
			end:        date(2026, time.January, 31),
			mode:       billing.ProrateModeFlat30,
			wantActive: 30, wantPeriod: 30, wantAmount: 100000, wantFull: true,
		},
		{
			name:       "Satu hari aktif",
			price:      100000,
			start:      date(2026, time.August, 20),
			end:        date(2026, time.August, 20),
			mode:       billing.ProrateModeCalendarDay,
			wantActive: 1, wantPeriod: 31, wantAmount: 3225.81,
		},
		{
			name:       "EndDate sebelum StartDate",
			price:      100000,
			start:      date(2026, time.August, 20),
			end:        date(2026, time.August, 19),
			mode:       billing.ProrateModeCalendarDay,
			wantActive: 0, wantPeriod: 0, wantAmount: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := billing.CalculateProrate(billing.ProrateParams{
				PlanPrice: tc.price,
				StartDate: tc.start,
				EndDate:   tc.end,
				Mode:      tc.mode,
			})
			assert.Equal(t, tc.wantActive, got.ActiveDays, "active days")
			assert.Equal(t, tc.wantPeriod, got.PeriodDays, "period days")
			assert.InDelta(t, tc.wantAmount, got.Amount, 0.005, "amount")
			assert.Equal(t, tc.wantFull, got.FullPeriod, "full period")
		})
	}
}
