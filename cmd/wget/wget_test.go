package wget

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCmd(t *testing.T) {
	cmd := Cmd()
	if cmd == nil {
		t.Error("Cmd returned nil")
	}
	if cmd.Name() != "wget" {
		t.Errorf("expected Name()='wget', got '%s'", cmd.Name())
	}
}

func TestFilenameFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com/file.zip", "file.zip"},
		{"https://example.com/path/to/document.pdf", "document.pdf"},
		{"https://example.com/", ""},
		{"https://example.com", ""},
		{"https://example.com/file.tar.gz", "file.tar.gz"},
		{"https://example.com/path/", "path"},
		{"https://example.com/download?file=test.zip", "download"},
		{"invalid-url", "invalid-url"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := filenameFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("filenameFromURL(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestParseContentRange(t *testing.T) {
	tests := []struct {
		header    string
		start     int64
		end       int64
		total     int64
		expectErr bool
	}{
		{"bytes 0-499/1234", 0, 499, 1234, false},
		{"bytes 500-999/1234", 500, 999, 1234, false},
		{"bytes 0-0/1", 0, 0, 1, false},
		{"", 0, 0, 0, true},
		{"invalid", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			start, end, total, err := parseContentRange(tt.header)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if start != tt.start || end != tt.end || total != tt.total {
				t.Errorf("parseContentRange(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.header, start, end, total, tt.start, tt.end, tt.total)
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	// Create a test server
	content := "Hello, World! This is test content for wget."
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "44")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	// Create temp directory
	dir := t.TempDir()
	outputFile := filepath.Join(dir, "downloaded.txt")

	params := &Params{
		URL:     server.URL + "/test.txt",
		Output:  outputFile,
		Quiet:   true,
		Timeout: 10,
		Retries: 1,
	}

	err := runWget(params)
	if err != nil {
		t.Fatalf("runWget failed: %v", err)
	}

	// Verify file contents
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}
}

func TestDownloadFileAutoFilename(t *testing.T) {
	content := "Test content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	// Change to temp directory to avoid polluting working directory
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	params := &Params{
		URL:     server.URL + "/myfile.txt",
		Quiet:   true,
		Timeout: 10,
		Retries: 1,
	}

	err := runWget(params)
	if err != nil {
		t.Fatalf("runWget failed: %v", err)
	}

	// Verify file was created with auto-detected name
	expectedFile := filepath.Join(dir, "myfile.txt")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("expected file %s to be created", expectedFile)
	}
}

func TestDownloadFile404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	dir := t.TempDir()
	params := &Params{
		URL:     server.URL + "/notfound.txt",
		Output:  filepath.Join(dir, "output.txt"),
		Quiet:   true,
		Timeout: 10,
		Retries: 1,
	}

	err := runWget(params)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestDownloadWithRedirect(t *testing.T) {
	content := "Redirected content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	dir := t.TempDir()
	outputFile := filepath.Join(dir, "output.txt")

	params := &Params{
		URL:     server.URL + "/redirect",
		Output:  outputFile,
		Quiet:   true,
		Timeout: 10,
		Retries: 1,
	}

	err := runWget(params)
	if err != nil {
		t.Fatalf("runWget failed: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}
}
