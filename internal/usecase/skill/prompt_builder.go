package skill

import (
	"fmt"
	"html"
	"strings"
	"text/template"

	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/pkg/logger"
)

// SkillsToolsHint adalah instruksi yang disisipkan ke system prompt saat skills_mode adalah "tools".
const SkillsToolsHint = "You have access to skills via the `request_skill` tool. " +
	"Call it with a skill name to retrieve the full skill instructions and SOP, then follow them to complete the task."

const defaultSkillsTemplate = `You can use the following skills to help with the task.
To request the skill, you need to use the ` + "`request_skill`" + ` tool. The skill name is the name of the skill you want to use.
<available_skills>
{{range .Skills}}
  <skill>
    <name>{{escapeXML .Name}}</name>
    {{if .Content}}<content>{{escapeXML .Content}}</content>{{else}}<description>{{escapeXML .Description}}</description>{{end}}
  </skill>
{{end}}
</available_skills>`

// RenderSkillsPrompt merender daftar skill menjadi format teks prompt XML untuk disuntikkan ke System Prompt LLM.
func RenderSkillsPrompt(skills []skill.SkillInfo, customTemplate string) string {
	if len(skills) == 0 {
		return ""
	}

	tmplText := customTemplate
	if tmplText == "" {
		tmplText = defaultSkillsTemplate
	}

	funcMap := template.FuncMap{
		"escapeXML": html.EscapeString,
	}

	tmpl, err := template.New("skills").Funcs(funcMap).Parse(tmplText)
	if err != nil {
		logger.WithComponent("SkillsPromptBuilder").WithError(err).Error("failed to parse skills template")
		// Fallback simple list
		var sb strings.Builder
		sb.WriteString("Available skills:\n")
		for _, s := range skills {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", s.Name, s.Description))
		}
		return sb.String()
	}

	data := map[string]any{
		"Skills": skills,
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		logger.WithComponent("SkillsPromptBuilder").WithError(err).Error("failed to execute skills template")
		return ""
	}

	return strings.TrimSpace(sb.String())
}

// RequestSkillArgs mendefinisikan argumen input untuk tool request_skill.
type RequestSkillArgs struct {
	SkillName string `json:"skill_name" jsonschema:"description=The name of the skill to request"`
}

// RequestSkillTool mengimplementasikan pemanggilan tool on-demand untuk mengambil konten lengkap SOP skill.
type RequestSkillTool struct {
	Skills []skill.SkillInfo
}

func (t RequestSkillTool) Run(args RequestSkillArgs) (string, error) {
	for _, s := range t.Skills {
		if s.Name == args.SkillName {
			body := s.Content
			if body == "" {
				body = s.Description
			}
			return fmt.Sprintf("Skill '%s':\n%s", s.Name, body), nil
		}
	}
	available := skillNames(t.Skills)
	return fmt.Sprintf("Skill '%s' not found. Available skills: %s", args.SkillName, available), nil
}

func skillNames(skills []skill.SkillInfo) string {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return strings.Join(names, ", ")
}

// FilterSkills menyaring daftar skill agar hanya mencakup skill yang terdaftar di selectedSkills.
// Jika selectedSkills kosong, seluruh skill dikembalikan.
func FilterSkills(all []skill.SkillInfo, selectedSkills []string) []skill.SkillInfo {
	if len(selectedSkills) == 0 {
		return all
	}

	selected := make(map[string]bool, len(selectedSkills))
	for _, s := range selectedSkills {
		selected[s] = true
	}

	var filtered []skill.SkillInfo
	for _, s := range all {
		if selected[s.Name] {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
