package whatsapp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatWhatsAppText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "   \n\t ", want: ""},
		{name: "bold", input: "**Status**", want: "*Status*"},
		{name: "italic", input: "*miring*", want: "_miring_"},
		{name: "bold italic", input: "***penting***", want: "*_penting_*"},
		{name: "strike", input: "~~lama~~", want: "~lama~"},
		{name: "heading", input: "## Status Jaringan", want: "*Status Jaringan*"},
		{name: "bullets", input: "- satu\n- dua", want: "• satu\n• dua"},
		{name: "numbered list stays readable", input: "1. satu\n2. dua", want: "1. satu\n2. dua"},
		{name: "link", input: "[Dashboard](https://example.com)", want: "Dashboard (https://example.com)"},
		{name: "image", input: "![Foto](https://example.com/a.png)", want: "Foto (https://example.com/a.png)"},
		{name: "inline code", input: "Gunakan `ping` sekarang", want: "Gunakan ```ping``` sekarang"},
		{name: "fenced code", input: "```go\nfmt.Println(\"ok\")\n```", want: "```\nfmt.Println(\"ok\")\n```"},
		{name: "table", input: "| A | B |\n|---|---|\n| 1 | 2 |", want: "A | B\n\n1 | 2"},
		{name: "html", input: "Baris<br>baru <b>tebal</b>", want: "Baris\nbaru tebal"},
		{name: "task list", input: "- [x] selesai\n- [ ] belum", want: "• ☑ selesai\n• ☐ belum"},
		{name: "horizontal rule", input: "---", want: "────────"},
		{
			name:  "mixed markdown",
			input: "**Status**\n- ONT `normal`\n- Sinyal *baik*",
			want:  "*Status*\n• ONT ```normal```\n• Sinyal _baik_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatWhatsAppText(tt.input))
		})
	}
}

func TestFormatWhatsAppTextKeepsLiteralAsterisks(t *testing.T) {
	input := "2 * 3 = 6"
	assert.Equal(t, input, formatWhatsAppText(input))
}

func TestFormatWhatsAppTextNormalizesBlankLines(t *testing.T) {
	assert.Equal(t, "satu\n\ndua", formatWhatsAppText("satu\n\n\n\ndua"))
}

func TestFormatWhatsAppTextDoesNotEscapeHTML(t *testing.T) {
	assert.Equal(t, "A & B < C", formatWhatsAppText("A &amp; B &lt; C"))
}

func TestFormatWhatsAppTextKeepsLongText(t *testing.T) {
	input := strings.Repeat("a", maxAssistantReplyRunes+100)
	assert.Len(t, []rune(formatWhatsAppText(input)), maxAssistantReplyRunes+100)
}
