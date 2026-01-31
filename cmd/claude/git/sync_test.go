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

func TestCopyAndLocalizeIndex(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src-index.json")
	dstPath := filepath.Join(tmpDir, "dst-index.json")

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Create source index with canonical paths
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/home/gigur/git/tofu",
				FullPath:    "/home/gigur/.claude/projects/-home-gigur-git-tofu/session-1.jsonl",
			},
		},
	}
	srcData, _ := json.MarshalIndent(srcIndex, "", "  ")
	os.WriteFile(srcPath, srcData, 0644)

	// Copy with localization (simulating Mac)
	localHome := "/Users/johkjo"
	err := copyAndLocalizeIndex(srcPath, dstPath, config, localHome)
	if err != nil {
		t.Fatalf("copyAndLocalizeIndex failed: %v", err)
	}

	// Read result
	resultData, _ := os.ReadFile(dstPath)
	var result conv.SessionsIndex
	json.Unmarshal(resultData, &result)

	if len(result.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(result.Entries))
	}

	entry := result.Entries[0]
	expectedProjectPath := "/Users/johkjo/git/personal/tofu"
	if entry.ProjectPath != expectedProjectPath {
		t.Errorf("ProjectPath not localized: got %q, want %q", entry.ProjectPath, expectedProjectPath)
	}

	expectedFullPath := "/Users/johkjo/.claude/projects/-home-gigur-git-tofu/session-1.jsonl"
	if entry.FullPath != expectedFullPath {
		t.Errorf("FullPath not localized: got %q, want %q", entry.FullPath, expectedFullPath)
	}
}

func TestCopyAndLocalizeIndex_AlreadyLocal(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src-index.json")
	dstPath := filepath.Join(tmpDir, "dst-index.json")

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
	}

	// Source already has canonical (Linux) paths
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/home/gigur/git/tofu",
			},
		},
	}
	srcData, _ := json.MarshalIndent(srcIndex, "", "  ")
	os.WriteFile(srcPath, srcData, 0644)

	// Copy with localization on Linux (canonical = local)
	localHome := "/home/gigur"
	err := copyAndLocalizeIndex(srcPath, dstPath, config, localHome)
	if err != nil {
		t.Fatalf("copyAndLocalizeIndex failed: %v", err)
	}

	// Read result
	resultData, _ := os.ReadFile(dstPath)
	var result conv.SessionsIndex
	json.Unmarshal(resultData, &result)

	entry := result.Entries[0]
	// Should stay as canonical since we're on the canonical machine
	if entry.ProjectPath != "/home/gigur/git/tofu" {
		t.Errorf("Path should stay canonical on Linux: got %q", entry.ProjectPath)
	}
}

func TestRoundTrip_LocalToSyncToLocal(t *testing.T) {
	// This tests the full round trip:
	// 1. Mac local paths → sync (canonicalized)
	// 2. sync → Mac local (localized back)

	tmpDir := t.TempDir()

	// Setup config
	configDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(configDir, 0755)
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	config := SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(configDir, "sync_config.json"), configData, 0644)

	// Create source (Mac local) index
	srcDir := filepath.Join(tmpDir, "local")
	os.MkdirAll(srcDir, 0755)
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

	// Create sync dir (starts empty)
	syncDir := filepath.Join(tmpDir, "sync")
	os.MkdirAll(syncDir, 0755)
	syncIndex := conv.SessionsIndex{Version: 1, Entries: []conv.SessionEntry{}}
	syncData, _ := json.MarshalIndent(syncIndex, "", "  ")
	os.WriteFile(filepath.Join(syncDir, "sessions-index.json"), syncData, 0644)

	// Step 1: Merge local → sync (should canonicalize)
	err := mergeSessionsIndex(srcDir, syncDir)
	if err != nil {
		t.Fatalf("mergeSessionsIndex failed: %v", err)
	}

	// Verify sync has canonical paths
	syncResultData, _ := os.ReadFile(filepath.Join(syncDir, "sessions-index.json"))
	var syncResult conv.SessionsIndex
	json.Unmarshal(syncResultData, &syncResult)

	if syncResult.Entries[0].ProjectPath != "/home/gigur/git/tofu" {
		t.Errorf("Sync should have canonical path: got %q", syncResult.Entries[0].ProjectPath)
	}

	// Step 2: Copy sync → local (should localize back to Mac paths)
	dstDir := filepath.Join(tmpDir, "local-copy")
	os.MkdirAll(dstDir, 0755)

	localHome := "/Users/johkjo"
	err = copyAndLocalizeIndex(
		filepath.Join(syncDir, "sessions-index.json"),
		filepath.Join(dstDir, "sessions-index.json"),
		&config,
		localHome,
	)
	if err != nil {
		t.Fatalf("copyAndLocalizeIndex failed: %v", err)
	}

	// Verify local copy has Mac paths
	localResultData, _ := os.ReadFile(filepath.Join(dstDir, "sessions-index.json"))
	var localResult conv.SessionsIndex
	json.Unmarshal(localResultData, &localResult)

	expectedPath := "/Users/johkjo/git/personal/tofu"
	if localResult.Entries[0].ProjectPath != expectedPath {
		t.Errorf("Local should have Mac path: got %q, want %q", localResult.Entries[0].ProjectPath, expectedPath)
	}
}
