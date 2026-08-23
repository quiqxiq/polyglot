package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// New returns "<prefix>-<16 hex chars>" — bentuk UUID-string konsisten
// dengan kolom TEXT PK skema ISP.
func New(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s-fallback-%x", prefix, [8]byte{})
	}
	return prefix + "-" + hex.EncodeToString(b)
}

// Digits returns a random n-digit numeric string (leading zeros allowed).
func Digits(n int) string {
	out := make([]byte, n)
	for i := range out {
		v, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			out[i] = '0'
			continue
		}
		out[i] = byte('0' + v.Int64())
	}
	return string(out)
}

// Slug normalizes a human name into an identifier-friendly uppercase token:
// non-alphanumeric → '-', trimmed, max 12 char.
func Slug(name string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(name) {
		switch {
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) > 12 {
		s = s[:12]
	}
	return s
}
