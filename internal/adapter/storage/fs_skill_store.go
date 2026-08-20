package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/internal/port"
)

var _ port.SkillFileStore = (*FSSkillStore)(nil)

// FSSkillStore mengelola pembacaan dan penulisan berkas fisik skill di direktori disk server.
type FSSkillStore struct {
	baseDir          string
	skillsDir        string
	globalPromptPath string
}

func NewFSSkillStore(baseDir string) (*FSSkillStore, error) {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "data"
	}
	skillsDir := filepath.Join(baseDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create skills directory: %w", err)
	}

	globalPromptPath := filepath.Join(baseDir, "system-prompt.md")
	return &FSSkillStore{
		baseDir:          baseDir,
		skillsDir:        skillsDir,
		globalPromptPath: globalPromptPath,
	}, nil
}

func (s *FSSkillStore) ScanAllSkillsFromDisk() ([]skill.Skill, error) {
	entries, err := os.ReadDir(s.skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []skill.Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		sk, err := s.ReadSkillFromDisk(slug)
		if err != nil {
			continue
		}
		if sk != nil {
			result = append(result, *sk)
		}
	}
	return result, nil
}

func (s *FSSkillStore) ReadSkillFromDisk(slug string) (*skill.Skill, error) {
	skillDir := filepath.Join(s.skillsDir, slug)
	stat, err := os.Stat(skillDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, skill.ErrSkillNotFound
		}
		return nil, err
	}
	if !stat.IsDir() {
		return nil, skill.ErrSkillNotFound
	}

	sk := &skill.Skill{
		Slug:      slug,
		Name:      slug,
		IsEnabled: true,
		UpdatedAt: stat.ModTime(),
	}

	var files []skill.SkillFile

	// 1. Read Root Files in skillDir
	rootEntries, err := os.ReadDir(skillDir)
	if err == nil {
		for _, e := range rootEntries {
			if e.IsDir() {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
				continue
			}
			fPath := filepath.Join(skillDir, e.Name())
			contentBytes, rErr := os.ReadFile(fPath)
			if rErr != nil {
				continue
			}
			content := string(contentBytes)
			fileInfo, _ := e.Info()
			modTime := time.Now()
			if fileInfo != nil {
				modTime = fileInfo.ModTime()
			}

			if e.Name() == "SKILL.md" {
				parsedName, parsedDesc, _, _ := skill.ParseFrontmatter(content)
				if parsedName != "" {
					sk.Name = parsedName
				}
				if parsedDesc != "" {
					sk.Description = parsedDesc
				}
			}

			files = append(files, skill.SkillFile{
				Name:        e.Name(),
				FilePath:    e.Name(),
				Content:     content,
				IsReference: false,
				UpdatedAt:   modTime,
			})
		}
	}

	// 2. Read References subfolder
	refDir := filepath.Join(skillDir, "references")
	refEntries, err := os.ReadDir(refDir)
	if err == nil {
		for _, e := range refEntries {
			if e.IsDir() {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
				continue
			}
			fPath := filepath.Join(refDir, e.Name())
			contentBytes, rErr := os.ReadFile(fPath)
			if rErr != nil {
				continue
			}
			fileInfo, _ := e.Info()
			modTime := time.Now()
			if fileInfo != nil {
				modTime = fileInfo.ModTime()
			}

			files = append(files, skill.SkillFile{
				Name:        e.Name(),
				FilePath:    filepath.Join("references", e.Name()),
				Content:     string(contentBytes),
				IsReference: true,
				UpdatedAt:   modTime,
			})
		}
	}

	sk.Files = files
	return sk, nil
}

func (s *FSSkillStore) WriteSkillToDisk(sk *skill.Skill) error {
	if sk == nil || strings.TrimSpace(sk.Slug) == "" {
		return skill.ErrInvalidSlug
	}
	skillDir := filepath.Join(s.skillsDir, sk.Slug)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(skillDir, "references"), 0755); err != nil {
		return err
	}

	for _, f := range sk.Files {
		if err := s.WriteSkillFileToDisk(sk.Slug, &f); err != nil {
			return err
		}
	}
	return nil
}

func (s *FSSkillStore) WriteSkillFileToDisk(slug string, f *skill.SkillFile) error {
	if f == nil || strings.TrimSpace(f.FilePath) == "" {
		return skill.ErrInvalidFileName
	}
	fullPath := filepath.Join(s.skillsDir, slug, f.FilePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(f.Content), 0644)
}

func (s *FSSkillStore) DeleteSkillFileFromDisk(slug string, filePath string) error {
	fullPath := filepath.Join(s.skillsDir, slug, filePath)
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *FSSkillStore) DeleteSkillFolderFromDisk(slug string) error {
	skillDir := filepath.Join(s.skillsDir, slug)
	return os.RemoveAll(skillDir)
}

func (s *FSSkillStore) ReadGlobalPromptFromDisk() (string, error) {
	// Priority 1: data/system-prompt.md
	if bytes, err := os.ReadFile(s.globalPromptPath); err == nil && len(bytes) > 0 {
		return string(bytes), nil
	}
	// Priority 2 fallback: data/skills/system-prompt.md
	skillsPromptPath := filepath.Join(s.skillsDir, "system-prompt.md")
	if bytes, err := os.ReadFile(skillsPromptPath); err == nil && len(bytes) > 0 {
		return string(bytes), nil
	}
	return "", nil
}

func (s *FSSkillStore) WriteGlobalPromptToDisk(content string) error {
	dir := filepath.Dir(s.globalPromptPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(s.globalPromptPath, []byte(content), 0644); err != nil {
		return err
	}

	// Clean up legacy duplicated file in data/skills/system-prompt.md if it exists
	skillsPromptPath := filepath.Join(s.skillsDir, "system-prompt.md")
	if _, statErr := os.Stat(skillsPromptPath); statErr == nil {
		_ = os.Remove(skillsPromptPath)
	}
	return nil
}
