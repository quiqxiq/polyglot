// DEVIASI: YAML frontmatter parsing for SKILL.md is co-located with domain skill metadata.
package skill

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var validSkillNameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// SkillMetadata merepresentasikan struktur metadata di dalam YAML frontmatter SKILL.md.
// SkillMetadata contains SKILL.md frontmatter.
//
//nolint:revive // Domain package context makes the explicit name clearer to callers.
type SkillMetadata struct {
	Name          string            `yaml:"name" json:"name"`
	Description   string            `yaml:"description" json:"description"`
	License       string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// BuildFrontmatter merender YAML Frontmatter standar untuk diletakkan di bagian atas file SKILL.md.
func BuildFrontmatter(name, description, license, compatibility, allowedTools string, metadata map[string]string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", strings.TrimSpace(name)))
	sb.WriteString(fmt.Sprintf("description: %s\n", strings.TrimSpace(description)))

	if strings.TrimSpace(license) != "" {
		sb.WriteString(fmt.Sprintf("license: %s\n", strings.TrimSpace(license)))
	}
	if strings.TrimSpace(compatibility) != "" {
		sb.WriteString(fmt.Sprintf("compatibility: %s\n", strings.TrimSpace(compatibility)))
	}
	if len(metadata) > 0 {
		sb.WriteString("metadata:\n")
		for k, v := range metadata {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", strings.TrimSpace(k), strings.TrimSpace(v)))
		}
	}
	if strings.TrimSpace(allowedTools) != "" {
		sb.WriteString(fmt.Sprintf("allowed-tools: %s\n", strings.TrimSpace(allowedTools)))
	}
	sb.WriteString("---\n\n")
	return sb.String()
}

// ParseFrontmatter memisahkan metadata YAML frontmatter dan body markdown dari file SKILL.md.
func ParseFrontmatter(rawContent string) (SkillMetadata, string, error) {
	trimmed := strings.TrimSpace(rawContent)
	if !strings.HasPrefix(trimmed, "---") {
		return SkillMetadata{}, trimmed, nil
	}

	rest := strings.TrimPrefix(trimmed, "---")
	idx := strings.Index(rest, "---")
	if idx == -1 {
		return SkillMetadata{}, trimmed, fmt.Errorf("invalid frontmatter: missing closing ---")
	}

	fmBlock := rest[:idx]
	body := strings.TrimSpace(rest[idx+3:])

	var meta SkillMetadata
	if err := yaml.Unmarshal([]byte(fmBlock), &meta); err != nil {
		// Fallback manual line-by-line parsing jika format YAML sederhana
		meta = parseFrontmatterFallback(fmBlock)
	}

	return meta, body, nil
}

func parseFrontmatterFallback(fmBlock string) SkillMetadata {
	var meta SkillMetadata
	meta.Metadata = make(map[string]string)
	lines := strings.Split(fmBlock, "\n")
	isInsideMetadata := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if isInsideMetadata {
			if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					k := strings.TrimSpace(parts[0])
					v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
					meta.Metadata[k] = v
				}
				continue
			}
			isInsideMetadata = false
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

		switch key {
		case "name":
			meta.Name = val
		case "description":
			meta.Description = val
		case "license":
			meta.License = val
		case "compatibility":
			meta.Compatibility = val
		case "allowed-tools", "allowed_tools":
			meta.AllowedTools = val
		case "metadata":
			isInsideMetadata = true
		}
	}
	return meta
}

// ValidateSkillName memverifikasi bahwa nama skill hanya berisi huruf kecil, angka, dan tanda hubung (-).
func ValidateSkillName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidSkillName
	}
	if !validSkillNameRegex.MatchString(name) {
		return ErrInvalidSkillName
	}
	return nil
}

// ValidateResourcePath memverifikasi bahwa path relatif resource aman dan tidak mengandung directory traversal.
func ValidateResourcePath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return ErrInvalidResourcePath
	}
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "\\") {
		return ErrInvalidResourcePath
	}
	return nil
}
