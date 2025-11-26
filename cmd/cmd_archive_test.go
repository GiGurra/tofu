package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveCmd(t *testing.T) {
	cmd := ArchiveCmd()
	if cmd == nil {
		t.Error("ArchiveCmd returned nil")
	}
	if cmd.Name() != "archive" {
		t.Errorf("expected Name()='archive', got '%s'", cmd.Name())
	}
}

func TestParseFormatString(t *testing.T) {
	tests := []struct {
		format    string
		expectErr bool
	}{
		{"tar", false},
		{"tar.gz", false},
		{"tgz", false},
		{"tar.bz2", false},
		{"tbz2", false},
		{"tar.xz", false},
		{"txz", false},
		{"tar.zst", false},
		{"tar.lz4", false},
		{"zip", false},
		{"7z", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			_, err := parseFormatString(tt.format)
			if tt.expectErr && err == nil {
				t.Errorf("expected error for format %q", tt.format)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error for format %q: %v", tt.format, err)
			}
		})
	}
}

func TestParseFormatFromExtension(t *testing.T) {
	tests := []struct {
		filename  string
		expectErr bool
	}{
		{"archive.tar", false},
		{"archive.tar.gz", false},
		{"archive.tgz", false},
		{"archive.tar.bz2", false},
		{"archive.tbz2", false},
		{"archive.tar.xz", false},
		{"archive.txz", false},
		{"archive.tar.zst", false},
		{"archive.tar.lz4", false},
		{"archive.zip", false},
		{"archive.7z", false},
		{"archive.txt", true},
		{"noextension", true},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			_, err := parseFormatFromExtension(tt.filename)
			if tt.expectErr && err == nil {
				t.Errorf("expected error for filename %q", tt.filename)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error for filename %q: %v", tt.filename, err)
			}
		})
	}
}

func TestArchiveCreateAndExtract_Tar(t *testing.T) {
	testArchiveRoundTrip(t, "tar")
}

func TestArchiveCreateAndExtract_TarGz(t *testing.T) {
	testArchiveRoundTrip(t, "tar.gz")
}

func TestArchiveCreateAndExtract_Zip(t *testing.T) {
	testArchiveRoundTrip(t, "zip")
}

func testArchiveRoundTrip(t *testing.T, format string) {
	dir := t.TempDir()

	// Create test files
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0755)

	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	os.WriteFile(file1, []byte("content of file 1"), 0644)
	os.WriteFile(file2, []byte("content of file 2"), 0644)

	// Create subdirectory with file
	subDir := filepath.Join(srcDir, "subdir")
	os.MkdirAll(subDir, 0755)
	file3 := filepath.Join(subDir, "file3.txt")
	os.WriteFile(file3, []byte("content of file 3"), 0644)

	// Determine archive extension
	ext := format
	if format == "tar.gz" {
		ext = "tar.gz"
	}
	archivePath := filepath.Join(dir, "archive."+ext)

	// Create archive
	createParams := &ArchiveCreateParams{
		Output:  archivePath,
		Files:   []string{srcDir},
		Format:  format,
		Verbose: false,
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	// Verify archive was created
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatal("archive file was not created")
	}

	// Extract archive
	extractDir := filepath.Join(dir, "extracted")
	extractParams := &ArchiveExtractParams{
		Archive: archivePath,
		Output:  extractDir,
		Verbose: false,
	}

	err = runArchiveExtract(extractParams)
	if err != nil {
		t.Fatalf("failed to extract archive: %v", err)
	}

	// Verify extracted files exist (path will include srcDir name)
	// The exact path depends on how files were added
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		t.Fatalf("failed to read extracted dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("no files were extracted")
	}
}

func TestArchiveList(t *testing.T) {
	dir := t.TempDir()

	// Create a simple test file
	srcFile := filepath.Join(dir, "test.txt")
	os.WriteFile(srcFile, []byte("test content"), 0644)

	archivePath := filepath.Join(dir, "test.tar")

	// Create archive
	createParams := &ArchiveCreateParams{
		Output: archivePath,
		Files:  []string{srcFile},
		Format: "tar",
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	// List archive (just verify it doesn't error)
	listParams := &ArchiveListParams{
		Archive: archivePath,
		Long:    true,
	}

	err = runArchiveList(listParams)
	if err != nil {
		t.Fatalf("failed to list archive: %v", err)
	}
}

func TestGetArchiveFormat_Override(t *testing.T) {
	// Format override should take precedence over filename
	format, err := getArchiveFormat("file.zip", "tar.gz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The result should be a compressed archive (tar.gz), not zip
	if format.Extension() != ".tar.gz" {
		t.Errorf("expected .tar.gz extension, got %s", format.Extension())
	}
}

func TestGetArchiveFormat_FromExtension(t *testing.T) {
	format, err := getArchiveFormat("backup.zip", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if format.Extension() != ".zip" {
		t.Errorf("expected .zip extension, got %s", format.Extension())
	}
}
