package bot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quixiq/polyglot/pkg/logger"
)

// MediaCleanerConfig configures the temporary file cleanup worker.
type MediaCleanerConfig struct {
	TargetDir string        // Directory to clean up (e.g., "./tmp" or os.TempDir())
	MaxAge    time.Duration // Threshold age after which files are deleted (e.g. 30 * time.Minute)
	Interval  time.Duration // Ticker interval for periodic cleaning (e.g. 1 * time.Hour)
}

// MediaCleanerWorker periodically scans and removes temporary media files.
type MediaCleanerWorker struct {
	cfg MediaCleanerConfig
}

// NewMediaCleanerWorker creates a new worker instance with sensible defaults if unconfigured.
func NewMediaCleanerWorker(cfg MediaCleanerConfig) *MediaCleanerWorker {
	if cfg.TargetDir == "" {
		cfg.TargetDir = filepath.Join(os.TempDir(), "polyglot_media")
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 30 * time.Minute
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 1 * time.Hour
	}
	return &MediaCleanerWorker{cfg: cfg}
}

// CleanOnce performs a single pass over TargetDir, removing files older than MaxAge.
func (w *MediaCleanerWorker) CleanOnce(_ context.Context) (int, error) {
	if _, err := os.Stat(w.cfg.TargetDir); os.IsNotExist(err) {
		return 0, nil
	}

	cutoff := time.Now().Add(-w.cfg.MaxAge)
	cleanedCount := 0

	err := filepath.Walk(w.cfg.TargetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip unreadable paths
		}
		if info.IsDir() {
			return nil
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".pdf", ".png", ".jpg", ".jpeg", ".tmp", ".html":
			if info.ModTime().Before(cutoff) {
				if removeErr := os.Remove(path); removeErr == nil {
					cleanedCount++
				} else {
					logger.WithFields(logger.Fields{
						"path":  path,
						"error": removeErr,
					}).Warn("[MediaCleaner] Failed to remove expired file")
				}
			}
		}

		return nil
	})

	if err != nil {
		return cleanedCount, fmt.Errorf("media cleaner walk failed: %w", err)
	}

	return cleanedCount, nil
}

// Start runs the cleanup loop periodically until ctx is done.
func (w *MediaCleanerWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n, err := w.CleanOnce(ctx); err != nil {
				logger.WithError(err).Error("[MediaCleaner] Error during cleanup pass")
			} else if n > 0 {
				logger.WithField("count", n).Info("[MediaCleaner] Cleaned expired media files")
			}
		}
	}
}
