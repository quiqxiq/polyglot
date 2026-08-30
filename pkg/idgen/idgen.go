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

// Initials extracts lowercase initials from a name (e.g. "Budi Santoso" -> "bs", "Ahmad" -> "ah").
func Initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "user"
	}
	if len(parts) == 1 {
		runes := []rune(strings.ToLower(parts[0]))
		if len(runes) >= 2 {
			return string(runes[:2])
		}
		return string(runes)
	}
	var b strings.Builder
	for _, p := range parts {
		r := []rune(p)
		if len(r) > 0 {
			b.WriteRune(r[0])
		}
	}
	res := strings.ToLower(b.String())
	if len(res) > 6 {
		res = res[:6]
	}
	return res
}

// GenerateUsername generates a username using initials/name and random digits based on pattern.
// Default pattern: "{initials}{digits4}" (e.g. "bs4829").
func GenerateUsername(name, pattern, prefix, customerCode string) string {
	if pattern == "" {
		pattern = "{initials}{digits4}"
	}
	initials := Initials(name)
	slug := strings.ToLower(Slug(name))
	if slug == "" {
		slug = "user"
	}
	res := pattern
	res = strings.ReplaceAll(res, "{initials}", initials)
	res = strings.ReplaceAll(res, "{name_slug}", slug)
	res = strings.ReplaceAll(res, "{customer_code}", strings.ToLower(customerCode))
	if strings.Contains(res, "{digits4}") {
		res = strings.ReplaceAll(res, "{digits4}", Digits(4))
	}
	if strings.Contains(res, "{digits6}") {
		res = strings.ReplaceAll(res, "{digits6}", Digits(6))
	}
	if prefix != "" {
		res = prefix + res
	}
	return res
}
