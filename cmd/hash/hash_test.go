package hash

import (
	"bytes"
	"strings"
	"testing"
)

func TestHashCommand(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		input    string
		algo     string
		expected string // partial match is enough (hash only)
	}{
		{
			name:     "md5",
			input:    "hello",
			algo:     "md5",
			expected: "5d41402abc4b2a76b9719d911017c592",
		},
		{
			name:     "sha1",
			input:    "hello",
			algo:     "sha1",
			expected: "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
		},
		{
			name:     "sha256",
			input:    "hello",
			algo:     "sha256",
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := &Params{
				Files: []string{"-"}, // Read from stdin
				Algo:  tc.algo,
			}
			stdin := strings.NewReader(tc.input)
			var stdout bytes.Buffer

			if err := runHash(params, &stdout, stdin); err != nil {
				t.Fatalf("runHash failed: %v", err)
			}

			output := stdout.String()
			if !strings.Contains(output, tc.expected) {
				t.Errorf("Expected output to contain hash %q, got %q", tc.expected, output)
			}
		})
	}
}

func TestHashInvalidAlgo(t *testing.T) {
	_, err := newHasher("invalid")
	if err == nil {
		t.Error("Expected error for invalid algorithm, got nil")
	}
}
