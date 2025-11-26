package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTailReader_Simple(t *testing.T) {
	input := "Line1\nLine2\nLine3\n"
	expected := "Line1\nLine2\nLine3\n"

	var stdout, stderr bytes.Buffer
	tailReader(strings.NewReader(input), &stdout, &stderr, 10)

	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestTailReader_Limit(t *testing.T) {
	input := "Line1\nLine2\nLine3\nLine4\n"
	expected := "Line3\nLine4\n"

	var stdout, stderr bytes.Buffer
	tailReader(strings.NewReader(input), &stdout, &stderr, 2)

	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestTailReader_ZeroLines(t *testing.T) {
	input := "Line1\nLine2\n"
	expected := ""

	var stdout, stderr bytes.Buffer
	tailReader(strings.NewReader(input), &stdout, &stderr, 0)

	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestRunTailStatic_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	content := "A\nB\nC\nD\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	params := &TailParams{
		Files: []string{file},
		Lines: 2,
	}

	var stdout, stderr bytes.Buffer
	runTailStatic(params, &stdout, &stderr, false)

	expected := "C\nD\n"
	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}

func TestRunTailStatic_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "1.txt")
	file2 := filepath.Join(tmpDir, "2.txt")

	if err := os.WriteFile(file1, []byte("A\nB\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("C\nD\n"), 0644); err != nil {
		t.Fatal(err)
	}

	params := &TailParams{
		Files: []string{file1, file2},
		Lines: 10,
	}

	var stdout, stderr bytes.Buffer
	// printHeaders = true for multiple files usually
	runTailStatic(params, &stdout, &stderr, true)

	// Note: We expect headers
	// implementation uses fmt.Fprintf(stdout, "==> %s <==\n", file)
	// and newline between files
	expectedSubstr1 := "==> " + file1 + " <==\nA\nB\n"
	expectedSubstr2 := "\n==> " + file2 + " <==\nC\nD\n"

	out := stdout.String()
	if !strings.Contains(out, expectedSubstr1) {
		t.Errorf("Output missing file1 content/header. Got: %q", out)
	}
	if !strings.Contains(out, expectedSubstr2) {
		t.Errorf("Output missing file2 content/header. Got: %q", out)
	}
}

func TestRunTailStatic_Quiet(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "1.txt")
	file2 := filepath.Join(tmpDir, "2.txt")

	if err := os.WriteFile(file1, []byte("A\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("B\n"), 0644); err != nil {
		t.Fatal(err)
	}

	params := &TailParams{
		Files: []string{file1, file2},
		Lines: 10,
		Quiet: true,
	}

	var stdout, stderr bytes.Buffer
	// Quiet means printHeaders passed to runTailStatic is false (logic in TailCmd)
	// But here we call runTailStatic directly. We must simulate what TailCmd passes.
	// logic: printHeaders := (len(params.Files) > 1 && !params.Quiet) || params.Verbose
	printHeaders := (len(params.Files) > 1 && !params.Quiet) || params.Verbose
	
	if printHeaders {
		t.Fatalf("Logic error in test setup: Quiet should force printHeaders false")
	}

	runTailStatic(params, &stdout, &stderr, printHeaders)

	expected := "A\nB\n"
	if stdout.String() != expected {
		t.Errorf("Expected %q, got %q", expected, stdout.String())
	}
}
