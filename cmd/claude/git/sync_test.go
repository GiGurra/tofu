package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gigurra/tofu/cmd/claude/conv"
)

// Test path constants - same as in repair_test.go
// These are just string values for testing path transformations, not actual filesystem paths.
const (
	syncTestCanonicalHome = "/home/canonical"
	syncTestLocalHome     = "/home/local"
	syncTestCanonicalGit  = "/home/canonical/git"
	syncTestLocalGit      = "/home/local/projects"
)

func syncTestConfig() *SyncConfig {
	return &SyncConfig{
		Homes: []string{syncTestCanonicalHome, syncTestLocalHome},
		Dirs:  [][]string{{syncTestCanonicalGit, syncTestLocalGit}},
	}
}

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
		Homes: []string{syncTestCanonicalHome, syncTestLocalHome},
		Dirs:  [][]string{{syncTestCanonicalGit, syncTestLocalGit}},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(configDir, "sync_config.json"), configData, 0644)

	// Create source index with non-canonical (local) paths
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/home/local/projects/myproject",
				FullPath:    "/home/local/.claude/projects/-home-local-projects-myproject/session-1.jsonl",
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
	if entry.ProjectPath != "/home/canonical/git/myproject" {
		t.Errorf("ProjectPath not canonicalized: got %q, want %q", entry.ProjectPath, "/home/canonical/git/myproject")
	}
	// FullPath home prefix should be canonicalized
	expectedPrefix := "/home/canonical/"
	if len(entry.FullPath) < len(expectedPrefix) || entry.FullPath[:len(expectedPrefix)] != expectedPrefix {
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
		Homes: []string{syncTestCanonicalHome, syncTestLocalHome},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(configDir, "sync_config.json"), configData, 0644)

	// Create source index with old non-canonical path and older timestamp
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/home/local/git/myproject",
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
				ProjectPath: "/home/canonical/git/myproject",
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
	if entry.ProjectPath != "/home/canonical/git/myproject" {
		t.Errorf("Should keep newer canonical path: got %q, want %q", entry.ProjectPath, "/home/canonical/git/myproject")
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
		Homes: []string{syncTestCanonicalHome, syncTestLocalHome},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(configDir, "sync_config.json"), configData, 0644)

	// Create source index with newer timestamp
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:    "session-1",
				ProjectPath:  "/home/local/git/myproject",
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
				ProjectPath:  "/home/canonical/git/myproject",
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
	if entry.ProjectPath != "/home/canonical/git/myproject" {
		t.Errorf("Path should be canonicalized: got %q, want %q", entry.ProjectPath, "/home/canonical/git/myproject")
	}
}

func TestFindLocalEquivalent_FindsExistingDir(t *testing.T) {
	localDirs := map[string]bool{
		"-home-local-projects-myproject": true,
	}

	// Looking for canonical name, should find the local equivalent
	result := findLocalEquivalent("-home-canonical-git-myproject", localDirs, syncTestConfig())
	if result != "-home-local-projects-myproject" {
		t.Errorf("Should find local equivalent: got %q, want %q", result, "-home-local-projects-myproject")
	}
}

func TestFindLocalEquivalent_ReturnsCanonicalIfNoLocal(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{syncTestCanonicalHome, syncTestLocalHome},
	}

	localDirs := map[string]bool{} // No local dirs

	result := findLocalEquivalent("-home-canonical-git-myproject", localDirs, config)
	if result != "-home-canonical-git-myproject" {
		t.Errorf("Should return canonical when no local: got %q, want %q", result, "-home-canonical-git-myproject")
	}
}

func TestFindLocalEquivalent_PrefersCanonicalIfExists(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{syncTestCanonicalHome, syncTestLocalHome},
	}

	localDirs := map[string]bool{
		"-home-canonical-git-myproject": true,
		"-home-local-git-myproject":     true,
	}

	// Both exist, should prefer canonical
	result := findLocalEquivalent("-home-canonical-git-myproject", localDirs, config)
	if result != "-home-canonical-git-myproject" {
		t.Errorf("Should prefer canonical: got %q, want %q", result, "-home-canonical-git-myproject")
	}
}

func TestCopyAndLocalizeIndex(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src-index.json")
	dstPath := filepath.Join(tmpDir, "dst-index.json")

	// Create source index with canonical paths
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/home/canonical/git/myproject",
				FullPath:    "/home/canonical/.claude/projects/-home-canonical-git-myproject/session-1.jsonl",
			},
		},
	}
	srcData, _ := json.MarshalIndent(srcIndex, "", "  ")
	os.WriteFile(srcPath, srcData, 0644)

	// Copy with localization (simulating local machine)
	err := copyAndLocalizeIndex(srcPath, dstPath, syncTestConfig(), syncTestLocalHome)
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
	expectedProjectPath := "/home/local/projects/myproject"
	if entry.ProjectPath != expectedProjectPath {
		t.Errorf("ProjectPath not localized: got %q, want %q", entry.ProjectPath, expectedProjectPath)
	}

	// FullPath should be fully localized (both home prefix AND embedded project dir)
	expectedFullPath := "/home/local/.claude/projects/-home-local-projects-myproject/session-1.jsonl"
	if entry.FullPath != expectedFullPath {
		t.Errorf("FullPath not localized: got %q, want %q", entry.FullPath, expectedFullPath)
	}
}

func TestCopyAndLocalizeIndex_AlreadyLocal(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src-index.json")
	dstPath := filepath.Join(tmpDir, "dst-index.json")

	config := &SyncConfig{
		Homes: []string{syncTestCanonicalHome, syncTestLocalHome},
	}

	// Source already has canonical paths
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/home/canonical/git/myproject",
			},
		},
	}
	srcData, _ := json.MarshalIndent(srcIndex, "", "  ")
	os.WriteFile(srcPath, srcData, 0644)

	// Copy with localization on canonical machine (canonical = local)
	err := copyAndLocalizeIndex(srcPath, dstPath, config, syncTestCanonicalHome)
	if err != nil {
		t.Fatalf("copyAndLocalizeIndex failed: %v", err)
	}

	// Read result
	resultData, _ := os.ReadFile(dstPath)
	var result conv.SessionsIndex
	json.Unmarshal(resultData, &result)

	entry := result.Entries[0]
	// Should stay as canonical since we're on the canonical machine
	if entry.ProjectPath != "/home/canonical/git/myproject" {
		t.Errorf("Path should stay canonical: got %q", entry.ProjectPath)
	}
}

func TestRoundTrip_LocalToSyncToLocal(t *testing.T) {
	// This tests the full round trip:
	// 1. Local paths → sync (canonicalized)
	// 2. sync → Local (localized back)

	tmpDir := t.TempDir()

	// Setup config
	configDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(configDir, 0755)
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	config := SyncConfig{
		Homes: []string{syncTestCanonicalHome, syncTestLocalHome},
		Dirs:  [][]string{{syncTestCanonicalGit, syncTestLocalGit}},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(configDir, "sync_config.json"), configData, 0644)

	// Create source (local machine) index
	srcDir := filepath.Join(tmpDir, "local")
	os.MkdirAll(srcDir, 0755)
	srcIndex := conv.SessionsIndex{
		Version: 1,
		Entries: []conv.SessionEntry{
			{
				SessionID:   "session-1",
				ProjectPath: "/home/local/projects/myproject",
				FullPath:    "/home/local/.claude/projects/-home-local-projects-myproject/session-1.jsonl",
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

	if syncResult.Entries[0].ProjectPath != "/home/canonical/git/myproject" {
		t.Errorf("Sync should have canonical path: got %q", syncResult.Entries[0].ProjectPath)
	}

	// Step 2: Copy sync → local (should localize back to local paths)
	dstDir := filepath.Join(tmpDir, "local-copy")
	os.MkdirAll(dstDir, 0755)

	err = copyAndLocalizeIndex(
		filepath.Join(syncDir, "sessions-index.json"),
		filepath.Join(dstDir, "sessions-index.json"),
		&config,
		syncTestLocalHome,
	)
	if err != nil {
		t.Fatalf("copyAndLocalizeIndex failed: %v", err)
	}

	// Verify local copy has local paths
	localResultData, _ := os.ReadFile(filepath.Join(dstDir, "sessions-index.json"))
	var localResult conv.SessionsIndex
	json.Unmarshal(localResultData, &localResult)

	expectedPath := "/home/local/projects/myproject"
	if localResult.Entries[0].ProjectPath != expectedPath {
		t.Errorf("Local should have local path: got %q, want %q", localResult.Entries[0].ProjectPath, expectedPath)
	}
}
