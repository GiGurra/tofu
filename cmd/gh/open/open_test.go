package open

import "testing"

func TestGitRemoteToHTTPS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SSH short format with .git",
			input:    "git@github.com:owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "SSH short format without .git",
			input:    "git@github.com:owner/repo",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "SSH full format with .git",
			input:    "ssh://git@github.com/owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "SSH full format without .git",
			input:    "ssh://git@github.com/owner/repo",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "HTTPS with .git",
			input:    "https://github.com/owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "HTTPS without .git",
			input:    "https://github.com/owner/repo",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "git protocol",
			input:    "git://github.com/owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "GitLab SSH",
			input:    "git@gitlab.com:owner/repo.git",
			expected: "https://gitlab.com/owner/repo",
		},
		{
			name:     "Bitbucket SSH",
			input:    "git@bitbucket.org:owner/repo.git",
			expected: "https://bitbucket.org/owner/repo",
		},
		{
			name:     "Self-hosted GitLab SSH",
			input:    "git@git.example.com:group/project.git",
			expected: "https://git.example.com/group/project",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := gitRemoteToHTTPS(tc.input)
			if result != tc.expected {
				t.Errorf("gitRemoteToHTTPS(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://github.com/owner/repo", true},
		{"http://example.com", true},
		{"github.com/owner/repo", false},
		{"./some/path", false},
		{"/absolute/path", false},
		{".", false},
		{"git@github.com:owner/repo.git", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := isURL(tc.input)
			if result != tc.expected {
				t.Errorf("isURL(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestLooksLikeRepoURL(t *testing.T) {
	// Note: This function is only called after checking if path exists locally.
	// So we only need to distinguish between URL-like strings and non-URL strings.
	tests := []struct {
		input    string
		expected bool
	}{
		{"github.com/owner/repo", true},
		{"gitlab.com/owner/repo", true},
		{"bitbucket.org/owner/repo", true},
		{"git.example.com/group/project", true},
		{"github.com/owner/repo/tree/main", true},
		{"github.com/owner", true}, // valid URL pattern (user profile page)
		{"github.com", false},      // just domain, no path
		{"./some/path", false},     // starts with dot
		{".hidden/path", false},    // starts with dot
		{".", false},               // current dir
		{"..", false},              // parent dir
		{"some-dir", false},        // simple name, no dot in domain
		{"owner/repo", false},      // no domain (no dot)
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := looksLikeRepoURL(tc.input)
			if result != tc.expected {
				t.Errorf("looksLikeRepoURL(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}
