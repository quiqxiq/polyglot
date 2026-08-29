package storage

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

// --- ZIP Export / Import ---

func (s *FSSkillStore) ExportSkillZip(name string) ([]byte, error) {
	if _, err := s.ReadSkillFromDisk(name); err != nil {
		return nil, err
	}

	skillDir := filepath.Join(s.skillsDir, name)
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err := filepath.Walk(skillDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return nil
		}

		relPath, relErr := filepath.Rel(skillDir, path)
		if relErr != nil {
			return nil
		}

		zipFileHeader, hErr := zip.FileInfoHeader(info)
		if hErr != nil {
			return hErr
		}
		zipFileHeader.Name = filepath.ToSlash(relPath)
		zipFileHeader.Method = zip.Deflate

		w, wErr := zipWriter.CreateHeader(zipFileHeader)
		if wErr != nil {
			return wErr
		}

		fileData, rErr := os.ReadFile(path)
		if rErr != nil {
			return rErr
		}
		_, writeErr := w.Write(fileData)
		return writeErr
	})

	if err != nil {
		return nil, err
	}
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *FSSkillStore) ImportSkillZip(archiveData []byte) (*skill.Skill, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return nil, fmt.Errorf("invalid zip archive: %w", err)
	}

	// 1. Detect root skill name from SKILL.md
	var targetSkillName string
	for _, f := range zipReader.File {
		cleanName := filepath.Clean(f.Name)
		if strings.HasSuffix(cleanName, "SKILL.md") {
			parts := strings.Split(filepath.ToSlash(cleanName), "/")
			if len(parts) > 1 {
				targetSkillName = parts[0]
			}
			break
		}
	}

	if targetSkillName == "" {
		// Read SKILL.md directly from root of zip
		for _, f := range zipReader.File {
			if f.Name == "SKILL.md" {
				rc, oErr := f.Open()
				if oErr == nil {
					data, readErr := io.ReadAll(rc)
					_ = rc.Close()
					if readErr != nil {
						break
					}
					meta, _, _ := skill.ParseFrontmatter(string(data))
					targetSkillName = meta.Name
				}
				break
			}
		}
	}

	if targetSkillName == "" {
		targetSkillName = fmt.Sprintf("imported-skill-%d", time.Now().Unix())
	}
	targetSkillName = strings.ToLower(strings.TrimSpace(targetSkillName))
	_ = skill.ValidateSkillName(targetSkillName)

	targetDir := filepath.Join(s.skillsDir, targetSkillName)
	_ = os.RemoveAll(targetDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, err
	}

	for _, f := range zipReader.File {
		cleanPath := filepath.Clean(f.Name)
		if strings.Contains(cleanPath, "..") {
			continue
		}

		// Strip root dir prefix if zip has top-level folder
		relPath := cleanPath
		if strings.HasPrefix(filepath.ToSlash(cleanPath), targetSkillName+"/") {
			relPath = strings.TrimPrefix(filepath.ToSlash(cleanPath), targetSkillName+"/")
		}

		destPath := filepath.Join(targetDir, relPath)
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(destPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			continue
		}

		rc, oErr := f.Open()
		if oErr != nil {
			continue
		}
		destFile, cErr := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if cErr == nil {
			_, _ = io.Copy(destFile, rc)
			_ = destFile.Close()
		}
		_ = rc.Close()
	}

	return s.ReadSkillFromDisk(targetSkillName)
}

// --- Git Repositories Configuration ---

func (s *FSSkillStore) ListGitRepos() ([]skill.GitRepoInfo, error) {
	if _, err := os.Stat(s.gitConfigPath); os.IsNotExist(err) {
		return []skill.GitRepoInfo{}, nil
	}

	data, err := os.ReadFile(s.gitConfigPath)
	if err != nil {
		return []skill.GitRepoInfo{}, nil
	}

	var repos []skill.GitRepoInfo
	if err := json.Unmarshal(data, &repos); err != nil {
		return []skill.GitRepoInfo{}, nil
	}
	return repos, nil
}

func (s *FSSkillStore) SaveGitRepo(repo skill.GitRepoInfo) error {
	repos, _ := s.ListGitRepos()
	found := false
	for i, r := range repos {
		if r.ID == repo.ID || r.URL == repo.URL {
			repos[i] = repo
			found = true
			break
		}
	}
	if !found {
		repos = append(repos, repo)
	}

	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.gitConfigPath, data, 0644)
}

func (s *FSSkillStore) DeleteGitRepo(id string) error {
	repos, _ := s.ListGitRepos()
	var updated []skill.GitRepoInfo
	var repoName string
	for _, r := range repos {
		if r.ID == id {
			repoName = r.Name
		} else {
			updated = append(updated, r)
		}
	}

	data, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.gitConfigPath, data, 0644); err != nil {
		return err
	}

	if repoName != "" {
		_ = os.RemoveAll(filepath.Join(s.skillsDir, repoName))
	}
	return nil
}

// --- Global System Prompt ---

func (s *FSSkillStore) ReadGlobalPromptFromDisk() (string, error) {
	data, err := os.ReadFile(s.globalPromptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (s *FSSkillStore) WriteGlobalPromptToDisk(content string) error {
	dir := filepath.Dir(s.globalPromptPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(s.globalPromptPath, []byte(content), 0644)
}
