package storage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

var _ port.SkillGitSyncer = (*GitSyncer)(nil)

// GitSyncer melakukan operasi clone atau pull repositori Git di disk server.
type GitSyncer struct {
	timeout time.Duration
}

func NewGitSyncer(timeout time.Duration) *GitSyncer {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	return &GitSyncer{timeout: timeout}
}

func (g *GitSyncer) SyncRepo(ctx context.Context, targetDir, repoURL string) error {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return fmt.Errorf("git repo url cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	gitDir := filepath.Join(targetDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Repo already cloned, perform git pull
		logger.WithComponent("GitSyncer").WithField("target_dir", targetDir).Info("pulling repository updates")
		cmd := exec.CommandContext(ctx, "git", "-C", targetDir, "pull", "--ff-only")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git pull failed: %w (output: %s)", err, string(out))
		}
		return nil
	}

	// New repo, perform git clone
	logger.WithComponent("GitSyncer").WithFields(map[string]any{"repo_url": repoURL, "target_dir": targetDir}).Info("cloning repository")
	_ = os.RemoveAll(targetDir)
	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoURL, targetDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w (output: %s)", err, string(out))
	}

	return nil
}
