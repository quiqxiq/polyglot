// Package phone menyediakan utilitas normalisasi nomor telepon Indonesia
// dan konversi ke JID WhatsApp agar perbandingan nomor di seluruh sistem
// konsisten dengan format pengirim yang diekstrak dari WhatsApp JID
// (digit internasional tanpa tanda tambah, contoh: 6281234567890).
package phone

import "strings"

// Normalize menyeragamkan nomor telepon mentah menjadi digit internasional
// tanpa "+" (contoh: "6281234567890"). Input yang diterima:
//   - 081234567890            -> 6281234567890
//   - +6281234567890          -> 6281234567890
//   - 0812-3456-7890          -> 6281234567890
//   - 6281234567890           -> 6281234567890
//   - 6281234567890:12@s.whatsapp.net -> 6281234567890
//   - 12036304xxx@g.us        -> diteruskan apa adanya (grup tidak dikonversi)
//
// Nomor non-Indonesia (tidak berawalan 0/62) diteruskan apa adanya setelah
// dibersihkan dari karakter non-digit. Input kosong atau tanpa digit
// mengembalikan string kosong.
func Normalize(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Grup WhatsApp diteruskan apa adanya tanpa konversi kode negara.
	if strings.HasSuffix(raw, "@g.us") {
		return raw
	}

	// Buang suffix domain JID ("@s.whatsapp.net", "@c.us", dst)
	// beserta device suffix (mis. ":12").
	if idx := strings.IndexByte(raw, '@'); idx != -1 {
		raw = raw[:idx]
	}
	if idx := strings.IndexByte(raw, ':'); idx != -1 {
		raw = raw[:idx]
	}

	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	if digits == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(digits, "62"):
		return digits
	case strings.HasPrefix(digits, "0"):
		return "62" + digits[1:]
	default:
		return digits
	}
}

// ToWhatsAppJID mengubah nomor telepon mentah menjadi JID tujuan kirim
// WhatsApp ("62xxx@s.whatsapp.net"). Grup (@g.us) diteruskan apa adanya.
// Mengembalikan string kosong bila input tidak memuat digit valid.
func ToWhatsAppJID(raw string) string {
	n := Normalize(raw)
	if n == "" || strings.HasSuffix(n, "@g.us") {
		return n
	}
	return n + "@s.whatsapp.net"
}
