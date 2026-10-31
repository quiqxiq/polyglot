package billing

// Kebijakan pembebanan biaya gateway (planning v2 D11 / analisis §6.1.B).
const (
	GatewayFeeBearerMerchant = "MERCHANT"
	GatewayFeeBearerCustomer = "CUSTOMER"
	GatewayFeeBearerSplit    = "SPLIT"
)

// GatewayCustomerFee menghitung porsi biaya gateway yang dibebankan ke
// pelanggan: MERCHANT = 0, CUSTOMER = penuh, SPLIT = separuh. Nominal
// dibulatkan 2 desimal dan tidak pernah negatif.
func GatewayCustomerFee(fee float64, bearer string) float64 {
	if fee <= 0 {
		return 0
	}
	switch bearer {
	case GatewayFeeBearerCustomer:
		return round2(fee)
	case GatewayFeeBearerSplit:
		return round2(fee / 2)
	default: // MERCHANT
		return 0
	}
}
