package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMatchExact(t *testing.T) {
	tests := []struct {
		name       string
		a          string
		b          string
		ignoreCase bool
		expected   bool
	}{
		{"exact match case sensitive", "file.txt", "file.txt", false, true},
		{"no match case sensitive", "file.txt", "File.txt", false, false},
		{"exact match case insensitive", "file.txt", "FILE.TXT", true, true},
		{"no match different strings", "file.txt", "other.txt", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchExact(tt.a, tt.b, tt.ignoreCase)
			if result != tt.expected {
				t.Errorf("matchExact(%q, %q, %v) = %v, expected %v", tt.a, tt.b, tt.ignoreCase, result, tt.expected)
			}
		})
	}
}

func TestMatchContains(t *testing.T) {
	tests := []struct {
		name       string
		tot        string
		substr     string
		ignoreCase bool
		expected   bool
	}{
		{"contains case sensitive", "file.txt", "file", false, true},
		{"contains middle", "my_file_name.txt", "file", false, true},
		{"no match case sensitive", "file.txt", "FILE", false, false},
		{"contains case insensitive", "file.txt", "FILE", true, true},
		{"no match different substring", "file.txt", "doc", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchContains(tt.tot, tt.substr, tt.ignoreCase)
			if result != tt.expected {
				t.Errorf("matchContains(%q, %q, %v) = %v, expected %v", tt.tot, tt.substr, tt.ignoreCase, result, tt.expected)
			}
		})
	}
}

func TestMatchPrefix(t *testing.T) {
	tests := []struct {
		name       string
		tot        string
		prefix     string
		ignoreCase bool
		expected   bool
	}{
		{"prefix match case sensitive", "file.txt", "file", false, true},
		{"no match case sensitive", "file.txt", "FILE", false, false},
		{"prefix match case insensitive", "file.txt", "FILE", true, true},
		{"no match wrong prefix", "file.txt", "doc", false, false},
		{"prefix empty string", "file.txt", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPrefix(tt.tot, tt.prefix, tt.ignoreCase)
			if result != tt.expected {
				t.Errorf("matchPrefix(%q, %q, %v) = %v, expected %v", tt.tot, tt.prefix, tt.ignoreCase, result, tt.expected)
			}
		})
	}
}

func TestMatchSuffix(t *testing.T) {
	tests := []struct {
		name       string
		tot        string
		suffix     string
		ignoreCase bool
		expected   bool
	}{
		{"suffix match case sensitive", "file.txt", ".txt", false, true},
		{"no match case sensitive", "file.txt", ".TXT", false, false},
		{"suffix match case insensitive", "file.txt", ".TXT", true, true},
		{"no match wrong suffix", "file.txt", ".doc", false, false},
		{"suffix empty string", "file.txt", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchSuffix(tt.tot, tt.suffix, tt.ignoreCase)
			if result != tt.expected {
				t.Errorf("matchSuffix(%q, %q, %v) = %v, expected %v", tt.tot, tt.suffix, tt.ignoreCase, result, tt.expected)
			}
		})
	}
}

func TestRunFind_ExactMatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "test.txt")
	file2 := filepath.Join(tmpDir, "Test.txt")
	file3 := filepath.Join(tmpDir, "other.txt")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file3, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &FindParams{
		SearchTerm: "test.txt",
		SearchType: SearchTypeExact,
		IgnoreCase: false,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeAll},
		Quiet:      false,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	output := stdout.String()
	if !strings.Contains(output, "test.txt") {
		t.Errorf("Expected output to contain 'test.txt', got %q", output)
	}
	if strings.Contains(output, "Test.txt") {
		t.Errorf("Expected output to NOT contain 'Test.txt' (case sensitive), got %q", output)
	}
	if strings.Contains(output, "other.txt") {
		t.Errorf("Expected output to NOT contain 'other.txt', got %q", output)
	}
}

func TestRunFind_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectories to ensure files are distinct on case-insensitive filesystems
	dir1 := filepath.Join(tmpDir, "1")
	dir2 := filepath.Join(tmpDir, "2")
	dir3 := filepath.Join(tmpDir, "3")

	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}
	if err := os.MkdirAll(dir3, 0755); err != nil {
		t.Fatalf("Failed to create dir3: %v", err)
	}

	file1 := filepath.Join(dir1, "test.txt")
	file2 := filepath.Join(dir2, "Test.txt")
	file3 := filepath.Join(dir3, "TEST.txt")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file3, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &FindParams{
		SearchTerm: "test.txt",
		SearchType: SearchTypeExact,
		IgnoreCase: true,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeAll},
		Quiet:      false,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	output := stdout.String()
	if !strings.Contains(output, filepath.Join("1", "test.txt")) {
		t.Errorf("Expected output to contain '1/test.txt', got %q", output)
	}
	if !strings.Contains(output, filepath.Join("2", "Test.txt")) {
		t.Errorf("Expected output to contain '2/Test.txt' (case insensitive), got %q", output)
	}
	if !strings.Contains(output, filepath.Join("3", "TEST.txt")) {
		t.Errorf("Expected output to contain '3/TEST.txt' (case insensitive), got %q", output)
	}
}

func TestRunFind_ContainsSearch(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test_file.txt")
	file2 := filepath.Join(tmpDir, "my_test_doc.txt")
	file3 := filepath.Join(tmpDir, "other.txt")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file3, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &FindParams{
		SearchTerm: "test",
		SearchType: SearchTypeContains,
		IgnoreCase: false,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeAll},
		Quiet:      false,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	output := stdout.String()
	if !strings.Contains(output, "test_file.txt") {
		t.Errorf("Expected output to contain 'test_file.txt', got %q", output)
	}
	if !strings.Contains(output, "my_test_doc.txt") {
		t.Errorf("Expected output to contain 'my_test_doc.txt', got %q", output)
	}
	if strings.Contains(output, "other.txt") {
		t.Errorf("Expected output to NOT contain 'other.txt', got %q", output)
	}
}

func TestRunFind_PrefixSearch(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test_file.txt")
	file2 := filepath.Join(tmpDir, "my_test.txt")
	file3 := filepath.Join(tmpDir, "testdoc.txt")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file3, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &FindParams{
		SearchTerm: "test",
		SearchType: SearchTypePrefix,
		IgnoreCase: false,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeAll},
		Quiet:      false,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	output := stdout.String()
	if !strings.Contains(output, "test_file.txt") {
		t.Errorf("Expected output to contain 'test_file.txt', got %q", output)
	}
	if !strings.Contains(output, "testdoc.txt") {
		t.Errorf("Expected output to contain 'testdoc.txt', got %q", output)
	}
	if strings.Contains(output, "my_test.txt") {
		t.Errorf("Expected output to NOT contain 'my_test.txt', got %q", output)
	}
}

func TestRunFind_SuffixSearch(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file.txt")
	file2 := filepath.Join(tmpDir, "doc.txt")
	file3 := filepath.Join(tmpDir, "readme.md")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file3, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &FindParams{
		SearchTerm: ".txt",
		SearchType: SearchTypeSuffix,
		IgnoreCase: false,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeAll},
		Quiet:      false,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	output := stdout.String()
	if !strings.Contains(output, "file.txt") {
		t.Errorf("Expected output to contain 'file.txt', got %q", output)
	}
	if !strings.Contains(output, "doc.txt") {
		t.Errorf("Expected output to contain 'doc.txt', got %q", output)
	}
	if strings.Contains(output, "readme.md") {
		t.Errorf("Expected output to NOT contain 'readme.md', got %q", output)
	}
}

func TestRunFind_RegexSearch(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")
	file3 := filepath.Join(tmpDir, "other.txt")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file3, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &FindParams{
		SearchTerm: `test\d\.txt`,
		SearchType: SearchTypeRegex,
		IgnoreCase: false,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeAll},
		Quiet:      false,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	output := stdout.String()
	if !strings.Contains(output, "test1.txt") {
		t.Errorf("Expected output to contain 'test1.txt', got %q", output)
	}
	if !strings.Contains(output, "test2.txt") {
		t.Errorf("Expected output to contain 'test2.txt', got %q", output)
	}
	if strings.Contains(output, "other.txt") {
		t.Errorf("Expected output to NOT contain 'other.txt', got %q", output)
	}
}

func TestRunFind_TypeFiltering_FilesOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files and directories
	file1 := filepath.Join(tmpDir, "test.txt")
	dir1 := filepath.Join(tmpDir, "test_dir")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	params := &FindParams{
		SearchTerm: "test",
		SearchType: SearchTypeContains,
		IgnoreCase: false,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeFile},
		Quiet:      false,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	output := stdout.String()
	if !strings.Contains(output, "test.txt") {
		t.Errorf("Expected output to contain 'test.txt', got %q", output)
	}
	if strings.Contains(output, "test_dir") {
		t.Errorf("Expected output to NOT contain 'test_dir' (dirs filtered out), got %q", output)
	}
}

func TestRunFind_TypeFiltering_DirsOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files and directories
	file1 := filepath.Join(tmpDir, "test.txt")
	dir1 := filepath.Join(tmpDir, "test_dir")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	params := &FindParams{
		SearchTerm: "test",
		SearchType: SearchTypeContains,
		IgnoreCase: false,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeDir},
		Quiet:      false,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	output := stdout.String()
	if strings.Contains(output, "test.txt") {
		t.Errorf("Expected output to NOT contain 'test.txt' (files filtered out), got %q", output)
	}
	if !strings.Contains(output, "test_dir") {
		t.Errorf("Expected output to contain 'test_dir', got %q", output)
	}
}

func TestRunFind_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	file1 := filepath.Join(tmpDir, "test.txt")
	file2 := filepath.Join(subDir, "test_nested.txt")

	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	params := &FindParams{
		SearchTerm: "test",
		SearchType: SearchTypeContains,
		IgnoreCase: false,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeAll},
		Quiet:      false,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	output := stdout.String()
	if !strings.Contains(output, "test.txt") {
		t.Errorf("Expected output to contain 'test.txt', got %q", output)
	}
	if !strings.Contains(output, "test_nested.txt") {
		t.Errorf("Expected output to contain 'test_nested.txt', got %q", output)
	}
}

func TestRunFind_QuietMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	file1 := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &FindParams{
		SearchTerm: "test",
		SearchType: SearchTypeContains,
		IgnoreCase: false,
		WorkDir:    tmpDir,
		Types:      []FsItemType{FsItemTypeAll},
		Quiet:      true,
	}

	var stdout, stderr bytes.Buffer
	runFind(params, &stdout, &stderr)

	// In quiet mode, errors should be suppressed
	if stderr.Len() > 0 {
		t.Errorf("Expected no stderr output in quiet mode, got %q", stderr.String())
	}
}

func TestExistsAccessibleDirDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Test existing directory
	if !ExistsAccessibleDirDir(tmpDir) {
		t.Errorf("Expected ExistsAccessibleDirDir to return true for existing directory")
	}

	// Test non-existent directory
	if ExistsAccessibleDirDir(filepath.Join(tmpDir, "nonexistent")) {
		t.Errorf("Expected ExistsAccessibleDirDir to return false for non-existent directory")
	}

	// Test file (not directory)
	file := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if ExistsAccessibleDirDir(file) {
		t.Errorf("Expected ExistsAccessibleDirDir to return false for file")
	}
}
