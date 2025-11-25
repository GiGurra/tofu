package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeDiff(t *testing.T) {
	tests := []struct {
		name          string
		lines1        []string
		lines2        []string
		wantInserts   int
		wantDeletes   int
		wantEqual     int
	}{
		{
			name:        "identical",
			lines1:      []string{"a", "b", "c"},
			lines2:      []string{"a", "b", "c"},
			wantInserts: 0,
			wantDeletes: 0,
			wantEqual:   3,
		},
		{
			name:        "all different",
			lines1:      []string{"a", "b"},
			lines2:      []string{"x", "y"},
			wantInserts: 2,
			wantDeletes: 2,
			wantEqual:   0,
		},
		{
			name:        "insertion",
			lines1:      []string{"a", "c"},
			lines2:      []string{"a", "b", "c"},
			wantInserts: 1,
			wantDeletes: 0,
			wantEqual:   2,
		},
		{
			name:        "deletion",
			lines1:      []string{"a", "b", "c"},
			lines2:      []string{"a", "c"},
			wantInserts: 0,
			wantDeletes: 1,
			wantEqual:   2,
		},
		{
			name:        "modification",
			lines1:      []string{"a", "b", "c"},
			lines2:      []string{"a", "x", "c"},
			wantInserts: 1,
			wantDeletes: 1,
			wantEqual:   2,
		},
		{
			name:        "empty first",
			lines1:      []string{},
			lines2:      []string{"a", "b"},
			wantInserts: 2,
			wantDeletes: 0,
			wantEqual:   0,
		},
		{
			name:        "empty second",
			lines1:      []string{"a", "b"},
			lines2:      []string{},
			wantInserts: 0,
			wantDeletes: 2,
			wantEqual:   0,
		},
		{
			name:        "both empty",
			lines1:      []string{},
			lines2:      []string{},
			wantInserts: 0,
			wantDeletes: 0,
			wantEqual:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := computeDiff(tt.lines1, tt.lines2)

			inserts, deletes, equal := 0, 0, 0
			for _, d := range diff {
				switch d.Op {
				case DiffInsert:
					inserts++
				case DiffDelete:
					deletes++
				case DiffEqual:
					equal++
				}
			}

			if inserts != tt.wantInserts {
				t.Errorf("insertions = %d, want %d", inserts, tt.wantInserts)
			}
			if deletes != tt.wantDeletes {
				t.Errorf("deletions = %d, want %d", deletes, tt.wantDeletes)
			}
			if equal != tt.wantEqual {
				t.Errorf("equal = %d, want %d", equal, tt.wantEqual)
			}
		})
	}
}

func TestToLowerLines(t *testing.T) {
	input := []string{"Hello", "WORLD", "MiXeD"}
	expected := []string{"hello", "world", "mixed"}
	result := toLowerLines(input)

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("toLowerLines[%d] = %q, want %q", i, result[i], exp)
		}
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	input := []string{"  hello   world  ", "\ttab\tseparated", "normal"}
	expected := []string{"hello world", "tab separated", "normal"}
	result := normalizeWhitespace(input)

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("normalizeWhitespace[%d] = %q, want %q", i, result[i], exp)
		}
	}
}

func TestRemoveBlankLines(t *testing.T) {
	input := []string{"hello", "", "world", "   ", "test"}
	expected := []string{"hello", "world", "test"}
	result := removeBlankLines(input)

	if len(result) != len(expected) {
		t.Fatalf("removeBlankLines length = %d, want %d", len(result), len(expected))
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("removeBlankLines[%d] = %q, want %q", i, result[i], exp)
		}
	}
}

func TestGroupIntoHunks(t *testing.T) {
	// Create a diff with changes at positions 2 and 7
	diff := []DiffLine{
		{Op: DiffEqual, Line: "line0"},
		{Op: DiffEqual, Line: "line1"},
		{Op: DiffDelete, Line: "line2"}, // change at 2
		{Op: DiffEqual, Line: "line3"},
		{Op: DiffEqual, Line: "line4"},
		{Op: DiffEqual, Line: "line5"},
		{Op: DiffEqual, Line: "line6"},
		{Op: DiffInsert, Line: "line7"}, // change at 7
		{Op: DiffEqual, Line: "line8"},
		{Op: DiffEqual, Line: "line9"},
	}

	// With context 1, should create 2 separate hunks
	hunks := groupIntoHunks(diff, 1)

	if len(hunks) != 2 {
		t.Errorf("got %d hunks, want 2", len(hunks))
	}
}

func TestReadFileLines(t *testing.T) {
	// Create a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "line1\nline2\nline3\n"
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	lines, err := readFileLines(path)
	if err != nil {
		t.Fatalf("readFileLines failed: %v", err)
	}

	if len(lines) != 3 {
		t.Errorf("got %d lines, want 3", len(lines))
	}

	expected := []string{"line1", "line2", "line3"}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], exp)
		}
	}
}

func TestReadFileLines_NotFound(t *testing.T) {
	_, err := readFileLines("/nonexistent/file/path")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestDiffCmd(t *testing.T) {
	cmd := DiffCmd()
	if cmd == nil {
		t.Error("DiffCmd returned nil")
	}
	if cmd.Name() != "diff" {
		t.Errorf("expected Name()='diff', got '%s'", cmd.Name())
	}
}

func TestRunDiff_IdenticalFiles(t *testing.T) {
	dir := t.TempDir()

	// Create two identical files
	content := "line1\nline2\nline3\n"
	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")

	os.WriteFile(file1, []byte(content), 0644)
	os.WriteFile(file2, []byte(content), 0644)

	params := &DiffParams{
		File1:   file1,
		File2:   file2,
		Unified: 3,
		Brief:   true,
	}

	// Should not produce output for identical files in brief mode
	err := runDiff(params)
	if err != nil {
		t.Errorf("runDiff failed: %v", err)
	}
}

func TestRunDiff_DifferentFiles(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")

	os.WriteFile(file1, []byte("line1\nline2\n"), 0644)
	os.WriteFile(file2, []byte("line1\nmodified\n"), 0644)

	params := &DiffParams{
		File1:   file1,
		File2:   file2,
		Unified: 3,
		NoColor: true,
	}

	err := runDiff(params)
	if err != nil {
		t.Errorf("runDiff failed: %v", err)
	}
}

func TestShouldUseColor(t *testing.T) {
	tests := []struct {
		name     string
		noColor  bool
		color    string
		wantAuto bool // true if we expect auto behavior
	}{
		{"no-color flag", true, "auto", false},
		{"color always", false, "always", false},
		{"color never", false, "never", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &DiffParams{
				NoColor: tt.noColor,
				Color:   tt.color,
			}

			result := shouldUseColor(params)

			if tt.noColor && result {
				t.Error("expected no color when NoColor is true")
			}
			if tt.color == "always" && !tt.noColor && !result {
				t.Error("expected color when Color is 'always'")
			}
			if tt.color == "never" && result {
				t.Error("expected no color when Color is 'never'")
			}
		})
	}
}
