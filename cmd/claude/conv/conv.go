package conv

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	CustomTitle  string `json:"customTitle,omitempty"`
	MessageCount int    `json:"messageCount"`
	Created      string `json:"created"`
	Modified     string `json:"modified"`
	GitBranch    string `json:"gitBranch"`
	ProjectPath  string `json:"projectPath"`
	IsSidechain  bool   `json:"isSidechain"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[boa.NoParams]{
		Use:   "conv",
		Short: "Manage Claude Code conversations",
		SubCmds: []*cobra.Command{
			ListCmd(),
			SearchCmd(),
			CpCmd(),
			MvCmd(),
		},
	}.ToCobra()
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
func LoadSessionsIndex(projectPath string) (*SessionsIndex, error) {
	indexPath := filepath.Join(projectPath, "sessions-index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &SessionsIndex{Version: 1, Entries: []SessionEntry{}}, nil
		}
		return nil, fmt.Errorf("failed to read sessions index: %w", err)
	}

	var index SessionsIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse sessions index: %w", err)
	}

	return &index, nil
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

