package leet

import (
	"testing"
)

func TestLeetifyBasic(t *testing.T) {
	tests := []struct {
		input    string
		level    int
		expected string
	}{
		{"hello", 1, "h3ll0"},
		{"HELLO", 1, "H3LL0"},
		{"aeio", 1, "4310"},
		{"test", 2, "7357"},
		{"leet", 3, "1337"},
		{"The quick brown fox", 2, "7h3 qu1ck br0wn f0x"},
	}

	for _, tc := range tests {
		result := leetify(tc.input, tc.level, false)
		if result != tc.expected {
			t.Errorf("leetify(%q, %d, false) = %q, want %q", tc.input, tc.level, result, tc.expected)
		}
	}
}

func TestLeetifyPreservesNonAlpha(t *testing.T) {
	input := "hello, world! 123"
	result := leetify(input, 2, false)
	expected := "h3ll0, w0rld! 123"
	if result != expected {
		t.Errorf("leetify(%q) = %q, want %q", input, result, expected)
	}
}
