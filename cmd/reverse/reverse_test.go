package reverse

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReverseLines_Simple(t *testing.T) {
	input := "line1\nline2\nline3\n"
	expected := "line3\nline2\nline1\n"

	var stdout bytes.Buffer
	err := reverseLines(strings.NewReader(input), &stdout)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestReverseLines_SingleLine(t *testing.T) {
	input := "only one line\n"
	expected := "only one line\n"

	var stdout bytes.Buffer
	err := reverseLines(strings.NewReader(input), &stdout)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestReverseLines_Empty(t *testing.T) {
	input := ""
	expected := ""

	var stdout bytes.Buffer
	err := reverseLines(strings.NewReader(input), &stdout)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestReverseLines_NoTrailingNewline(t *testing.T) {
	input := "line1\nline2\nline3"
	expected := "line3\nline2\nline1\n"

	var stdout bytes.Buffer
	err := reverseLines(strings.NewReader(input), &stdout)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestRun_Stdin(t *testing.T) {
	input := "a\nb\nc\n"
	expected := "c\nb\na\n"

	params := &Params{
		Files: []string{"-"},
	}

	var stdout, stderr bytes.Buffer
	exitCode := Run(params, strings.NewReader(input), &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}
	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestRun_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	content := "first\nsecond\nthird\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	params := &Params{
		Files: []string{file},
	}

	var stdout, stderr bytes.Buffer
	exitCode := Run(params, nil, &stdout, &stderr)

	expected := "third\nsecond\nfirst\n"
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}
	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestRun_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "1.txt")
	file2 := filepath.Join(tmpDir, "2.txt")

	if err := os.WriteFile(file1, []byte("a\nb\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("c\nd\n"), 0644); err != nil {
		t.Fatal(err)
	}

	params := &Params{
		Files: []string{file1, file2},
	}

	var stdout, stderr bytes.Buffer
	exitCode := Run(params, nil, &stdout, &stderr)

	expected := "b\na\nd\nc\n"
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}
	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestRun_NonexistentFile(t *testing.T) {
	params := &Params{
		Files: []string{"/nonexistent/file.txt"},
	}

	var stdout, stderr bytes.Buffer
	exitCode := Run(params, nil, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "cannot open") {
		t.Errorf("Expected error message about opening file, got: %s", stderr.String())
	}
}
