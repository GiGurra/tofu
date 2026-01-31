package conv

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gigurra/tofu/cmd/claude/session"
)

var (
	wSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("62"))
	wHeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	wHelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	wSearchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	wConfirmStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	wActiveStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Green for active session indicator
)

type watchTickMsg time.Time

// watchConfirmMode represents confirmation dialogs
type watchConfirmMode int

const (
	watchConfirmNone watchConfirmMode = iota
	watchConfirmAttachForce // Session already attached, confirm force attach
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
	return tea.Batch(watchTickCmd(), tea.EnterAltScreen)
}

func watchTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return watchTickMsg(t)
	})
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

func (m watchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle confirmation dialogs first
		if m.confirmMode != watchConfirmNone {
			switch msg.String() {
			case "y", "Y":
				if m.confirmMode == watchConfirmAttachForce && m.cursor < len(m.filtered) {
					m.selectedConv = &m.filtered[m.cursor]
					m.shouldCreate = false
					m.forceAttach = true
					m.confirmMode = watchConfirmNone
					return m, tea.Quit
				}
			case "n", "N", "esc", "q":
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
					if existing.Attached > 0 {
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
		case "h", "?":
			m.helpView = true
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate viewport height (terminal height - header - search - footer - padding)
		m.viewportHeight = max(msg.Height-7, 5)
		m = m.ensureCursorVisible()

	case watchTickMsg:
		// Only refresh session state (lightweight), not full conversation list
		m = m.refreshActiveSessions()
		return m, watchTickCmd()
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

	// Header
	width := m.width
	if width < 10 {
		width = 120
	}

	var header string
	if m.global {
		header = fmt.Sprintf("  %-10s %-30s %-40s %-15s %s", "ID", "PROJECT", "TITLE/PROMPT", "BRANCH", "MODIFIED")
	} else {
		header = fmt.Sprintf("  %-10s %-50s %-15s %s", "ID", "TITLE/PROMPT", "BRANCH", "MODIFIED")
	}
	b.WriteString(wHeaderStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(wHeaderStyle.Render(strings.Repeat("─", min(width-2, 100))))
	b.WriteString("\n")

	// Render visible rows only
	endIdx := min(m.viewportOffset+m.viewportHeight, len(m.filtered))

	for i := m.viewportOffset; i < endIdx; i++ {
		e := m.filtered[i]

		// Session indicator
		sessionMark := "  "
		if state, ok := m.activeSessions[e.SessionID]; ok {
			if state.Attached > 0 {
				sessionMark = wActiveStyle.Render("⚡")
			} else {
				sessionMark = wActiveStyle.Render("○ ")
			}
		}

		// Format row
		id := e.SessionID[:8]
		title := e.DisplayTitle()
		if title == "" {
			title = truncatePrompt(e.FirstPrompt, 48)
		} else {
			title = truncatePrompt("["+title+"]", 48)
		}
		branch := e.GitBranch
		if len(branch) > 15 {
			branch = branch[:14] + "…"
		}
		modified := formatDate(e.Modified)

		var row string
		if m.global {
			project := shortenPath(e.ProjectPath, 28)
			row = fmt.Sprintf("%s%-10s %-30s %-40s %-15s %s", sessionMark, id, project, title, branch, modified)
		} else {
			row = fmt.Sprintf("%s%-10s %-50s %-15s %s", sessionMark, id, title, branch, modified)
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

	// Footer / confirmation dialog
	b.WriteString("\n")
	if m.confirmMode == watchConfirmAttachForce && m.cursor < len(m.filtered) {
		b.WriteString(wConfirmStyle.Render("  Session already attached. Detach others? [y/n]"))
	} else {
		b.WriteString(wHelpStyle.Render("  h help • ↑/↓ navigate • enter create/attach • r refresh • q quit"))
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
	b.WriteString("    r         Refresh conversation list\n")
	b.WriteString("\n")

	b.WriteString(wHeaderStyle.Render("  Indicators"))
	b.WriteString("\n")
	b.WriteString("    ⚡        Conversation has attached session\n")
	b.WriteString("    ○         Conversation has active session (not attached)\n")
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

// RunConvWatch runs the interactive watch mode and returns the result
func RunConvWatch(global bool, since, before string) (WatchResult, error) {
	m := initialWatchModel(global, since, before)
	m = m.loadConversations()

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return WatchResult{}, err
	}

	fm := finalModel.(watchModel)
	return WatchResult{
		Conv:         fm.selectedConv,
		ShouldCreate: fm.shouldCreate,
		ForceAttach:  fm.forceAttach,
	}, nil
}

// RunConvWatchMode runs the interactive watch mode with create/attach loop
func RunConvWatchMode(global bool, since, before string) error {
	for {
		result, err := RunConvWatch(global, since, before)
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
			state := findSessionForConv(result.Conv.SessionID)
			if state == nil {
				fmt.Fprintf(os.Stderr, "Session not found, creating new...\n")
				if err := createSessionForConv(result.Conv); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating session: %v\n", err)
					continue
				}
			} else {
				if result.ForceAttach {
					fmt.Printf("Attaching to %s (detaching others)... (Ctrl+B D to detach)\n", state.ID)
				} else {
					fmt.Printf("Attaching to %s... (Ctrl+B D to detach)\n", state.ID)
				}

				if err := attachToTmuxSession(state.TmuxSession, result.ForceAttach); err != nil {
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
