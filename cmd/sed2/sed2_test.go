package sed2

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestProcessReader_LiteralReplace(t *testing.T) {
	input := "Hello World\nWorld of Go\nGoodbye"
	params := &Params{
		From:       "World",
		To:         "Universe",
		SearchType: PatternTypeLiteral,
		Global:     false,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := output.String()
	expectedLines := []string{
		"Hello Universe",
		"Universe of Go", // Only first occurrence on line
		"Goodbye",
	}
	expected := strings.Join(expectedLines, "\n") + "\n"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessReader_LiteralReplaceGlobal(t *testing.T) {
	input := "test test test\nother test"
	params := &Params{
		From:       "test",
		To:         "TEST",
		SearchType: PatternTypeLiteral,
		Global:     true,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := output.String()
	expectedLines := []string{
		"TEST TEST TEST",
		"other TEST",
	}
	expected := strings.Join(expectedLines, "\n") + "\n"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessReader_RegexReplace(t *testing.T) {
	input := "test123\ntest456\nother"
	params := &Params{
		From:       `test\d+`,
		To:         "NUM",
		SearchType: PatternTypeRegex,
		Global:     false,
	}

	pattern, err := regexp.Compile(params.From)
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := output.String()
	expectedLines := []string{
		"NUM",
		"NUM",
		"other",
	}
	expected := strings.Join(expectedLines, "\n") + "\n"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessReader_RegexCaptureGroups(t *testing.T) {
	input := "John Doe\nJane Smith"
	params := &Params{
		From:       `(\w+) (\w+)`,
		To:         "$2, $1",
		SearchType: PatternTypeRegex,
		Global:     false,
	}

	pattern, err := regexp.Compile(params.From)
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := output.String()
	expectedLines := []string{
		"Doe, John",
		"Smith, Jane",
	}
	expected := strings.Join(expectedLines, "\n") + "\n"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessReader_IgnoreCase(t *testing.T) {
	input := "Hello WORLD\nworld hello"
	params := &Params{
		From:       "world",
		To:         "UNIVERSE",
		SearchType: PatternTypeLiteral,
		IgnoreCase: true,
		Global:     false,
	}

	pattern, err := regexp.Compile("(?i)" + regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := output.String()
	expectedLines := []string{
		"Hello UNIVERSE",
		"UNIVERSE hello",
	}
	expected := strings.Join(expectedLines, "\n") + "\n"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessReader_NoMatch(t *testing.T) {
	input := "Hello World\nGoodbye"
	params := &Params{
		From:       "test",
		To:         "TEST",
		SearchType: PatternTypeLiteral,
		Global:     false,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := output.String()
	expected := input + "\n" // Lines should be unchanged with newlines added

	// Split and compare line by line to handle newline differences
	resultLines := strings.Split(strings.TrimSpace(result), "\n")
	expectedLines := strings.Split(strings.TrimSpace(expected), "\n")

	if len(resultLines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(resultLines))
	}

	for i := range expectedLines {
		if i < len(resultLines) && resultLines[i] != expectedLines[i] {
			t.Errorf("Line %d: expected %q, got %q", i, expectedLines[i], resultLines[i])
		}
	}
}

func TestReplaceFirst(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		pattern     string
		replacement string
		expected    string
	}{
		{
			name:        "single occurrence",
			line:        "hello world",
			pattern:     "world",
			replacement: "universe",
			expected:    "hello universe",
		},
		{
			name:        "multiple occurrences",
			line:        "test test test",
			pattern:     "test",
			replacement: "TEST",
			expected:    "TEST test test",
		},
		{
			name:        "no match",
			line:        "hello world",
			pattern:     "foo",
			replacement: "bar",
			expected:    "hello world",
		},
		{
			name:        "regex pattern",
			line:        "hello123world",
			pattern:     `\d+`,
			replacement: "XXX",
			expected:    "helloXXXworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := regexp.MustCompile(tt.pattern)
			result := ReplaceFirst(tt.line, pattern, tt.replacement)
			if result != tt.expected {
				t.Errorf("ReplaceFirst() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestReplaceFirst_CaptureGroups(t *testing.T) {
	line := "John Doe is a person"
	pattern := regexp.MustCompile(`(\w+) (\w+)`)
	replacement := "$2, $1"

	result := ReplaceFirst(line, pattern, replacement)
	expected := "Doe, John is a person"

	if result != expected {
		t.Errorf("ReplaceFirst() = %q, expected %q", result, expected)
	}
}

func TestProcessFile_ReadOnly(t *testing.T) {
	tmpDir := t.TempDir()

	file := filepath.Join(tmpDir, "test.txt")
	content := "Hello World\nWorld of Go"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &Params{
		From:       "World",
		To:         "Universe",
		SearchType: PatternTypeLiteral,
		Global:     false,
		InPlace:    false,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = ProcessFile(file, pattern, params)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	if !strings.Contains(output, "Universe") {
		t.Errorf("Expected output to contain 'Universe', got %q", output)
	}

	// Verify original file is unchanged
	fileContent, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(fileContent) != content {
		t.Errorf("File should be unchanged, got %q", string(fileContent))
	}
}

func TestProcessFile_InPlace(t *testing.T) {
	tmpDir := t.TempDir()

	file := filepath.Join(tmpDir, "test.txt")
	content := "Hello World\nWorld of Go"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &Params{
		From:       "World",
		To:         "Universe",
		SearchType: PatternTypeLiteral,
		Global:     false,
		InPlace:    true,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	err = ProcessFile(file, pattern, params)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	// Verify file was modified
	fileContent, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	result := string(fileContent)
	if !strings.Contains(result, "Universe") {
		t.Errorf("Expected file to contain 'Universe', got %q", result)
	}
	if strings.Contains(result, "Hello World") {
		t.Errorf("Expected 'Hello World' to be replaced, got %q", result)
	}
}

func TestProcessFile_FileNotFound(t *testing.T) {
	params := &Params{
		From:       "test",
		To:         "TEST",
		SearchType: PatternTypeLiteral,
		Global:     false,
		InPlace:    false,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	err = ProcessFile("/nonexistent/file.txt", pattern, params)
	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}
}

func TestRun_InvalidPattern(t *testing.T) {
	params := &Params{
		From:       "[invalid",
		To:         "test",
		SearchType: PatternTypeRegex,
		Global:     false,
		Files:      []string{"-"},
	}

	// Redirect stderr to capture error message
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := Run(params)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 for invalid pattern, got %d", exitCode)
	}
	if !strings.Contains(stderrOutput, "invalid pattern") {
		t.Errorf("Expected stderr to contain 'invalid pattern', got %q", stderrOutput)
	}
}

func TestProcessReader_EmptyLine(t *testing.T) {
	input := "line1\n\nline3"
	params := &Params{
		From:       "line",
		To:         "LINE",
		SearchType: PatternTypeLiteral,
		Global:     false,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := output.String()
	lines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	if lines[1] != "" {
		t.Errorf("Expected empty line at index 1, got %q", lines[1])
	}
}

func TestProcessReader_SpecialCharacters(t *testing.T) {
	input := "test.txt\ntest*txt"
	params := &Params{
		From:       "test.txt",
		To:         "result",
		SearchType: PatternTypeLiteral,
		Global:     false,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := output.String()
	lines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")

	if lines[0] != "result" {
		t.Errorf("Expected first line to be 'result', got %q", lines[0])
	}
	// Second line should not match because . is literal in literal mode
	if lines[1] != "test*txt" {
		t.Errorf("Expected second line to be unchanged, got %q", lines[1])
	}
}

func TestProcessReader_MultipleGlobalReplacements(t *testing.T) {
	input := "a a a a a"
	params := &Params{
		From:       "a",
		To:         "b",
		SearchType: PatternTypeLiteral,
		Global:     true,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := strings.TrimSpace(output.String())
	expected := "b b b b b"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessFile_InPlace_MultipleLines(t *testing.T) {
	tmpDir := t.TempDir()

	file := filepath.Join(tmpDir, "test.txt")
	content := "line1 test\nline2 test\nline3 test test"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &Params{
		From:       "test",
		To:         "PASS",
		SearchType: PatternTypeLiteral,
		Global:     true,
		InPlace:    true,
	}

	pattern, err := regexp.Compile(regexp.QuoteMeta(params.From))
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	err = ProcessFile(file, pattern, params)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	// Verify file was modified
	fileContent, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	result := string(fileContent)
	expectedLines := []string{
		"line1 PASS",
		"line2 PASS",
		"line3 PASS PASS",
	}
	expected := strings.Join(expectedLines, "\n") + "\n"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessReader_RegexIgnoreCase(t *testing.T) {
	input := "Hello WORLD\nworld hello"
	params := &Params{
		From:       "w[oO]rld",
		To:         "UNIVERSE",
		SearchType: PatternTypeRegex,
		IgnoreCase: true,
		Global:     false,
	}

	pattern, err := regexp.Compile("(?i)" + params.From)
	if err != nil {
		t.Fatalf("Failed to compile pattern: %v", err)
	}

	var output bytes.Buffer
	err = ProcessReader(strings.NewReader(input), &output, pattern, params)
	if err != nil {
		t.Fatalf("ProcessReader failed: %v", err)
	}

	result := output.String()
	expectedLines := []string{
		"Hello UNIVERSE",
		"UNIVERSE hello",
	}
	expected := strings.Join(expectedLines, "\n") + "\n"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
