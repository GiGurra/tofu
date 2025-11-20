package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWhichCmd(t *testing.T) {
	// Create a temporary directory to act as our PATH
	tempDir := t.TempDir()

	// Create a dummy executable
	exeName := "mytestexe"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}
	exePath := filepath.Join(tempDir, exeName)
	
	// Create the file and make it executable
	f, err := os.Create(exePath)
	if err != nil {
		t.Fatalf("Failed to create executable: %v", err)
	}
	f.Close()
	if err := os.Chmod(exePath, 0755); err != nil {
		t.Fatalf("Failed to chmod executable: %v", err)
	}

	// Add tempDir to PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	
	newPath := tempDir + string(os.PathListSeparator) + oldPath
	os.Setenv("PATH", newPath)

	tests := []struct {
		name         string
		programs     []string
		wantExitCode int
		wantStdout   string // simple check if it contains path
		wantStderr   string
	}{
		{
			name:         "Find existing executable",
			programs:     []string{exeName},
			wantExitCode: 0,
			wantStdout:   exePath,
			wantStderr:   "",
		},
		{
			name:         "Find non-existing executable",
			programs:     []string{"nonexistentcmd_12345"},
			wantExitCode: 1,
			wantStdout:   "",
			wantStderr:   "not found",
		},
		{
			name:         "Find mixed existing and non-existing",
			programs:     []string{exeName, "nonexistentcmd_12345"},
			wantExitCode: 1,
			wantStdout:   exePath,
			wantStderr:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			params := &WhichParams{
				Programs: tt.programs,
			}
			
			exitCode := runWhich(params, &stdout, &stderr)
			
			if exitCode != tt.wantExitCode {
				t.Errorf("runWhich() exitCode = %v, want %v", exitCode, tt.wantExitCode)
			}

			if tt.wantStdout != "" {
				if !strings.Contains(stdout.String(), tt.wantStdout) {
					t.Errorf("runWhich() stdout = %v, want substring %v", stdout.String(), tt.wantStdout)
				}
			}

			if tt.wantStderr != "" {
				if !strings.Contains(stderr.String(), tt.wantStderr) {
					t.Errorf("runWhich() stderr = %v, want substring %v", stderr.String(), tt.wantStderr)
				}
			}
		})
	}
}
