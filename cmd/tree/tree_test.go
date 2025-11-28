package tree

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestTree(t *testing.T, root string) {
	// root
	// ├── dir1
	// │   ├── file1.txt
	// │   └── .hidden_file
	// └── dir2
	//     ├── file2.txt
	//     └── subdir3
	//         └── file3.txt
	//     └── .hidden_dir
	//         └── file_in_hidden
	// └── .config
	//     └── config.txt

	mkdir := func(path string) {
		if err := os.MkdirAll(filepath.Join(root, path), 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", path, err)
		}
	}
	createFile := func(path, content string) {
		if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	mkdir("dir1")
	createFile("dir1/file1.txt", "content")
	createFile("dir1/.hidden_file", "content")
	mkdir("dir2")
	createFile("dir2/file2.txt", "content")
	mkdir("dir2/subdir3")
	createFile("dir2/subdir3/file3.txt", "content")
	mkdir("dir2/.hidden_dir")
	createFile("dir2/.hidden_dir/file_in_hidden", "content")
	mkdir(".config")
	createFile(".config/config.txt", "content")
}

func TestTreeCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tofu-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	createTestTree(t, tmpDir)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test 1: Default behavior (no hidden files, infinite depth)
	params := &Params{
		Dir:   tmpDir,
		Depth: -1,
		All:   false,
	}
	if err := Run(params); err != nil {
		t.Errorf("Run default failed: %v", err)
	}

	_ = w.Close()
	out, _ := io.ReadAll(r)
	_ = r.Close()

	expected := tmpDir + `
├── dir1
│   └── file1.txt
└── dir2
    ├── file2.txt
    └── subdir3
        └── file3.txt

4 directories, 3 files`
	if strings.TrimSpace(string(out)) != strings.TrimSpace(expected) {
		t.Fatalf("Default tree output mismatch. Expected:\n%s\nGot:\n%s", expected, string(out))
	}

	// Reset stdout for next test
	r, w, _ = os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test 2: With -a (all files, including hidden)
	params.All = true
	if err := Run(params); err != nil {
		t.Errorf("Run -a failed: %v", err)
	}

	_ = w.Close()
	out, _ = io.ReadAll(r)
	_ = r.Close()

	expectedAll := tmpDir + `
├── .config
│   └── config.txt
├── dir1
│   ├── .hidden_file
│   └── file1.txt
└── dir2
    ├── .hidden_dir
    │   └── file_in_hidden
    ├── file2.txt
    └── subdir3
        └── file3.txt

6 directories, 6 files
`
	if strings.TrimSpace(string(out)) != strings.TrimSpace(expectedAll) {
		t.Fatalf("Tree -a output mismatch. Expected:\n%s\nGot:\n%s", expectedAll, string(out))
	}

	// Reset stdout for next test
	r, w, _ = os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test 3: With -L 1 (depth 1)
	params.All = false // reset
	params.Depth = 1
	if err := Run(params); err != nil {
		t.Errorf("Run -L 1 failed: %v", err)
	}

	_ = w.Close()
	out, _ = io.ReadAll(r)
	_ = r.Close()

	expectedDepth1 := tmpDir + `
├── dir1
└── dir2

3 directories, 0 files
`
	if strings.TrimSpace(string(out)) != strings.TrimSpace(expectedDepth1) {
		t.Fatalf("Tree -L 1 output mismatch. Expected:\n%s\nGot:\n%s", expectedDepth1, string(out))
	}
}
