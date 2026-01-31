package conv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

// SessionsIndex represents the sessions-index.json file
type SessionsIndex struct {
	Version int            `json:"version"`
	Entries []SessionEntry `json:"entries"`
}

// SessionEntry represents a single session/conversation in the index
type SessionEntry struct {
	SessionID    string `json:"sessionId"`
	FullPath     string `json:"fullPath"`
	FileMtime    int64  `json:"fileMtime"`
	FirstPrompt  string `json:"firstPrompt"`
	Summary      string `json:"summary,omitempty"`     // New field in recent Claude versions
	CustomTitle  string `json:"customTitle,omitempty"` // Legacy field, kept for backwards compatibility
	MessageCount int    `json:"messageCount"`
	Created      string `json:"created"`
	Modified     string `json:"modified"`
	GitBranch    string `json:"gitBranch"`
	ProjectPath  string `json:"projectPath"`
	IsSidechain  bool   `json:"isSidechain"`
}

// DisplayTitle returns the best available title for display
// Priority: CustomTitle (legacy) -> Summary (new) -> FirstPrompt
func (e *SessionEntry) DisplayTitle() string {
	if e.CustomTitle != "" {
		return e.CustomTitle
	}
	if e.Summary != "" {
		return e.Summary
	}
	return e.FirstPrompt
}

// HasTitle returns true if the entry has a custom title or summary
func (e *SessionEntry) HasTitle() bool {
	return e.CustomTitle != "" || e.Summary != ""
}

func Cmd() *cobra.Command {
	cmd := boa.CmdT[boa.NoParams]{
		Use:   "conv",
		Short: "Manage Claude Code conversations",
		SubCmds: []*cobra.Command{
			ListCmd(),
			SearchCmd(),
			AISearchCmd(),
			ResumeCmd(),
			CpCmd(),
			MvCmd(),
			DeleteCmd(),
			PruneEmptyCmd(),
		},
	}.ToCobra()
	cmd.Aliases = []string{"convs", "conversation", "conversations"}
	return cmd
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
	// Clean and normalize the path
	absPath, err := filepath.Abs(realPath)
	if err != nil {
		absPath = realPath
	}
	// Replace path separators with dashes
	projectDir := strings.ReplaceAll(absPath, string(filepath.Separator), "-")
	// On Windows, also replace colons (from drive letters)
	projectDir = strings.ReplaceAll(projectDir, ":", "")
	return projectDir
}

// GetClaudeProjectPath returns the full path to a Claude project directory
func GetClaudeProjectPath(realPath string) string {
	return filepath.Join(ClaudeProjectsDir(), PathToProjectDir(realPath))
}

// LoadSessionsIndex loads the sessions index from a Claude project directory
// It also scans for unindexed .jsonl files and merges them, deduplicating by sessionId
// Additionally, it re-scans entries with missing display data (no prompt, summary, or title)
func LoadSessionsIndex(projectPath string) (*SessionsIndex, error) {
	indexPath := filepath.Join(projectPath, "sessions-index.json")
	data, err := os.ReadFile(indexPath)

	var index SessionsIndex
	if err != nil {
		if os.IsNotExist(err) {
			index = SessionsIndex{Version: 1, Entries: []SessionEntry{}}
		} else {
			return nil, fmt.Errorf("failed to read sessions index: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &index); err != nil {
			return nil, fmt.Errorf("failed to parse sessions index: %w", err)
		}
	}

	// Re-scan entries with missing display data
	for i := range index.Entries {
		if index.Entries[i].DisplayTitle() == "" {
			// Try to get data from the file
			filePath := filepath.Join(projectPath, index.Entries[i].SessionID+".jsonl")
			if scanned := parseJSONLSession(filePath, index.Entries[i].SessionID); scanned != nil {
				// Update missing fields from scanned data
				if scanned.Summary != "" && index.Entries[i].Summary == "" {
					index.Entries[i].Summary = scanned.Summary
				}
				if scanned.FirstPrompt != "" && index.Entries[i].FirstPrompt == "" {
					index.Entries[i].FirstPrompt = scanned.FirstPrompt
				}
				if scanned.ProjectPath != "" && index.Entries[i].ProjectPath == "" {
					index.Entries[i].ProjectPath = scanned.ProjectPath
				}
				if scanned.GitBranch != "" && index.Entries[i].GitBranch == "" {
					index.Entries[i].GitBranch = scanned.GitBranch
				}
			}
		}
	}

	// Scan for unindexed .jsonl files and merge them
	unindexed := scanUnindexedSessions(projectPath)
	if len(unindexed) > 0 {
		index.Entries = mergeSessionEntries(index.Entries, unindexed)
	}

	return &index, nil
}

// scanUnindexedSessions scans for .jsonl files directly in the project directory
// These are conversations that haven't been indexed yet by Claude Code
func scanUnindexedSessions(projectPath string) []SessionEntry {
	var entries []SessionEntry

	files, err := os.ReadDir(projectPath)
	if err != nil {
		return entries
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".jsonl") {
			continue
		}

		// Extract session ID from filename (e.g., "0789725a-bc71-47dd-9ca5-1b4fe7aead9b.jsonl")
		sessionID := strings.TrimSuffix(file.Name(), ".jsonl")
		if len(sessionID) != 36 { // UUID length
			continue
		}

		filePath := filepath.Join(projectPath, file.Name())
		entry := parseJSONLSession(filePath, sessionID)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	return entries
}

// jsonlMessage represents a line in the .jsonl conversation file
type jsonlMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
	Timestamp string `json:"timestamp"`
	Cwd       string `json:"cwd"`
	GitBranch string `json:"gitBranch"`
	Summary   string `json:"summary"` // For type="summary" messages
	Message   struct {
		Role    string `json:"role"`
		Content any    `json:"content"` // Can be string or array
	} `json:"message"`
}

// parseJSONLSession parses a .jsonl file and extracts session metadata
// Only reads the first few lines to find the first user message for the prompt
func parseJSONLSession(filePath, sessionID string) *SessionEntry {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil
	}

	var entry SessionEntry
	entry.SessionID = sessionID
	entry.FullPath = filePath
	entry.FileMtime = info.ModTime().Unix()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var firstTimestamp string

	// Scan the entire file looking for:
	// 1. Summary messages (keep the last one as it's most up-to-date)
	// 2. First user message with actual text content
	// Stop early once we have both a summary/title AND a firstPrompt
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg jsonlMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		// Track first timestamp
		if firstTimestamp == "" && msg.Timestamp != "" {
			firstTimestamp = msg.Timestamp
		}

		// Capture summaries (keep the last one as it's most up-to-date)
		if msg.Type == "summary" && msg.Summary != "" {
			entry.Summary = msg.Summary
			// If we already have a firstPrompt, we're done
			if entry.FirstPrompt != "" {
				break
			}
			continue
		}

		// Get project path and git branch from first message with cwd
		if entry.ProjectPath == "" && msg.Cwd != "" {
			entry.ProjectPath = msg.Cwd
		}
		if entry.GitBranch == "" && msg.GitBranch != "" {
			entry.GitBranch = msg.GitBranch
		}

		// Capture first user message with actual text content as the prompt
		if entry.FirstPrompt == "" && msg.Type == "user" && msg.Message.Role == "user" {
			text := extractMessageContent(msg.Message.Content)
			// Skip messages without text (e.g., tool_result blocks from resumed sessions)
			// Also skip system-generated messages like "[Request interrupted by user...]"
			if text != "" && !strings.HasPrefix(text, "[Request interrupted") {
				entry.FirstPrompt = text
				if msg.Timestamp != "" {
					firstTimestamp = msg.Timestamp
				}
				// If we already have a summary, we're done
				if entry.Summary != "" {
					break
				}
			}
		}
	}

	if firstTimestamp == "" {
		// No valid data found
		return nil
	}

	entry.Created = firstTimestamp
	// Use file mtime for Modified since we're not reading the whole file
	entry.Modified = info.ModTime().UTC().Format(time.RFC3339)
	entry.MessageCount = 0 // Unknown for unindexed sessions

	return &entry
}

// extractMessageContent extracts text content from a message
// Content can be a string or an array of content blocks
func extractMessageContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		// Array of content blocks - look for text type
		for _, block := range v {
			if m, ok := block.(map[string]any); ok {
				if text, ok := m["text"].(string); ok {
					return text
				}
			}
		}
	}
	return ""
}

// mergeSessionEntries merges indexed and unindexed entries, deduplicating by sessionId
func mergeSessionEntries(indexed, unindexed []SessionEntry) []SessionEntry {
	// Build a set of indexed session IDs
	indexedIDs := make(map[string]bool)
	for _, e := range indexed {
		indexedIDs[e.SessionID] = true
	}

	// Add unindexed entries that aren't already in the index
	result := indexed
	for _, e := range unindexed {
		if !indexedIDs[e.SessionID] {
			result = append(result, e)
		}
	}

	return result
}

// SaveSessionsIndex saves the sessions index to a Claude project directory
func SaveSessionsIndex(projectPath string, index *SessionsIndex) error {
	indexPath := filepath.Join(projectPath, "sessions-index.json")
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions index: %w", err)
	}

	if err := os.WriteFile(indexPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write sessions index: %w", err)
	}

	return nil
}

// FindSessionByID finds a session entry by its ID (full or prefix)
func FindSessionByID(index *SessionsIndex, sessionID string) (*SessionEntry, int) {
	// First try exact match
	for i, entry := range index.Entries {
		if entry.SessionID == sessionID {
			return &index.Entries[i], i
		}
	}
	// Then try prefix match
	var matches []int
	for i, entry := range index.Entries {
		if strings.HasPrefix(entry.SessionID, sessionID) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 1 {
		return &index.Entries[matches[0]], matches[0]
	}
	return nil, -1
}

// RemoveSessionByID removes a session from the index by its ID
func RemoveSessionByID(index *SessionsIndex, sessionID string) bool {
	for i, entry := range index.Entries {
		if entry.SessionID == sessionID {
			index.Entries = append(index.Entries[:i], index.Entries[i+1:]...)
			return true
		}
	}
	return false
}

// CopyDir recursively copies a directory
func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile copies a single file
func CopyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, srcInfo.Mode())
}

// ListSessions returns all sessions from a project directory
func ListSessions(projectPath string) ([]SessionEntry, error) {
	index, err := LoadSessionsIndex(projectPath)
	if err != nil {
		return nil, err
	}
	return index.Entries, nil
}

// ParseTimeParam parses a time parameter string into a time.Time
// Supports formats: "2024-01-15", "2024-01-15T10:30", "24h", "7d", "2w", or any time.Duration
func ParseTimeParam(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	// Try standard time.Duration first (e.g., "1h30m", "2h45m30s")
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(-d), nil
	}

	// Try extended duration with days/weeks (e.g., "7d", "2w")
	if len(s) >= 2 {
		unit := s[len(s)-1]
		numStr := s[:len(s)-1]
		if num, err := strconv.Atoi(numStr); err == nil {
			var duration time.Duration
			switch unit {
			case 'd':
				duration = time.Duration(num) * 24 * time.Hour
			case 'w':
				duration = time.Duration(num) * 7 * 24 * time.Hour
			}
			if duration > 0 {
				return time.Now().Add(-duration), nil
			}
		}
	}

	// Try various date/time formats
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
		"2006/01/02",
		"01-02-2006",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s (try formats like 2024-01-15, 1h30m, 7d)", s)
}

// FilterEntriesByTime filters session entries by time range
func FilterEntriesByTime(entries []SessionEntry, since, before string) ([]SessionEntry, error) {
	sinceTime, err := ParseTimeParam(since)
	if err != nil {
		return nil, fmt.Errorf("invalid --since value: %w", err)
	}

	beforeTime, err := ParseTimeParam(before)
	if err != nil {
		return nil, fmt.Errorf("invalid --before value: %w", err)
	}

	if sinceTime.IsZero() && beforeTime.IsZero() {
		return entries, nil
	}

	var filtered []SessionEntry
	for _, e := range entries {
		modTime, err := time.Parse(time.RFC3339, e.Modified)
		if err != nil {
			continue
		}

		if !sinceTime.IsZero() && modTime.Before(sinceTime) {
			continue
		}
		if !beforeTime.IsZero() && modTime.After(beforeTime) {
			continue
		}
		filtered = append(filtered, e)
	}

	return filtered, nil
}
