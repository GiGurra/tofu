package git

import (
	"testing"
)

func TestProjectDirToPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-home-gigur-git-tofu", "/home/gigur/git/tofu"},
		{"-Users-johkjo-git-personal-tofu", "/Users/johkjo/git/personal/tofu"},
		{"home-gigur-git-tofu", "/home/gigur/git/tofu"}, // without leading dash
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ProjectDirToPath(tt.input)
			if result != tt.expected {
				t.Errorf("ProjectDirToPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPathToProjectDir(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/home/gigur/git/tofu", "-home-gigur-git-tofu"},
		{"/Users/johkjo/git/personal/tofu", "-Users-johkjo-git-personal-tofu"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := PathToProjectDir(tt.input)
			if result != tt.expected {
				t.Errorf("PathToProjectDir(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCanonicalizeProjectDir_HomesOnly(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already canonical",
			input:    "-home-gigur-git-tofu",
			expected: "-home-gigur-git-tofu",
		},
		{
			name:     "mac to linux",
			input:    "-Users-johkjo-git-tofu",
			expected: "-home-gigur-git-tofu",
		},
		{
			name:     "unrelated path unchanged",
			input:    "-var-log-something",
			expected: "-var-log-something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.CanonicalizeProjectDir(tt.input)
			if result != tt.expected {
				t.Errorf("CanonicalizeProjectDir(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCanonicalizeProjectDir_DirsOnly(t *testing.T) {
	config := &SyncConfig{
		Dirs: [][]string{
			{"/home/gigur/git", "/Users/johkjo/git/personal"},
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already canonical",
			input:    "-home-gigur-git-tofu",
			expected: "-home-gigur-git-tofu",
		},
		{
			name:     "mac personal to linux git",
			input:    "-Users-johkjo-git-personal-tofu",
			expected: "-home-gigur-git-tofu",
		},
		{
			name:     "unrelated mac path unchanged",
			input:    "-Users-johkjo-Documents-notes",
			expected: "-Users-johkjo-Documents-notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.CanonicalizeProjectDir(tt.input)
			if result != tt.expected {
				t.Errorf("CanonicalizeProjectDir(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCanonicalizeProjectDir_DirsAndHomes(t *testing.T) {
	// Real-world scenario: dirs mapping first, then homes
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs: [][]string{
			{"/home/gigur/git", "/Users/johkjo/git/personal"},
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already canonical",
			input:    "-home-gigur-git-tofu",
			expected: "-home-gigur-git-tofu",
		},
		{
			name:     "mac personal git to linux git",
			input:    "-Users-johkjo-git-personal-tofu",
			expected: "-home-gigur-git-tofu",
		},
		{
			name:     "mac home other dir uses homes mapping",
			input:    "-Users-johkjo-Documents-notes",
			expected: "-home-gigur-Documents-notes",
		},
		{
			name:     "mac non-personal git uses homes mapping only",
			input:    "-Users-johkjo-git-work-project",
			expected: "-home-gigur-git-work-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.CanonicalizeProjectDir(tt.input)
			if result != tt.expected {
				t.Errorf("CanonicalizeProjectDir(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFindEquivalentProjectDirs(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs: [][]string{
			{"/home/gigur/git", "/Users/johkjo/git/personal"},
		},
	}

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "linux git project",
			input: "-home-gigur-git-tofu",
			expected: []string{
				"-home-gigur-git-tofu",
				"-Users-johkjo-git-tofu",          // from homes mapping
				"-Users-johkjo-git-personal-tofu", // from dirs mapping
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.FindEquivalentProjectDirs(tt.input)

			// Check that all expected are present
			for _, exp := range tt.expected {
				found := false
				for _, r := range result {
					if r == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("FindEquivalentProjectDirs(%q) missing %q, got %v", tt.input, exp, result)
				}
			}
		})
	}
}

func TestNilConfig(t *testing.T) {
	var config *SyncConfig
	result := config.CanonicalizeProjectDir("-home-gigur-git-tofu")
	if result != "-home-gigur-git-tofu" {
		t.Errorf("nil config should return input unchanged, got %q", result)
	}
}

func TestEmptyConfig(t *testing.T) {
	config := &SyncConfig{}
	result := config.CanonicalizeProjectDir("-home-gigur-git-tofu")
	if result != "-home-gigur-git-tofu" {
		t.Errorf("empty config should return input unchanged, got %q", result)
	}
}

func TestLocalizePath_LinuxToMac(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	tests := []struct {
		name      string
		path      string
		localHome string
		expected  string
	}{
		{
			name:      "canonical to mac with dirs mapping",
			path:      "/home/gigur/git/tofu",
			localHome: "/Users/johkjo",
			expected:  "/Users/johkjo/git/personal/tofu",
		},
		{
			name:      "canonical to mac with homes only",
			path:      "/home/gigur/Documents/notes",
			localHome: "/Users/johkjo",
			expected:  "/Users/johkjo/Documents/notes",
		},
		{
			name:      "already local path unchanged",
			path:      "/Users/johkjo/git/personal/tofu",
			localHome: "/Users/johkjo",
			expected:  "/Users/johkjo/git/personal/tofu",
		},
		{
			name:      "unrelated path unchanged",
			path:      "/var/log/something",
			localHome: "/Users/johkjo",
			expected:  "/var/log/something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.LocalizePath(tt.path, tt.localHome)
			if result != tt.expected {
				t.Errorf("LocalizePath(%q, %q) = %q, want %q", tt.path, tt.localHome, result, tt.expected)
			}
		})
	}
}

func TestLocalizePath_MacToLinux(t *testing.T) {
	// Same config but localHome is Linux
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	tests := []struct {
		name      string
		path      string
		localHome string
		expected  string
	}{
		{
			name:      "canonical stays canonical on linux",
			path:      "/home/gigur/git/tofu",
			localHome: "/home/gigur",
			expected:  "/home/gigur/git/tofu",
		},
		{
			name:      "mac path stays unchanged on linux (no reverse needed)",
			path:      "/Users/johkjo/git/personal/tofu",
			localHome: "/home/gigur",
			expected:  "/Users/johkjo/git/personal/tofu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.LocalizePath(tt.path, tt.localHome)
			if result != tt.expected {
				t.Errorf("LocalizePath(%q, %q) = %q, want %q", tt.path, tt.localHome, result, tt.expected)
			}
		})
	}
}

func TestLocalizePath_NilConfig(t *testing.T) {
	var config *SyncConfig
	result := config.LocalizePath("/home/gigur/git/tofu", "/Users/johkjo")
	if result != "/home/gigur/git/tofu" {
		t.Errorf("nil config should return path unchanged, got %q", result)
	}
}

func TestLocalizePath_EmptyConfig(t *testing.T) {
	config := &SyncConfig{}
	result := config.LocalizePath("/home/gigur/git/tofu", "/Users/johkjo")
	if result != "/home/gigur/git/tofu" {
		t.Errorf("empty config should return path unchanged, got %q", result)
	}
}

func TestLocalizePath_FullPathWithProjectDir(t *testing.T) {
	config := &SyncConfig{
		Homes: []string{"/home/gigur", "/Users/johkjo"},
		Dirs:  [][]string{{"/home/gigur/git", "/Users/johkjo/git/personal"}},
	}

	// FullPath contains the .claude path which should also be localized
	path := "/home/gigur/.claude/projects/-home-gigur-git-tofu/session.jsonl"
	localHome := "/Users/johkjo"

	result := config.LocalizePath(path, localHome)
	expected := "/Users/johkjo/.claude/projects/-home-gigur-git-tofu/session.jsonl"

	if result != expected {
		t.Errorf("LocalizePath(%q, %q) = %q, want %q", path, localHome, result, expected)
	}
}
