package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type MockProcessRunner struct {
	StartFunc func() error
	WaitFunc  func() error
	KillFunc  func() error
}

func (m *MockProcessRunner) Start() error {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	return nil
}

func (m *MockProcessRunner) Wait() error {
	if m.WaitFunc != nil {
		return m.WaitFunc()
	}
	return nil
}

func (m *MockProcessRunner) Kill() error {
	if m.KillFunc != nil {
		return m.KillFunc()
	}
	return nil
}

func TestWatchCommand(t *testing.T) {
	// Create a temp directory for watching
	tmpDir, err := os.MkdirTemp("", "watch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file to watch
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Prepare watch params
	params := &WatchParams{
		Dirs:             []string{tmpDir},
		Execute:          "echo test", // Command string is irrelevant for mock
		Recursive:        true,
		PatternType:      WatchPatternTypeGlob,
		RestartPolicy:    "exponential-backoff",
		MinBackoffMillis: 100,
		MaxBackoffMillis: 1000,
		MaxRestarts:      10,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to verify execution
	executed := make(chan struct{}, 10)

	factory := func() ProcessRunner {
		return &MockProcessRunner{
			StartFunc: func() error {
				executed <- struct{}{}
				return nil
			},
			WaitFunc: func() error {
				return nil
			},
			KillFunc: func() error {
				return nil
			},
		}
	}

	// Run watch in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- runWatch(ctx, params, factory)
	}()

	// Give it some time to start watching
	time.Sleep(200 * time.Millisecond)

	// Verify initial start (it runs once at startup)
	select {
	case <-executed:
		// Good
	case <-time.After(1 * time.Second):
		t.Errorf("Command did not run initially")
	}

	// Modify the file
	if err := os.WriteFile(filePath, []byte("modified"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Wait for the command to execute (debounce 100ms)
	select {
	case <-executed:
		// Good
	case <-time.After(1 * time.Second):
		t.Errorf("Command did not run after file change")
	}

	// Cancel the context to stop the watcher
	cancel()

	// Wait for runWatch to return
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("runWatch returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
			t.Errorf("runWatch did not exit after context cancellation")
	}
}

func TestWatchCommandNoPatterns(t *testing.T) {
	// Create a temp directory for watching
	tmpDir, err := os.MkdirTemp("", "watch-test-no-patterns")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file to watch
	filePath1 := filepath.Join(tmpDir, "test1.txt")
	if err := os.WriteFile(filePath1, []byte("initial1"), 0644); err != nil {
		t.Fatalf("Failed to create test file1: %v", err)
	}

	filePath2 := filepath.Join(tmpDir, "another.log")
	if err := os.WriteFile(filePath2, []byte("initial2"), 0644); err != nil {
		t.Fatalf("Failed to create test file2: %v", err)
	}

	// Prepare watch params with empty Patterns
	params := &WatchParams{
		Dirs:      []string{tmpDir},
		Execute:   "echo file changed",
		Recursive: true,
		// Patterns is intentionally left empty
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	executed := make(chan struct{}, 10)
	factory := func() ProcessRunner {
		return &MockProcessRunner{
			StartFunc: func() error {
				executed <- struct{}{}
				return nil
			},
			WaitFunc: func() error {
				return nil
			},
			KillFunc: func() error {
				return nil
			},
		}
	}

	// Run watch in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- runWatch(ctx, params, factory)
	}()

	// Give it some time to start watching
	time.Sleep(200 * time.Millisecond)

	// Verify initial start
	select {
	case <-executed:
		// Good
	case <-time.After(1 * time.Second):
		t.Errorf("Command did not run initially")
	}

	// Modify first file
	if err := os.WriteFile(filePath1, []byte("modified1"), 0644); err != nil {
		t.Fatalf("Failed to modify test file1: %v", err)
	}

	// Wait for the command to execute
	select {
	case <-executed:
		// Good
	case <-time.After(1 * time.Second):
		t.Errorf("Command did not run after first file change")
	}

	// Modify second file
	if err := os.WriteFile(filePath2, []byte("modified2"), 0644); err != nil {
		t.Fatalf("Failed to modify test file2: %v", err)
	}

	// Wait for the command to execute
	select {
	case <-executed:
		// Good
	case <-time.After(1 * time.Second):
		t.Errorf("Command did not run after second file change")
	}

	// Cancel the context to stop the watcher
	cancel()

	// Wait for runWatch to return
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("runWatch returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("runWatch did not exit after context cancellation")
	}
}