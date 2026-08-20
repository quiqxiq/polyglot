package skill

import (
	"context"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// ManageSkillUseCase mengorkestrasikan seluruh operasi manajemen skill (LocalAI Write-Through pattern).
type ManageSkillUseCase struct {
	repo      port.SkillRepository
	store     port.SkillStore
	gitSyncer port.SkillGitSyncer
}

func NewManageSkillUseCase(repo port.SkillRepository, store port.SkillStore, gitSyncer port.SkillGitSyncer) *ManageSkillUseCase {
	return &ManageSkillUseCase{
		repo:      repo,
		store:     store,
		gitSyncer: gitSyncer,
	}
}

// Provider mengembalikan port.SkillProvider untuk bot engine runtime.
func (u *ManageSkillUseCase) Provider() port.SkillProvider {
	return &skillProviderAdapter{uc: u}
}

type skillProviderAdapter struct {
	uc *ManageSkillUseCase
}

func (a *skillProviderAdapter) ListSkills(ctx context.Context) ([]skill.SkillInfo, error) {
	return a.uc.ListSkillsForProvider(ctx)
}

func (a *skillProviderAdapter) GetSkillContent(ctx context.Context, name string) (string, error) {
	return a.uc.GetSkillContent(ctx, name)
}

// --- Skills CRUD ---

func (u *ManageSkillUseCase) ListSkills(ctx context.Context, userID string) ([]skill.Skill, error) {
	// 1. Coba baca dari PostgreSQL jika repo tersedia
	if u.repo != nil {
		records, err := u.repo.ListSkills(ctx, userID)
		if err == nil && len(records) > 0 {
			skills := make([]skill.Skill, 0, len(records))
			for _, r := range records {
				if !r.Enabled {
					continue
				}
				// Baca detail penuh dari disk jika ada
				sk, readErr := u.store.ReadSkillFromDisk(r.Name)
				if readErr == nil && sk != nil {
					sk.SourceType = r.SourceType
					sk.SourceURL = r.SourceURL
					skills = append(skills, *sk)
				} else {
					skills = append(skills, skill.Skill{
						ID:          r.Name,
						Name:        r.Name,
						Description: r.Definition,
						SourceType:  r.SourceType,
						SourceURL:   r.SourceURL,
					})
				}
			}
			return skills, nil
		}
	}

	// 2. Fallback baca langsung dari Filesystem disk
	fsSkills, err := u.store.ListSkillsFromDisk()
	if err != nil {
		return nil, err
	}

	// Auto-seed metadata ke PostgreSQL jika kosong
	if u.repo != nil && len(fsSkills) > 0 {
		for _, s := range fsSkills {
			u.persistMetadata(ctx, userID, s.Name, s.SourceType, s.SourceURL)
		}
	}

	return fsSkills, nil
}

func (u *ManageSkillUseCase) GetSkill(ctx context.Context, userID, name string) (*skill.Skill, error) {
	name = strings.TrimSpace(name)
	if err := skill.ValidateSkillName(name); err != nil {
		return nil, err
	}
	return u.store.ReadSkillFromDisk(name)
}

func (u *ManageSkillUseCase) SearchSkills(ctx context.Context, query string) ([]skill.Skill, error) {
	allSkills, err := u.store.ListSkillsFromDisk()
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))
	if queryLower == "" {
		return allSkills, nil
	}

	var results []skill.Skill
	for _, s := range allSkills {
		if strings.Contains(strings.ToLower(s.Name), queryLower) ||
			strings.Contains(strings.ToLower(s.Description), queryLower) ||
			strings.Contains(strings.ToLower(s.Content), queryLower) {
			results = append(results, s)
		}
	}
	return results, nil
}

func (u *ManageSkillUseCase) CreateSkill(ctx context.Context, userID, name, description, content, license, compatibility, allowedTools string, metadata map[string]string) (*skill.Skill, error) {
	sk, err := u.store.CreateSkillOnDisk(name, description, content, license, compatibility, allowedTools, metadata)
	if err != nil {
		return nil, err
	}

	u.persistMetadata(ctx, userID, name, "inline", "")
	return sk, nil
}

func (u *ManageSkillUseCase) UpdateSkill(ctx context.Context, userID, name, description, content, license, compatibility, allowedTools string, metadata map[string]string) (*skill.Skill, error) {
	sk, err := u.store.UpdateSkillOnDisk(name, description, content, license, compatibility, allowedTools, metadata)
	if err != nil {
		return nil, err
	}

	u.persistMetadata(ctx, userID, name, sk.SourceType, sk.SourceURL)
	return sk, nil
}

func (u *ManageSkillUseCase) DeleteSkill(ctx context.Context, userID, name string) error {
	if err := u.store.DeleteSkillFromDisk(name); err != nil {
		return err
	}
	u.removeMetadata(ctx, userID, name)
	return nil
}

func (u *ManageSkillUseCase) ExportSkill(name string) ([]byte, error) {
	return u.store.ExportSkillZip(name)
}

func (u *ManageSkillUseCase) ImportSkill(ctx context.Context, userID string, archiveData []byte) (*skill.Skill, error) {
	sk, err := u.store.ImportSkillZip(archiveData)
	if err != nil {
		return nil, err
	}
	u.persistMetadata(ctx, userID, sk.Name, "inline", "")
	return sk, nil
}

func (u *ManageSkillUseCase) ToggleSkillEnabled(ctx context.Context, userID, name string, enabled bool) error {
	if u.repo == nil {
		return nil
	}
	return u.repo.ToggleSkillEnabled(ctx, userID, name, enabled)
}

// --- Global System Prompt ---

func (u *ManageSkillUseCase) GetGlobalSystemPrompt(ctx context.Context) (string, error) {
	if u.repo != nil {
		prompt, err := u.repo.GetGlobalSystemPrompt(ctx)
		if err == nil && strings.TrimSpace(prompt) != "" {
			return prompt, nil
		}
	}
	return u.store.ReadGlobalPromptFromDisk()
}

func (u *ManageSkillUseCase) SaveGlobalSystemPrompt(ctx context.Context, content string) error {
	if u.repo != nil {
		_ = u.repo.SaveGlobalSystemPrompt(ctx, content)
	}
	return u.store.WriteGlobalPromptToDisk(content)
}

// --- SkillProvider Implementation for Agent Runtime ---

func (u *ManageSkillUseCase) ListSkillsForProvider(ctx context.Context) ([]skill.SkillInfo, error) {
	skills, err := u.store.ListSkillsFromDisk()
	if err != nil {
		return nil, err
	}

	out := make([]skill.SkillInfo, len(skills))
	for i, s := range skills {
		out[i] = skill.SkillInfo{
			Name:        s.Name,
			Description: s.Description,
			Content:     s.Content,
		}
	}
	return out, nil
}

func (u *ManageSkillUseCase) GetSkillContent(ctx context.Context, name string) (string, error) {
	sk, err := u.store.ReadSkillFromDisk(name)
	if err != nil {
		return "", err
	}
	if sk.Content != "" {
		return sk.Content, nil
	}
	return sk.Description, nil
}

// --- Internal Helpers ---

func (u *ManageSkillUseCase) persistMetadata(ctx context.Context, userID, name, sourceType, sourceURL string) {
	if u.repo == nil {
		return
	}

	definition := ""
	if sk, err := u.store.ReadSkillFromDisk(name); err == nil && sk != nil {
		definition = sk.Content
		if len(definition) > 500 {
			definition = definition[:500]
		}
		if definition == "" {
			definition = sk.Description
		}
	}

	rec := &skill.SkillMetadataRecord{
		UserID:     userID,
		Name:       name,
		Definition: definition,
		SourceType: sourceType,
		SourceURL:  sourceURL,
		Enabled:    true,
	}

	if err := u.repo.SaveSkillMetadata(ctx, rec); err != nil {
		logger.WithComponent("ManageSkill").WithError(err).Warnf("Failed to persist skill metadata for %s", name)
	}
}

func (u *ManageSkillUseCase) removeMetadata(ctx context.Context, userID, name string) {
	if u.repo == nil {
		return
	}
	if err := u.repo.DeleteSkillMetadata(ctx, userID, name); err != nil {
		logger.WithComponent("ManageSkill").WithError(err).Warnf("Failed to delete skill metadata for %s", name)
	}
}

func extractRepoName(repoURL string) string {
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) > 0 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return "repo"
}
