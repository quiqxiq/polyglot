package port

import (
	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
)

// CharSet defines character set types used for generating voucher usernames and passwords.
// Aliased to domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type CharSet = domainHotspot.CharSet

// Supported character sets for voucher code generation (re-exported from domain).
const (
	// CharSetNumeric uses digits only (0-9).
	CharSetNumeric = domainHotspot.CharSetNumeric
	// CharSetLower uses lowercase letters only (a-z).
	CharSetLower = domainHotspot.CharSetLower
	// CharSetUpper uses uppercase letters only (A-Z).
	CharSetUpper = domainHotspot.CharSetUpper
	// CharSetLowerNum uses lowercase letters and digits.
	CharSetLowerNum = domainHotspot.CharSetLowerNum
	// CharSetUpperNum uses uppercase letters and digits.
	CharSetUpperNum = domainHotspot.CharSetUpperNum
	// CharSetMixed uses mixed-case letters and digits.
	CharSetMixed = domainHotspot.CharSetMixed
)

// VoucherGenerateParams holds parameters for mass-generating vouchers.
// Aliased to domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type VoucherGenerateParams = domainHotspot.VoucherGenerateParams

// GeneratedVoucher alias to domain model.
type GeneratedVoucher = domainHotspot.GeneratedVoucher

// VoucherBatch alias to domain model.
type VoucherBatch = domainHotspot.VoucherBatch

// MikhmonComment alias to domain model.
type MikhmonComment = domainHotspot.MikhmonComment
