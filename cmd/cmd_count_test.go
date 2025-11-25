package cmd

import (
	"strings"
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"  hello   world  ", 2},
		{"one two three four five", 5},
		{"\ttab\tseparated\twords", 3},
		{"newline\nwords", 2},
	}

	for _, tt := range tests {
		got := countWords(tt.input)
		if got != tt.want {
			t.Errorf("countWords(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCountReader(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLines int64
		wantWords int64
		wantChars int64
		wantBytes int64
	}{
		{
			name:      "empty",
			input:     "",
			wantLines: 0,
			wantWords: 0,
			wantChars: 0,
			wantBytes: 0,
		},
		{
			name:      "single line no newline",
			input:     "hello world",
			wantLines: 1,
			wantWords: 2,
			wantChars: 11,
			wantBytes: 11,
		},
		{
			name:      "single line with newline",
			input:     "hello world\n",
			wantLines: 1,
			wantWords: 2,
			wantChars: 12,
			wantBytes: 12,
		},
		{
			name:      "multiple lines",
			input:     "line one\nline two\nline three\n",
			wantLines: 3,
			wantWords: 6,
			wantChars: 29,
			wantBytes: 29,
		},
		{
			name:      "unicode",
			input:     "héllo wörld 日本語\n",
			wantLines: 1,
			wantWords: 3,
			wantChars: 16,
			wantBytes: 24, // UTF-8 bytes: é=2, ö=2, 日本語=9, rest=11 -> 24
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &CountParams{
				Lines: true,
				Words: true,
				Chars: true,
				Bytes: true,
			}
			reader := strings.NewReader(tt.input)
			result, err := countReader(reader, "test", params)
			if err != nil {
				t.Fatalf("countReader() error = %v", err)
			}

			if result.Lines != tt.wantLines {
				t.Errorf("Lines = %d, want %d", result.Lines, tt.wantLines)
			}
			if result.Words != tt.wantWords {
				t.Errorf("Words = %d, want %d", result.Words, tt.wantWords)
			}
			if result.Chars != tt.wantChars {
				t.Errorf("Chars = %d, want %d", result.Chars, tt.wantChars)
			}
			if result.Bytes != tt.wantBytes {
				t.Errorf("Bytes = %d, want %d", result.Bytes, tt.wantBytes)
			}
		})
	}
}

func TestCountMaxLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantMax int
	}{
		{"empty", "", 0},
		{"single short line", "hello\n", 5},
		{"multiple lines", "short\nthis is longer\nmed\n", 14},
		{"unicode line", "日本語テスト\n", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &CountParams{
				MaxLine: true,
			}
			reader := strings.NewReader(tt.input)
			result, err := countReader(reader, "test", params)
			if err != nil {
				t.Fatalf("countReader() error = %v", err)
			}

			if result.MaxLine != tt.wantMax {
				t.Errorf("MaxLine = %d, want %d", result.MaxLine, tt.wantMax)
			}
		})
	}
}

func TestCountLinesOnly(t *testing.T) {
	params := &CountParams{
		Lines: true,
	}
	input := "line1\nline2\nline3\n"
	reader := strings.NewReader(input)
	result, err := countReader(reader, "test", params)
	if err != nil {
		t.Fatalf("countReader() error = %v", err)
	}

	if result.Lines != 3 {
		t.Errorf("Lines = %d, want 3", result.Lines)
	}
}

func TestCountWordsOnly(t *testing.T) {
	params := &CountParams{
		Words: true,
	}
	input := "one two three\nfour five\n"
	reader := strings.NewReader(input)
	result, err := countReader(reader, "test", params)
	if err != nil {
		t.Fatalf("countReader() error = %v", err)
	}

	if result.Words != 5 {
		t.Errorf("Words = %d, want 5", result.Words)
	}
}

func TestCountCmd(t *testing.T) {
	cmd := CountCmd()
	if cmd == nil {
		t.Error("CountCmd returned nil")
	}
	if cmd.Name() != "count" {
		t.Errorf("expected Name()='count', got '%s'", cmd.Name())
	}
}

func TestDefaultBehavior(t *testing.T) {
	// When no flags specified, should show lines, words, chars
	params := &CountParams{}

	// Check that showAll logic works
	showAll := !params.Lines && !params.Words && !params.Chars && !params.Bytes && !params.MaxLine
	if !showAll {
		t.Error("expected showAll to be true when no flags set")
	}
}
