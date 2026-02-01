package conv

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
	"github.com/gigurra/tofu/cmd/claude/session"
	"github.com/gigurra/tofu/cmd/claude/syncutil"
)

var (
	wSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("62"))
	wHeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	wHelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	wSearchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	wConfirmStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	wActiveStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))  // Green for attached tmux session
	wDetachedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))  // Green for detached tmux session
	wNoTmuxStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow for non-tmux session
)

type watchTickMsg time.Time
type watchFileChangeMsg struct {
	sessionID string // ID of the session that changed (empty for delete)
	deleted   bool   // true if the file was deleted
}

// watchConfirmMode represents confirmation dialogs
type watchConfirmMode int

const (
	watchConfirmNone watchConfirmMode = iota
	watchConfirmAttachForce       // Session already attached, confirm force attach
	watchConfirmDelete            // Delete conversation (no active session)
	watchConfirmDeleteWithSession // Delete conversation that has an active session
	watchConfirmNoTmux            // Session has no tmux, cannot attach
)

type watchModel struct {
	// Data
	entries        []SessionEntry
	filtered       []SessionEntry                    // After search filter
	activeSessions map[string]*session.SessionState  // convID -> session

	// Navigation
	cursor         int
	viewportOffset int
	viewportHeight int

	// Search
	searchInput   string
	searchFocused bool

	// UI state
	width       int
	height      int
	confirmMode watchConfirmMode
	helpView    bool

	// Settings
	global      bool   // Search all projects
	projectPath string // Current project path
	since       string // Filter: modified after
	before      string // Filter: modified before

	// Result
	selectedConv *SessionEntry
	shouldCreate bool  // true = create new session, false = attach to existing
	forceAttach  bool

	// Status message (shown briefly after actions)
	statusMsg string
}

func initialWatchModel(global bool, since, before string) watchModel {
	cwd, _ := os.Getwd()
	return watchModel{
		entries:        []SessionEntry{},
		filtered:       []SessionEntry{},
		activeSessions: make(map[string]*session.SessionState),
		global:         global,
		projectPath:    cwd,
		since:          since,
		before:         before,
		viewportHeight: 20, // Will be adjusted based on terminal size
	}
}

func (m watchModel) Init() tea.Cmd {
	return tea.Batch(watchTickCmd(), watchSessionFilesCmd(), tea.EnterAltScreen)
}

func watchTickCmd() tea.Cmd {
	// Fallback tick - slower now that we have file watching
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return watchTickMsg(t)
	})
}

// watchSessionFilesCmd starts a file watcher on the session state directory
// and returns a command that sends watchFileChangeMsg when session files change
func watchSessionFilesCmd() tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil // Fall back to polling
		}

		dir := session.SessionsDir()
		if dir == "" {
			watcher.Close()
			return nil
		}

		// Ensure directory exists before watching
		if err := session.EnsureSessionsDir(); err != nil {
			watcher.Close()
			return nil
		}

		if err := watcher.Add(dir); err != nil {
			watcher.Close()
			return nil
		}

		// Wait for a relevant file change
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				// Only care about .json files (session state files)
				if filepath.Ext(event.Name) == ".json" {
					// Extract session ID from filename (e.g., "abc123.json" -> "abc123")
					sessionID := filepath.Base(event.Name)
					sessionID = sessionID[:len(sessionID)-5] // remove .json

					if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
						// For remove events, trigger immediately
						if event.Op&fsnotify.Remove != 0 {
							watcher.Close()
							return watchFileChangeMsg{sessionID: sessionID, deleted: true}
						}
						// For write/create, verify the file is complete and parseable
						if isValidSessionStateFile(event.Name) {
							watcher.Close()
							return watchFileChangeMsg{sessionID: sessionID, deleted: false}
						}
						// File not yet complete, wait for next event
					}
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				// On error, close and fall back to polling
				watcher.Close()
				return nil
			}
		}
	}
}

// isValidSessionStateFile checks if a session state file is complete and parseable
func isValidSessionStateFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var state session.SessionState
	return json.Unmarshal(data, &state) == nil
}

// loadConversations loads all conversations based on settings
func (m watchModel) loadConversations() watchModel {
	var allEntries []SessionEntry

	if m.global {
		projectsDir := ClaudeProjectsDir()
		entries, err := os.ReadDir(projectsDir)
		if err != nil {
			return m
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			projectPath := projectsDir + "/" + entry.Name()
			index, err := LoadSessionsIndex(projectPath)
			if err != nil {
				continue
			}
			allEntries = append(allEntries, index.Entries...)
		}
	} else {
		projectPath := GetClaudeProjectPath(m.projectPath)
		index, err := LoadSessionsIndex(projectPath)
		if err != nil {
			m.entries = []SessionEntry{}
			m.filtered = []SessionEntry{}
			return m
		}
		allEntries = index.Entries
	}

	// Filter by time if specified
	allEntries, _ = FilterEntriesByTime(allEntries, m.since, m.before)

	// Sort by modified date descending (most recent first)
	sortEntries(allEntries, "modified", false)

	m.entries = allEntries
	m = m.applySearchFilter()
	m = m.refreshActiveSessions()
	return m
}

// applySearchFilter filters entries based on search input
func (m watchModel) applySearchFilter() watchModel {
	if m.searchInput == "" {
		m.filtered = m.entries
		return m
	}

	query := strings.ToLower(m.searchInput)
	var filtered []SessionEntry
	for _, e := range m.entries {
		if matchesSearch(e, query) {
			filtered = append(filtered, e)
		}
	}
	m.filtered = filtered

	// Reset cursor/viewport if needed
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
	}
	m.viewportOffset = 0

	return m
}

func matchesSearch(e SessionEntry, query string) bool {
	return strings.Contains(strings.ToLower(e.DisplayTitle()), query) ||
		strings.Contains(strings.ToLower(e.FirstPrompt), query) ||
		strings.Contains(strings.ToLower(e.ProjectPath), query) ||
		strings.Contains(strings.ToLower(e.GitBranch), query) ||
		strings.Contains(strings.ToLower(e.SessionID), query)
}

// refreshActiveSessions updates which conversations have active sessions
func (m watchModel) refreshActiveSessions() watchModel {
	states, _ := session.ListSessionStates()
	m.activeSessions = make(map[string]*session.SessionState)

	for _, state := range states {
		session.RefreshSessionStatus(state)
		if state.Status != session.StatusExited && state.ConvID != "" {
			m.activeSessions[state.ConvID] = state
		}
	}

	return m
}

// updateSingleSession updates just one session in the activeSessions map
func (m watchModel) updateSingleSession(sessionID string, deleted bool) watchModel {
	if deleted {
		// Find and remove any session with this ID
		for convID, state := range m.activeSessions {
			if state.ID == sessionID {
				delete(m.activeSessions, convID)
				break
			}
		}
		return m
	}

	// Load the updated session state
	state, err := session.LoadSessionState(sessionID)
	if err != nil {
		return m
	}

	session.RefreshSessionStatus(state)

	// Update or remove from map based on status
	if state.Status == session.StatusExited {
		// Remove from active sessions
		if state.ConvID != "" {
			delete(m.activeSessions, state.ConvID)
		}
	} else if state.ConvID != "" {
		// Update in active sessions
		m.activeSessions[state.ConvID] = state
	}

	return m
}

// ensureCursorVisible scrolls viewport to keep cursor visible
func (m watchModel) ensureCursorVisible() watchModel {
	if m.cursor < m.viewportOffset {
		m.viewportOffset = m.cursor
	}
	if m.cursor >= m.viewportOffset+m.viewportHeight {
		m.viewportOffset = m.cursor - m.viewportHeight + 1
	}
	return m
}

// deleteConversation deletes a conversation's files and removes it from the index
func (m watchModel) deleteConversation(conv *SessionEntry) error {
	// Determine project path
	var projectPath string
	if m.global {
		// For global mode, derive project path from the conversation's ProjectPath
		projectPath = GetClaudeProjectPath(conv.ProjectPath)
	} else {
		projectPath = GetClaudeProjectPath(m.projectPath)
	}

	// Load index
	index, err := LoadSessionsIndex(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Delete conversation file
	convFile := projectPath + "/" + conv.SessionID + ".jsonl"
	if err := os.Remove(convFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Delete conversation directory if it exists
	convDir := projectPath + "/" + conv.SessionID
	if info, err := os.Stat(convDir); err == nil && info.IsDir() {
		if err := os.RemoveAll(convDir); err != nil {
			return fmt.Errorf("failed to delete directory: %w", err)
		}
	}

	// Remove from index and save
	RemoveSessionByID(index, conv.SessionID)
	if err := SaveSessionsIndex(projectPath, index); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	// Add tombstone if sync is initialized
	if syncutil.IsInitialized() {
		if err := AddTombstoneForProject(projectPath, conv.SessionID); err != nil {
			// Log but don't fail - tombstone is best-effort
			fmt.Printf("Warning: failed to add tombstone: %v\n", err)
		}
	}

	return nil
}

// stopSession stops a session's tmux session and removes its state file
func (m watchModel) stopSession(state *session.SessionState) error {
	// Kill tmux session if alive
	if session.IsTmuxSessionAlive(state.TmuxSession) {
		cmd := exec.Command("tmux", "kill-session", "-t", state.TmuxSession)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to kill tmux session: %w", err)
		}
	}

	// Remove state file
	if err := session.DeleteSessionState(state.ID); err != nil {
		return fmt.Errorf("failed to delete session state: %w", err)
	}

	return nil
}

func (m watchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle confirmation dialogs first
		if m.confirmMode != watchConfirmNone {
			switch msg.String() {
			case "y", "Y":
				if m.cursor < len(m.filtered) {
					conv := m.filtered[m.cursor]
					switch m.confirmMode {
					case watchConfirmAttachForce:
						m.selectedConv = &conv
						m.shouldCreate = false
						m.forceAttach = true
						m.confirmMode = watchConfirmNone
						return m, tea.Quit
					case watchConfirmDelete:
						// Delete conversation (no session)
						if err := m.deleteConversation(&conv); err != nil {
							m.statusMsg = "Error: " + err.Error()
						} else {
							m.statusMsg = "Deleted conversation " + conv.SessionID[:8]
						}
						m.confirmMode = watchConfirmNone
						m = m.loadConversations()
					case watchConfirmDeleteWithSession:
						// Stop session AND delete conversation
						if state, ok := m.activeSessions[conv.SessionID]; ok {
							if err := m.stopSession(state); err != nil {
								m.statusMsg = "Error stopping session: " + err.Error()
							}
						}
						if err := m.deleteConversation(&conv); err != nil {
							m.statusMsg = "Error: " + err.Error()
						} else {
							m.statusMsg = "Stopped session and deleted conversation " + conv.SessionID[:8]
						}
						m.confirmMode = watchConfirmNone
						m = m.loadConversations()
					}
				}
			case "s", "S":
				// Stop session only (when there's an active session)
				if m.confirmMode == watchConfirmDeleteWithSession && m.cursor < len(m.filtered) {
					conv := m.filtered[m.cursor]
					if state, ok := m.activeSessions[conv.SessionID]; ok {
						if err := m.stopSession(state); err != nil {
							m.statusMsg = "Error: " + err.Error()
						} else {
							m.statusMsg = "Stopped session " + state.ID
						}
					}
					m.confirmMode = watchConfirmNone
					m = m.refreshActiveSessions()
				}
			case "n", "N", "esc", " ":
				m.confirmMode = watchConfirmNone
			}
			// Any key dismisses the no-tmux message
			if m.confirmMode == watchConfirmNoTmux {
				m.confirmMode = watchConfirmNone
			}
			return m, nil
		}

		// Handle help view
		if m.helpView {
			m.helpView = false
			return m, nil
		}

		// Handle search mode
		if m.searchFocused {
			switch msg.String() {
			case "esc":
				if m.searchInput != "" {
					m.searchInput = ""
					m = m.applySearchFilter()
				} else {
					m.searchFocused = false
				}
			case "enter":
				m.searchFocused = false
			case "up":
				// Exit search and navigate up
				m.searchFocused = false
				if m.cursor > 0 {
					m.cursor--
					m = m.ensureCursorVisible()
				}
			case "down":
				// Exit search and navigate down
				m.searchFocused = false
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
					m = m.ensureCursorVisible()
				}
			case "backspace":
				if len(m.searchInput) > 0 {
					m.searchInput = m.searchInput[:len(m.searchInput)-1]
					m = m.applySearchFilter()
				}
			case "ctrl+u":
				m.searchInput = ""
				m = m.applySearchFilter()
			default:
				// Add printable characters to search
				if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] < 127 {
					m.searchInput += msg.String()
					m = m.applySearchFilter()
				}
			}
			return m, nil
		}

		// Normal mode
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.searchInput != "" {
				m.searchInput = ""
				m = m.applySearchFilter()
			}
		case "/":
			m.searchFocused = true
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m = m.ensureCursorVisible()
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m = m.ensureCursorVisible()
			}
		case "pgup", "ctrl+b":
			m.cursor -= m.viewportHeight
			if m.cursor < 0 {
				m.cursor = 0
			}
			m = m.ensureCursorVisible()
		case "pgdown", "ctrl+f":
			m.cursor += m.viewportHeight
			if m.cursor >= len(m.filtered) {
				m.cursor = len(m.filtered) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
			m = m.ensureCursorVisible()
		case "home", "g":
			m.cursor = 0
			m = m.ensureCursorVisible()
		case "end", "G":
			if len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1
			}
			m = m.ensureCursorVisible()
		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				conv := m.filtered[m.cursor]
				// Check if session already exists for this conversation
				if existing, ok := m.activeSessions[conv.SessionID]; ok {
					tmuxAlive := existing.TmuxSession != "" && session.IsTmuxSessionAlive(existing.TmuxSession)
					if !tmuxAlive {
						// Non-tmux or dead tmux session, cannot attach
						m.confirmMode = watchConfirmNoTmux
					} else if existing.Attached > 0 {
						m.confirmMode = watchConfirmAttachForce
					} else {
						m.selectedConv = &conv
						m.shouldCreate = false // Just attach to existing
						return m, tea.Quit
					}
				} else {
					m.selectedConv = &conv
					m.shouldCreate = true // Create new session
					return m, tea.Quit
				}
			}
		case "r":
			// Force refresh conversations
			m = m.loadConversations()
			m.statusMsg = ""
		case "h", "?":
			m.helpView = true
		case "delete", "backspace", "x":
			// Delete conversation (with confirmation)
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				conv := m.filtered[m.cursor]
				if _, hasSession := m.activeSessions[conv.SessionID]; hasSession {
					m.confirmMode = watchConfirmDeleteWithSession
				} else {
					m.confirmMode = watchConfirmDelete
				}
				m.statusMsg = ""
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate viewport height (terminal height - search box - header - separator - scroll - footer)
		m.viewportHeight = max(msg.Height-10, 5)
		m = m.ensureCursorVisible()

	case watchTickMsg:
		// Only refresh session state (lightweight), not full conversation list
		m = m.refreshActiveSessions()
		return m, watchTickCmd()

	case watchFileChangeMsg:
		// Session state file changed - update just that session
		m = m.updateSingleSession(msg.sessionID, msg.deleted)
		return m, watchSessionFilesCmd() // Restart watcher for next change
	}

	return m, nil
}

func (m watchModel) View() string {
	// Help view overlay
	if m.helpView {
		return m.renderHelpView()
	}

	var b strings.Builder

	// Search box
	b.WriteString("\n  ")
	if m.searchFocused {
		b.WriteString(wSearchStyle.Render("Search: "))
		b.WriteString(wSearchStyle.Render("[" + m.searchInput + "_]"))
	} else if m.searchInput != "" {
		b.WriteString(wSearchStyle.Render("Search: [" + m.searchInput + "]"))
	} else {
		b.WriteString(wHelpStyle.Render("/ to search"))
	}

	// Show count
	if len(m.filtered) != len(m.entries) {
		b.WriteString(wHelpStyle.Render(fmt.Sprintf("  [showing %d of %d]", len(m.filtered), len(m.entries))))
	} else if len(m.entries) > 0 {
		b.WriteString(wHelpStyle.Render(fmt.Sprintf("  [%d conversations]", len(m.entries))))
	}
	b.WriteString("\n\n")

	// Show empty state
	if len(m.filtered) == 0 {
		if len(m.entries) == 0 {
			b.WriteString("  No conversations found\n")
		} else {
			b.WriteString("  No matches for \"" + m.searchInput + "\"\n")
		}
		b.WriteString("\n")
		b.WriteString(wHelpStyle.Render("  r refresh • / search • q quit"))
		b.WriteString("\n")
		return b.String()
	}

	// Header - calculate dynamic title width based on terminal width
	termWidth := m.width
	if termWidth < 80 {
		termWidth = 120
	}

	// Fixed column widths: sessionMark(2) + id(10) + branch(15) + modified(16) + spaces(4) = 47
	// Global adds: project(30) + space(1) = 31 more
	// Title gets the rest
	var fixedWidth int
	if m.global {
		fixedWidth = 3 + 10 + 1 + 30 + 1 + 1 + 15 + 1 + 16 // 78 (icon + id + project + title margin + branch + modified)
	} else {
		fixedWidth = 3 + 10 + 1 + 1 + 15 + 1 + 16 // 47 (icon + id + title margin + branch + modified)
	}
	titleWidth := termWidth - fixedWidth - 2 // -2 for some margin
	if titleWidth < 30 {
		titleWidth = 30
	}
	if titleWidth > 80 {
		titleWidth = 80
	}

	var header string
	rowWidth := fixedWidth + titleWidth
	if m.global {
		header = fmt.Sprintf("   %-10s %-30s %-*s %-15s %-16s", "ID", "PROJECT", titleWidth, "TITLE/PROMPT", "BRANCH", "MODIFIED")
	} else {
		header = fmt.Sprintf("   %-10s %-*s %-15s %-16s", "ID", titleWidth, "TITLE/PROMPT", "BRANCH", "MODIFIED")
	}
	b.WriteString(wHeaderStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(wHeaderStyle.Render(strings.Repeat("─", rowWidth)))
	b.WriteString("\n")

	// Render visible rows only
	endIdx := min(m.viewportOffset+m.viewportHeight, len(m.filtered))

	for i := m.viewportOffset; i < endIdx; i++ {
		e := m.filtered[i]

		// Session indicator
		// ⚡ = tmux session with attached clients
		// ○ = tmux session, detached
		// ◉ = non-tmux or dead tmux session (in-terminal, can't attach)
		sessionMark := "  "
		if state, ok := m.activeSessions[e.SessionID]; ok {
			tmuxAlive := state.TmuxSession != "" && session.IsTmuxSessionAlive(state.TmuxSession)
			if !tmuxAlive {
				sessionMark = " " + wNoTmuxStyle.Render("◉")
			} else if state.Attached > 0 {
				sessionMark = wActiveStyle.Render("⚡")
			} else {
				sessionMark = " " + wDetachedStyle.Render("▷")
			}
		}

		// Format row
		id := e.SessionID[:8]
		branch := e.GitBranch
		if len(branch) > 15 {
			branch = branch[:14] + "…"
		}
		modified := formatDate(e.Modified)

		// Use dynamic title width (titleWidth calculated above)
		title := e.DisplayTitle()
		if title == "" {
			title = truncatePrompt(e.FirstPrompt, titleWidth-2)
		} else {
			title = truncatePrompt("["+title+"]", titleWidth-2)
		}

		var row string
		if m.global {
			project := shortenPath(e.ProjectPath, 28)
			row = fmt.Sprintf("%s %-10s %-30s %-*s %-15s %s", sessionMark, id, project, titleWidth, title, branch, modified)
		} else {
			row = fmt.Sprintf("%s %-10s %-*s %-15s %s", sessionMark, id, titleWidth, title, branch, modified)
		}

		if i == m.cursor {
			b.WriteString(wSelectedStyle.Render(row))
		} else {
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.filtered) > m.viewportHeight {
		scrollPct := float64(m.viewportOffset) / float64(len(m.filtered)-m.viewportHeight) * 100
		b.WriteString(wHelpStyle.Render(fmt.Sprintf("  ── scroll: %.0f%% ──", scrollPct)))
		b.WriteString("\n")
	}

	// Footer / confirmation dialog / status message
	b.WriteString("\n")
	switch m.confirmMode {
	case watchConfirmAttachForce:
		b.WriteString(wConfirmStyle.Render("  Session already attached. Detach others? [y/n]"))
	case watchConfirmDelete:
		b.WriteString(wConfirmStyle.Render("  Delete conversation? [y/n]"))
	case watchConfirmDeleteWithSession:
		b.WriteString(wConfirmStyle.Render("  Has active session. Delete+stop (y), stop only (s), cancel (n)?"))
	case watchConfirmNoTmux:
		b.WriteString(wConfirmStyle.Render("  Session was started outside tofu/tmux (◉) - already in its terminal. [press any key]"))
	default:
		if m.statusMsg != "" {
			b.WriteString(wSearchStyle.Render("  " + m.statusMsg))
		} else {
			b.WriteString(wHelpStyle.Render("  h help • ↑/↓ navigate • enter attach • del delete • q quit"))
		}
	}
	b.WriteString("\n")

	return b.String()
}

func (m watchModel) renderHelpView() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(wSearchStyle.Render("  Conversation Watch - Keyboard Shortcuts"))
	b.WriteString("\n\n")

	b.WriteString(wHeaderStyle.Render("  Navigation"))
	b.WriteString("\n")
	b.WriteString("    ↑/k       Move cursor up\n")
	b.WriteString("    ↓/j       Move cursor down\n")
	b.WriteString("    PgUp/^B   Page up\n")
	b.WriteString("    PgDn/^F   Page down\n")
	b.WriteString("    g/Home    Go to first\n")
	b.WriteString("    G/End     Go to last\n")
	b.WriteString("    enter     Create session or attach to existing\n")
	b.WriteString("    q/esc     Quit watch mode\n")
	b.WriteString("\n")

	b.WriteString(wHeaderStyle.Render("  Search"))
	b.WriteString("\n")
	b.WriteString("    /         Start search\n")
	b.WriteString("    esc       Clear search / exit search mode\n")
	b.WriteString("    ^U        Clear search input\n")
	b.WriteString("\n")

	b.WriteString(wHeaderStyle.Render("  Actions"))
	b.WriteString("\n")
	b.WriteString("    del/x     Delete conversation (with confirmation)\n")
	b.WriteString("              If has session: y=delete+stop, s=stop only, n=cancel\n")
	b.WriteString("    r         Refresh conversation list\n")
	b.WriteString("\n")

	b.WriteString(wHeaderStyle.Render("  Indicators"))
	b.WriteString("\n")
	b.WriteString("    ⚡        Tmux session with attached clients\n")
	b.WriteString("    ▷         Tmux session, detached (can attach)\n")
	b.WriteString("    ◉         Non-tmux session (in-terminal, can't attach)\n")
	b.WriteString("\n")

	b.WriteString(wHelpStyle.Render("  Press any key to close"))
	b.WriteString("\n")
	return b.String()
}

// WatchResult holds the result of the watch mode selection
type WatchResult struct {
	Conv        *SessionEntry
	ShouldCreate bool  // true = create new session, false = attach to existing
	ForceAttach  bool  // Detach other clients when attaching
}

// ConvWatchState holds state that persists between attach cycles
type ConvWatchState struct {
	SearchInput    string
	Cursor         int
	ViewportOffset int
}

// RunConvWatch runs the interactive watch mode and returns the result
func RunConvWatch(global bool, since, before string, state ConvWatchState) (WatchResult, ConvWatchState, error) {
	m := initialWatchModel(global, since, before)

	// Restore previous state
	m.searchInput = state.SearchInput
	m.cursor = state.Cursor
	m.viewportOffset = state.ViewportOffset

	m = m.loadConversations()

	// Ensure cursor is still valid after loading (list may have changed)
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	m = m.ensureCursorVisible()

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return WatchResult{}, state, err
	}

	fm := finalModel.(watchModel)
	newState := ConvWatchState{
		SearchInput:    fm.searchInput,
		Cursor:         fm.cursor,
		ViewportOffset: fm.viewportOffset,
	}
	return WatchResult{
		Conv:         fm.selectedConv,
		ShouldCreate: fm.shouldCreate,
		ForceAttach:  fm.forceAttach,
	}, newState, nil
}

// RunConvWatchMode runs the interactive watch mode with create/attach loop
func RunConvWatchMode(global bool, since, before string) error {
	var watchState ConvWatchState
	for {
		result, newState, err := RunConvWatch(global, since, before, watchState)
		watchState = newState // Preserve state between cycles
		if err != nil {
			return err
		}

		if result.Conv == nil {
			// User quit without selecting
			return nil
		}

		if result.ShouldCreate {
			// Create a new session for this conversation
			if err := createSessionForConv(result.Conv); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating session: %v\n", err)
				// Continue back to watch mode
				continue
			}
		} else {
			// Attach to existing session
			sessState := findSessionForConv(result.Conv.SessionID)
			if sessState == nil {
				fmt.Fprintf(os.Stderr, "Session not found, creating new...\n")
				if err := createSessionForConv(result.Conv); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating session: %v\n", err)
					continue
				}
			} else {
				if result.ForceAttach {
					fmt.Printf("Attaching to %s (detaching others)... (Ctrl+B D to detach)\n", sessState.ID)
				} else {
					fmt.Printf("Attaching to %s... (Ctrl+B D to detach)\n", sessState.ID)
				}

				if err := attachToTmuxSession(sessState.TmuxSession, result.ForceAttach); err != nil {
					// Session may have exited, continue to watch mode
					continue
				}
			}
		}

		// After detach or session end, loop back to watch mode
	}
}

// createSessionForConv creates a new session for a conversation
func createSessionForConv(conv *SessionEntry) error {
	if err := session.CheckTmuxInstalled(); err != nil {
		return err
	}

	// Ensure hooks are installed
	session.EnsureHooksInstalled(false, os.Stdout, os.Stderr)

	cwd := conv.ProjectPath
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Use conv ID prefix as session ID
	sessionID := conv.SessionID[:8]
	tmuxSession := "tofu-claude-" + sessionID

	// Build claude command with TOFU_SESSION_ID env var
	claudeCmd := fmt.Sprintf("TOFU_SESSION_ID=%s claude --resume %s", sessionID, conv.SessionID)

	// Create tmux session
	tmuxArgs := []string{
		"new-session",
		"-d",
		"-s", tmuxSession,
		"-c", cwd,
		"sh", "-c", claudeCmd,
	}

	cmd := exec.Command("tmux", tmuxArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	// Get PID and save state
	pid := session.ParsePIDFromTmux(tmuxSession)
	state := &session.SessionState{
		ID:          sessionID,
		TmuxSession: tmuxSession,
		PID:         pid,
		Cwd:         cwd,
		ConvID:      conv.SessionID,
		Status:      session.StatusIdle,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	if err := session.SaveSessionState(state); err != nil {
		return fmt.Errorf("failed to save session state: %w", err)
	}

	fmt.Printf("Created session %s\n", sessionID)
	fmt.Println("Attaching... (Ctrl+B D to detach)")

	return attachToTmuxSession(tmuxSession, false)
}

// findSessionForConv finds an existing session for a conversation ID
func findSessionForConv(convID string) *session.SessionState {
	states, _ := session.ListSessionStates()
	for _, state := range states {
		session.RefreshSessionStatus(state)
		if state.Status != session.StatusExited && state.ConvID == convID {
			return state
		}
	}
	return nil
}

// attachToTmuxSession attaches to a tmux session
func attachToTmuxSession(tmuxSession string, forceDetach bool) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	// Run tmux attach as a subprocess (not exec, so we return here after detach)
	args := []string{"attach-session", "-t", tmuxSession}
	if forceDetach {
		args = []string{"attach-session", "-d", "-t", tmuxSession}
	}

	cmd := exec.Command(tmuxPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() != 0 {
					// Session might have ended, return without error to go back to watch
					return nil
				}
			}
		}
	}

	return nil
}
