package session

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

// SessionState represents the state of a Claude session
type SessionState struct {
	ID           string    `json:"id"`
	TmuxSession  string    `json:"tmuxSession"`
	PID          int       `json:"pid"`
	Cwd          string    `json:"cwd"`
	ConvID       string    `json:"convId,omitempty"`
	Status       string    `json:"status"`
	StatusDetail string    `json:"statusDetail,omitempty"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
}

// Status constants
const (
	StatusRunning           = "running"
	StatusWaitingInput      = "waiting_input"
	StatusWaitingPermission = "waiting_permission"
	StatusExited            = "exited"
)

func Cmd() *cobra.Command {
	return boa.CmdT[boa.NoParams]{
		Use:   "session",
		Short: "Manage Claude Code sessions (tmux-based)",
		Long:  "Multiplex and manage multiple Claude Code sessions with detach/reattach support.",
		SubCmds: []*cobra.Command{
			NewCmd(),
			ListCmd(),
			AttachCmd(),
			KillCmd(),
		},
	}.ToCobra()
}

// SessionsDir returns the directory where session state files are stored
func SessionsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".tofu", "claude-sessions")
}

// EnsureSessionsDir creates the sessions directory if it doesn't exist
func EnsureSessionsDir() error {
	dir := SessionsDir()
	return os.MkdirAll(dir, 0755)
}

// SessionStatePath returns the path to a session's state file
func SessionStatePath(id string) string {
	return filepath.Join(SessionsDir(), id+".json")
}

// SaveSessionState saves session state to disk
func SaveSessionState(state *SessionState) error {
	if err := EnsureSessionsDir(); err != nil {
		return err
	}
	state.Updated = time.Now()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(SessionStatePath(state.ID), data, 0644)
}

// LoadSessionState loads session state from disk
func LoadSessionState(id string) (*SessionState, error) {
	data, err := os.ReadFile(SessionStatePath(id))
	if err != nil {
		return nil, err
	}
	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// DeleteSessionState removes a session's state file
func DeleteSessionState(id string) error {
	return os.Remove(SessionStatePath(id))
}

// ListSessionStates returns all session states
func ListSessionStates() ([]*SessionState, error) {
	dir := SessionsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var states []*SessionState
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := entry.Name()[:len(entry.Name())-5] // remove .json
		state, err := LoadSessionState(id)
		if err != nil {
			continue
		}
		states = append(states, state)
	}
	return states, nil
}

// IsProcessAlive checks if a process with the given PID is still running
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds, so we need to send signal 0
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// IsTmuxSessionAlive checks if a tmux session exists
func IsTmuxSessionAlive(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

// CheckTmuxInstalled verifies tmux is available
func CheckTmuxInstalled() error {
	_, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux is required but not installed. Install it with:\n  Ubuntu/Debian: sudo apt install tmux\n  macOS: brew install tmux")
	}
	return nil
}

// GenerateSessionID creates a short unique session ID
func GenerateSessionID() string {
	// Use last 8 hex chars of unix nano time
	hex := fmt.Sprintf("%016x", time.Now().UnixNano())
	return hex[len(hex)-8:]
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

// RefreshSessionStatus updates the session status based on actual state
func RefreshSessionStatus(state *SessionState) {
	if !IsTmuxSessionAlive(state.TmuxSession) {
		state.Status = StatusExited
	}
}

// ShortID returns the first 8 characters of an ID for display
func ShortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// ShortenPath shortens a path for display
func ShortenPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	// Show last part of path
	parts := filepath.SplitList(path)
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if len(last) <= maxLen-3 {
			return "…" + string(filepath.Separator) + last
		}
	}
	return "…" + path[len(path)-maxLen+1:]
}

// ParsePIDFromTmux gets the PID of the main process in a tmux session
func ParsePIDFromTmux(sessionName string) int {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "#{pane_pid}")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(string(output[:len(output)-1])) // trim newline
	return pid
}
