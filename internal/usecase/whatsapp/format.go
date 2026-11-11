package whatsapp

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// formatWhatsAppText converts the Markdown that language models commonly emit
// into the small formatting subset WhatsApp renders. It deliberately keeps the
// result plain enough for a chat bubble: no headings, tables, HTML, or
// Markdown-style links.
func formatWhatsAppText(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	code := &placeholderStore{prefix: "C"}
	text = replaceAllStringSubmatchFunc(fencedCodeRe, text, func(match []string) string {
		body := strings.TrimRight(match[2], "\n")
		return code.put("```\n" + body + "\n```")
	})
	text = replaceAllStringSubmatchFunc(inlineCodeRe, text, func(match []string) string {
		return code.put("```" + match[1] + "```")
	})

	text = replaceAllStringSubmatchFunc(markdownImageRe, text, func(match []string) string {
		alt := strings.TrimSpace(match[1])
		if alt == "" {
			return match[2]
		}
		return alt + " (" + match[2] + ")"
	})
	text = replaceAllStringSubmatchFunc(markdownLinkRe, text, func(match []string) string {
		return match[1] + " (" + match[2] + ")"
	})

	headings := &placeholderStore{prefix: "H"}
	text = replaceAllStringSubmatchFunc(markdownHeadingRe, text, func(match []string) string {
		title := strings.TrimSpace(match[1])
		title = strings.Trim(title, "*_")
		title = strings.ReplaceAll(title, "**", "")
		title = strings.ReplaceAll(title, "__", "")
		return headings.put("*" + title + "*")
	})

	emphasis := &placeholderStore{prefix: "E"}
	text = replaceAllStringSubmatchFunc(markdownBoldItalicRe, text, func(match []string) string {
		return emphasis.put("*_" + match[1] + "_*")
	})
	text = replaceAllStringSubmatchFunc(markdownBoldStarRe, text, func(match []string) string {
		return emphasis.put("*" + match[1] + "*")
	})
	text = replaceAllStringSubmatchFunc(markdownBoldUnderscoreRe, text, func(match []string) string {
		return emphasis.put("*" + match[1] + "*")
	})
	text = markdownItalicRe.ReplaceAllString(text, "${1}_${2}_${3}")
	text = replaceAllStringSubmatchFunc(markdownStrikeRe, text, func(match []string) string {
		return "~" + match[1] + "~"
	})

	text = replaceAllStringSubmatchFunc(markdownTaskRe, text, func(match []string) string {
		box := "☐"
		if strings.EqualFold(match[2], "x") {
			box = "☑"
		}
		return match[1] + "• " + box + " "
	})
	text = markdownBulletRe.ReplaceAllString(text, "${1}• ")
	text = markdownTableSeparatorRe.ReplaceAllString(text, "")
	text = normalizeTableRows(text)
	text = markdownHorizontalRuleRe.ReplaceAllString(text, "────────")
	text = htmlBreakRe.ReplaceAllString(text, "\n")
	text = htmlTagRe.ReplaceAllString(text, "")
	text = html.UnescapeString(text)

	text = emphasis.restore(text)
	text = headings.restore(text)
	text = code.restore(text)

	text = markdownBlankLinesRe.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

var (
	fencedCodeRe             = regexp.MustCompile("(?s)```([^\\n`]*)\\n?(.*?)```")
	inlineCodeRe             = regexp.MustCompile("`([^`\\n]+)`")
	markdownImageRe          = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)\)`)
	markdownLinkRe           = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
	markdownHeadingRe        = regexp.MustCompile(`(?m)^#{1,6}[ \t]+(.+?)[ \t]*#*[ \t]*$`)
	markdownBoldItalicRe     = regexp.MustCompile(`\*\*\*([^*\n]+)\*\*\*`)
	markdownBoldStarRe       = regexp.MustCompile(`\*\*([^*\n]+)\*\*`)
	markdownBoldUnderscoreRe = regexp.MustCompile(`__([^_\n]+)__`)
	markdownItalicRe         = regexp.MustCompile(`(^|[^*])\*([^*\n]+)\*([^*]|$)`)
	markdownStrikeRe         = regexp.MustCompile(`~~([^~\n]+)~~`)
	markdownTaskRe           = regexp.MustCompile(`(?m)^([ \t]*)[-*+][ \t]+\[([ xX])\][ \t]+`)
	markdownBulletRe         = regexp.MustCompile(`(?m)^([ \t]*)[-*+][ \t]+`)
	markdownTableSeparatorRe = regexp.MustCompile(`(?m)^[ \t]*\|?[ \t]*:?-{3,}:?[ \t]*(\|[ \t]*:?-{3,}:?[ \t]*)+\|?[ \t]*$`)
	markdownHorizontalRuleRe = regexp.MustCompile(`(?m)^[ \t]*(?:-{3,}|\*{3,}|_{3,})[ \t]*$`)
	htmlBreakRe              = regexp.MustCompile(`(?i)<br\s*/?>`)
	htmlTagRe                = regexp.MustCompile(`<[^>]+>`)
	markdownBlankLinesRe     = regexp.MustCompile(`\n{3,}`)
)

type placeholderStore struct {
	prefix string
	values []string
}

func (p *placeholderStore) put(value string) string {
	p.values = append(p.values, value)
	return fmt.Sprintf("\x00%s%d\x00", p.prefix, len(p.values)-1)
}

func (p *placeholderStore) restore(text string) string {
	for i, value := range p.values {
		text = strings.ReplaceAll(text, fmt.Sprintf("\x00%s%d\x00", p.prefix, i), value)
	}
	return text
}

// replaceAllStringSubmatchFunc is a small helper that lets the formatter build
// replacement strings per match without the limitations of ${n} expansion.
func replaceAllStringSubmatchFunc(re *regexp.Regexp, text string, fn func(match []string) string) string {
	indexes := re.FindAllStringSubmatchIndex(text, -1)
	if len(indexes) == 0 {
		return text
	}
	var b strings.Builder
	last := 0
	for _, index := range indexes {
		b.WriteString(text[last:index[0]])
		match := make([]string, 0, len(index)/2)
		for i := 0; i < len(index); i += 2 {
			if index[i] < 0 {
				match = append(match, "")
				continue
			}
			match = append(match, text[index[i]:index[i+1]])
		}
		b.WriteString(fn(match))
		last = index[1]
	}
	b.WriteString(text[last:])
	return b.String()
}

func normalizeTableRows(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, "|") {
			continue
		}
		trimmed = strings.TrimPrefix(trimmed, "|")
		trimmed = strings.TrimSuffix(trimmed, "|")
		cells := strings.Split(trimmed, "|")
		for j, cell := range cells {
			cells[j] = strings.TrimSpace(cell)
		}
		lines[i] = strings.Join(cells, " | ")
	}
	return strings.Join(lines, "\n")
}
