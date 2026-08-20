package skill

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/pkg/logger"
)

// --- Git Repositories Integration ---

func (u *ManageSkillUseCase) ListGitRepos() ([]skill.GitRepoInfo, error) {
	return u.store.ListGitRepos()
}

func (u *ManageSkillUseCase) AddGitRepo(ctx context.Context, repoURL string) (*skill.GitRepoInfo, error) {
	repoURL = strings.TrimSpace(repoURL)
	if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") && !strings.HasPrefix(repoURL, "git@") {
		return nil, skill.ErrInvalidGitURL
	}

	repos, _ := u.store.ListGitRepos()
	for _, r := range repos {
		if r.URL == repoURL {
			return nil, skill.ErrGitRepoAlreadyExists
		}
	}

	repoName := extractRepoName(repoURL)
	repoID := fmt.Sprintf("git-%s", repoName)

	newRepo := skill.GitRepoInfo{
		ID:           repoID,
		URL:          repoURL,
		Name:         repoName,
		Enabled:      true,
		LastSyncedAt: time.Now(),
	}

	if err := u.store.SaveGitRepo(newRepo); err != nil {
		return nil, err
	}

	// Trigger background clone
	if u.gitSyncer != nil {
		go func() {
			targetDir := filepath.Join(u.store.GetSkillsDir(), repoName)
			if err := u.gitSyncer.SyncRepo(context.Background(), targetDir, repoURL); err != nil {
				logger.WithComponent("ManageSkill").WithError(err).Errorf("Background git clone failed for %s", repoURL)
				return
			}
			u.persistMetadata(context.Background(), "", repoName, "git", repoURL)
		}()
	}

	return &newRepo, nil
}

func (u *ManageSkillUseCase) UpdateGitRepo(id, repoURL string, enabled *bool) (*skill.GitRepoInfo, error) {
	repos, err := u.store.ListGitRepos()
	if err != nil {
		return nil, err
	}

	for _, r := range repos {
		if r.ID == id {
			if repoURL != "" {
				r.URL = repoURL
			}
			if enabled != nil {
				r.Enabled = *enabled
			}
			if err := u.store.SaveGitRepo(r); err != nil {
				return nil, err
			}
			return &r, nil
		}
	}
	return nil, skill.ErrGitRepoNotFound
}

func (u *ManageSkillUseCase) DeleteGitRepo(ctx context.Context, id string) error {
	repos, _ := u.store.ListGitRepos()
	var repoName string
	for _, r := range repos {
		if r.ID == id {
			repoName = r.Name
			break
		}
	}

	if err := u.store.DeleteGitRepo(id); err != nil {
		return err
	}

	if repoName != "" {
		u.removeMetadata(ctx, "", repoName)
	}
	return nil
}

func (u *ManageSkillUseCase) SyncGitRepo(ctx context.Context, id string) error {
	repos, err := u.store.ListGitRepos()
	if err != nil {
		return err
	}

	var repo *skill.GitRepoInfo
	for _, r := range repos {
		if r.ID == id {
			repo = &r
			break
		}
	}
	if repo == nil {
		return skill.ErrGitRepoNotFound
	}

	if u.gitSyncer != nil {
		go func() {
			targetDir := filepath.Join(u.store.GetSkillsDir(), repo.Name)
			if err := u.gitSyncer.SyncRepo(context.Background(), targetDir, repo.URL); err != nil {
				logger.WithComponent("ManageSkill").WithError(err).Errorf("Manual git sync failed for %s", repo.URL)
				return
			}
			u.persistMetadata(context.Background(), "", repo.Name, "git", repo.URL)
		}()
	}
	return nil
}

func (u *ManageSkillUseCase) ToggleGitRepo(id string) (*skill.GitRepoInfo, error) {
	repos, err := u.store.ListGitRepos()
	if err != nil {
		return nil, err
	}

	for _, r := range repos {
		if r.ID == id {
			r.Enabled = !r.Enabled
			if err := u.store.SaveGitRepo(r); err != nil {
				return nil, err
			}
			return &r, nil
		}
	}
	return nil, skill.ErrGitRepoNotFound
}
