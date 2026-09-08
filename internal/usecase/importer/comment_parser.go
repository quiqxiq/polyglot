package importer

import (
	"strconv"
	"strings"
	"unicode"
)

// extractPriceFromComment mencoba membaca harga dari komentar secret/user/binding
// (konvensi Mikhmon / ISP menyimpan metadata harga pada comment).
func extractPriceFromComment(comment string) float64 {
	normalized := strings.ReplaceAll(comment, "Rp ", "Rp")
	normalized = strings.ReplaceAll(normalized, "rp ", "rp")
	normalized = strings.ReplaceAll(normalized, "RP ", "RP")
	for _, part := range strings.Fields(normalized) {
		lower := strings.ToLower(part)
		if strings.HasPrefix(lower, "rp") {
			clean := strings.NewReplacer("rp.", "", "rp", "", ".", "", ",", ".").Replace(lower)
			if v, err := strconv.ParseFloat(clean, 64); err == nil {
				return v
			}
		}
	}
	return 0
}

// extractRateFromProfileHint tidak tersedia di level secret (rate ada di profil).
func extractRateFromProfileHint(string) string { return "" }

// guessPhone mencari pola nomor di komentar.
func guessPhone(comment string) string {
	for _, part := range strings.Fields(comment) {
		digits := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' || r == '+' {
				return r
			}
			return -1
		}, part)
		if len(digits) >= 10 && strings.HasPrefix(digits, "08") ||
			len(digits) >= 11 && (strings.HasPrefix(digits, "628") || strings.HasPrefix(digits, "+628")) {
			return digits
		}
	}
	return ""
}

// guessCustomerName mengekstrak nama manusia/pelanggan dari komentar router.
// Jika komentar kosong atau merupakan voucher/IP, fallback ke username router.
func guessCustomerName(username, comment string) string {
	c := strings.TrimSpace(comment)
	if c == "" {
		return FormatAutoName(username)
	}

	lower := strings.ToLower(c)
	// Abaikan pola komentar voucher Mikhmon
	if strings.HasPrefix(lower, "vc-") ||
		strings.Contains(lower, "exp:") ||
		strings.Contains(lower, "validity") ||
		strings.Contains(lower, "mikhmon") ||
		strings.Contains(lower, "uptime") {
		return FormatAutoName(username)
	}

	// Pola prefix eksplisit: "Nama: Budi Santoso"
	for _, prefix := range []string{"nama:", "nama :", "name:", "name :"} {
		if idx := strings.Index(lower, prefix); idx != -1 {
			candidate := strings.TrimSpace(c[idx+len(prefix):])
			if cutIdx := strings.IndexAny(candidate, "|,-/"); cutIdx != -1 {
				candidate = strings.TrimSpace(candidate[:cutIdx])
			}
			if len(candidate) >= 2 && !isNumericOrIP(candidate) {
				return candidate
			}
		}
	}

	// Pola delimiter: "Budi Santoso - 08123456789 - Jl Mawar"
	parts := strings.FieldsFunc(c, func(r rune) bool {
		return r == '-' || r == '|' || r == '/' || r == ','
	})
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if len(trimmed) >= 2 &&
			!strings.HasPrefix(strings.ToLower(trimmed), "rp") &&
			guessPhone(trimmed) == "" &&
			!isNumericOrIP(trimmed) {
			return trimmed
		}
	}

	return FormatAutoName(username)
}

func isNumericOrIP(s string) bool {
	cleaned := strings.ReplaceAll(s, ".", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.TrimSpace(cleaned)
	_, err := strconv.ParseInt(cleaned, 10, 64)
	return err == nil
}

// FormatAutoName memformat username atau string teknis (misal: "susanto_madek" atau "susanto-madek")
// menjadi nama rapi berformat Title Case (misal: "Susanto Madek").
// Jika string tidak memiliki delimiter '_' atau '-', nilai asli dipertahankan seperti biasa.
func FormatAutoName(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	hasDelimiter := strings.ContainsAny(s, "_-")
	if !hasDelimiter {
		return s
	}

	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)
	for i, w := range words {
		runes := []rune(strings.ToLower(w))
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// GuessCustomerName mengekspos guessCustomerName untuk pengujian dan pemetaan data.
func GuessCustomerName(username, comment string) string {
	return guessCustomerName(username, comment)
}

// GuessPhone mengekspos guessPhone untuk pengujian.
func GuessPhone(comment string) string {
	return guessPhone(comment)
}

// ExtractPriceFromComment mengekspos extractPriceFromComment untuk pengujian.
func ExtractPriceFromComment(comment string) float64 {
	return extractPriceFromComment(comment)
}
