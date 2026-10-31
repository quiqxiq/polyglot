package ont

import (
	"regexp"
	"strings"
)

var macCleanRe = regexp.MustCompile(`[^0-9a-fA-F]`)

// NormalizeMAC menormalkan MAC ONT/CPE ke format "aa:bb:cc:dd:ee:ff".
// Menerima bentuk "aa-bb-...", "aabb.ccdd.eeff", atau campuran separator.
func NormalizeMAC(raw string) (string, bool) {
	clean := macCleanRe.ReplaceAllString(strings.TrimSpace(raw), "")
	if len(clean) != 12 {
		return "", false
	}
	clean = strings.ToLower(clean)
	parts := make([]string, 6)
	for i := 0; i < 6; i++ {
		parts[i] = clean[i*2 : i*2+2]
	}
	return strings.Join(parts, ":"), true
}

// NormalizeSerial membersihkan serial number ONT (uppercase, tanpa spasi).
func NormalizeSerial(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}
