package skill

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type ManageSkillUseCase struct {
	repo port.SkillRepository
	fs   port.SkillFileStore
}

func NewManageSkillUseCase(repo port.SkillRepository, fs port.SkillFileStore) *ManageSkillUseCase {
	return &ManageSkillUseCase{
		repo: repo,
		fs:   fs,
	}
}

func (u *ManageSkillUseCase) ListSkills(ctx context.Context) ([]skill.Skill, error) {
	return u.repo.ListSkills(ctx)
}

func (u *ManageSkillUseCase) GetSkill(ctx context.Context, slug string) (*skill.Skill, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, skill.ErrInvalidSlug
	}
	return u.repo.GetSkillBySlug(ctx, slug)
}

func (u *ManageSkillUseCase) CreateSkill(ctx context.Context, slug, name, description string) (*skill.Skill, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugRegex.MatchString(slug) {
		return nil, skill.ErrInvalidSlug
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, skill.ErrInvalidSkillName
	}

	existing, _ := u.repo.GetSkillBySlug(ctx, slug)
	if existing != nil {
		return nil, skill.ErrSkillAlreadyExists
	}

	initialContent := fmt.Sprintf(`---
name: %s
description: '%s'
---

# %s

Jelaskan tujuan utama skill ini secara ringkas di sini.

## 1. Ruang Lingkup & Batasan
- Hal yang **boleh** dijawab oleh skill ini: ...
- Hal yang **di luar cakupan** dan harus dialihkan: ...

## 2. Alur Prosedur / SOP
Jelaskan langkah demi langkah instruksi untuk bot:
1. Langkah 1: ...
2. Langkah 2: ...

## 3. Dokumen Referensi Pendukung
Hubungkan berkas yang ada di folder `+"`references/`"+`:

| Kasus / Topik Pelanggan | Dokumen Referensi yang Dibaca |
|---|---|
| Info & Detail Khusus | `+"`references/panduan.md`"+` |

## 4. Kriteria Eskalasi ke Manusia
Tentukan kapan bot harus menyerah dan mengalihkan ke CS manusia:
- Pelanggan meminta berbicara langsung dengan petugas / teknisi.
- Kendala teknis tidak terselesaikan setelah mengikuti langkah SOP di atas.
`, slug, description, name)

	sk := &skill.Skill{
		Slug:        slug,
		Name:        name,
		Description: description,
		IsEnabled:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Files: []skill.SkillFile{
			{
				Name:        "SKILL.md",
				FilePath:    "SKILL.md",
				Content:     initialContent,
				IsReference: false,
				UpdatedAt:   time.Now(),
			},
		},
	}

	// 1. Simpan ke PostgreSQL
	if err := u.repo.CreateSkill(ctx, sk); err != nil {
		return nil, fmt.Errorf("failed to save skill to database: %w", err)
	}

	// 2. Tulis ke Filesystem Disk
	if u.fs != nil {
		if err := u.fs.WriteSkillToDisk(sk); err != nil {
			logger.WithComponent("ManageSkill").WithError(err).Warn("failed to write skill to disk")
		}
	}

	return sk, nil
}

func (u *ManageSkillUseCase) SaveSkillFile(ctx context.Context, slug, filePath, content string, isReference bool) (*skill.SkillFile, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	sk, err := u.repo.GetSkillBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	filePath = filepath.Clean(filePath)
	fileName := filepath.Base(filePath)
	if isReference && !strings.HasPrefix(filePath, "references/") && !strings.HasPrefix(filePath, "references\\") {
		filePath = filepath.Join("references", fileName)
	}

	f := &skill.SkillFile{
		SkillID:     sk.ID,
		Name:        fileName,
		FilePath:    filePath,
		Content:     content,
		IsReference: isReference,
		UpdatedAt:   time.Now(),
	}

	// 1. Update frontmatter metadata jika mengedit SKILL.md
	if fileName == "SKILL.md" {
		parsedName, parsedDesc, _, err := skill.ParseFrontmatter(content)
		if err == nil {
			if parsedName != "" {
				sk.Name = parsedName
			}
			if parsedDesc != "" {
				sk.Description = parsedDesc
			}
			_ = u.repo.UpdateSkill(ctx, sk)
		}
	}

	// 2. Simpan ke Database
	if err := u.repo.SaveSkillFile(ctx, sk.ID, f); err != nil {
		return nil, fmt.Errorf("failed to save file to database: %w", err)
	}

	// 3. Tulis ke Filesystem Disk
	if u.fs != nil {
		if err := u.fs.WriteSkillFileToDisk(slug, f); err != nil {
			logger.WithComponent("ManageSkill").WithError(err).Warn("failed to write file to disk")
		}
	}

	return f, nil
}

func (u *ManageSkillUseCase) DeleteSkillFile(ctx context.Context, slug string, fileID uint, filePath string) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	sk, _ := u.repo.GetSkillBySlug(ctx, slug)
	if sk != nil {
		_ = u.repo.DeleteSkillFileByPath(ctx, sk.ID, filePath)
	}
	if fileID > 0 {
		_ = u.repo.DeleteSkillFile(ctx, fileID)
	}
	if u.fs != nil {
		_ = u.fs.DeleteSkillFileFromDisk(slug, filePath)
	}
	return nil
}

func (u *ManageSkillUseCase) DeleteSkill(ctx context.Context, id uint, slug string) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if id == 0 && slug != "" {
		existing, _ := u.repo.GetSkillBySlug(ctx, slug)
		if existing != nil {
			id = existing.ID
		}
	}
	if id > 0 {
		if err := u.repo.DeleteSkill(ctx, id); err != nil {
			return err
		}
	}
	if u.fs != nil {
		_ = u.fs.DeleteSkillFolderFromDisk(slug)
	}
	return nil
}

func (u *ManageSkillUseCase) ToggleSkillEnabled(ctx context.Context, slug string, enabled bool) error {
	return u.repo.ToggleSkillEnabled(ctx, slug, enabled)
}

func (u *ManageSkillUseCase) SyncFromDisk(ctx context.Context) (int, error) {
	if u.fs == nil {
		return 0, nil
	}
	diskSkills, err := u.fs.ScanAllSkillsFromDisk()
	if err != nil {
		return 0, fmt.Errorf("failed to scan skills from disk: %w", err)
	}

	// 1. Dapatkan daftar skill saat ini di DB untuk pruning skill yang sudah dihapus dari disk
	dbSkills, _ := u.repo.ListSkills(ctx)
	diskSlugMap := make(map[string]bool)
	for _, ds := range diskSkills {
		diskSlugMap[ds.Slug] = true
	}
	for _, dbs := range dbSkills {
		if !diskSlugMap[dbs.Slug] {
			_ = u.repo.DeleteSkill(ctx, dbs.ID)
		}
	}

	// 2. Sinkronkan dan perbarui seluruh file skill
	syncedCount := 0
	for _, ds := range diskSkills {
		existing, _ := u.repo.GetSkillBySlug(ctx, ds.Slug)
		if existing == nil {
			if err := u.repo.CreateSkill(ctx, &ds); err == nil {
				syncedCount++
			}
		} else {
			// Prune berkas DB yang sudah dihapus dari disk
			diskFileMap := make(map[string]bool)
			for _, f := range ds.Files {
				diskFileMap[f.FilePath] = true
				diskFileMap[f.Name] = true
			}
			for _, dbf := range existing.Files {
				if !diskFileMap[dbf.FilePath] && !diskFileMap[dbf.Name] {
					_ = u.repo.DeleteSkillFile(ctx, dbf.ID)
				}
			}

			// Simpan atau update berkas disk
			for _, f := range ds.Files {
				_ = u.repo.SaveSkillFile(ctx, existing.ID, &f)
			}
			syncedCount++
		}
	}

	// 3. Sync global prompt dari data/system-prompt.md
	diskPrompt, err := u.fs.ReadGlobalPromptFromDisk()
	if err == nil && strings.TrimSpace(diskPrompt) != "" {
		_ = u.repo.SaveGlobalSystemPrompt(ctx, diskPrompt)
	}

	return syncedCount, nil
}

func (u *ManageSkillUseCase) GetGlobalSystemPrompt(ctx context.Context) (string, error) {
	content, err := u.repo.GetGlobalSystemPrompt(ctx)
	if err == nil && content != "" {
		return content, nil
	}
	if u.fs != nil {
		return u.fs.ReadGlobalPromptFromDisk()
	}
	return "", nil
}

func (u *ManageSkillUseCase) SaveGlobalSystemPrompt(ctx context.Context, content string) error {
	if err := u.repo.SaveGlobalSystemPrompt(ctx, content); err != nil {
		return err
	}
	if u.fs != nil {
		_ = u.fs.WriteGlobalPromptToDisk(content)
	}
	return nil
}

// BuildCompositeSystemPrompt merakit System Prompt Lengkap: Base System Prompt + Seluruh Skill yang Aktif.
func (u *ManageSkillUseCase) BuildCompositeSystemPrompt(ctx context.Context) (string, error) {
	var sb strings.Builder

	// 1. Base / Global System Prompt
	basePrompt, _ := u.GetGlobalSystemPrompt(ctx)
	if strings.TrimSpace(basePrompt) != "" {
		sb.WriteString(strings.TrimSpace(basePrompt))
		sb.WriteString("\n\n")
	}

	// 2. Active Modular Skills
	skills, err := u.repo.ListSkills(ctx)
	if err == nil && len(skills) > 0 {
		hasActive := false
		for _, sk := range skills {
			if !sk.IsEnabled {
				continue
			}
			if !hasActive {
				sb.WriteString("## MODUL SKILL & PROSEDUR SOP AKTIF:\n\n")
				hasActive = true
			}
			sb.WriteString(fmt.Sprintf("### SKILL: %s (%s)\n", sk.Name, sk.Slug))
			if sk.Description != "" {
				sb.WriteString(fmt.Sprintf("**Deskripsi Pemicu**: %s\n\n", sk.Description))
			}

			// Render isi berkas (utamakan SKILL.md lalu berkas referensi)
			for _, f := range sk.Files {
				content := strings.TrimSpace(f.Content)
				if content == "" {
					continue
				}
				sb.WriteString(fmt.Sprintf("#### [BERKAS: %s]\n", f.FilePath))
				sb.WriteString(content)
				sb.WriteString("\n\n")
			}
			sb.WriteString("---\n\n")
		}
	}

	return strings.TrimSpace(sb.String()), nil
}
