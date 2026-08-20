package skill

import (
	"fmt"
	"strings"
	"time"
)

// SkillFile merepresentasikan sebuah berkas teks di dalam skill (misal: SKILL.md atau references/profil.md).
type SkillFile struct {
	ID          uint      `json:"id"`
	SkillID     uint      `json:"skill_id"`
	Name        string    `json:"name"`        // e.g. "SKILL.md", "profil-perusahaan.md"
	FilePath    string    `json:"file_path"`   // e.g. "SKILL.md", "references/profil-perusahaan.md"
	Content     string    `json:"content"`     // Isi markdown
	IsReference bool      `json:"is_reference"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Skill merepresentasikan satu paket kemampuan bot modular.
type Skill struct {
	ID          uint        `json:"id"`
	Slug        string      `json:"slug"`        // identifier unik, nama folder (e.g. "ghaib-network-cs")
	Name        string      `json:"name"`        // e.g. "Ghaib Network — Customer Service Jaringan"
	Description string      `json:"description"` // Penjelasan kapan skill ini dipicu
	IsEnabled   bool        `json:"is_enabled"`  // Toggle aktif/nonaktif
	Files       []SkillFile `json:"files"`       // Berkas-berkas pendukung
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// ParseFrontmatter mengekstrak metadata 'name' dan 'description' dari bagian YAML frontmatter (antara --- dan ---) pada SKILL.md.
func ParseFrontmatter(content string) (name, description string, body string, err error) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return "", "", trimmed, nil
	}

	rest := strings.TrimPrefix(trimmed, "---")
	idx := strings.Index(rest, "---")
	if idx == -1 {
		return "", "", trimmed, fmt.Errorf("invalid frontmatter: missing closing ---")
	}

	fmBlock := rest[:idx]
	body = strings.TrimSpace(rest[idx+3:])

	lines := strings.Split(fmBlock, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

		switch key {
		case "name":
			name = val
		case "description":
			description = val
		}
	}

	return name, description, body, nil
}
