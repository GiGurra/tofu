package base64

import (
	"bytes"
	"strings"
	"testing"
)

func TestBase64Command(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		params   Params
		expected string
		wantErr  bool
	}{
		{
			name:     "Standard Encode",
			input:    "hello",
			params:   Params{},
			expected: "aGVsbG8=\n",
		},
		{
			name:     "Standard Decode",
			input:    "aGVsbG8=\n",
			params:   Params{Decode: true},
			expected: "hello",
		},
		{
			name:     "URL Safe Encode",
			input:    "hello world?",
			params:   Params{UrlSafe: true},
			expected: "aGVsbG8gd29ybGQ_\n",
		},
		{
			name:     "URL Safe Decode",
			input:    "aGVsbG8gd29ybGQ_\n",
			params:   Params{UrlSafe: true, Decode: true},
			expected: "hello world?",
		},
		{
			name:     "No Padding Encode",
			input:    "hello",
			params:   Params{NoPadding: true},
			expected: "aGVsbG8\n",
		},
		{
			name:     "Custom Alphabet",
			input:    "hello",
			params:   Params{Alphabet: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"},
			expected: "aGVsbG8=\n", // Same as std but with different chars if they differed
		},
		{
			name:    "Invalid Alphabet Length",
			input:   "hello",
			params:  Params{Alphabet: "short"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := strings.NewReader(tt.input)
			var stdout bytes.Buffer

			err := runBase64(&tt.params, &stdout, stdin)

			if (err != nil) != tt.wantErr {
				t.Errorf("runBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got := stdout.String(); got != tt.expected {
					t.Errorf("runBase64() = %q, want %q", got, tt.expected)
				}
			}
		})
	}
}
