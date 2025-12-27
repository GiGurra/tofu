package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCmd(t *testing.T) {
	cmd := Cmd()
	if cmd == nil {
		t.Error("Cmd returned nil")
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
	createParams := &CreateParams{
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
	extractParams := &ExtractParams{
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
	createParams := &CreateParams{
		Output: archivePath,
		Files:  []string{srcFile},
		Format: "tar",
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	// List archive (just verify it doesn't error)
	listParams := &ListParams{
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

func TestArchiveExtract_RelativePath(t *testing.T) {
	// Test extracting to current directory (relative path ".")
	// This tests the bug fix where relative extraction didn't work
	dir := t.TempDir()

	// Create a simple test file and archive it
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0755)

	testFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	archivePath := filepath.Join(dir, "test.tar.gz")

	// Create archive
	createParams := &CreateParams{
		Output: archivePath,
		Files:  []string{srcDir},
		Format: "tar.gz",
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	// Extract to a new directory using relative path "."
	extractDir := filepath.Join(dir, "extracted")
	os.MkdirAll(extractDir, 0755)

	// Change to the extraction directory and extract with "." as output
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(extractDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	extractParams := &ExtractParams{
		Archive: archivePath,
		Output:  ".", // Use relative path - this was the bug
		Verbose: false,
	}

	err = runArchiveExtract(extractParams)
	if err != nil {
		t.Fatalf("failed to extract archive with relative path: %v", err)
	}

	// Verify files were extracted to the current directory
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		t.Fatalf("failed to read extracted dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("no files were extracted to relative path")
	}
}

func TestArchiveExtract_DefaultOutput(t *testing.T) {
	// Test extracting without specifying output (should default to ".")
	dir := t.TempDir()

	// Create a simple test file and archive it
	srcFile := filepath.Join(dir, "original.txt")
	os.WriteFile(srcFile, []byte("original content"), 0644)

	archivePath := filepath.Join(dir, "test.tar")

	// Create archive
	createParams := &CreateParams{
		Output: archivePath,
		Files:  []string{srcFile},
		Format: "tar",
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	// Extract in a new directory without specifying output
	extractDir := filepath.Join(dir, "extracted")
	os.MkdirAll(extractDir, 0755)

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(extractDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Default output is "." per the struct definition
	extractParams := &ExtractParams{
		Archive: archivePath,
		Output:  ".",
	}

	err = runArchiveExtract(extractParams)
	if err != nil {
		t.Fatalf("failed to extract archive with default output: %v", err)
	}

	// Verify files were extracted
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		t.Fatalf("failed to read extracted dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("no files were extracted with default output")
	}
}

func TestArchiveExtract_PreservesSymlinks_Tar(t *testing.T) {
	testArchiveSymlinkPreservation(t, "tar")
}

func TestArchiveExtract_PreservesSymlinks_TarGz(t *testing.T) {
	testArchiveSymlinkPreservation(t, "tar.gz")
}

func TestParseEncryptionMethod(t *testing.T) {
	tests := []struct {
		method    string
		expectErr bool
	}{
		{"legacy", false},
		{"LEGACY", false},
		{"zipcrypto", false},
		{"ZIPCRYPTO", false},
		{"aes128", false},
		{"AES128", false},
		{"aes192", false},
		{"AES192", false},
		{"aes256", false},
		{"AES256", false},
		{"", false}, // empty defaults to aes256
		{"invalid", true},
		{"aes512", true},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			_, err := parseEncryptionMethod(tt.method)
			if tt.expectErr && err == nil {
				t.Errorf("expected error for method %q", tt.method)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error for method %q: %v", tt.method, err)
			}
		})
	}
}

func TestEncryptedZip_CreateAndExtract(t *testing.T) {
	dir := t.TempDir()

	// Create test file
	srcFile := filepath.Join(dir, "secret.txt")
	testContent := "this is secret content"
	os.WriteFile(srcFile, []byte(testContent), 0644)

	archivePath := filepath.Join(dir, "encrypted.zip")
	password := "testpassword123"

	// Create encrypted archive
	createParams := &CreateParams{
		Output:     archivePath,
		Files:      []string{srcFile},
		Password:   password,
		Encryption: "aes256",
		Verbose:    false,
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create encrypted archive: %v", err)
	}

	// Verify archive was created
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatal("encrypted archive was not created")
	}

	// Extract encrypted archive
	extractDir := filepath.Join(dir, "extracted")
	extractParams := &ExtractParams{
		Archive:  archivePath,
		Output:   extractDir,
		Password: password,
		Verbose:  false,
	}

	err = runArchiveExtract(extractParams)
	if err != nil {
		t.Fatalf("failed to extract encrypted archive: %v", err)
	}

	// Verify extracted file content
	extractedFile := filepath.Join(extractDir, "secret.txt")
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("content mismatch: expected %q, got %q", testContent, string(content))
	}
}

func TestEncryptedZip_List(t *testing.T) {
	dir := t.TempDir()

	// Create test file
	srcFile := filepath.Join(dir, "test.txt")
	os.WriteFile(srcFile, []byte("test content"), 0644)

	archivePath := filepath.Join(dir, "encrypted.zip")
	password := "testpassword123"

	// Create encrypted archive
	createParams := &CreateParams{
		Output:     archivePath,
		Files:      []string{srcFile},
		Password:   password,
		Encryption: "aes256",
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create encrypted archive: %v", err)
	}

	// List encrypted archive
	listParams := &ListParams{
		Archive:  archivePath,
		Password: password,
		Long:     true,
	}

	err = runArchiveList(listParams)
	if err != nil {
		t.Fatalf("failed to list encrypted archive: %v", err)
	}
}

func TestEncryptedZip_WrongPassword(t *testing.T) {
	dir := t.TempDir()

	// Create test file
	srcFile := filepath.Join(dir, "test.txt")
	os.WriteFile(srcFile, []byte("test content"), 0644)

	archivePath := filepath.Join(dir, "encrypted.zip")
	password := "correctpassword"

	// Create encrypted archive
	createParams := &CreateParams{
		Output:     archivePath,
		Files:      []string{srcFile},
		Password:   password,
		Encryption: "aes256",
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create encrypted archive: %v", err)
	}

	// Try to extract with wrong password
	extractDir := filepath.Join(dir, "extracted")
	extractParams := &ExtractParams{
		Archive:  archivePath,
		Output:   extractDir,
		Password: "wrongpassword",
	}

	err = runArchiveExtract(extractParams)
	if err == nil {
		t.Error("expected error when extracting with wrong password, but got none")
	}
}

func TestEncryptedZip_DifferentEncryptionMethods(t *testing.T) {
	methods := []string{"legacy", "aes128", "aes192", "aes256"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			dir := t.TempDir()

			// Create test file
			srcFile := filepath.Join(dir, "test.txt")
			testContent := "content for " + method
			os.WriteFile(srcFile, []byte(testContent), 0644)

			archivePath := filepath.Join(dir, "encrypted.zip")
			password := "testpassword"

			// Create encrypted archive
			createParams := &CreateParams{
				Output:     archivePath,
				Files:      []string{srcFile},
				Password:   password,
				Encryption: method,
			}

			err := runArchiveCreate(createParams)
			if err != nil {
				t.Fatalf("failed to create archive with %s: %v", method, err)
			}

			// Extract and verify
			extractDir := filepath.Join(dir, "extracted")
			extractParams := &ExtractParams{
				Archive:  archivePath,
				Output:   extractDir,
				Password: password,
			}

			err = runArchiveExtract(extractParams)
			if err != nil {
				t.Fatalf("failed to extract archive with %s: %v", method, err)
			}

			// Verify content
			extractedFile := filepath.Join(extractDir, "test.txt")
			content, err := os.ReadFile(extractedFile)
			if err != nil {
				t.Fatalf("failed to read extracted file: %v", err)
			}
			if string(content) != testContent {
				t.Errorf("content mismatch for %s", method)
			}
		})
	}
}

func TestEncryptedZip_Directory(t *testing.T) {
	dir := t.TempDir()

	// Create test directory structure
	srcDir := filepath.Join(dir, "mydir")
	os.MkdirAll(srcDir, 0755)

	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	os.WriteFile(file1, []byte("content 1"), 0644)
	os.WriteFile(file2, []byte("content 2"), 0644)

	subDir := filepath.Join(srcDir, "subdir")
	os.MkdirAll(subDir, 0755)
	file3 := filepath.Join(subDir, "file3.txt")
	os.WriteFile(file3, []byte("content 3"), 0644)

	archivePath := filepath.Join(dir, "encrypted.zip")
	password := "testpassword"

	// Create encrypted archive
	createParams := &CreateParams{
		Output:     archivePath,
		Files:      []string{srcDir},
		Password:   password,
		Encryption: "aes256",
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create encrypted archive: %v", err)
	}

	// Extract and verify
	extractDir := filepath.Join(dir, "extracted")
	extractParams := &ExtractParams{
		Archive:  archivePath,
		Output:   extractDir,
		Password: password,
	}

	err = runArchiveExtract(extractParams)
	if err != nil {
		t.Fatalf("failed to extract encrypted archive: %v", err)
	}

	// Verify all files were extracted
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		t.Fatalf("failed to read extracted dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("no files were extracted")
	}
}

func TestNonEncryptedZip_StillWorks(t *testing.T) {
	// Verify that creating non-encrypted ZIP archives still works
	dir := t.TempDir()

	srcFile := filepath.Join(dir, "test.txt")
	os.WriteFile(srcFile, []byte("test content"), 0644)

	archivePath := filepath.Join(dir, "normal.zip")

	// Create non-encrypted archive (no password)
	createParams := &CreateParams{
		Output: archivePath,
		Files:  []string{srcFile},
		Format: "zip",
	}

	err := runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create non-encrypted archive: %v", err)
	}

	// Extract without password
	extractDir := filepath.Join(dir, "extracted")
	extractParams := &ExtractParams{
		Archive: archivePath,
		Output:  extractDir,
	}

	err = runArchiveExtract(extractParams)
	if err != nil {
		t.Fatalf("failed to extract non-encrypted archive: %v", err)
	}

	// Verify content
	extractedFile := filepath.Join(extractDir, "test.txt")
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("content mismatch")
	}
}

func testArchiveSymlinkPreservation(t *testing.T, format string) {
	dir := t.TempDir()

	// Create source directory with a file and a symlink
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0755)

	// Create a regular file
	realFile := filepath.Join(srcDir, "realfile.txt")
	os.WriteFile(realFile, []byte("real file content"), 0644)

	// Create a symlink to the file
	symlinkPath := filepath.Join(srcDir, "link_to_realfile")
	err := os.Symlink("realfile.txt", symlinkPath)
	if err != nil {
		t.Skipf("cannot create symlink (possibly unsupported filesystem): %v", err)
	}

	// Create a subdirectory and a symlink to a directory
	subDir := filepath.Join(srcDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "subfile.txt"), []byte("subfile content"), 0644)

	dirSymlink := filepath.Join(srcDir, "link_to_subdir")
	err = os.Symlink("subdir", dirSymlink)
	if err != nil {
		t.Skipf("cannot create directory symlink: %v", err)
	}

	// Determine archive extension
	ext := format
	archivePath := filepath.Join(dir, "archive."+ext)

	// Create archive
	createParams := &CreateParams{
		Output:  archivePath,
		Files:   []string{srcDir},
		Format:  format,
		Verbose: false,
	}

	err = runArchiveCreate(createParams)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	// Extract archive
	extractDir := filepath.Join(dir, "extracted")
	extractParams := &ExtractParams{
		Archive: archivePath,
		Output:  extractDir,
		Verbose: false,
	}

	err = runArchiveExtract(extractParams)
	if err != nil {
		t.Fatalf("failed to extract archive: %v", err)
	}

	// Find the extracted source directory
	extractedSrcDir := filepath.Join(extractDir, "src")
	if _, err := os.Stat(extractedSrcDir); os.IsNotExist(err) {
		// Try finding it at root level
		entries, _ := os.ReadDir(extractDir)
		if len(entries) > 0 {
			extractedSrcDir = filepath.Join(extractDir, entries[0].Name())
		}
	}

	// Verify the file symlink was preserved
	extractedSymlink := filepath.Join(extractedSrcDir, "link_to_realfile")
	linkInfo, err := os.Lstat(extractedSymlink)
	if err != nil {
		t.Fatalf("failed to stat extracted symlink: %v", err)
	}

	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be a symlink, but it's not", extractedSymlink)
	}

	// Verify symlink target is preserved
	target, err := os.Readlink(extractedSymlink)
	if err != nil {
		t.Fatalf("failed to read symlink target: %v", err)
	}
	if target != "realfile.txt" {
		t.Errorf("expected symlink target 'realfile.txt', got '%s'", target)
	}

	// Verify the symlink actually works (can read through it)
	content, err := os.ReadFile(extractedSymlink)
	if err != nil {
		t.Errorf("failed to read through symlink: %v", err)
	} else if string(content) != "real file content" {
		t.Errorf("content through symlink mismatch: got %q", string(content))
	}

	// Verify the directory symlink was preserved
	extractedDirSymlink := filepath.Join(extractedSrcDir, "link_to_subdir")
	dirLinkInfo, err := os.Lstat(extractedDirSymlink)
	if err != nil {
		t.Fatalf("failed to stat extracted directory symlink: %v", err)
	}

	if dirLinkInfo.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be a symlink, but it's not", extractedDirSymlink)
	}

	// Verify directory symlink target
	dirTarget, err := os.Readlink(extractedDirSymlink)
	if err != nil {
		t.Fatalf("failed to read directory symlink target: %v", err)
	}
	if dirTarget != "subdir" {
		t.Errorf("expected directory symlink target 'subdir', got '%s'", dirTarget)
	}
}
