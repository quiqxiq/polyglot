package billing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

func TestApplyDiscount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		base  float64
		typ   billing.DiscountType
		value float64
		want  float64
	}{
		{name: "persen 10 dari 100000", base: 100000, typ: billing.DiscountTypePercent, value: 10, want: 10000},
		{name: "persen 12.5 dari 80000", base: 80000, typ: billing.DiscountTypePercent, value: 12.5, want: 10000},
		{name: "nominal 25000", base: 100000, typ: billing.DiscountTypeAmount, value: 25000, want: 25000},
		{name: "nominal melebihi base dijepit", base: 100000, typ: billing.DiscountTypeAmount, value: 150000, want: 100000},
		{name: "persen melebihi 100 dijepit", base: 50000, typ: billing.DiscountTypePercent, value: 250, want: 50000},
		{name: "tipe NONE selalu nol", base: 100000, typ: billing.DiscountTypeNone, value: 5000, want: 0},
		{name: "nilai nol", base: 100000, typ: billing.DiscountTypePercent, value: 0, want: 0},
		{name: "nilai negatif", base: 100000, typ: billing.DiscountTypeAmount, value: -5000, want: 0},
		{name: "base nol", base: 0, typ: billing.DiscountTypeAmount, value: 5000, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := billing.ApplyDiscount(tc.base, tc.typ, tc.value)
			assert.InDelta(t, tc.want, got, 0.005)
		})
	}
}
