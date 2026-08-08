package templates

import "embed"

// FS embeds all voucher templates statically into the Go binary.
//
//go:embed *.txt
var FS embed.FS
