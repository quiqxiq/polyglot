package port

// CharSet defines character set types used for generating voucher usernames and passwords.
type CharSet string

// VoucherGenerateParams holds parameters for mass-generating Mikhmon vouchers.
type VoucherGenerateParams struct {
	Server      string  // Hotspot server name (empty = "all")
	Profile     string  // Hotspot user profile name (required)
	Prefix      string  // Code prefix (e.g. "vc")
	UserLength  int     // Length of generated username
	PassLength  int     // Length of generated password (0 = username is password)
	CharSet     CharSet // Character set to use
	LimitUptime string  // Time limit (e.g. "1d", "3h")
	LimitBytes  string  // Data quota in bytes
	CommentTag  string  // Label tag in comment
}

// GeneratedVoucher holds the details of one generated voucher.
type GeneratedVoucher struct {
	Username string
	Password string
	Comment  string
}

// VoucherBatch holds the result of a mass voucher generation.
type VoucherBatch struct {
	Vouchers []GeneratedVoucher
}

// MikhmonComment represents parsed metadata stored in a MikroTik Hotspot user comment
// by Mikhmon v4. Mikhmon uses two comment formats:
//
//  1. Pre-login (created during voucher generation):
//     "<type>-<code>-<date>-<tag>" e.g. "vc-A3X-08.03.26-Voucher_1_Hari"
//     - Type : "vc" (voucher) or "up" (username/password user)
//     - Code : 3-4 character random string generated during creation
//     - CreatedDate : creation date formatted as MM.DD.YY
//     - Tag  : optional user/batch label
//
//  2. Post-login (updated automatically by profile on-login script on first login):
//     "DD/MM/YYYY HH:MM:SS <mode> <old-comment>" e.g. "03/08/2026 15:30:00 N vc-A3X-08.03.26-Voucher_1_Hari"
//     - ExpireDate : expiry date string (DD/MM/YYYY)
//     - ExpireTime : expiry time string (HH:MM:SS)
//     - ExpireMode : "N" (Notify / set limit-uptime=1s) or "X" (Remove / delete user)
//     - IsActivated: true when post-login expiry date is present
type MikhmonComment struct {
	Type        string // "vc" or "up"
	Code        string // e.g. "A3X"
	CreatedDate string // MM.DD.YY
	Tag         string // e.g. "Voucher_1_Hari"

	IsActivated bool
	ExpireDate  string // DD/MM/YYYY
	ExpireTime  string // HH:MM:SS
	ExpireMode  string // "N" (Notify) or "X" (Remove)
	RawComment  string
}
