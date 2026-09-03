package skill_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/skill"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
)

type fakeSkillStore struct {
	skills map[string]*skill.Skill
}

func newFakeSkillStore() *fakeSkillStore {
	return &fakeSkillStore{skills: make(map[string]*skill.Skill)}
}

func (s *fakeSkillStore) ListSkillsFromDisk() ([]skill.Skill, error) {
	var list []skill.Skill
	for _, sk := range s.skills {
		list = append(list, *sk)
	}
	return list, nil
}

func (s *fakeSkillStore) ReadSkillFromDisk(name string) (*skill.Skill, error) {
	sk, ok := s.skills[name]
	if !ok {
		return nil, skill.ErrSkillNotFound
	}
	return sk, nil
}

func (s *fakeSkillStore) CreateSkillOnDisk(name, description, content, license, compatibility, allowedTools string, metadata map[string]string) (*skill.Skill, error) {
	sk := &skill.Skill{
		Name:          name,
		Description:   description,
		Content:       content,
		License:       license,
		Compatibility: compatibility,
		AllowedTools:  allowedTools,
		Metadata:      metadata,
	}
	s.skills[name] = sk
	return sk, nil
}

func (s *fakeSkillStore) UpdateSkillOnDisk(name, description, content, license, compatibility, allowedTools string, metadata map[string]string) (*skill.Skill, error) {
	sk, ok := s.skills[name]
	if !ok {
		return nil, skill.ErrSkillNotFound
	}
	sk.Description = description
	sk.Content = content
	sk.License = license
	sk.Compatibility = compatibility
	sk.AllowedTools = allowedTools
	sk.Metadata = metadata
	return sk, nil
}

func (s *fakeSkillStore) DeleteSkillFromDisk(name string) error {
	delete(s.skills, name)
	return nil
}

func (s *fakeSkillStore) ExportSkillZip(name string) ([]byte, error) { return nil, nil }
func (s *fakeSkillStore) ImportSkillZip(archiveData []byte) (*skill.Skill, error) {
	return nil, nil
}
func (s *fakeSkillStore) ListResources(skillName string) ([]skill.SkillResource, error) {
	return nil, nil
}
func (s *fakeSkillStore) ReadResource(skillName, path string) (*skill.ResourceContent, *skill.SkillResource, error) {
	return nil, nil, nil
}
func (s *fakeSkillStore) WriteResource(skillName, path string, data []byte) error { return nil }
func (s *fakeSkillStore) DeleteResource(skillName, path string) error             { return nil }
func (s *fakeSkillStore) ListGitRepos() ([]skill.GitRepoInfo, error)              { return nil, nil }
func (s *fakeSkillStore) SaveGitRepo(repo skill.GitRepoInfo) error                { return nil }
func (s *fakeSkillStore) DeleteGitRepo(id string) error                           { return nil }
func (s *fakeSkillStore) GetSkillsDir() string                                    { return "/tmp/skills" }
func (s *fakeSkillStore) ReadGlobalPromptFromDisk() (string, error)               { return "", nil }
func (s *fakeSkillStore) WriteGlobalPromptToDisk(content string) error            { return nil }

type fakeSkillRepo struct {
	records      map[string]*skill.SkillMetadataRecord
	globalPrompt string
}

func newFakeSkillRepo() *fakeSkillRepo {
	return &fakeSkillRepo{
		records: make(map[string]*skill.SkillMetadataRecord),
	}
}

func (f *fakeSkillRepo) ListSkills(ctx context.Context, userID string) ([]skill.SkillMetadataRecord, error) {
	var list []skill.SkillMetadataRecord
	for _, r := range f.records {
		if userID == "" || r.UserID == userID {
			list = append(list, *r)
		}
	}
	return list, nil
}

func (f *fakeSkillRepo) GetSkill(ctx context.Context, userID, name string) (*skill.SkillMetadataRecord, error) {
	if r, ok := f.records[name]; ok {
		return r, nil
	}
	return nil, skill.ErrSkillNotFound
}

func (f *fakeSkillRepo) SaveSkillMetadata(ctx context.Context, rec *skill.SkillMetadataRecord) error {
	f.records[rec.Name] = rec
	return nil
}

func (f *fakeSkillRepo) DeleteSkillMetadata(ctx context.Context, userID, name string) error {
	delete(f.records, name)
	return nil
}

func (f *fakeSkillRepo) ListGitSkills(ctx context.Context) ([]skill.SkillMetadataRecord, error) {
	var list []skill.SkillMetadataRecord
	for _, r := range f.records {
		if r.SourceType == "git" && r.Enabled {
			list = append(list, *r)
		}
	}
	return list, nil
}

func (f *fakeSkillRepo) ToggleSkillEnabled(ctx context.Context, userID, name string, enabled bool) error {
	if r, ok := f.records[name]; ok {
		r.Enabled = enabled
		return nil
	}
	return skill.ErrSkillNotFound
}

func (f *fakeSkillRepo) GetGlobalSystemPrompt(ctx context.Context) (string, error) {
	return f.globalPrompt, nil
}

func (f *fakeSkillRepo) SaveGlobalSystemPrompt(ctx context.Context, content string) error {
	f.globalPrompt = content
	return nil
}

func TestManageSkillUseCase_Lifecycle(t *testing.T) {
	fsStore := newFakeSkillStore()
	repo := newFakeSkillRepo()
	uc := skillUC.NewManageSkillUseCase(repo, fsStore, nil)
	ctx := context.Background()

	// 1. Create skill (Write-Through)
	sk, err := uc.CreateSkill(
		ctx,
		"user-1",
		"cs-jaringan",
		"Panduan CS untuk menangani keluhan internet",
		"# SOP CS\n1. Sapa pelanggan.\n2. Minta nomor ID pelanggan.",
		"Apache-2.0",
		">=1.0.0",
		"mcp_device",
		map[string]string{"category": "cs"},
	)
	require.NoError(t, err)
	assert.Equal(t, "cs-jaringan", sk.Name)

	// Verify metadata persisted to repo
	meta, err := repo.GetSkill(ctx, "user-1", "cs-jaringan")
	require.NoError(t, err)
	assert.Equal(t, "cs-jaringan", meta.Name)
	assert.Equal(t, "inline", meta.SourceType)

	// 2. List skills
	skills, err := uc.ListSkills(ctx, "user-1")
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "cs-jaringan", skills[0].Name)

	// 3. Search skill
	searchResults, err := uc.SearchSkills(ctx, "keluhan")
	require.NoError(t, err)
	require.Len(t, searchResults, 1)

	// 4. Update skill
	updated, err := uc.UpdateSkill(
		ctx,
		"user-1",
		"cs-jaringan",
		"Panduan CS Baru",
		"# SOP CS Updated\n1. Sapa pelanggan dengan ramah.",
		"Apache-2.0",
		">=1.0.0",
		"mcp_device",
		map[string]string{"category": "cs"},
	)
	require.NoError(t, err)
	assert.Equal(t, "Panduan CS Baru", updated.Description)

	// 5. Provider integration
	provider := uc.Provider()
	providerSkills, err := provider.ListSkills(ctx)
	require.NoError(t, err)
	require.Len(t, providerSkills, 1)
	assert.Equal(t, "cs-jaringan", providerSkills[0].Name)

	content, err := provider.GetSkillContent(ctx, "cs-jaringan")
	require.NoError(t, err)
	assert.Contains(t, content, "SOP CS Updated")

	// 6. Delete skill
	err = uc.DeleteSkill(ctx, "user-1", "cs-jaringan")
	require.NoError(t, err)

	_, err = repo.GetSkill(ctx, "user-1", "cs-jaringan")
	assert.ErrorIs(t, err, skill.ErrSkillNotFound)
}
