package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gigurra/tofu/cmd/claude/conv"
)

func TestUpdateIndexFile_Canonicalize(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "-Users-johkjo-git-personal-tofu")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a sessions-index.json with Mac paths
	index := conv.SessionsIndex{
		Entries: []conv.SessionEntry{
			{
				SessionID:   "abc123",
				ProjectPath: "/Users/johkjo/git/personal/tofu",
				FullPath:    "/Users/johkjo/.claude/projects/-Users-johkjo-git-personal-tofu/abc123.jsonl",
			},
		},
	}
	indexData, _ := json.MarshalIndent(index, "", "  ")
	indexPath := filepath.Join(projectDir, "sessions-index.json")
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a dummy session file so LoadSessionsIndex works
	sessionFile := filepath.Join(projectDir, "abc123.jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"type":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Canonicalize (empty localHome)
	wasFixed, err := updateIndexFile(indexPath, config, "")
	if err != nil {
		t.Fatalf("updateIndexFile failed: %v", err)
	}
	if !wasFixed {
		t.Error("expected file to be modified")
	}

	// Read and verify
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}

	var result conv.SessionsIndex
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}

	// Should be canonicalized to Linux paths (including embedded project dir in fullPath)
	if result.Entries[0].ProjectPath != "/home/gigur/git/tofu" {
		t.Errorf("ProjectPath not canonicalized: got %q, want %q",
			result.Entries[0].ProjectPath, "/home/gigur/git/tofu")
	}

	// FullPath should have both the home prefix AND embedded project dir canonicalized
	expectedFullPath := "/home/gigur/.claude/projects/-home-gigur-git-tofu/abc123.jsonl"
	if result.Entries[0].FullPath != expectedFullPath {
		t.Errorf("FullPath not canonicalized: got %q, want %q",
			result.Entries[0].FullPath, expectedFullPath)
	}
}

func TestUpdateIndexFile_Localize(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "-home-gigur-git-tofu")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a sessions-index.json with canonical (Linux) paths
	index := conv.SessionsIndex{
		Entries: []conv.SessionEntry{
			{
				SessionID:   "abc123",
				ProjectPath: "/home/gigur/git/tofu",
				FullPath:    "/home/gigur/.claude/projects/-home-gigur-git-tofu/abc123.jsonl",
			},
		},
	}
	indexData, _ := json.MarshalIndent(index, "", "  ")
	indexPath := filepath.Join(projectDir, "sessions-index.json")
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a dummy session file so LoadSessionsIndex works
	sessionFile := filepath.Join(projectDir, "abc123.jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"type":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Localize to Mac
	wasFixed, err := updateIndexFile(indexPath, config, "/Users/johkjo")
	if err != nil {
		t.Fatalf("updateIndexFile failed: %v", err)
	}
	if !wasFixed {
		t.Error("expected file to be modified")
	}

	// Read and verify
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}

	var result conv.SessionsIndex
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}

	// Should be localized to Mac paths
	if result.Entries[0].ProjectPath != "/Users/johkjo/git/personal/tofu" {
		t.Errorf("ProjectPath not localized: got %q, want %q",
			result.Entries[0].ProjectPath, "/Users/johkjo/git/personal/tofu")
	}

	if result.Entries[0].FullPath != "/Users/johkjo/.claude/projects/-home-gigur-git-tofu/abc123.jsonl" {
		t.Errorf("FullPath not localized: got %q, want %q",
			result.Entries[0].FullPath, "/Users/johkjo/.claude/projects/-home-gigur-git-tofu/abc123.jsonl")
	}
}

func TestUpdateIndexFile_NoChange(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "-home-gigur-git-tofu")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a sessions-index.json that's already canonical
	index := conv.SessionsIndex{
		Entries: []conv.SessionEntry{
			{
				SessionID:   "abc123",
				ProjectPath: "/home/gigur/git/tofu",
				FullPath:    "/home/gigur/.claude/projects/-home-gigur-git-tofu/abc123.jsonl",
			},
		},
	}
	indexData, _ := json.MarshalIndent(index, "", "  ")
	indexPath := filepath.Join(projectDir, "sessions-index.json")
	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a dummy session file
	sessionFile := filepath.Join(projectDir, "abc123.jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"type":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Canonicalize - should report no change since it's already canonical
	wasFixed, err := updateIndexFile(indexPath, config, "")
	if err != nil {
		t.Fatalf("updateIndexFile failed: %v", err)
	}
	if wasFixed {
		t.Error("expected no changes to already canonical file")
	}
}

func TestUpdateSessionPaths_MultipleProjects(t *testing.T) {
	tempDir := t.TempDir()

	// Create two project directories with Mac paths
	projects := []string{"-Users-johkjo-git-personal-tofu", "-Users-johkjo-git-personal-other"}
	for _, proj := range projects {
		projectDir := filepath.Join(tempDir, proj)
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			t.Fatal(err)
		}

		index := conv.SessionsIndex{
			Entries: []conv.SessionEntry{
				{
					SessionID:   "session1",
					ProjectPath: "/Users/johkjo/git/personal/something",
				},
			},
		}
		indexData, _ := json.MarshalIndent(index, "", "  ")
		if err := os.WriteFile(filepath.Join(projectDir, "sessions-index.json"), indexData, 0644); err != nil {
			t.Fatal(err)
		}

		// Create dummy session file
		if err := os.WriteFile(filepath.Join(projectDir, "session1.jsonl"), []byte(`{}`), 0644); err != nil {
			t.Fatal(err)
		}
	}

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Canonicalize both
	fixed, err := updateSessionPaths(tempDir, config, "")
	if err != nil {
		t.Fatalf("updateSessionPaths failed: %v", err)
	}
	if fixed != 2 {
		t.Errorf("expected 2 files fixed, got %d", fixed)
	}

	// Verify both are canonicalized
	for _, proj := range projects {
		indexPath := filepath.Join(tempDir, proj, "sessions-index.json")
		data, _ := os.ReadFile(indexPath)
		var result conv.SessionsIndex
		json.Unmarshal(data, &result)

		if result.Entries[0].ProjectPath != "/home/gigur/git/something" {
			t.Errorf("project %s not canonicalized: %q", proj, result.Entries[0].ProjectPath)
		}
	}
}

func TestRepairDirectory_CanonicalizesSyncDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create a project dir with Mac path naming
	macProjectDir := filepath.Join(tempDir, "-Users-johkjo-git-personal-tofu")
	if err := os.MkdirAll(macProjectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create index with Mac paths
	index := conv.SessionsIndex{
		Entries: []conv.SessionEntry{
			{SessionID: "test1", ProjectPath: "/Users/johkjo/git/personal/tofu"},
		},
	}
	indexData, _ := json.MarshalIndent(index, "", "  ")
	if err := os.WriteFile(filepath.Join(macProjectDir, "sessions-index.json"), indexData, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(macProjectDir, "test1.jsonl"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Repair as sync dir (canonicalize - empty localHome)
	count, err := repairDirectory(tempDir, config, false, "")
	if err != nil {
		t.Fatalf("repairDirectory failed: %v", err)
	}

	// Should have merged and created canonical directory
	if count != 1 {
		t.Errorf("expected 1 merge, got %d", count)
	}

	// Canonical dir should exist
	canonicalDir := filepath.Join(tempDir, "-home-gigur-git-tofu")
	if _, err := os.Stat(canonicalDir); os.IsNotExist(err) {
		t.Error("canonical directory was not created")
	}

	// Mac dir should be removed
	if _, err := os.Stat(macProjectDir); !os.IsNotExist(err) {
		t.Error("Mac directory was not removed after merge")
	}

	// Check paths in index are canonicalized
	indexPath := filepath.Join(canonicalDir, "sessions-index.json")
	data, _ := os.ReadFile(indexPath)
	var result conv.SessionsIndex
	json.Unmarshal(data, &result)

	if result.Entries[0].ProjectPath != "/home/gigur/git/tofu" {
		t.Errorf("path not canonicalized: %q", result.Entries[0].ProjectPath)
	}
}

func TestRepairDirectory_LocalizesLocalDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create a project dir with canonical path naming
	canonicalProjectDir := filepath.Join(tempDir, "-home-gigur-git-tofu")
	if err := os.MkdirAll(canonicalProjectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create index with canonical (Linux) paths
	index := conv.SessionsIndex{
		Entries: []conv.SessionEntry{
			{SessionID: "test1", ProjectPath: "/home/gigur/git/tofu"},
		},
	}
	indexData, _ := json.MarshalIndent(index, "", "  ")
	if err := os.WriteFile(filepath.Join(canonicalProjectDir, "sessions-index.json"), indexData, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canonicalProjectDir, "test1.jsonl"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Repair as local dir (localize - with localHome set to Mac)
	_, err := repairDirectory(tempDir, config, false, "/Users/johkjo")
	if err != nil {
		t.Fatalf("repairDirectory failed: %v", err)
	}

	// Check paths in index are localized to Mac
	indexPath := filepath.Join(canonicalProjectDir, "sessions-index.json")
	data, _ := os.ReadFile(indexPath)
	var result conv.SessionsIndex
	json.Unmarshal(data, &result)

	if result.Entries[0].ProjectPath != "/Users/johkjo/git/personal/tofu" {
		t.Errorf("path not localized: got %q, want %q",
			result.Entries[0].ProjectPath, "/Users/johkjo/git/personal/tofu")
	}
}

func TestRepairDirectory_DryRun(t *testing.T) {
	tempDir := t.TempDir()

	// Create a project dir with Mac path naming
	macProjectDir := filepath.Join(tempDir, "-Users-johkjo-git-personal-tofu")
	if err := os.MkdirAll(macProjectDir, 0755); err != nil {
		t.Fatal(err)
	}

	index := conv.SessionsIndex{
		Entries: []conv.SessionEntry{
			{SessionID: "test1", ProjectPath: "/Users/johkjo/git/personal/tofu"},
		},
	}
	indexData, _ := json.MarshalIndent(index, "", "  ")
	if err := os.WriteFile(filepath.Join(macProjectDir, "sessions-index.json"), indexData, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(macProjectDir, "test1.jsonl"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Dry run
	count, err := repairDirectory(tempDir, config, true, "")
	if err != nil {
		t.Fatalf("repairDirectory dry-run failed: %v", err)
	}

	// Should report the merge would happen
	if count != 1 {
		t.Errorf("expected 1 merge to be reported, got %d", count)
	}

	// But Mac dir should still exist (dry run doesn't modify)
	if _, err := os.Stat(macProjectDir); os.IsNotExist(err) {
		t.Error("Mac directory should still exist in dry-run mode")
	}

	// And canonical dir should NOT exist
	canonicalDir := filepath.Join(tempDir, "-home-gigur-git-tofu")
	if _, err := os.Stat(canonicalDir); !os.IsNotExist(err) {
		t.Error("canonical directory should not be created in dry-run mode")
	}
}

func TestCanonicalizePath(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "mac git personal to canonical",
			path:     "/Users/johkjo/git/personal/tofu",
			expected: "/home/gigur/git/tofu",
		},
		{
			name:     "mac home other to canonical",
			path:     "/Users/johkjo/Documents/notes",
			expected: "/home/gigur/Documents/notes",
		},
		{
			name:     "already canonical",
			path:     "/home/gigur/git/tofu",
			expected: "/home/gigur/git/tofu",
		},
		{
			name:     "unrelated path unchanged",
			path:     "/var/log/something",
			expected: "/var/log/something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canonicalizePath(tt.path, config)
			if result != tt.expected {
				t.Errorf("canonicalizePath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCanonicalizeEmbeddedProjectDir(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "fullPath with embedded mac project dir",
			path:     "/home/gigur/.claude/projects/-Users-johkjo-git-personal-tofu/session.jsonl",
			expected: "/home/gigur/.claude/projects/-home-gigur-git-tofu/session.jsonl",
		},
		{
			name:     "fullPath with already canonical project dir",
			path:     "/home/gigur/.claude/projects/-home-gigur-git-tofu/session.jsonl",
			expected: "/home/gigur/.claude/projects/-home-gigur-git-tofu/session.jsonl",
		},
		{
			name:     "fullPath without session file (project dir only)",
			path:     "/home/gigur/.claude/projects/-Users-johkjo-git-personal-tofu",
			expected: "/home/gigur/.claude/projects/-home-gigur-git-tofu",
		},
		{
			name:     "path without .claude/projects marker unchanged",
			path:     "/home/gigur/some/other/path/session.jsonl",
			expected: "/home/gigur/some/other/path/session.jsonl",
		},
		{
			name:     "complete mac fullPath gets fully canonicalized",
			path:     "/Users/johkjo/.claude/projects/-Users-johkjo-git-personal-tofu/abc.jsonl",
			expected: "/home/gigur/.claude/projects/-home-gigur-git-tofu/abc.jsonl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canonicalizePath(tt.path, config)
			if result != tt.expected {
				t.Errorf("canonicalizePath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestRepairRoundTrip_SyncThenLocal(t *testing.T) {
	// Simulate the full workflow: Mac paths -> sync (canonicalize) -> local on Mac (localize)
	syncDir := t.TempDir()
	localDir := t.TempDir()

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Step 1: Create project with Mac paths (as if on Mac)
	macProjectDir := filepath.Join(syncDir, "-Users-johkjo-git-personal-tofu")
	if err := os.MkdirAll(macProjectDir, 0755); err != nil {
		t.Fatal(err)
	}

	index := conv.SessionsIndex{
		Entries: []conv.SessionEntry{
			{
				SessionID:   "test1",
				ProjectPath: "/Users/johkjo/git/personal/tofu",
				FullPath:    "/Users/johkjo/.claude/projects/-Users-johkjo-git-personal-tofu/test1.jsonl",
			},
		},
	}
	indexData, _ := json.MarshalIndent(index, "", "  ")
	if err := os.WriteFile(filepath.Join(macProjectDir, "sessions-index.json"), indexData, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(macProjectDir, "test1.jsonl"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Step 2: Repair sync dir (canonicalize)
	_, err := repairDirectory(syncDir, config, false, "")
	if err != nil {
		t.Fatalf("sync repair failed: %v", err)
	}

	// Verify canonical dir was created in sync
	canonicalDir := filepath.Join(syncDir, "-home-gigur-git-tofu")
	if _, err := os.Stat(canonicalDir); os.IsNotExist(err) {
		t.Fatal("canonical directory not created")
	}

	// Check sync paths are canonical
	syncIndexPath := filepath.Join(canonicalDir, "sessions-index.json")
	data, _ := os.ReadFile(syncIndexPath)
	var syncResult conv.SessionsIndex
	json.Unmarshal(data, &syncResult)

	if syncResult.Entries[0].ProjectPath != "/home/gigur/git/tofu" {
		t.Errorf("sync not canonicalized: %q", syncResult.Entries[0].ProjectPath)
	}

	// Step 3: Copy to local dir (simulating sync -> local copy)
	localProjectDir := filepath.Join(localDir, "-home-gigur-git-tofu")
	if err := os.MkdirAll(localProjectDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Copy files
	if err := os.WriteFile(filepath.Join(localProjectDir, "sessions-index.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localProjectDir, "test1.jsonl"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Step 4: Repair local dir with Mac localHome (localize)
	_, err = repairDirectory(localDir, config, false, "/Users/johkjo")
	if err != nil {
		t.Fatalf("local repair failed: %v", err)
	}

	// Check local paths are localized to Mac
	localIndexPath := filepath.Join(localProjectDir, "sessions-index.json")
	localData, _ := os.ReadFile(localIndexPath)
	var localResult conv.SessionsIndex
	json.Unmarshal(localData, &localResult)

	if localResult.Entries[0].ProjectPath != "/Users/johkjo/git/personal/tofu" {
		t.Errorf("local not localized: got %q, want %q",
			localResult.Entries[0].ProjectPath, "/Users/johkjo/git/personal/tofu")
	}

	expectedFullPath := "/Users/johkjo/.claude/projects/-home-gigur-git-tofu/test1.jsonl"
	if localResult.Entries[0].FullPath != expectedFullPath {
		t.Errorf("local fullPath not localized: got %q, want %q",
			localResult.Entries[0].FullPath, expectedFullPath)
	}
}

func TestRepairDirectory_MergeEquivalentDirs(t *testing.T) {
	tempDir := t.TempDir()

	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// Create TWO equivalent directories
	linuxDir := filepath.Join(tempDir, "-home-gigur-git-tofu")
	macDir := filepath.Join(tempDir, "-Users-johkjo-git-personal-tofu")

	for _, dir := range []string{linuxDir, macDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Linux dir has session1
	linuxIndex := conv.SessionsIndex{
		Entries: []conv.SessionEntry{{SessionID: "session1", ProjectPath: "/home/gigur/git/tofu"}},
	}
	linuxData, _ := json.MarshalIndent(linuxIndex, "", "  ")
	os.WriteFile(filepath.Join(linuxDir, "sessions-index.json"), linuxData, 0644)
	os.WriteFile(filepath.Join(linuxDir, "session1.jsonl"), []byte(`{"linux":true}`), 0644)

	// Mac dir has session2
	macIndex := conv.SessionsIndex{
		Entries: []conv.SessionEntry{{SessionID: "session2", ProjectPath: "/Users/johkjo/git/personal/tofu"}},
	}
	macData, _ := json.MarshalIndent(macIndex, "", "  ")
	os.WriteFile(filepath.Join(macDir, "sessions-index.json"), macData, 0644)
	os.WriteFile(filepath.Join(macDir, "session2.jsonl"), []byte(`{"mac":true}`), 0644)

	// Repair (canonicalize)
	count, err := repairDirectory(tempDir, config, false, "")
	if err != nil {
		t.Fatalf("repairDirectory failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 merge group, got %d", count)
	}

	// Only canonical dir should exist
	if _, err := os.Stat(linuxDir); os.IsNotExist(err) {
		t.Error("canonical directory should still exist")
	}
	if _, err := os.Stat(macDir); !os.IsNotExist(err) {
		t.Error("mac directory should be removed after merge")
	}

	// Canonical dir should have both session files
	if _, err := os.Stat(filepath.Join(linuxDir, "session1.jsonl")); os.IsNotExist(err) {
		t.Error("session1.jsonl should exist in merged dir")
	}
	if _, err := os.Stat(filepath.Join(linuxDir, "session2.jsonl")); os.IsNotExist(err) {
		t.Error("session2.jsonl should exist in merged dir")
	}
}
