package system

import (
	"fmt"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// BackupSaveConfig holds parameters for saving a binary backup.
type BackupSaveConfig struct {
	Name        string
	Password    string
	DontEncrypt bool
	Encryption  string // e.g. "aes-sha256"
}

// ExportConfig holds parameters for exporting configuration text.
type ExportConfig struct {
	File          string
	Terse         bool
	ShowSensitive bool
	Compact       bool
	HideSensitive bool
}

// NewBackupSaveCommand builds a RouterOS command to create a binary backup.
func NewBackupSaveCommand(cfg BackupSaveConfig) command.Command {
	args := make(map[string]string)
	if cfg.Name != "" {
		args["name"] = cfg.Name
	}
	if cfg.Password != "" {
		args["password"] = cfg.Password
	}
	if cfg.DontEncrypt {
		args["dont-encrypt"] = "yes"
	}
	if cfg.Encryption != "" {
		args["encryption"] = cfg.Encryption
	}
	return command.Command{
		Raw:  "/system/backup/save",
		Args: args,
	}
}

// NewBackupLoadCommand builds a RouterOS command to restore a binary backup.
func NewBackupLoadCommand(name, password string) command.Command {
	args := map[string]string{
		"name": name,
	}
	if password != "" {
		args["password"] = password
	}
	return command.Command{
		Raw:  "/system/backup/load",
		Args: args,
	}
}

// NewExportCommand builds a RouterOS command to export router configuration to file.
func NewExportCommand(cfg ExportConfig) command.Command {
	args := make(map[string]string)
	if cfg.File != "" {
		args["file"] = cfg.File
	}
	if cfg.Terse {
		args["terse"] = ""
	}
	if cfg.ShowSensitive {
		args["show-sensitive"] = ""
	}
	if cfg.Compact {
		args["compact"] = ""
	}
	if cfg.HideSensitive {
		args["hide-sensitive"] = ""
	}
	return command.Command{
		Raw:  "/export",
		Args: args,
	}
}

// Settled polls until a target file reaches a stable size across consecutive checks.
func Settled(getFileSize func(name string) (int, error), name string, timeout time.Duration, sleep func(time.Duration)) (int, error) {
	deadline := time.Now().Add(timeout)
	lastSize := -1
	stableCount := 0

	for time.Now().Before(deadline) {
		sz, err := getFileSize(name)
		if err == nil && sz > 0 {
			if sz == lastSize {
				stableCount++
				if stableCount >= 2 {
					return sz, nil
				}
			} else {
				stableCount = 0
			}
			lastSize = sz
		}
		sleep(300 * time.Millisecond)
	}
	return 0, fmt.Errorf("timed out waiting for file %s to settle", name)
}

// Sweep removes leftover temporary backup files using the provided removal function.
func Sweep(listNames func() ([]string, error), removeFile func(name string) error, filterPrefix string) (int, error) {
	names, err := listNames()
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, n := range names {
		if filterPrefix != "" && !strings.HasPrefix(n, filterPrefix) {
			continue
		}
		if err := removeFile(n); err == nil {
			removed++
		}
	}
	return removed, nil
}
