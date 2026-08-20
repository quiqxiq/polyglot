package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

func TestFSSkillStore_ReadWrite(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "polyglot_skills_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewFSSkillStore(tempDir)
	if err != nil {
		t.Fatalf("NewFSSkillStore error: %v", err)
	}

	// 1. Test Global Prompt
	globalContent := "# Global System Prompt\nNia CS Ghaib"
	if err := store.WriteGlobalPromptToDisk(globalContent); err != nil {
		t.Fatalf("WriteGlobalPromptToDisk error: %v", err)
	}
	readGlobal, err := store.ReadGlobalPromptFromDisk()
	if err != nil {
		t.Fatalf("ReadGlobalPromptFromDisk error: %v", err)
	}
	if readGlobal != globalContent {
		t.Fatalf("expected global prompt %s, got %s", globalContent, readGlobal)
	}

	// 2. Test Write and Read Skill
	sk := &skill.Skill{
		Slug:        "test-skill",
		Name:        "Test Skill",
		Description: "Testing skill description",
		IsEnabled:   true,
		Files: []skill.SkillFile{
			{
				Name:     "SKILL.md",
				FilePath: "SKILL.md",
				Content:  "---\nname: Test Skill\ndescription: Test Desc\n---\n# Test",
			},
			{
				Name:        "ref1.md",
				FilePath:    filepath.Join("references", "ref1.md"),
				Content:     "## Reference Document 1",
				IsReference: true,
			},
		},
	}

	if err := store.WriteSkillToDisk(sk); err != nil {
		t.Fatalf("WriteSkillToDisk error: %v", err)
	}

	readSk, err := store.ReadSkillFromDisk("test-skill")
	if err != nil {
		t.Fatalf("ReadSkillFromDisk error: %v", err)
	}
	if readSk.Slug != "test-skill" {
		t.Fatalf("expected slug test-skill, got %s", readSk.Slug)
	}
	if len(readSk.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(readSk.Files))
	}

	// 3. Test Scan All
	allSkills, err := store.ScanAllSkillsFromDisk()
	if err != nil {
		t.Fatalf("ScanAllSkillsFromDisk error: %v", err)
	}
	if len(allSkills) != 1 {
		t.Fatalf("expected 1 scanned skill, got %d", len(allSkills))
	}

	// 4. Test Delete
	if err := store.DeleteSkillFolderFromDisk("test-skill"); err != nil {
		t.Fatalf("DeleteSkillFolderFromDisk error: %v", err)
	}
	_, err = store.ReadSkillFromDisk("test-skill")
	if err == nil {
		t.Fatalf("expected error after deletion, got nil")
	}
}
