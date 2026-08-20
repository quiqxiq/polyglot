package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/internal/port"
)

var _ port.SkillStore = (*FSSkillStore)(nil)

// FSSkillStore mengelola penyimpanan fisik skill, resources, zip archive, dan git config di disk lokal.
type FSSkillStore struct {
	baseDir          string
	skillsDir        string
	gitConfigPath    string
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

	gitConfigPath := filepath.Join(skillsDir, "git_repos.json")
	globalPromptPath := filepath.Join(baseDir, "system-prompt.md")

	return &FSSkillStore{
		baseDir:          baseDir,
		skillsDir:        skillsDir,
		gitConfigPath:    gitConfigPath,
		globalPromptPath: globalPromptPath,
	}, nil
}

func (s *FSSkillStore) GetSkillsDir() string {
	return s.skillsDir
}

// --- Skills CRUD ---

func (s *FSSkillStore) ListSkillsFromDisk() ([]skill.Skill, error) {
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
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		sk, err := s.ReadSkillFromDisk(name)
		if err != nil || sk == nil {
			continue
		}
		result = append(result, *sk)
	}
	return result, nil
}

func (s *FSSkillStore) ReadSkillFromDisk(name string) (*skill.Skill, error) {
	if err := skill.ValidateSkillName(name); err != nil {
		return nil, err
	}

	skillDir := filepath.Join(s.skillsDir, name)
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

	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	contentBytes, err := os.ReadFile(skillMDPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, skill.ErrSkillNotFound
		}
		return nil, err
	}

	meta, body, _ := skill.ParseFrontmatter(string(contentBytes))
	skillName := meta.Name
	if skillName == "" {
		skillName = name
	}

	// Check if this skill is sourced from Git
	readOnly := false
	sourceType := "inline"
	sourceURL := ""
	gitRepos, _ := s.ListGitRepos()
	for _, gr := range gitRepos {
		if gr.Name == name || gr.ID == name {
			readOnly = true
			sourceType = "git"
			sourceURL = gr.URL
			break
		}
	}

	return &skill.Skill{
		ID:            name,
		Name:          skillName,
		Description:   meta.Description,
		Content:       body,
		License:       meta.License,
		Compatibility: meta.Compatibility,
		AllowedTools:  meta.AllowedTools,
		Metadata:      meta.Metadata,
		ReadOnly:      readOnly,
		SourceType:    sourceType,
		SourceURL:     sourceURL,
		CreatedAt:     stat.ModTime(),
		UpdatedAt:     stat.ModTime(),
	}, nil
}

func (s *FSSkillStore) CreateSkillOnDisk(name, description, content, license, compatibility, allowedTools string, metadata map[string]string) (*skill.Skill, error) {
	if err := skill.ValidateSkillName(name); err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "---") {
		parsedMeta, body, parseErr := skill.ParseFrontmatter(trimmed)
		if parseErr == nil {
			if parsedMeta.Name != "" {
				name = parsedMeta.Name
			}
			if parsedMeta.Description != "" {
				description = parsedMeta.Description
			}
			if parsedMeta.License != "" {
				license = parsedMeta.License
			}
			if parsedMeta.Compatibility != "" {
				compatibility = parsedMeta.Compatibility
			}
			if parsedMeta.AllowedTools != "" {
				allowedTools = parsedMeta.AllowedTools
			}
			if len(parsedMeta.Metadata) > 0 {
				metadata = parsedMeta.Metadata
			}
			content = body
		}
	}

	skillDir := filepath.Join(s.skillsDir, name)
	if _, err := os.Stat(skillDir); err == nil {
		return nil, skill.ErrSkillAlreadyExists
	}

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return nil, err
	}

	fm := skill.BuildFrontmatter(name, description, license, compatibility, allowedTools, metadata)
	fullDoc := fm + strings.TrimSpace(content) + "\n"

	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillMDPath, []byte(fullDoc), 0644); err != nil {
		_ = os.RemoveAll(skillDir)
		return nil, err
	}

	return s.ReadSkillFromDisk(name)
}

func (s *FSSkillStore) UpdateSkillOnDisk(name, description, content, license, compatibility, allowedTools string, metadata map[string]string) (*skill.Skill, error) {
	existing, err := s.ReadSkillFromDisk(name)
	if err != nil {
		return nil, err
	}
	if existing.ReadOnly {
		return nil, skill.ErrReadOnlySkill
	}

	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "---") {
		parsedMeta, body, parseErr := skill.ParseFrontmatter(trimmed)
		if parseErr == nil {
			if parsedMeta.Description != "" {
				description = parsedMeta.Description
			}
			if parsedMeta.License != "" {
				license = parsedMeta.License
			}
			if parsedMeta.Compatibility != "" {
				compatibility = parsedMeta.Compatibility
			}
			if parsedMeta.AllowedTools != "" {
				allowedTools = parsedMeta.AllowedTools
			}
			if len(parsedMeta.Metadata) > 0 {
				metadata = parsedMeta.Metadata
			}
			content = body
		}
	}

	skillDir := filepath.Join(s.skillsDir, name)
	fm := skill.BuildFrontmatter(name, description, license, compatibility, allowedTools, metadata)
	fullDoc := fm + strings.TrimSpace(content) + "\n"

	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillMDPath, []byte(fullDoc), 0644); err != nil {
		return nil, err
	}

	return s.ReadSkillFromDisk(name)
}

func (s *FSSkillStore) DeleteSkillFromDisk(name string) error {
	existing, err := s.ReadSkillFromDisk(name)
	if err != nil {
		return err
	}
	if existing.ReadOnly {
		return skill.ErrReadOnlySkill
	}

	skillDir := filepath.Join(s.skillsDir, name)
	return os.RemoveAll(skillDir)
}
