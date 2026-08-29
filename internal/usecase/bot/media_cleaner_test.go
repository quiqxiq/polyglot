package bot_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quixiq/polyglot/internal/usecase/bot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaCleanerWorker_CleanOnce(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "media_cleaner_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	oldFile := filepath.Join(tempDir, "old_voucher.pdf")
	newFile := filepath.Join(tempDir, "new_voucher.png")
	nonMediaFile := filepath.Join(tempDir, "old_data.json")

	require.NoError(t, os.WriteFile(oldFile, []byte("old pdf content"), 0644))
	require.NoError(t, os.WriteFile(newFile, []byte("new png content"), 0644))
	require.NoError(t, os.WriteFile(nonMediaFile, []byte("{}"), 0644))

	// Backdate old files by 1 hour
	pastTime := time.Now().Add(-1 * time.Hour)
	require.NoError(t, os.Chtimes(oldFile, pastTime, pastTime))
	require.NoError(t, os.Chtimes(nonMediaFile, pastTime, pastTime))

	worker := bot.NewMediaCleanerWorker(bot.MediaCleanerConfig{
		TargetDir: tempDir,
		MaxAge:    30 * time.Minute,
		Interval:  1 * time.Hour,
	})

	ctx := context.Background()
	n, err := worker.CleanOnce(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "Should clean exactly 1 expired media file (.pdf)")

	_, oldErr := os.Stat(oldFile)
	assert.True(t, os.IsNotExist(oldErr), "Old PDF should be deleted")

	_, newErr := os.Stat(newFile)
	assert.NoError(t, newErr, "New PNG should not be deleted")

	_, jsonErr := os.Stat(nonMediaFile)
	assert.NoError(t, jsonErr, "Non-media file should not be deleted")
}

func TestMediaCleanerWorker_NonExistentDir(t *testing.T) {
	worker := bot.NewMediaCleanerWorker(bot.MediaCleanerConfig{
		TargetDir: filepath.Join(os.TempDir(), "non_existent_dir_12345"),
		MaxAge:    30 * time.Minute,
	})

	ctx := context.Background()
	n, err := worker.CleanOnce(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}
