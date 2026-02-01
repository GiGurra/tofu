package session

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	Attached     int       `json:"-"` // Number of attached clients (runtime only, not persisted)
}

// Status constants
const (
	StatusRunning           = "running"
	StatusWaitingInput      = "waiting_input"
	StatusWaitingPermission = "waiting_permission"
	StatusExited            = "exited"
)

// SortColumn represents which column to sort by
type SortColumn int

const (
	SortNone SortColumn = iota
	SortID
	SortDirectory
	SortStatus
	SortAge
	SortUpdated
)

// SortDirection represents ascending or descending
type SortDirection int

const (
	SortAsc SortDirection = iota
	SortDesc
)

// SortState tracks current sort settings
type SortState struct {
	Column    SortColumn
	Direction SortDirection
}

// NextState cycles through: none -> asc -> desc -> none
func (s *SortState) Toggle(col SortColumn) {
	if s.Column != col {
		// New column: start with ascending
		s.Column = col
		s.Direction = SortAsc
	} else if s.Direction == SortAsc {
		// Same column, was asc: switch to desc
		s.Direction = SortDesc
	} else {
		// Same column, was desc: reset to none
		s.Column = SortNone
	}
}

// SortIndicator returns a display indicator for the column header
func (s *SortState) Indicator(col SortColumn) string {
	if s.Column != col {
		return ""
	}
	if s.Direction == SortAsc {
		return " ▼"
	}
	return " ▲"
}

// SortSessions sorts sessions according to the current sort state
func SortSessions(sessions []*SessionState, state SortState) {
	if state.Column == SortNone || len(sessions) < 2 {
		return
	}

	// Simple bubble sort for small lists
	for i := 0; i < len(sessions)-1; i++ {
		for j := 0; j < len(sessions)-i-1; j++ {
			if shouldSwap(sessions[j], sessions[j+1], state) {
				sessions[j], sessions[j+1] = sessions[j+1], sessions[j]
			}
		}
	}
}

func shouldSwap(a, b *SessionState, state SortState) bool {
	var less bool

	switch state.Column {
	case SortID:
		less = a.ID < b.ID
	case SortDirectory:
		less = a.Cwd < b.Cwd
	case SortStatus:
		// Custom status priority: red (needs attention) first, then yellow, then rest
		less = statusPriority(a.Status) < statusPriority(b.Status)
	case SortAge:
		less = a.Created.Before(b.Created)
	case SortUpdated:
		less = a.Updated.Before(b.Updated)
	default:
		return false
	}

	if state.Direction == SortDesc {
		return less // swap if a < b (to get descending)
	}
	return !less // swap if a > b (to get ascending)
}

// statusPriority returns sort priority for status (lower = shown first when ascending)
// Red (needs attention) = 0, Yellow (idle) = 1, Green (working) = 2, Gray (exited) = 3
func statusPriority(status string) int {
	switch status {
	case StatusAwaitingPermission, StatusAwaitingInput:
		return 0 // Red - needs attention, show first
	case StatusIdle:
		return 1 // Yellow
	case StatusWorking:
		return 2 // Green
	case StatusExited:
		return 3 // Gray
	default:
		return 0 // Unknown = needs attention
	}
}

func Cmd() *cobra.Command {
	cmd := boa.CmdT[boa.NoParams]{
		Use:   "session",
		Short: "Manage Claude Code sessions (tmux-based)",
		Long:  "Multiplex and manage multiple Claude Code sessions with detach/reattach support.",
		SubCmds: []*cobra.Command{
			NewCmd(),
			ListCmd(),
			AttachCmd(),
			KillCmd(),
			PruneCmd(),
			InstallHooksCmd(),
			StatusCallbackCmd(),
		},
	}.ToCobra()
	cmd.Aliases = []string{"sessions"}
	return cmd
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

// DefaultCleanupAge is the default max age for exited sessions in prune command
const DefaultCleanupAge = 7 * 24 * time.Hour // 1 week

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

// CleanupOldExitedSessions removes exited session states older than maxAge
func CleanupOldExitedSessions(maxAge time.Duration) error {
	dir := SessionsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := entry.Name()[:len(entry.Name())-5]
		state, err := LoadSessionState(id)
		if err != nil {
			// Can't load, maybe corrupted - delete it
			_ = DeleteSessionState(id)
			continue
		}

		// Refresh status to ensure we have current state
		RefreshSessionStatus(state)

		// Delete if exited and older than cutoff
		if state.Status == StatusExited && state.Updated.Before(cutoff) {
			_ = DeleteSessionState(id)
		}
	}
	return nil
}

// IsTmuxSessionAlive checks if a tmux session exists
func IsTmuxSessionAlive(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

// GetTmuxSessionAttachedCount returns the number of clients attached to a tmux session
// Returns 0 if session doesn't exist or on error
func GetTmuxSessionAttachedCount(sessionName string) int {
	cmd := exec.Command("tmux", "display-message", "-t", sessionName, "-p", "#{session_attached}")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	count, _ := strconv.Atoi(strings.TrimSpace(string(output)))
	return count
}

// IsTmuxSessionAttached checks if a tmux session has any clients attached
func IsTmuxSessionAttached(sessionName string) bool {
	return GetTmuxSessionAttachedCount(sessionName) > 0
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
	// For tmux-backed sessions, check if tmux session is alive
	if state.TmuxSession != "" {
		if IsTmuxSessionAlive(state.TmuxSession) {
			state.Attached = GetTmuxSessionAttachedCount(state.TmuxSession)
			return
		}
		// Tmux session is dead - fall through to check PID
		state.Attached = 0
	}

	// Check if PID is alive (works for both non-tmux sessions and
	// sessions where tmux died but the process is still running)
	if state.PID > 0 {
		if !IsProcessAlive(state.PID) {
			state.Status = StatusExited
		}
		// If PID is alive, keep the current status (updated by hooks)
		return
	}

	// No tmux session and no PID - mark as exited
	state.Status = StatusExited
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

// getSessionCompletions returns completions for session IDs
// If includeExited is true, includes exited sessions (for kill command)
func getSessionCompletions(includeExited bool) []string {
	states, err := ListSessionStates()
	if err != nil || len(states) == 0 {
		return nil
	}

	var completions []string
	for _, state := range states {
		RefreshSessionStatus(state)
		if !includeExited && state.Status == StatusExited {
			continue
		}

		// Format: ID_status_directory
		dir := state.Cwd
		if len(dir) > 30 {
			dir = "…" + dir[len(dir)-29:]
		}
		dir = strings.ReplaceAll(dir, " ", "_")

		completion := fmt.Sprintf("%s_%s_%s", state.ID, state.Status, dir)
		completions = append(completions, completion)
	}

	return completions
}
