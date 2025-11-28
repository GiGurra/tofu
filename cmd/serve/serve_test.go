package serve

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServeCommand(t *testing.T) {
	// Create a temp dir with some files
	tmpDir, err := os.MkdirTemp("", "serve-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create index.html
	indexContent := "<html>index</html>"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to create index.html: %v", err)
	}

	// Create other.html
	otherContent := "<html>other</html>"
	if err := os.WriteFile(filepath.Join(tmpDir, "other.html"), []byte(otherContent), 0644); err != nil {
		t.Fatalf("Failed to create other.html: %v", err)
	}

	// Find a free port? or just use a random one.
	// Let's try 0 to let OS choose, but my code uses int port.
	// Let's pick a random high port.
	port := 45678

	params := &Params{
		Port:               port,
		Dir:                tmpDir,
		Host:               "localhost",
		SpaMode:            true,
		NoCache:            true,
		ReadTimeoutMillis:  1000,
		WriteTimeoutMillis: 1000,
		IdleTimeoutMillis:  1000,
		MaxHeaderBytes:     1024,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- Run(ctx, params)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// Test 1: Get index.html
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Fatalf("Failed to get root: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != indexContent {
		t.Errorf("Expected index content, got %s", string(body))
	}

	// Test 2: Get other.html
	resp, err = http.Get(baseURL + "/other.html")
	if err != nil {
		t.Fatalf("Failed to get other.html: %v", err)
	}
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	if string(body) != otherContent {
		t.Errorf("Expected other content, got %s", string(body))
	}

	// Test 3: SPA Fallback (non-existent file)
	resp, err = http.Get(baseURL + "/missing-page")
	if err != nil {
		t.Fatalf("Failed to get missing page: %v", err)
	}
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	if string(body) != indexContent {
		t.Errorf("Expected SPA fallback to index content, got %s", string(body))
	}

	// Test 4: No Cache Headers
	if resp.Header.Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
		t.Errorf("Expected Cache-Control header, got %s", resp.Header.Get("Cache-Control"))
	}

	// Shutdown
	cancel()
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Run did not exit")
	}
}
