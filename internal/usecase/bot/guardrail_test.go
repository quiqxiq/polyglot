package bot_test

import (
	"testing"

	"github.com/quixiq/polyglot/internal/usecase/bot"
	"github.com/stretchr/testify/assert"
)

func TestGuardrail_SanitizeResponse(t *testing.T) {
	g := bot.NewGuardrail()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal clean response",
			input:    "Halo Kak, ada yang bisa dibantu?",
			expected: "Halo Kak, ada yang bisa dibantu?",
		},
		{
			name:     "Complete <think> tag stripped",
			input:    "<think>User wants help with package info. Let's provide it.</think>Halo Kak! Ini paket internet kami.",
			expected: "Halo Kak! Ini paket internet kami.",
		},
		{
			name:     "Complete <thought> tag stripped",
			input:    "<thought>\nStep 1: Check greeting\nStep 2: Respond\n</thought>\nSelamat pagi Kak!",
			expected: "Selamat pagi Kak!",
		},
		{
			name:     "Unclosed <think> tag stripped",
			input:    "<think>Internal reasoning step 1, 2, 3...",
			expected: "",
		},
		{
			name: "Reasoning monologue with final output marker",
			input: `1. **Acknowledge & Empathize:** Friendly greeting.
2. **Drafting:** Halo kak.
**Final Output Generation:**
Halo Kak! Selamat datang di layanan kami.`,
			expected: "Halo Kak! Selamat datang di layanan kami.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := g.SanitizeResponse(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGuardrail_MarkdownToWhatsApp(t *testing.T) {
	g := bot.NewGuardrail()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bold double asterisk",
			input:    "**Teks Tebal**",
			expected: "*Teks Tebal*",
		},
		{
			name:     "italic single asterisk",
			input:    "*Teks Miring*",
			expected: "_Teks Miring_",
		},
		{
			name:     "strikethrough",
			input:    "harga ~~coret~~ final",
			expected: "harga ~coret~ final",
		},
		{
			name:     "header h3",
			input:    "### Bagian Tagihan",
			expected: "*Bagian Tagihan*",
		},
		{
			name:     "link with title",
			input:    "Cek di [Portal Login](http://10.10.10.1) ya Kak.",
			expected: "Cek di Portal Login (http://10.10.10.1) ya Kak.",
		},
		{
			name:     "bullet dash list",
			input:    "- Item satu\n- Item dua\n- Item tiga",
			expected: "• Item satu\n• Item dua\n• Item tiga",
		},
		{
			name:     "bullet star list",
			input:    "* poin a\n* poin b",
			expected: "• poin a\n• poin b",
		},
		{
			name:     "horizontal rule dashes",
			input:    "---",
			expected: "━━━━━━━━━━━━━━━━━━━━━",
		},
		{
			name:     "markdown table to bullet rows",
			input:    "| Paket | Harga |\n|---|---|\n| Home | 150rb |\n| Business | 300rb |",
			expected: "• Paket  —  Harga\n• Home  —  150rb\n• Business  —  300rb",
		},
		{
			name:     "bold and italic mixed order safe",
			input:    "**Bold** dan *italic* sekaligus",
			expected: "*Bold* dan _italic_ sekaligus",
		},
		{
			name:     "bold italic triple asterisk",
			input:    "***penting banget***",
			expected: "*_penting banget_*",
		},
		{
			name:     "inline code untouched",
			input:    "jalankan `ping 8.8.8.8` dulu ya",
			expected: "jalankan `ping 8.8.8.8` dulu ya",
		},
		{
			name:     "fenced code block untouched",
			input:    "```go\nfmt.Println(\"halo\")\n```",
			expected: "```go\nfmt.Println(\"halo\")\n```",
		},
		{
			name:     "multiple blank lines collapsed",
			input:    "paragraf satu\n\n\n\n\nparagraf dua",
			expected: "paragraf satu\n\nparagraf dua",
		},
		{
			name:     "plain text passthrough",
			input:    "Halo Kak, internet sudah normal kembali.",
			expected: "Halo Kak, internet sudah normal kembali.",
		},
		{
			name:     "full response mix",
			input:    "### Status Gangguan\n\n**Wilayah:** Desa Sukamaju\n\n- Tim sedang menuju lokasi\n- Estimasi: 1 jam\n\nHubungi [CS Ghaib](https://wa.me/62812) bila ada pertanyaan.",
			expected: "*Status Gangguan*\n\n*Wilayah:* Desa Sukamaju\n\n• Tim sedang menuju lokasi\n• Estimasi: 1 jam\n\nHubungi CS Ghaib (https://wa.me/62812) bila ada pertanyaan.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, g.MarkdownToWhatsApp(tc.input))
		})
	}
}

func TestGuardrail_SanitizeResponse_ConvertsMarkdown(t *testing.T) {
	g := bot.NewGuardrail()

	result := g.SanitizeResponse("<think>reasoning...</think>### Info Tagihan\n\n**Total:** Rp50.000")
	assert.Equal(t, "*Info Tagihan*\n\n*Total:* Rp50.000", result)
}
