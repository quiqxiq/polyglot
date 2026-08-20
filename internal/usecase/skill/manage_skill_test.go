package skill

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

type mockSkillRepo struct {
	skills       map[string]*skill.Skill
	globalPrompt string
}

func newMockSkillRepo() *mockSkillRepo {
	return &mockSkillRepo{
		skills: make(map[string]*skill.Skill),
	}
}

func (m *mockSkillRepo) ListSkills(ctx context.Context) ([]skill.Skill, error) {
	var list []skill.Skill
	for _, s := range m.skills {
		list = append(list, *s)
	}
	return list, nil
}

func (m *mockSkillRepo) GetSkillByID(ctx context.Context, id uint) (*skill.Skill, error) {
	for _, s := range m.skills {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, skill.ErrSkillNotFound
}

func (m *mockSkillRepo) GetSkillBySlug(ctx context.Context, slug string) (*skill.Skill, error) {
	s, ok := m.skills[slug]
	if !ok {
		return nil, skill.ErrSkillNotFound
	}
	return s, nil
}

func (m *mockSkillRepo) CreateSkill(ctx context.Context, s *skill.Skill) error {
	s.ID = uint(len(m.skills) + 1)
	m.skills[s.Slug] = s
	return nil
}

func (m *mockSkillRepo) UpdateSkill(ctx context.Context, s *skill.Skill) error {
	m.skills[s.Slug] = s
	return nil
}

func (m *mockSkillRepo) DeleteSkill(ctx context.Context, id uint) error {
	for slug, s := range m.skills {
		if s.ID == id {
			delete(m.skills, slug)
			return nil
		}
	}
	return nil
}

func (m *mockSkillRepo) ToggleSkillEnabled(ctx context.Context, slug string, enabled bool) error {
	s, ok := m.skills[slug]
	if !ok {
		return skill.ErrSkillNotFound
	}
	s.IsEnabled = enabled
	return nil
}

func (m *mockSkillRepo) SaveSkillFile(ctx context.Context, skillID uint, f *skill.SkillFile) error {
	for _, s := range m.skills {
		if s.ID == skillID {
			f.ID = uint(len(s.Files) + 1)
			f.SkillID = skillID
			s.Files = append(s.Files, *f)
			return nil
		}
	}
	return skill.ErrSkillNotFound
}

func (m *mockSkillRepo) DeleteSkillFile(ctx context.Context, fileID uint) error {
	return nil
}

func (m *mockSkillRepo) DeleteSkillFileByPath(ctx context.Context, skillID uint, filePath string) error {
	return nil
}

func (m *mockSkillRepo) GetGlobalSystemPrompt(ctx context.Context) (string, error) {
	return m.globalPrompt, nil
}

func (m *mockSkillRepo) SaveGlobalSystemPrompt(ctx context.Context, content string) error {
	m.globalPrompt = content
	return nil
}

func TestManageSkillUseCase_CreateAndCompositePrompt(t *testing.T) {
	repo := newMockSkillRepo()
	uc := NewManageSkillUseCase(repo, nil)

	ctx := context.Background()

	// 1. Set global prompt
	err := uc.SaveGlobalSystemPrompt(ctx, "Kamu adalah asisten Nia Ghaib Network.")
	if err != nil {
		t.Fatalf("SaveGlobalSystemPrompt error: %v", err)
	}

	// 2. Create skill with invalid slug
	_, err = uc.CreateSkill(ctx, "Invalid Slug!!", "CS", "Desc")
	if err == nil {
		t.Fatalf("expected error for invalid slug, got nil")
	}

	// 3. Create valid skill
	sk, err := uc.CreateSkill(ctx, "cs-jaringan", "CS Jaringan", "Menangani gangguan internet")
	if err != nil {
		t.Fatalf("CreateSkill error: %v", err)
	}
	if sk.Slug != "cs-jaringan" {
		t.Fatalf("expected slug cs-jaringan, got %s", sk.Slug)
	}

	// 4. Save reference file
	_, err = uc.SaveSkillFile(ctx, "cs-jaringan", "references/troubleshooting.md", "## Cara Cek Lampu ONT", true)
	if err != nil {
		t.Fatalf("SaveSkillFile error: %v", err)
	}

	// 5. Test composite prompt
	compositePrompt, err := uc.BuildCompositeSystemPrompt(ctx)
	if err != nil {
		t.Fatalf("BuildCompositeSystemPrompt error: %v", err)
	}

	if compositePrompt == "" {
		t.Fatalf("expected non-empty composite prompt")
	}
}
