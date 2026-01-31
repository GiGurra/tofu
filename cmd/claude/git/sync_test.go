package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gigurra/tofu/cmd/claude/conv"
)

func TestMergeSessionsIndex_CanonicalizesSourcePaths(t *testing.T) {
	// Setup temp directories
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(dstDir, 0755)

	// Create a config file with path mappings
	configDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(configDir, 0755)

	// Temporarily override the config path for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	config := SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(configDir, "sync_config.json"), configData, 0644)

	// Create source index with non-canonical paths
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/Users/johkjo/git/personal/tofu",
				FullPath:    "/Users/johkjo/.claude/projects/-Users-johkjo-git-personal-tofu/session-1.jsonl",
				Modified:    "2024-01-15T12:00:00Z",
			},
		},
	}
	srcData, _ := json.MarshalIndent(srcIndex, "", "  ")
	os.WriteFile(filepath.Join(srcDir, "sessions-index.json"), srcData, 0644)

	// Create empty dst index
	dstIndex := conv.SessionsIndex{Version: 1, Entries: []conv.SessionEntry{}}
	dstData, _ := json.MarshalIndent(dstIndex, "", "  ")
	os.WriteFile(filepath.Join(dstDir, "sessions-index.json"), dstData, 0644)

	// Run merge
	err := mergeSessionsIndex(srcDir, dstDir)
	if err != nil {
		t.Fatalf("mergeSessionsIndex failed: %v", err)
	}

	// Read result
	resultData, _ := os.ReadFile(filepath.Join(dstDir, "sessions-index.json"))
	var result conv.SessionsIndex
	json.Unmarshal(resultData, &result)

	// Check that paths are canonicalized
	if len(result.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(result.Entries))
	}

	entry := result.Entries[0]
	if entry.ProjectPath != "/home/gigur/git/tofu" {
		t.Errorf("ProjectPath not canonicalized: got %q, want %q", entry.ProjectPath, "/home/gigur/git/tofu")
	}
	// FullPath home prefix should be canonicalized (project dir name in path is a separate concern)
	if entry.FullPath[:12] != "/home/gigur/" {
		t.Errorf("FullPath home prefix not canonicalized: got %q", entry.FullPath)
	}
}

func TestMergeSessionsIndex_PreservesExistingCanonicalPaths(t *testing.T) {
	// Setup temp directories
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(dstDir, 0755)

	// Create config
	configDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(configDir, 0755)
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	config := SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(configDir, "sync_config.json"), configData, 0644)

	// Create source index with old non-canonical path and older timestamp
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/Users/johkjo/git/tofu",
				Modified:    "2024-01-10T12:00:00Z", // Older
			},
		},
	}
	srcData, _ := json.MarshalIndent(srcIndex, "", "  ")
	os.WriteFile(filepath.Join(srcDir, "sessions-index.json"), srcData, 0644)

	// Create dst index with canonical path and newer timestamp
	dstIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/home/gigur/git/tofu",
				Modified:    "2024-01-15T12:00:00Z", // Newer
			},
		},
	}
	dstData, _ := json.MarshalIndent(dstIndex, "", "  ")
	os.WriteFile(filepath.Join(dstDir, "sessions-index.json"), dstData, 0644)

	// Run merge
	err := mergeSessionsIndex(srcDir, dstDir)
	if err != nil {
		t.Fatalf("mergeSessionsIndex failed: %v", err)
	}

	// Read result
	resultData, _ := os.ReadFile(filepath.Join(dstDir, "sessions-index.json"))
	var result conv.SessionsIndex
	json.Unmarshal(resultData, &result)

	// Check that the newer (canonical) path is preserved
	if len(result.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(result.Entries))
	}

	entry := result.Entries[0]
	if entry.ProjectPath != "/home/gigur/git/tofu" {
		t.Errorf("Should keep newer canonical path: got %q, want %q", entry.ProjectPath, "/home/gigur/git/tofu")
	}
}

func TestMergeSessionsIndex_NewerLocalWins(t *testing.T) {
	// Setup temp directories
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(dstDir, 0755)

	// Create config
	configDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(configDir, 0755)
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	config := SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(configDir, "sync_config.json"), configData, 0644)

	// Create source index with newer timestamp
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:    "session-1",
				ProjectPath:  "/Users/johkjo/git/tofu",
				MessageCount: 100, // More messages
				Modified:     "2024-01-20T12:00:00Z", // Newer
			},
		},
	}
	srcData, _ := json.MarshalIndent(srcIndex, "", "  ")
	os.WriteFile(filepath.Join(srcDir, "sessions-index.json"), srcData, 0644)

	// Create dst index with older timestamp
	dstIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:    "session-1",
				ProjectPath:  "/home/gigur/git/tofu",
				MessageCount: 50, // Fewer messages
				Modified:     "2024-01-15T12:00:00Z", // Older
			},
		},
	}
	dstData, _ := json.MarshalIndent(dstIndex, "", "  ")
	os.WriteFile(filepath.Join(dstDir, "sessions-index.json"), dstData, 0644)

	// Run merge
	err := mergeSessionsIndex(srcDir, dstDir)
	if err != nil {
		t.Fatalf("mergeSessionsIndex failed: %v", err)
	}

	// Read result
	resultData, _ := os.ReadFile(filepath.Join(dstDir, "sessions-index.json"))
	var result conv.SessionsIndex
	json.Unmarshal(resultData, &result)

	// Check that the newer entry wins (but with canonicalized path)
	if len(result.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(result.Entries))
	}

	entry := result.Entries[0]
	if entry.MessageCount != 100 {
		t.Errorf("Should use newer entry: got MessageCount %d, want 100", entry.MessageCount)
	}
	if entry.ProjectPath != "/home/gigur/git/tofu" {
		t.Errorf("Path should be canonicalized: got %q, want %q", entry.ProjectPath, "/home/gigur/git/tofu")
	}
}

func TestFindLocalEquivalent_FindsExistingDir(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	localDirs := map[string]bool{
		"-Users-johkjo-git-personal-tofu": true,
	}

	// Looking for canonical name, should find the local equivalent
	result := findLocalEquivalent("-home-gigur-git-tofu", localDirs, config)
	if result != "-Users-johkjo-git-personal-tofu" {
		t.Errorf("Should find local equivalent: got %q, want %q", result, "-Users-johkjo-git-personal-tofu")
	}
}

func TestFindLocalEquivalent_ReturnsCanonicalIfNoLocal(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
	}

	localDirs := map[string]bool{} // No local dirs

	result := findLocalEquivalent("-home-gigur-git-tofu", localDirs, config)
	if result != "-home-gigur-git-tofu" {
		t.Errorf("Should return canonical when no local: got %q, want %q", result, "-home-gigur-git-tofu")
	}
}

func TestFindLocalEquivalent_PrefersCanonicalIfExists(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
	}

	localDirs := map[string]bool{
		"-home-gigur-git-tofu":   true,
		"-Users-johkjo-git-tofu": true,
	}

	// Both exist, should prefer canonical
	result := findLocalEquivalent("-home-gigur-git-tofu", localDirs, config)
	if result != "-home-gigur-git-tofu" {
		t.Errorf("Should prefer canonical: got %q, want %q", result, "-home-gigur-git-tofu")
	}
}
