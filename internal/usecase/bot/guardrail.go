package bot

import (
	"regexp"
	"strings"
)

var (
	thinkTagRegex    = regexp.MustCompile(`(?is)<(?:think|thought)>.*?</(?:think|thought)>`)
	unclosedThinkReg = regexp.MustCompile(`(?is)<(?:think|thought)>.*$`)
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
	res := strings.TrimSpace(response)
	// 1. Strip complete <think>...</think> and <thought>...</thought> blocks
	res = thinkTagRegex.ReplaceAllString(res, "")
	// 2. Strip unclosed <think>... blocks if model reached token limit during reasoning
	if strings.Contains(strings.ToLower(res), "<think>") || strings.Contains(strings.ToLower(res), "<thought>") {
		res = unclosedThinkReg.ReplaceAllString(res, "")
	}
	// 3. Remove stray tags
	res = strings.ReplaceAll(res, "</think>", "")
	res = strings.ReplaceAll(res, "</thought>", "")
	res = strings.ReplaceAll(res, "<think>", "")
	res = strings.ReplaceAll(res, "<thought>", "")
	res = strings.TrimSpace(res)

	// 4. Strip internal monologue reasoning if model outputs draft steps directly
	for _, marker := range []string{
		"**Final Output Generation:**",
		"**Constructing the final Indonesian response:**",
		"**Final Response:**",
		"**Final Output:**",
	} {
		if idx := strings.Index(res, marker); idx != -1 {
			res = strings.TrimSpace(res[idx+len(marker):])
		}
	}

	return strings.TrimSpace(res)
}
