package cert

// Certificate represents an X.509 certificate managed in RouterOS (/certificate).
type Certificate struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CommonName    string `json:"common_name"`
	Fingerprint   string `json:"fingerprint"`
	InvalidBefore string `json:"invalid_before"`
	InvalidAfter  string `json:"invalid_after"`
	ExpiresAfter  string `json:"expires_after"`
	CA            bool   `json:"ca"`
	Trusted       bool   `json:"trusted"`
	Issued        bool   `json:"issued"`
}
