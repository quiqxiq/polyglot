package port

import (
	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
)

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

// GeneratedVoucher alias to domain model.
type GeneratedVoucher = domainHotspot.GeneratedVoucher

// VoucherBatch alias to domain model.
type VoucherBatch = domainHotspot.VoucherBatch

// MikhmonComment alias to domain model.
type MikhmonComment = domainHotspot.MikhmonComment
