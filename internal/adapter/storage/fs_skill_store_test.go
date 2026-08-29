package storage_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/storage"
	"github.com/quixiq/polyglot/internal/domain/skill"
)

func TestFSSkillStore_CRUDAndResources(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_skill_store_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	store, err := storage.NewFSSkillStore(tempDir)
	require.NoError(t, err)

	// 1. Create skill
	created, err := store.CreateSkillOnDisk(
		"troubleshoot-los",
		"Panduan penanganan lampu LOS merah",
		"# Langkah 1\nCek kabel fiber optik.",
		"Apache-2.0",
		">=1.0.0",
		"mcp_device",
		map[string]string{"category": "network"},
	)
	require.NoError(t, err)
	assert.Equal(t, "troubleshoot-los", created.Name)
	assert.Equal(t, "Panduan penanganan lampu LOS merah", created.Description)

	// 2. Read skill
	read, err := store.ReadSkillFromDisk("troubleshoot-los")
	require.NoError(t, err)
	assert.Equal(t, "troubleshoot-los", read.Name)
	assert.Contains(t, read.Content, "Cek kabel fiber optik")

	// 3. Write Resource File (e.g. scripts/diag.rsc)
	rscScript := []byte("/interface/ethernet print detail\n/system/health print")
	err = store.WriteResource("troubleshoot-los", "scripts/diag.rsc", rscScript)
	require.NoError(t, err)

	// 4. List Resources
	resources, err := store.ListResources("troubleshoot-los")
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Equal(t, "scripts/diag.rsc", resources[0].Path)
	assert.Equal(t, "script", resources[0].Type)

	// 5. Read Resource
	resContent, resInfo, err := store.ReadResource("troubleshoot-los", "scripts/diag.rsc")
	require.NoError(t, err)
	assert.Equal(t, string(rscScript), resContent.Content)
	assert.Equal(t, "raw", resContent.Encoding)
	assert.True(t, resInfo.Readable)

	// 6. Export to ZIP
	zipBytes, err := store.ExportSkillZip("troubleshoot-los")
	require.NoError(t, err)
	assert.NotEmpty(t, zipBytes)

	// 7. Delete skill
	err = store.DeleteSkillFromDisk("troubleshoot-los")
	require.NoError(t, err)
	_, err = store.ReadSkillFromDisk("troubleshoot-los")
	assert.ErrorIs(t, err, skill.ErrSkillNotFound)

	// 8. Import from ZIP
	imported, err := store.ImportSkillZip(zipBytes)
	require.NoError(t, err)
	assert.Equal(t, "troubleshoot-los", imported.Name)

	importedResources, err := store.ListResources("troubleshoot-los")
	require.NoError(t, err)
	assert.Len(t, importedResources, 1)
}

func TestFSSkillStore_GitRepoConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs_skill_git_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	store, err := storage.NewFSSkillStore(tempDir)
	require.NoError(t, err)

	repo := skill.GitRepoInfo{
		ID:      "repo-1",
		URL:     "https://github.com/example/netops-skills.git",
		Name:    "netops-skills",
		Enabled: true,
	}

	err = store.SaveGitRepo(repo)
	require.NoError(t, err)

	repos, err := store.ListGitRepos()
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "repo-1", repos[0].ID)

	err = store.DeleteGitRepo("repo-1")
	require.NoError(t, err)

	reposAfter, err := store.ListGitRepos()
	require.NoError(t, err)
	assert.Empty(t, reposAfter)
}
