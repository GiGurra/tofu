// Package convindex provides minimal conversation index lookup functionality.
// This package exists to avoid import cycles between session and conv packages.
package convindex

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// SessionsIndex represents the sessions-index.json file
type SessionsIndex struct {
	Version int            `json:"version"`
	Entries []SessionEntry `json:"entries"`
}

// SessionEntry represents a single session/conversation in the index
type SessionEntry struct {
	SessionID   string `json:"sessionId"`
	FirstPrompt string `json:"firstPrompt"`
	Summary     string `json:"summary,omitempty"`
	CustomTitle string `json:"customTitle,omitempty"`
}

// DisplayTitle returns the best available title for display
// Priority: CustomTitle -> Summary -> FirstPrompt
func (e *SessionEntry) DisplayTitle() string {
	if e.CustomTitle != "" {
		return e.CustomTitle
	}
	if e.Summary != "" {
		return e.Summary
	}
	return e.FirstPrompt
}

// ClaudeProjectsDir returns the Claude projects directory path
func ClaudeProjectsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// PathToProjectDir converts a real path to the Claude project directory name
// e.g., /home/gigur/git/tofu -> -home-gigur-git-tofu
func PathToProjectDir(realPath string) string {
	absPath, err := filepath.Abs(realPath)
	if err != nil {
		absPath = realPath
	}
	projectDir := strings.ReplaceAll(absPath, string(filepath.Separator), "-")
	projectDir = strings.ReplaceAll(projectDir, ":", "") // Windows drive letters
	return projectDir
}

// GetClaudeProjectPath returns the full path to a Claude project directory
func GetClaudeProjectPath(realPath string) string {
	return filepath.Join(ClaudeProjectsDir(), PathToProjectDir(realPath))
}

// LoadSessionsIndexFast loads just the sessions index JSON without scanning.
func LoadSessionsIndexFast(projectPath string) (*SessionsIndex, error) {
	indexPath := filepath.Join(projectPath, "sessions-index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &SessionsIndex{Version: 1, Entries: []SessionEntry{}}, nil
		}
		return nil, err
	}

	var index SessionsIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

// FindSessionByID finds a session entry by its ID (full or prefix)
func FindSessionByID(index *SessionsIndex, sessionID string) *SessionEntry {
	if index == nil {
		return nil
	}
	// First try exact match
	for i, entry := range index.Entries {
		if entry.SessionID == sessionID {
			return &index.Entries[i]
		}
	}
	// Then try prefix match
	var matches []*SessionEntry
	for i, entry := range index.Entries {
		if strings.HasPrefix(entry.SessionID, sessionID) {
			matches = append(matches, &index.Entries[i])
		}
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return nil
}

// GetConvTitle is a convenience function to look up a conversation title.
// It checks the index first, then falls back to parsing the .jsonl file directly.
func GetConvTitle(convID, cwd string) string {
	if convID == "" || cwd == "" {
		return ""
	}

	projectPath := GetClaudeProjectPath(cwd)

	// Try index first
	index, _ := LoadSessionsIndexFast(projectPath)
	if index != nil {
		if entry := FindSessionByID(index, convID); entry != nil {
			if title := entry.DisplayTitle(); title != "" {
				return cleanTitle(title)
			}
		}
	}

	// Fallback: parse .jsonl file directly for unindexed conversations
	return cleanTitle(parseFirstPromptFromJSONL(projectPath, convID))
}

// cleanTitle removes XML-like tags and truncates the title for display.
func cleanTitle(title string) string {
	if title == "" {
		return ""
	}

	// Remove XML-like tags (e.g., <local-command-caveat>...</local-command-caveat>)
	result := stripXMLTags(title)

	// Trim whitespace
	result = strings.TrimSpace(result)

	// Truncate to reasonable length for notifications
	const maxLen = 80
	if len(result) > maxLen {
		result = result[:maxLen-3] + "..."
	}

	return result
}

// stripXMLTags removes XML-like tags from a string.
func stripXMLTags(s string) string {
	var result strings.Builder
	inTag := false

	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// jsonlMessage represents a message in the JSONL transcript
type jsonlMessage struct {
	Type    string `json:"type"`
	Message struct {
		Role    string `json:"role"`
		Content any    `json:"content"` // Can be string or array
	} `json:"message"`
	Summary string `json:"summary,omitempty"` // Some entries have summary
}

// parseFirstPromptFromJSONL extracts the first user prompt from a .jsonl file
func parseFirstPromptFromJSONL(projectPath, sessionID string) string {
	filePath := filepath.Join(projectPath, sessionID+".jsonl")
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer for large lines
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var msg jsonlMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		// Check for summary field first
		if msg.Summary != "" {
			return msg.Summary
		}

		// Look for first user message
		if msg.Type == "user" && msg.Message.Role == "user" {
			return extractTextContent(msg.Message.Content)
		}
	}

	return ""
}

// extractTextContent extracts text from message content (can be string or array)
func extractTextContent(content any) string {
	// Direct string
	if s, ok := content.(string); ok {
		return s
	}

	// Array of content blocks
	if arr, ok := content.([]any); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				if m["type"] == "text" {
					if text, ok := m["text"].(string); ok {
						return text
					}
				}
			}
		}
	}

	return ""
}
