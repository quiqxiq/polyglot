package bot

import (
	"strings"
)

type Guardrail struct {
	allowedTopics []string
}

func NewGuardrail(allowedTopics []string) *Guardrail {
	return &Guardrail{
		allowedTopics: allowedTopics,
	}
}


func (g *Guardrail) IsTopicAllowed(message string) bool {
	msgLower := strings.ToLower(strings.TrimSpace(message))
	if msgLower == "" {
		return false
	}

	greetings := []string{
		"halo", "hi", "hey", "pagi", "siang", "sore", "malam",
		"ping", "p", "tes", "test", "cek", "bantu", "tanya",
		"admin", "min", "gan", "kak", "bro", "terima kasih", "makasih", "thanks",
		"ya", "iya", "y", "ok", "oke", "okay", "siap", "baik", "boleh", "setuju",
		"tolong", "buat", "buatkan", "lapor", "laporan", "laporkan", "teknisi",
		"mohon", "bisa", "proses", "hubungi", "nama", "lokasi", "alamat",
		"kendala", "masalah", "dusun", "desa", "rt", "rw", "jalan", "jl", "no",
	}
	for _, greeting := range greetings {
		if msgLower == greeting || strings.HasPrefix(msgLower, greeting+" ") || strings.HasSuffix(msgLower, " "+greeting) || strings.Contains(msgLower, " "+greeting+" ") {
			return true
		}
	}

	ispKeywords := []string{
		"internet", "wifi", "wi-fi", "koneksi", "jaringan", "lemot", "lambat",
		"down", "mati", "rusak", "sinyal", "net", "speed", "kecepatan",
		"paket", "harga", "tagihan", "gangguan", "pemasangan", "coverage",
		"area", "pembayaran", "promo", "kontak", "alamat", "server", "los",
		"router", "modem", "kabel", "gnet", "ghaib", "lapor", "teknisi",
	}

	for _, kw := range ispKeywords {
		if strings.Contains(msgLower, kw) {
			return true
		}
	}

	for _, topic := range g.allowedTopics {
		if strings.Contains(msgLower, strings.ToLower(topic)) {
			return true
		}
	}

	return false
}

func (g *Guardrail) FormatOffTopicResponse() string {
	return "Maaf, saya hanya dapat membantu menjawab pertanyaan seputar layanan internet GNET (paket, harga, gangguan, coverage, dan pembayaran). Ada yang bisa saya bantu mengenai layanan GNET?"
}

func (g *Guardrail) SanitizeResponse(response string) string {
	return strings.TrimSpace(response)
}
