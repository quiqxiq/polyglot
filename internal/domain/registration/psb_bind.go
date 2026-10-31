package registration

// PSBBindInput adalah data instalasi PSB yang dipakai untuk menautkan ONT
// ke pelanggan/langganan setelah konversi (Fase E PLAN-ACS-OLT-HIOSO.md).
type PSBBindInput struct {
	RegistrationID string
	OLTDeviceID    string
	PONIndex       int
	ONUIndex       int
	MACAddress     string
	SerialNumber   string
	CustomerID     string
	SubscriptionID string
	PPPoEUsername  string
	Label          string
}
