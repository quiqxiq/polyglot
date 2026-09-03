package skill_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/skill"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
)

type fakeGitStore struct {
	fakeSkillStore
	repos map[string]skill.GitRepoInfo
}

func newFakeGitStore() *fakeGitStore {
	return &fakeGitStore{
		fakeSkillStore: *newFakeSkillStore(),
		repos:          make(map[string]skill.GitRepoInfo),
	}
}

func (s *fakeGitStore) ListGitRepos() ([]skill.GitRepoInfo, error) {
	var list []skill.GitRepoInfo
	for _, r := range s.repos {
		list = append(list, r)
	}
	return list, nil
}

func (s *fakeGitStore) SaveGitRepo(repo skill.GitRepoInfo) error {
	s.repos[repo.ID] = repo
	return nil
}

func (s *fakeGitStore) DeleteGitRepo(id string) error {
	delete(s.repos, id)
	return nil
}

func TestManageSkillUseCase_GitRepos(t *testing.T) {
	ctx := context.Background()
	store := newFakeGitStore()
	repo := newFakeSkillRepo()
	uc := skillUC.NewManageSkillUseCase(repo, store, nil)

	// 1. Invalid git URL
	_, err := uc.AddGitRepo(ctx, "ftp://invalid-url.com")
	assert.ErrorIs(t, err, skill.ErrInvalidGitURL)

	// 2. Add valid git repo with trailing slash
	info, err := uc.AddGitRepo(ctx, "https://github.com/myorg/awesome-skills/")
	require.NoError(t, err)
	assert.Equal(t, "awesome-skills", info.Name)
	assert.Equal(t, "git-awesome-skills", info.ID)
	assert.True(t, info.Enabled)

	// 3. Duplicate git repo error
	_, err = uc.AddGitRepo(ctx, "https://github.com/myorg/awesome-skills/")
	assert.ErrorIs(t, err, skill.ErrGitRepoAlreadyExists)

	// 4. List git repos
	list, err := uc.ListGitRepos()
	require.NoError(t, err)
	require.Len(t, list, 1)

	// 5. Toggle git repo
	toggled, err := uc.ToggleGitRepo(info.ID)
	require.NoError(t, err)
	assert.False(t, toggled.Enabled)

	// 6. Delete git repo
	err = uc.DeleteGitRepo(ctx, info.ID)
	require.NoError(t, err)

	listAfter, err := uc.ListGitRepos()
	require.NoError(t, err)
	assert.Empty(t, listAfter)
}
