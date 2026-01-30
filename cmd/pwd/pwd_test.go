package pwd

import (
	"os"
	"strings"
	"testing"
)

func TestPwd(t *testing.T) {
	// Create temp files for capturing output
	stdout, err := os.CreateTemp("", "pwd_stdout")
	if err != nil {
		t.Fatalf("failed to create temp stdout: %v", err)
	}
	defer os.Remove(stdout.Name())
	defer stdout.Close()

	stderr, err := os.CreateTemp("", "pwd_stderr")
	if err != nil {
		t.Fatalf("failed to create temp stderr: %v", err)
	}
	defer os.Remove(stderr.Name())
	defer stderr.Close()

	// Test default (logical) mode
	params := &Params{}
	exitCode := Run(params, stdout, stderr)
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Read output
	stdout.Seek(0, 0)
	buf := make([]byte, 1024)
	n, _ := stdout.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))

	if output == "" {
		t.Error("expected non-empty output")
	}

	// Verify it's a valid path
	if _, err := os.Stat(output); err != nil {
		t.Errorf("output is not a valid path: %s", output)
	}
}

func TestPwdPhysical(t *testing.T) {
	stdout, err := os.CreateTemp("", "pwd_stdout")
	if err != nil {
		t.Fatalf("failed to create temp stdout: %v", err)
	}
	defer os.Remove(stdout.Name())
	defer stdout.Close()

	stderr, err := os.CreateTemp("", "pwd_stderr")
	if err != nil {
		t.Fatalf("failed to create temp stderr: %v", err)
	}
	defer os.Remove(stderr.Name())
	defer stderr.Close()

	params := &Params{Physical: true}
	exitCode := Run(params, stdout, stderr)
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	stdout.Seek(0, 0)
	buf := make([]byte, 1024)
	n, _ := stdout.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))

	if output == "" {
		t.Error("expected non-empty output")
	}

	// Physical path should not contain symlinks
	if _, err := os.Stat(output); err != nil {
		t.Errorf("output is not a valid path: %s", output)
	}
}
