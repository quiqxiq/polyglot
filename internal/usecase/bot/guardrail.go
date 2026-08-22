package bot

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	thinkTagRegex    = regexp.MustCompile(`(?is)<(?:think|thought)>.*?</(?:think|thought)>`)
	unclosedThinkReg = regexp.MustCompile(`(?is)<(?:think|thought)>.*$`)

	fencedCodeRe     = regexp.MustCompile("(?s)```.*?```")
	inlineCodeRe     = regexp.MustCompile("`[^`\n]*`")
	linkRe           = regexp.MustCompile(`\[([^\]\n]*)\]\(([^()\s]+)\)`)
	boldItalicStarRe = regexp.MustCompile(`\*\*\*(.+?)\*\*\*`)
	boldStarRe       = regexp.MustCompile(`\*\*(.+?)\*\*`)
	boldUnderscoreRe = regexp.MustCompile(`__(.+?)__`)
	italicStarRe     = regexp.MustCompile(`\*([^*\n]+?)\*`)
	strikeRe         = regexp.MustCompile(`~~(.+?)~~`)
	headerRe         = regexp.MustCompile(`(?m)^#{1,6}[ \t]+(.+?)[ \t]*#*[ \t]*$`)
	hrLineRe         = regexp.MustCompile(`(?m)^[ \t]*([-*_][ \t]*){3,}$`)
	bulletRe         = regexp.MustCompile(`(?m)^([ \t]*)[*+-][ \t]+`)
	blankLinesRe     = regexp.MustCompile(`\n{3,}`)
	sepCellRe        = regexp.MustCompile(`^:?-+:?$`)
)

const whatsappHorizontalRule = "━━━━━━━━━━━━━━━━━━━━━"

type Guardrail struct{}

func NewGuardrail() *Guardrail {
	return &Guardrail{}
}

// SanitizeResponse membersihkan output model dari tag <think>, reasoning
// monologue, draf internal, lalu mengubah sisa sintaks Markdown menjadi
// format teks yang didukung WhatsApp (*bold*, _italic_, ~strike~, bullet •).
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

	// 5. Konversi sisa Markdown ke format WhatsApp sebelum pesan dikirim.
	res = g.MarkdownToWhatsApp(res)

	return strings.TrimSpace(res)
}

// MarkdownToWhatsApp mengubah sintaks Markdown standar (heading, bold,
// italic, strikethrough, link, bullet, tabel, horizontal rule) menjadi
// format khusus WhatsApp sehingga tidak ada simbol mentah seperti **,
// ###, ~~ atau [judul](url) yang sampai ke pelanggan.
//
// Urutan transformasi penting:
//  1. Blok kode ``` dan inline `code` dilindungi (tetap monospace di WA).
//  2. Link [judul](url) -> "judul (url)".
//  3. ***bold-italic*** -> *_text_*.
//  4. **bold** / __bold__ dipindah ke stash (agar aman dari konversi italic).
//  5. *italic* tunggal -> _italic_.
//  6. ~~strike~~ -> ~strike~.
//  7. Heading ### -> *Heading*.
//  8. Horizontal rule --- / *** -> garis ━━━.
//  9. Baris tabel markdown -> baris bullet terstruktur.
//
// 10. Bullet * / - / + -> •, lalu rapikan baris kosong berlebih.
func (g *Guardrail) MarkdownToWhatsApp(md string) string {
	res := strings.TrimSpace(md)
	if res == "" || !strings.ContainsAny(res, "*`#~_[|-") {
		return blankLinesRe.ReplaceAllString(res, "\n\n")
	}

	var stash []string
	keep := func(final string) string {
		tok := fmt.Sprintf("\x00%d\x00", len(stash))
		stash = append(stash, final)
		return tok
	}
	restore := func(s string) string {
		for i, v := range stash {
			s = strings.ReplaceAll(s, fmt.Sprintf("\x00%d\x00", i), v)
		}
		return s
	}

	// 1. Lindungi blok & inline code.
	res = fencedCodeRe.ReplaceAllStringFunc(res, func(m string) string {
		return keep(strings.TrimSpace(m))
	})
	res = inlineCodeRe.ReplaceAllStringFunc(res, keep)

	// 2. Link markdown.
	res = linkRe.ReplaceAllStringFunc(res, func(m string) string {
		parts := linkRe.FindStringSubmatch(m)
		title, url := strings.TrimSpace(parts[1]), parts[2]
		switch {
		case title == "", strings.EqualFold(title, url):
			return keep(url)
		default:
			return keep(title + " (" + url + ")")
		}
	})

	// 3. Bold-italic gabungan.
	res = boldItalicStarRe.ReplaceAllStringFunc(res, func(m string) string {
		sub := boldItalicStarRe.FindStringSubmatch(m)
		return keep("*_" + sub[1] + "_*")
	})

	// 4. Bold ke stash sebelum italic diproses.
	res = boldStarRe.ReplaceAllStringFunc(res, func(m string) string {
		sub := boldStarRe.FindStringSubmatch(m)
		return keep("*" + sub[1] + "*")
	})
	res = boldUnderscoreRe.ReplaceAllStringFunc(res, func(m string) string {
		sub := boldUnderscoreRe.FindStringSubmatch(m)
		return keep("*" + sub[1] + "*")
	})

	// 5. Italic tunggal (aman karena semua bold sudah distash).
	// Catatan: gunakan ${1} agar underscore tidak dianggap bagian nama grup.
	res = italicStarRe.ReplaceAllString(res, "_${1}_")

	// 6. Strikethrough.
	res = strikeRe.ReplaceAllString(res, "~$1~")

	// 7. Heading.
	res = headerRe.ReplaceAllString(res, "*$1*")

	// 8. Horizontal rule.
	res = hrLineRe.ReplaceAllString(res, whatsappHorizontalRule)

	// 9. Tabel markdown -> baris bullet.
	res = convertMarkdownTables(res)

	// 10. Bullet list.
	res = bulletRe.ReplaceAllString(res, "$1• ")

	// Kembalikan hasil bold/link/code dari stash.
	res = restore(res)

	res = blankLinesRe.ReplaceAllString(res, "\n\n")
	return strings.TrimSpace(res)
}

// convertMarkdownTables mengubah setiap baris tabel pipa menjadi baris
// bullet "• sel1 — sel2", membuang baris pemisah |---|---|.
func convertMarkdownTables(s string) string {
	if !strings.Contains(s, "|") {
		return s
	}
	const dropRowMarker = "\x03"
	lines := strings.Split(s, "\n")
	for i, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") {
			continue
		}
		cells := strings.Split(trimmed, "|")
		inner := cells[1 : len(cells)-1]

		hasCell, isSeparator := false, true
		for _, c := range inner {
			ct := strings.TrimSpace(c)
			if ct == "" {
				continue
			}
			hasCell = true
			if !sepCellRe.MatchString(ct) {
				isSeparator = false
				break
			}
		}
		switch {
		case hasCell && isSeparator:
			lines[i] = dropRowMarker
		case len(keptCells(inner)) == 0:
			lines[i] = dropRowMarker
		default:
			lines[i] = "• " + strings.Join(keptCells(inner), "  —  ")
		}
	}
	var out []string
	for _, ln := range lines {
		if ln != dropRowMarker {
			out = append(out, ln)
		}
	}
	return strings.Join(out, "\n")
}

// keptCells mengekstrak sel tabel yang berisi teks.
func keptCells(inner []string) []string {
	kept := make([]string, 0, len(inner))
	for _, c := range inner {
		if ct := strings.TrimSpace(c); ct != "" {
			kept = append(kept, ct)
		}
	}
	return kept
}
