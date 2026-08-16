package port

// MikhmonTransaction represents one transaction record parsed from RouterOS
// /system/script entries created by Mikhmon v4's on-login script.
//
// Field notes:
//   - RosID   : internal RouterOS script ID (e.g. "*1")
//   - Date    : transaction date string (e.g. "dec/25/2024")
//   - Time    : transaction time string (e.g. "10:30:15")
//   - Username: subscriber username
//   - Price   : price recorded
//   - Address : IP address assigned to subscriber during login
//   - MAC     : MAC address of subscriber device
//   - Validity: validity string (e.g. "1d")
//   - Profile : profile name
//   - Comment : user comment
type MikhmonTransaction struct {
	RosID    string
	Date     string
	Time     string
	Username string
	Price    string
	Address  string
	MAC      string
	Validity string
	Profile  string
	Comment  string
	RawName  string
}
