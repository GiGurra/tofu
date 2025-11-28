package cat

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCatReader_Simple(t *testing.T) {
	input := "Hello\nWorld\n"
	expected := "Hello\nWorld\n"

	var stdout bytes.Buffer
	params := &Params{}
	lineNum := 0

	err := catReader(strings.NewReader(input), &stdout, params, &lineNum)
	if err != nil {
		t.Fatalf("catReader failed: %v", err)
	}

	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestCatReader_Number(t *testing.T) {
	input := "One\nTwo\n"
	expected := "     1\tOne\n     2\tTwo\n"

	var stdout bytes.Buffer
	params := &Params{Number: true}
	lineNum := 0

	err := catReader(strings.NewReader(input), &stdout, params, &lineNum)
	if err != nil {
		t.Fatalf("catReader failed: %v", err)
	}

	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestCatReader_NumberNonBlank(t *testing.T) {
	input := "One\n\nTwo\n"
	// Standard cat -b numbers non-blank lines.
	// Our implementation prints "      \t" for empty lines when -b is set.
	expected := "     1\tOne\n      \t\n     2\tTwo\n"

	var stdout bytes.Buffer
	params := &Params{NumberNonblank: true}
	lineNum := 0

	err := catReader(strings.NewReader(input), &stdout, params, &lineNum)
	if err != nil {
		t.Fatalf("catReader failed: %v", err)
	}

	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestCatReader_ShowEnds(t *testing.T) {
	input := "Line"
	expected := "Line$\n"

	var stdout bytes.Buffer
	params := &Params{ShowEnds: true}
	lineNum := 0

	err := catReader(strings.NewReader(input), &stdout, params, &lineNum)
	if err != nil {
		t.Fatalf("catReader failed: %v", err)
	}

	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestCatReader_SqueezeBlank(t *testing.T) {
	input := "One\n\n\nTwo\n"
	expected := "One\n\nTwo\n"

	var stdout bytes.Buffer
	params := &Params{SqueezeBlank: true}
	lineNum := 0

	err := catReader(strings.NewReader(input), &stdout, params, &lineNum)
	if err != nil {
		t.Fatalf("catReader failed: %v", err)
	}

	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestRunCat_Integration(t *testing.T) {
	// t.TempDir() automatically creates a temporary directory that is
	// deleted (along with its contents) when the test finishes.
	// This is safer and cleaner than manual defer os.RemoveAll(...).
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("Content1\n"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("Content2\n"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	params := &Params{
		Files: []string{file1, file2},
	}

	var stdout, stderr bytes.Buffer
	exitCode := Run(params, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	expected := "Content1\nContent2\n"
	if stdout.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, stdout.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("Expected no stderr output, got %q", stderr.String())
	}
}

func TestRunCat_FileNotFound(t *testing.T) {
	params := &Params{
		Files: []string{"non_existent_file.txt"},
	}

	var stdout, stderr bytes.Buffer
	exitCode := Run(params, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() > 0 {
		t.Errorf("Expected no stdout output, got %q", stdout.String())
	}
	expectedErrorSubstr := "no such file or directory"
	if runtime.GOOS == "windows" {
		expectedErrorSubstr = "The system cannot find the file specified."
	}

	if !strings.Contains(stderr.String(), expectedErrorSubstr) {
		t.Errorf("Expected stderr to contain %q, got %q", expectedErrorSubstr, stderr.String())
	}
}
