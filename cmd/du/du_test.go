package du

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// setupTestDir creates a temp directory with the given structure
// structure is a map of relative paths to file contents (empty string = directory)
func setupTestDir(t *testing.T, structure map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	for path, content := range structure {
		fullPath := filepath.Join(dir, path)
		if content == "" {
			// It's a directory
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				t.Fatalf("failed to create dir %s: %v", path, err)
			}
		} else {
			// It's a file
			parentDir := filepath.Dir(fullPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				t.Fatalf("failed to create parent dir for %s: %v", path, err)
			}
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Fatalf("failed to write file %s: %v", path, err)
			}
		}
	}

	return dir
}

func TestWalkDir_TreeStructure(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"root.txt":         "root file",
		"subdir/child.txt": "child file",
	})

	// Call walkDir directly in tree mode
	rootNode, err := walkDir(dir, true, true, nil, nil)
	if err != nil {
		t.Fatalf("walkDir failed: %v", err)
	}

	// Check that subdir is in ChildDirs
	if len(rootNode.ChildDirs) != 1 {
		t.Errorf("expected 1 child dir, got %d", len(rootNode.ChildDirs))
	} else {
		subdir := rootNode.ChildDirs[0]
		if !strings.HasSuffix(subdir.Path, "subdir") {
			t.Errorf("expected subdir path to end with 'subdir', got %s", subdir.Path)
		}
		// Check that child.txt is in subdir's ChildFiles
		if len(subdir.ChildFiles) != 1 {
			t.Errorf("expected 1 child file in subdir, got %d", len(subdir.ChildFiles))
		}
	}

	// Check that root.txt is in ChildFiles
	if len(rootNode.ChildFiles) != 1 {
		t.Errorf("expected 1 child file at root, got %d", len(rootNode.ChildFiles))
	}
}

func TestFlattenAndSort(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"root.txt":         "root file",
		"subdir/child.txt": "child file",
	})

	// Build tree with apparentSize=false to match Run behavior
	rootNode, err := walkDir(dir, false, true, nil, nil)
	if err != nil {
		t.Fatalf("walkDir failed: %v", err)
	}

	// Flatten and sort
	entries := flattenTree(rootNode, true)
	sortEntries(entries, "size", false)

	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}

	// Verify entries are sorted by size (ascending)
	for i := 1; i < len(entries); i++ {
		if entries[i].Size < entries[i-1].Size {
			t.Errorf("entries not sorted: %d (%s) < %d (%s)",
				entries[i].Size, entries[i].Path,
				entries[i-1].Size, entries[i-1].Path)
		}
	}
}

func TestDu_BasicDirectory(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"file1.txt": "hello",
		"file2.txt": "world!",
	})

	output := captureOutput(func() {
		Run(&Params{MaxDepth: -1,
			Paths: []string{dir},
			Human: true,
			Sort:  "none",
		})
	})

	// Should only show the directory, not individual files
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line (directory only), got %d: %v", len(lines), lines)
	}
	if !strings.HasSuffix(lines[0], dir) {
		t.Errorf("expected output to end with dir path, got: %s", lines[0])
	}
}

func TestDu_AllFlag_ShowsFiles(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"file1.txt": "hello",
		"file2.txt": "world!",
	})

	output := captureOutput(func() {
		Run(&Params{MaxDepth: -1,
			Paths: []string{dir},
			Human: true,
			All:   true,
			Sort:  "none",
		})
	})

	// Should show files AND the directory
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (2 files + 1 directory), got %d: %v", len(lines), lines)
	}

	// Check that both files appear
	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "file1.txt") {
		t.Errorf("expected file1.txt in output, got: %s", output)
	}
	if !strings.Contains(combined, "file2.txt") {
		t.Errorf("expected file2.txt in output, got: %s", output)
	}
}

func TestDu_AllFlag_NestedDirectories(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"root.txt":         "root file",
		"subdir/child.txt": "child file",
	})

	// Test tree mode (sorted) via Run
	output := captureOutput(func() {
		Run(&Params{
			Paths:    []string{dir},
			Human:    true,
			All:      true,
			Sort:     "size",
			MaxDepth: -1,
		})
	})

	// Should show: root.txt, child.txt, subdir/, root dir
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines (2 files, 2 dirs), got %d:\n%s", len(lines), output)
	}

	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "root.txt") {
		t.Errorf("expected root.txt in output:\n%s", output)
	}
	if !strings.Contains(combined, "child.txt") {
		t.Errorf("expected child.txt in output:\n%s", output)
	}
	if !strings.Contains(combined, "subdir") {
		t.Errorf("expected subdir in output:\n%s", output)
	}
}

func TestDu_AllFlag_NestedDirectories_Streaming(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"root.txt":         "root file",
		"subdir/child.txt": "child file",
	})

	// Test streaming mode (no sort)
	output := captureOutput(func() {
		Run(&Params{MaxDepth: -1,
			Paths: []string{dir},
			Human: true,
			All:   true,
			Sort:  "none",
		})
	})

	// Should show: root.txt, child.txt, subdir/, root dir
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines (2 files, 2 dirs), got %d:\n%s", len(lines), output)
	}

	combined := strings.Join(lines, "\n")
	if !strings.Contains(combined, "root.txt") {
		t.Errorf("expected root.txt in output:\n%s", output)
	}
	if !strings.Contains(combined, "child.txt") {
		t.Errorf("expected child.txt in output:\n%s", output)
	}
	if !strings.Contains(combined, "subdir") {
		t.Errorf("expected subdir in output:\n%s", output)
	}
}

func TestDu_AllFlag_SortBySize(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"small.txt":  "x",
		"medium.txt": strings.Repeat("x", 100),
		"large.txt":  strings.Repeat("x", 1000),
	})

	output := captureOutput(func() {
		Run(&Params{MaxDepth: -1,
			Paths: []string{dir},
			Bytes: true, // Use bytes for precise size comparison
			All:   true,
			Sort:  "size",
		})
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d: %v", len(lines), lines)
	}

	// With size sort (ascending = smallest first), small should come before large
	// Last line should be the directory
	smallIdx := -1
	largeIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "small.txt") {
			smallIdx = i
		}
		if strings.Contains(line, "large.txt") {
			largeIdx = i
		}
	}

	if smallIdx == -1 || largeIdx == -1 {
		t.Fatalf("could not find small.txt or large.txt in output: %s", output)
	}
	if smallIdx >= largeIdx {
		t.Errorf("expected small.txt before large.txt in size-sorted output, got small at %d, large at %d\nOutput:\n%s", smallIdx, largeIdx, output)
	}
}

func TestDu_AllFlag_SortByName(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"zebra.txt":    "z",
		"apple.txt":    "a",
		"mango.txt":    "m",
	})

	output := captureOutput(func() {
		Run(&Params{MaxDepth: -1,
			Paths: []string{dir},
			Human: true,
			All:   true,
			Sort:  "name",
		})
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")

	appleIdx := -1
	mangoIdx := -1
	zebraIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "apple.txt") {
			appleIdx = i
		}
		if strings.Contains(line, "mango.txt") {
			mangoIdx = i
		}
		if strings.Contains(line, "zebra.txt") {
			zebraIdx = i
		}
	}

	if appleIdx == -1 || mangoIdx == -1 || zebraIdx == -1 {
		t.Fatalf("could not find all files in output: %s", output)
	}
	if !(appleIdx < mangoIdx && mangoIdx < zebraIdx) {
		t.Errorf("expected alphabetical order: apple < mango < zebra, got indices %d, %d, %d", appleIdx, mangoIdx, zebraIdx)
	}
}

func TestDu_GlobalSizeSort(t *testing.T) {
	// Regression test: size sorting should be global, not per-directory
	// Without global sort, a large subdir's contents would appear before smaller
	// files at the root level, breaking the size order
	dir := setupTestDir(t, map[string]string{
		"small_root.txt":          strings.Repeat("x", 10),      // ~10 bytes
		"large_dir/big_file.txt":  strings.Repeat("x", 500),     // ~500 bytes
		"medium_root.txt":         strings.Repeat("x", 100),     // ~100 bytes
	})

	output := captureOutput(func() {
		Run(&Params{
			Paths:    []string{dir},
			Bytes:    true, // Use bytes for precise comparison
			All:      true,
			Sort:     "size",
			MaxDepth: -1,
		})
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Extract sizes from output
	sizes := make([]int64, 0, len(lines))
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			size, _ := strconv.ParseInt(parts[0], 10, 64)
			sizes = append(sizes, size)
		}
	}

	// Verify sizes are in ascending order (global sort)
	for i := 1; i < len(sizes); i++ {
		if sizes[i] < sizes[i-1] {
			t.Errorf("sizes not globally sorted at position %d: %d < %d\nFull output:\n%s",
				i, sizes[i], sizes[i-1], output)
			break
		}
	}
}

func TestDu_AllFlag_InterleavedSorting(t *testing.T) {
	// Create a structure where files and dirs should interleave by size
	dir := setupTestDir(t, map[string]string{
		"tiny.txt":             "x",                         // ~1 byte
		"small_dir/file.txt":   strings.Repeat("x", 50),     // dir total ~50 bytes
		"big.txt":              strings.Repeat("x", 500),    // ~500 bytes
		"medium_dir/file.txt":  strings.Repeat("x", 200),    // dir total ~200 bytes
	})

	output := captureOutput(func() {
		Run(&Params{MaxDepth: -1,
			Paths: []string{dir},
			Bytes: true, // Use bytes for precise sizes
			All:   true,
			Sort:  "size",
		})
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Find positions of key entries
	positions := make(map[string]int)
	for i, line := range lines {
		if strings.Contains(line, "tiny.txt") {
			positions["tiny"] = i
		}
		if strings.HasSuffix(line, "small_dir") {
			positions["small_dir"] = i
		}
		if strings.HasSuffix(line, "medium_dir") {
			positions["medium_dir"] = i
		}
		if strings.Contains(line, "big.txt") {
			positions["big"] = i
		}
	}

	// Expected order by size: tiny < small_dir < medium_dir < big
	if positions["tiny"] >= positions["small_dir"] {
		t.Errorf("expected tiny before small_dir")
	}
	if positions["small_dir"] >= positions["medium_dir"] {
		t.Errorf("expected small_dir before medium_dir")
	}
	if positions["medium_dir"] >= positions["big"] {
		t.Errorf("expected medium_dir before big")
	}
}

func TestDu_AllFlag_ReverseSort(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"small.txt": "x",
		"large.txt": strings.Repeat("x", 1000),
	})

	output := captureOutput(func() {
		Run(&Params{MaxDepth: -1,
			Paths:   []string{dir},
			Human:   true,
			All:     true,
			Sort:    "size",
			Reverse: true,
		})
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")

	smallIdx := -1
	largeIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "small.txt") {
			smallIdx = i
		}
		if strings.Contains(line, "large.txt") {
			largeIdx = i
		}
	}

	// With reverse sort, large should come before small
	if largeIdx >= smallIdx {
		t.Errorf("expected large.txt before small.txt in reverse-sorted output, got large at %d, small at %d", largeIdx, smallIdx)
	}
}

func TestDu_MaxDepth_WithAllFlag(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"root.txt":                  "root",
		"level1/file.txt":           "level1",
		"level1/level2/file.txt":    "level2",
	})

	// Depth 0 should only show root dir
	output := captureOutput(func() {
		Run(&Params{
			Paths:    []string{dir},
			Human:    true,
			All:      true,
			MaxDepth: 0,
			Sort:     "none",
		})
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	// With depth 0, we should see root files and root dir only
	// Files at root level + the root dir itself
	if !strings.Contains(output, "root.txt") {
		t.Errorf("expected root.txt at depth 0")
	}
	// Should NOT see level1 directory or its contents listed separately
	for _, line := range lines {
		if strings.Contains(line, "level1") && !strings.HasSuffix(line, dir) {
			t.Errorf("unexpected level1 entry at depth 0: %s", line)
		}
	}
}

func TestDu_Summarize_IgnoresAllFlag(t *testing.T) {
	dir := setupTestDir(t, map[string]string{
		"file1.txt": "hello",
		"file2.txt": "world",
	})

	output := captureOutput(func() {
		Run(&Params{MaxDepth: -1,
			Paths:     []string{dir},
			Human:     true,
			All:       true,
			Summarize: true,
			Sort:      "none",
		})
	})

	// Summarize (-s) is equivalent to -d 0, should only show root
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// Should show files at root + the root dir (since All is true)
	// Actually with summarize + all, we should still see files at depth 0
	if !strings.Contains(output, "file1.txt") || !strings.Contains(output, "file2.txt") {
		t.Errorf("expected files to appear with --all even with --summarize: %s", output)
	}
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (2 files + root dir) with -s -a, got %d: %v", len(lines), lines)
	}
}
