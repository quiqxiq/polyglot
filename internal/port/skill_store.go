package port

import (
	"github.com/quixiq/polyglot/internal/domain/skill"
)

// SkillStore mengelola operasi baca/tulis berkas fisik skill, resource files, zip archive, dan konfigurasi Git di disk lokal.
type SkillStore interface {
	// Skills CRUD di Filesystem
	ListSkillsFromDisk() ([]skill.Skill, error)
	ReadSkillFromDisk(name string) (*skill.Skill, error)
	CreateSkillOnDisk(name, description, content, license, compatibility, allowedTools string, metadata map[string]string) (*skill.Skill, error)
	UpdateSkillOnDisk(name, description, content, license, compatibility, allowedTools string, metadata map[string]string) (*skill.Skill, error)
	DeleteSkillFromDisk(name string) error

	// Export / Import ZIP
	ExportSkillZip(name string) ([]byte, error)
	ImportSkillZip(archiveData []byte) (*skill.Skill, error)

	// Resource Management (scripts, templates, references)
	ListResources(skillName string) ([]skill.SkillResource, error)
	ReadResource(skillName, path string) (*skill.ResourceContent, *skill.SkillResource, error)
	WriteResource(skillName, path string, data []byte) error
	DeleteResource(skillName, path string) error

	// Git Repository Configurations
	ListGitRepos() ([]skill.GitRepoInfo, error)
	SaveGitRepo(repo skill.GitRepoInfo) error
	DeleteGitRepo(id string) error

	// Path & Global System Prompt
	GetSkillsDir() string
	ReadGlobalPromptFromDisk() (string, error)
	WriteGlobalPromptToDisk(content string) error
}
