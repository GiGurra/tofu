package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ConvEntry represents a conversation entry for completions
type ConvEntry struct {
	SessionID   string `json:"sessionId"`
	FirstPrompt string `json:"firstPrompt"`
	Summary     string `json:"summary,omitempty"`
	CustomTitle string `json:"customTitle,omitempty"`
	ProjectPath string `json:"projectPath"`
	Modified    string `json:"modified"`
}

// DisplayTitle returns the best available title
func (e *ConvEntry) DisplayTitle() string {
	if e.CustomTitle != "" {
		return e.CustomTitle
	}
	if e.Summary != "" {
		return e.Summary
	}
	return e.FirstPrompt
}

// HasTitle returns true if the entry has a title or summary
func (e *ConvEntry) HasTitle() bool {
	return e.CustomTitle != "" || e.Summary != ""
}

// ClaudeProjectsDir returns the Claude projects directory path
func ClaudeProjectsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// GetConversationCompletions returns completions for conversation IDs
func GetConversationCompletions(global bool) []string {
	var entries []ConvEntry

	if global {
		projectsDir := ClaudeProjectsDir()
		dirEntries, err := os.ReadDir(projectsDir)
		if err != nil {
			return nil
		}

		for _, dirEntry := range dirEntries {
			if !dirEntry.IsDir() {
				continue
			}
			projPath := filepath.Join(projectsDir, dirEntry.Name())
			loaded := loadConvEntries(projPath)
			entries = append(entries, loaded...)
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil
		}

		projectPath := getClaudeProjectPath(cwd)
		entries = loadConvEntries(projectPath)
	}

	// Sort by modified date descending (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Modified > entries[j].Modified
	})

	// Format completions
	results := make([]string, len(entries))
	for i, e := range entries {
		results[i] = FormatConvCompletion(e)
	}

	return results
}

func loadConvEntries(projectPath string) []ConvEntry {
	indexPath := filepath.Join(projectPath, "sessions-index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil
	}

	var index struct {
		Entries []ConvEntry `json:"entries"`
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return nil
	}

	return index.Entries
}

func getClaudeProjectPath(realPath string) string {
	absPath, err := filepath.Abs(realPath)
	if err != nil {
		absPath = realPath
	}
	projectDir := strings.ReplaceAll(absPath, string(filepath.Separator), "-")
	projectDir = strings.ReplaceAll(projectDir, ":", "")
	return filepath.Join(ClaudeProjectsDir(), projectDir)
}

// ExtractIDFromCompletion extracts just the ID from autocomplete format
// e.g., "0459cd73_[title]_prompt..." -> "0459cd73"
func ExtractIDFromCompletion(s string) string {
	if idx := strings.Index(s, "_"); idx > 0 {
		return s[:idx]
	}
	return s
}

// ConvInfo contains resolved conversation information
type ConvInfo struct {
	SessionID   string // Full UUID
	ProjectPath string // Original project directory
}

// ResolveConvID resolves a short conversation ID prefix to full info
// Returns the full ID and project path if found
func ResolveConvID(shortID string) *ConvInfo {
	if shortID == "" {
		return nil
	}

	projectsDir := ClaudeProjectsDir()
	dirEntries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}
		projPath := filepath.Join(projectsDir, dirEntry.Name())
		entries := loadConvEntries(projPath)

		for _, e := range entries {
			// Exact match
			if e.SessionID == shortID {
				return &ConvInfo{SessionID: e.SessionID, ProjectPath: e.ProjectPath}
			}
			// Prefix match
			if strings.HasPrefix(e.SessionID, shortID) {
				return &ConvInfo{SessionID: e.SessionID, ProjectPath: e.ProjectPath}
			}
		}
	}

	return nil
}

// FormatConvCompletion formats a conversation entry for shell completion
func FormatConvCompletion(e ConvEntry) string {
	sanitize := func(s string) string {
		s = strings.ReplaceAll(s, "\t", "__")
		s = strings.ReplaceAll(s, " ", "_")
		s = strings.ReplaceAll(s, "\n", "_")
		s = strings.ReplaceAll(s, "\r", "")
		return s
	}

	id := e.SessionID
	if len(id) > 8 {
		id = id[:8]
	}

	var namePart string
	if e.HasTitle() {
		namePart = "[" + sanitize(e.DisplayTitle()) + "]_"
	}

	prompt := sanitize(e.FirstPrompt)
	if len(prompt) > 40 {
		prompt = prompt[:37] + "..."
	}

	modified := e.Modified
	if len(modified) >= 16 {
		modified = strings.ReplaceAll(modified[:16], "T", "_")
	}

	return id + "_" + namePart + prompt + "__" + modified
}
