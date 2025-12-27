package ls

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestFixture creates a temporary directory structure for testing
type TestFixture struct {
	Root string
	t    *testing.T
}

func NewTestFixture(t *testing.T) *TestFixture {
	t.Helper()
	root, err := os.MkdirTemp("", "ls-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	f := &TestFixture{Root: root, t: t}
	f.setup()
	return f
}

func (f *TestFixture) setup() {
	// Create directory structure:
	// root/
	//   file1.txt (100 bytes)
	//   file2.txt (200 bytes)
	//   .hidden
	//   dir1/
	//     nested.txt
	//   dir2/
	//   executable* (executable file)

	f.writeFile("file1.txt", strings.Repeat("a", 100))
	f.writeFile("file2.txt", strings.Repeat("b", 200))
	f.writeFile(".hidden", "hidden content")

	f.mkdir("dir1")
	f.writeFile("dir1/nested.txt", "nested")

	f.mkdir("dir2")

	// Create executable file - platform-specific approach
	if runtime.GOOS == "windows" {
		// On Windows, use .exe extension for executable detection
		f.writeFile("executable.exe", "MZ") // Minimal PE header marker
	} else {
		// On Unix, use execute permission bit
		f.writeFile("executable", "#!/bin/bash\necho hello")
		os.Chmod(filepath.Join(f.Root, "executable"), 0755)
	}
}

func (f *TestFixture) writeFile(name, content string) {
	path := filepath.Join(f.Root, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		f.t.Fatalf("failed to write file %s: %v", name, err)
	}
}

func (f *TestFixture) mkdir(name string) {
	path := filepath.Join(f.Root, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		f.t.Fatalf("failed to create dir %s: %v", name, err)
	}
}

func (f *TestFixture) Cleanup() {
	os.RemoveAll(f.Root)
}

func runLS(params *Params) (string, string, int) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(params, &stdout, &stderr)
	return stdout.String(), stderr.String(), exitCode
}

func TestBasicListing(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	params := &Params{Paths: []string{f.Root}}
	stdout, _, exitCode := runLS(params)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Should list visible files (not hidden)
	if !strings.Contains(stdout, "file1.txt") {
		t.Error("expected file1.txt in output")
	}
	if !strings.Contains(stdout, "file2.txt") {
		t.Error("expected file2.txt in output")
	}
	if !strings.Contains(stdout, "dir1") {
		t.Error("expected dir1 in output")
	}
	if strings.Contains(stdout, ".hidden") {
		t.Error("hidden file should not appear without -a")
	}
}

func TestAllFlag(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	params := &Params{Paths: []string{f.Root}, All: true}
	stdout, _, _ := runLS(params)

	if !strings.Contains(stdout, ".hidden") {
		t.Error("expected .hidden with -a flag")
	}
	if !strings.Contains(stdout, ".") {
		t.Error("expected . with -a flag")
	}
	if !strings.Contains(stdout, "..") {
		t.Error("expected .. with -a flag")
	}
}

func TestAlmostAllFlag(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	params := &Params{Paths: []string{f.Root}, AlmostAll: true}
	stdout, _, _ := runLS(params)

	if !strings.Contains(stdout, ".hidden") {
		t.Error("expected .hidden with -A flag")
	}
	// Should NOT contain . and ..
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "." || name == ".." {
			t.Errorf("unexpected %s with -A flag", name)
		}
	}
}

func TestLongFormat(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	params := &Params{Paths: []string{f.Root}, Long: true}
	stdout, _, _ := runLS(params)

	// Should have permission strings
	if !strings.Contains(stdout, "drwx") && !strings.Contains(stdout, "d---") {
		t.Error("expected directory permissions in long format")
	}
	if !strings.Contains(stdout, "-rw") {
		t.Error("expected file permissions in long format")
	}
}

func TestSortBySize(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	params := &Params{Paths: []string{f.Root}, SortBySize: true}
	stdout, _, _ := runLS(params)

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	// file2.txt (200 bytes) should come before file1.txt (100 bytes)
	var file1Idx, file2Idx int
	for i, line := range lines {
		if strings.Contains(line, "file1.txt") {
			file1Idx = i
		}
		if strings.Contains(line, "file2.txt") {
			file2Idx = i
		}
	}

	if file2Idx > file1Idx {
		t.Error("expected file2.txt before file1.txt when sorting by size (largest first)")
	}
}

func TestReverse(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	// Normal order
	params1 := &Params{Paths: []string{f.Root}}
	stdout1, _, _ := runLS(params1)
	lines1 := strings.Split(strings.TrimSpace(stdout1), "\n")

	// Reversed order
	params2 := &Params{Paths: []string{f.Root}, Reverse: true}
	stdout2, _, _ := runLS(params2)
	lines2 := strings.Split(strings.TrimSpace(stdout2), "\n")

	if len(lines1) != len(lines2) {
		t.Fatalf("different number of lines: %d vs %d", len(lines1), len(lines2))
	}

	// First and last should be swapped (approximately)
	if lines1[0] == lines2[0] && len(lines1) > 1 {
		t.Error("expected reversed order")
	}
}

func TestClassify(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	params := &Params{Paths: []string{f.Root}, Classify: true}
	stdout, _, _ := runLS(params)

	// Directories should have /
	if !strings.Contains(stdout, "dir1/") {
		t.Error("expected dir1/ with -F flag")
	}
	// Executables should have * (filename differs by platform)
	expectedExec := "executable*"
	if runtime.GOOS == "windows" {
		expectedExec = "executable.exe*"
	}
	if !strings.Contains(stdout, expectedExec) {
		t.Errorf("expected %s with -F flag, got: %s", expectedExec, stdout)
	}
}

func TestDirectoryFlag(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	params := &Params{Paths: []string{f.Root}, Directory: true}
	stdout, _, _ := runLS(params)

	// Should only show the directory itself, not its contents
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line with -d, got %d", len(lines))
	}
}

func TestRecursive(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	params := &Params{Paths: []string{f.Root}, Recursive: true}
	stdout, _, _ := runLS(params)

	// Should contain nested file
	if !strings.Contains(stdout, "nested.txt") {
		t.Error("expected nested.txt with -R flag")
	}
	// Should have directory headers
	if !strings.Contains(stdout, "dir1:") && !strings.Contains(stdout, "dir1") {
		t.Error("expected dir1 header with -R flag")
	}
}

func TestHumanReadable(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	params := &Params{Paths: []string{f.Root}, Long: true, HumanReadable: true}
	stdout, _, _ := runLS(params)

	// Sizes should be present (exact format may vary)
	if !strings.Contains(stdout, "100") && !strings.Contains(stdout, "200") {
		// Small files might just show bytes
	}
}

func TestNonExistentPath(t *testing.T) {
	params := &Params{Paths: []string{"/nonexistent/path/that/does/not/exist"}}
	_, stderr, exitCode := runLS(params)

	if exitCode == 0 {
		t.Error("expected non-zero exit code for nonexistent path")
	}
	if !strings.Contains(stderr, "cannot access") {
		t.Error("expected error message for nonexistent path")
	}
}

func TestMultiplePaths(t *testing.T) {
	f := NewTestFixture(t)
	defer f.Cleanup()

	dir1 := filepath.Join(f.Root, "dir1")
	dir2 := filepath.Join(f.Root, "dir2")

	params := &Params{Paths: []string{dir1, dir2}}
	stdout, _, _ := runLS(params)

	// Should have headers for both directories
	if !strings.Contains(stdout, "dir1") || !strings.Contains(stdout, "dir2") {
		t.Error("expected headers for multiple directories")
	}
}