package billing

// DiscountType adalah jenis diskon tagihan (planning v2 §1.7).
type DiscountType string

const (
	DiscountTypeNone    DiscountType = "NONE"
	DiscountTypePercent DiscountType = "PERCENT"
	DiscountTypeAmount  DiscountType = "AMOUNT"
)

// ApplyDiscount menghitung nominal diskon dari base. Hasil selalu >= 0 dan
// dibatasi maksimal sebesar base agar total faktur tidak menjadi negatif.
func ApplyDiscount(base float64, typ DiscountType, value float64) float64 {
	if base <= 0 || value <= 0 {
		return 0
	}
	var d float64
	switch typ {
	case DiscountTypePercent:
		d = base * value / 100
	case DiscountTypeAmount:
		d = value
	default:
		return 0
	}
	d = round2(d)
	if d > base {
		return round2(base)
	}
	if d < 0 {
		return 0
	}
	return d
}
