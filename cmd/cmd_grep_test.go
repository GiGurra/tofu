package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCompilePattern_Basic(t *testing.T) {
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
		IgnoreCase:  false,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	if !pattern.MatchString("test") {
		t.Errorf("Expected pattern to match 'test'")
	}
	if pattern.MatchString("TEST") {
		t.Errorf("Expected pattern to NOT match 'TEST' (case sensitive)")
	}
}

func TestCompilePattern_IgnoreCase(t *testing.T) {
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
		IgnoreCase:  true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	if !pattern.MatchString("test") {
		t.Errorf("Expected pattern to match 'test'")
	}
	if !pattern.MatchString("TEST") {
		t.Errorf("Expected pattern to match 'TEST' (case insensitive)")
	}
	if !pattern.MatchString("TeSt") {
		t.Errorf("Expected pattern to match 'TeSt' (case insensitive)")
	}
}

func TestCompilePattern_Fixed(t *testing.T) {
	params := &GrepParams{
		Pattern:     "test.txt",
		PatternType: PatternTypeFixed,
		IgnoreCase:  false,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	if !pattern.MatchString("test.txt") {
		t.Errorf("Expected pattern to match 'test.txt'")
	}
	// In fixed mode, . should be literal, not a wildcard
	if pattern.MatchString("testXtxt") {
		t.Errorf("Expected pattern to NOT match 'testXtxt' (dot should be literal)")
	}
}

func TestCompilePattern_WordRegexp(t *testing.T) {
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
		WordRegexp:  true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	if !pattern.MatchString("test") {
		t.Errorf("Expected pattern to match 'test'")
	}
	if !pattern.MatchString("a test here") {
		t.Errorf("Expected pattern to match 'a test here'")
	}
	if pattern.MatchString("testing") {
		t.Errorf("Expected pattern to NOT match 'testing' (word boundary)")
	}
	if pattern.MatchString("attest") {
		t.Errorf("Expected pattern to NOT match 'attest' (word boundary)")
	}
}

func TestCompilePattern_LineRegexp(t *testing.T) {
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
		LineRegexp:  true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	if !pattern.MatchString("test") {
		t.Errorf("Expected pattern to match 'test'")
	}
	if pattern.MatchString("test line") {
		t.Errorf("Expected pattern to NOT match 'test line' (whole line only)")
	}
	if pattern.MatchString("a test") {
		t.Errorf("Expected pattern to NOT match 'a test' (whole line only)")
	}
}

func TestConvertBasicToExtended(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"escaped plus", `test\+`, "test+"},
		{"escaped question", `test\?`, "test?"},
		{"escaped braces", `test\{1,3\}`, "test{1,3}"},
		{"escaped parens", `\(group\)`, "(group)"},
		{"mixed", `\(test\)\+`, "(test)+"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertBasicToExtended(tt.input)
			if result != tt.expected {
				t.Errorf("convertBasicToExtended(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGrepReader_SimpleMatch(t *testing.T) {
	input := "Hello\nWorld\ntest line\nAnother"
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepReader(strings.NewReader(input), "test.txt", pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("grepReader failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find match")
	}
	// Check for 'line' since 'test' will be wrapped in color codes
	if !strings.Contains(output, "line") {
		t.Errorf("Expected output to contain 'line', got %q", output)
	}
	// Verify color codes are present (highlighting is working)
	if !strings.Contains(output, colorRed) {
		t.Errorf("Expected output to contain color codes for highlighting")
	}
}

func TestGrepReader_InvertMatch(t *testing.T) {
	input := "Hello\nWorld\ntest line\nAnother"
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
		InvertMatch: true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepReader(strings.NewReader(input), "test.txt", pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("grepReader failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find match (inverted)")
	}
	if strings.Contains(output, "test line") {
		t.Errorf("Expected output to NOT contain 'test line', got %q", output)
	}
	if !strings.Contains(output, "Hello") {
		t.Errorf("Expected output to contain 'Hello', got %q", output)
	}
}

func TestGrepReader_Count(t *testing.T) {
	input := "test1\ntest2\nother\ntest3"
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
		Count:       true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepReader(strings.NewReader(input), "test.txt", pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if err != nil {
		t.Fatalf("grepReader failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find matches")
	}
	if output != "3" {
		t.Errorf("Expected count to be '3', got %q", output)
	}
}

func TestGrepReader_LineNumber(t *testing.T) {
	input := "line1\ntest line\nline3"
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
		LineNumber:  true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepReader(strings.NewReader(input), "test.txt", pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("grepReader failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find match")
	}
	if !strings.Contains(output, "2:") {
		t.Errorf("Expected output to contain line number '2:', got %q", output)
	}
}

func TestGrepReader_MaxCount(t *testing.T) {
	input := "test1\ntest2\ntest3\ntest4"
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
		MaxCount:    2,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepReader(strings.NewReader(input), "test.txt", pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("grepReader failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find matches")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines of output (MaxCount=2), got %d", len(lines))
	}
}

func TestGrepReader_Quiet(t *testing.T) {
	input := "test line\nother"
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
		Quiet:       true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepReader(strings.NewReader(input), "test.txt", pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("grepReader failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find match")
	}
	if len(output) > 0 {
		t.Errorf("Expected no output in quiet mode, got %q", output)
	}
}

func TestGrepFile_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test.txt")
	content := "line1\ntest line\nline3\nanother test\n"
	if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepFile(file1, pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("grepFile failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find match")
	}
	// Check for parts of the line that won't be highlighted
	if !strings.Contains(output, " line") {
		t.Errorf("Expected output to contain ' line', got %q", output)
	}
	if !strings.Contains(output, "another") {
		t.Errorf("Expected output to contain 'another', got %q", output)
	}
	// Verify we got 2 lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines of output, got %d", len(lines))
	}
}

func TestGrepFile_WithFilename(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test.txt")
	content := "test line\n"
	if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	params := &GrepParams{
		Pattern:      "test",
		PatternType:  PatternTypeExtended,
		WithFilename: true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepFile(file1, pattern, params, true)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("grepFile failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find match")
	}
	if !strings.Contains(output, file1) {
		t.Errorf("Expected output to contain filename %q, got %q", file1, output)
	}
}

func TestShouldSearchFile_Include(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		include  []string
		exclude  []string
		expected bool
	}{
		{"match include pattern", "test.txt", []string{"*.txt"}, nil, true},
		{"no match include pattern", "test.md", []string{"*.txt"}, nil, false},
		{"match exclude pattern", "test.txt", nil, []string{"*.txt"}, false},
		{"no match exclude pattern", "test.md", nil, []string{"*.txt"}, true},
		{"match include and exclude", "test.txt", []string{"*.txt"}, []string{"test.*"}, false},
		{"no patterns", "test.txt", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSearchFile(tt.filename, tt.include, tt.exclude)
			if result != tt.expected {
				t.Errorf("shouldSearchFile(%q, %v, %v) = %v, expected %v",
					tt.filename, tt.include, tt.exclude, result, tt.expected)
			}
		})
	}
}

func TestShouldExcludeDir(t *testing.T) {
	tests := []struct {
		name     string
		dirname  string
		patterns []string
		expected bool
	}{
		{"match pattern", "node_modules", []string{"node_modules"}, true},
		{"no match pattern", "src", []string{"node_modules"}, false},
		{"match wildcard", "test_dir", []string{"test_*"}, true},
		{"no patterns", "any_dir", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldExcludeDir(tt.dirname, tt.patterns)
			if result != tt.expected {
				t.Errorf("shouldExcludeDir(%q, %v) = %v, expected %v",
					tt.dirname, tt.patterns, result, tt.expected)
			}
		})
	}
}

func TestHighlightMatches(t *testing.T) {
	line := "this is a test line with test word"
	pattern := regexp.MustCompile("test")

	result := highlightMatches(line, pattern)

	// Check that the result contains color codes
	if !strings.Contains(result, colorRed) {
		t.Errorf("Expected highlighted output to contain red color code")
	}
	if !strings.Contains(result, colorReset) {
		t.Errorf("Expected highlighted output to contain color reset code")
	}
	if !strings.Contains(result, "test") {
		t.Errorf("Expected highlighted output to contain 'test'")
	}
}

func TestHighlightMatches_NoMatch(t *testing.T) {
	line := "this is a line"
	pattern := regexp.MustCompile("test")

	result := highlightMatches(line, pattern)

	if result != line {
		t.Errorf("Expected unchanged line when no match, got %q", result)
	}
}

func TestGrepFile_FileNotFound(t *testing.T) {
	params := &GrepParams{
		Pattern:     "test",
		PatternType: PatternTypeExtended,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	found, err := grepFile("/nonexistent/file.txt", pattern, params, false)

	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}
	if found {
		t.Errorf("Expected found=false for non-existent file")
	}
}

func TestIsFileBinary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a text file
	textFile := filepath.Join(tmpDir, "text.txt")
	if err := os.WriteFile(textFile, []byte("plain text content"), 0644); err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	file, err := os.Open(textFile)
	if err != nil {
		t.Fatalf("Failed to open text file: %v", err)
	}
	defer file.Close()

	isBinary, err := isFileBinary(file)
	if err != nil {
		t.Fatalf("isFileBinary failed: %v", err)
	}
	if isBinary {
		t.Errorf("Expected text file to not be detected as binary")
	}

	// Create a binary file (with NUL bytes)
	binaryFile := filepath.Join(tmpDir, "binary.bin")
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0x00}
	if err := os.WriteFile(binaryFile, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	file2, err := os.Open(binaryFile)
	if err != nil {
		t.Fatalf("Failed to open binary file: %v", err)
	}
	defer file2.Close()

	isBinary2, err := isFileBinary(file2)
	if err != nil {
		t.Fatalf("isFileBinary failed: %v", err)
	}
	if !isBinary2 {
		t.Errorf("Expected binary file to be detected as binary")
	}
}

func TestGrepReader_Context(t *testing.T) {
	input := "line1\nline2\ntest line\nline4\nline5"
	params := &GrepParams{
		Pattern:       "test",
		PatternType:   PatternTypeExtended,
		BeforeContext: 1,
		AfterContext:  1,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepReader(strings.NewReader(input), "test.txt", pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("grepReader failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find match")
	}
	if !strings.Contains(output, "line2") {
		t.Errorf("Expected output to contain 'line2' (before context), got %q", output)
	}
	// Check for part of the line that won't be colored
	if !strings.Contains(output, " line") {
		t.Errorf("Expected output to contain ' line', got %q", output)
	}
	if !strings.Contains(output, "line4") {
		t.Errorf("Expected output to contain 'line4' (after context), got %q", output)
	}
	// Verify we got 3 lines (before, match, after)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines of output (before, match, after), got %d", len(lines))
	}
}

func TestGrepReader_FilesWithMatch(t *testing.T) {
	input := "line1\ntest line\nline3"
	params := &GrepParams{
		Pattern:        "test",
		PatternType:    PatternTypeExtended,
		FilesWithMatch: true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepReader(strings.NewReader(input), "test.txt", pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if err != nil {
		t.Fatalf("grepReader failed: %v", err)
	}
	if !found {
		t.Errorf("Expected to find match")
	}
	if output != "test.txt" {
		t.Errorf("Expected output to be 'test.txt', got %q", output)
	}
}

func TestGrepReader_FilesWithoutMatch(t *testing.T) {
	input := "line1\nline2\nline3"
	params := &GrepParams{
		Pattern:           "test",
		PatternType:       PatternTypeExtended,
		FilesWithoutMatch: true,
	}

	pattern, err := compilePattern(params)
	if err != nil {
		t.Fatalf("compilePattern failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	found, err := grepReader(strings.NewReader(input), "test.txt", pattern, params, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if err != nil {
		t.Fatalf("grepReader failed: %v", err)
	}
	if found {
		t.Errorf("Expected to not find match")
	}
	if output != "test.txt" {
		t.Errorf("Expected output to be 'test.txt', got %q", output)
	}
}
