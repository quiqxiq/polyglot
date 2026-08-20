package port

import (
	"github.com/quixiq/polyglot/internal/domain/skill"
)

// SkillFileStore mengelola operasi baca/tulis berkas fisik di filesystem server (data/skills/ dan data/system-prompt.md).
type SkillFileStore interface {
	ScanAllSkillsFromDisk() ([]skill.Skill, error)
	ReadSkillFromDisk(slug string) (*skill.Skill, error)
	WriteSkillToDisk(s *skill.Skill) error
	WriteSkillFileToDisk(slug string, f *skill.SkillFile) error
	DeleteSkillFileFromDisk(slug string, filePath string) error
	DeleteSkillFolderFromDisk(slug string) error

	ReadGlobalPromptFromDisk() (string, error)
	WriteGlobalPromptToDisk(content string) error
}
