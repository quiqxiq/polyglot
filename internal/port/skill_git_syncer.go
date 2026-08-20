package port

import (
	"context"
)

// SkillGitSyncer mengabstraksikan sinkronisasi background git clone / git pull untuk repositori skill.
type SkillGitSyncer interface {
	SyncRepo(ctx context.Context, targetDir, repoURL string) error
}
