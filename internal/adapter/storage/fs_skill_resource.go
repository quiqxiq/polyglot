package storage

import (
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

// --- Resource Management ---

func (s *FSSkillStore) ListResources(skillName string) ([]skill.SkillResource, error) {
	if err := skill.ValidateSkillName(skillName); err != nil {
		return nil, err
	}

	skillDir := filepath.Join(s.skillsDir, skillName)
	if _, err := os.Stat(skillDir); err != nil {
		return nil, skill.ErrSkillNotFound
	}

	var resources []skill.SkillResource
	err := filepath.Walk(skillDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return nil
		}

		relPath, relErr := filepath.Rel(skillDir, path)
		if relErr != nil || relPath == "SKILL.md" {
			return nil
		}

		resType := "asset"
		lower := strings.ToLower(relPath)
		if strings.HasPrefix(lower, "scripts/") || strings.HasPrefix(lower, "scripts\\") || strings.HasSuffix(lower, ".rsc") || strings.HasSuffix(lower, ".sh") || strings.HasSuffix(lower, ".py") {
			resType = "script"
		} else if strings.HasPrefix(lower, "references/") || strings.HasPrefix(lower, "references\\") || strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".txt") {
			resType = "reference"
		}

		// Detect MIME and readability
		readable := true
		mimeType := "text/plain; charset=utf-8"
		if buf, rErr := os.ReadFile(path); rErr == nil && len(buf) > 0 {
			detected := http.DetectContentType(buf)
			mimeType = detected
			if !strings.HasPrefix(detected, "text/") && !strings.Contains(detected, "json") && !strings.Contains(detected, "xml") {
				readable = false
			}
		}

		resources = append(resources, skill.SkillResource{
			Path:     filepath.ToSlash(relPath),
			Name:     info.Name(),
			Type:     resType,
			Size:     info.Size(),
			MimeType: mimeType,
			Readable: readable,
			Modified: info.ModTime(),
		})
		return nil
	})

	return resources, err
}

func (s *FSSkillStore) ReadResource(skillName, path string) (*skill.ResourceContent, *skill.SkillResource, error) {
	if err := skill.ValidateSkillName(skillName); err != nil {
		return nil, nil, err
	}
	if err := skill.ValidateResourcePath(path); err != nil {
		return nil, nil, err
	}

	fullPath := filepath.Join(s.skillsDir, skillName, filepath.Clean(path))
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, nil, skill.ErrResourceNotFound
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil, err
	}

	mimeType := http.DetectContentType(data)
	isText := strings.HasPrefix(mimeType, "text/") || strings.Contains(mimeType, "json") || strings.Contains(mimeType, "xml")

	encoding := "raw"
	contentStr := string(data)
	if !isText {
		encoding = "base64"
		contentStr = base64.StdEncoding.EncodeToString(data)
	}

	resInfo := &skill.SkillResource{
		Path:     filepath.ToSlash(path),
		Name:     info.Name(),
		Type:     "asset",
		Size:     info.Size(),
		MimeType: mimeType,
		Readable: isText,
		Modified: info.ModTime(),
	}

	resContent := &skill.ResourceContent{
		Content:  contentStr,
		Encoding: encoding,
		MimeType: mimeType,
		Size:     info.Size(),
	}

	return resContent, resInfo, nil
}

func (s *FSSkillStore) WriteResource(skillName, path string, data []byte) error {
	sk, err := s.ReadSkillFromDisk(skillName)
	if err != nil {
		return err
	}
	if sk.ReadOnly {
		return skill.ErrReadOnlySkill
	}
	if err := skill.ValidateResourcePath(path); err != nil {
		return err
	}

	fullPath := filepath.Join(s.skillsDir, skillName, filepath.Clean(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (s *FSSkillStore) DeleteResource(skillName, path string) error {
	sk, err := s.ReadSkillFromDisk(skillName)
	if err != nil {
		return err
	}
	if sk.ReadOnly {
		return skill.ErrReadOnlySkill
	}
	if err := skill.ValidateResourcePath(path); err != nil {
		return err
	}

	fullPath := filepath.Join(s.skillsDir, skillName, filepath.Clean(path))
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
