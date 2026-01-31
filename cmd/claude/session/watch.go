package session

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("62"))
	idleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))     // Yellow
	workingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))      // Green
	needsInput    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))     // Bright red
	exitedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))     // Gray
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	searchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

type tickMsg time.Time

// Confirmation modes
type confirmMode int

const (
	confirmNone confirmMode = iota
	confirmKill
	confirmAttachForce // Session already attached, confirm force attach
)

// Filter options for the checkbox menu
var filterOptions = []struct {
	key    string
	label  string
	status string // internal status constant (empty for "all")
}{
	{"all", "All (no filter)", ""},
	{StatusIdle, "Idle", StatusIdle},
	{StatusWorking, "Working", StatusWorking},
	{StatusAwaitingPermission, "Awaiting permission", StatusAwaitingPermission},
	{StatusAwaitingInput, "Awaiting input", StatusAwaitingInput},
	{StatusExited, "Exited", StatusExited},
}

type model struct {
	allSessions   []*SessionState // all sessions before search filter
	sessions      []*SessionState // sessions after search filter
	cursor        int
	width         int
	height        int
	shouldAttach  string // session ID to attach to after quitting
	forceAttach   bool   // detach other clients when attaching
	includeAll    bool
	sort          SortState
	statusFilter  []string        // which statuses to show (empty = all)
	hideFilter    []string        // which statuses to hide
	confirmMode   confirmMode     // current confirmation dialog
	filterMenu    bool            // showing filter menu
	filterCursor  int             // cursor position in filter menu
	filterChecked map[string]bool // checked items in filter menu
	helpView      bool            // showing help view
	searchInput   string          // current search query
	searchFocused bool            // whether search box is focused
}

func initialModel(includeAll bool, statusFilter, hideFilter []string) model {
	return model{
		sessions:     []*SessionState{},
		cursor:       0,
		includeAll:   includeAll,
		sort:         SortState{Column: SortNone},
		statusFilter: statusFilter,
		hideFilter:   hideFilter,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), tea.EnterAltScreen)
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) refreshSessions() model {
	states, err := ListSessionStates()
	if err != nil {
		return m
	}

	var filtered []*SessionState
	for _, state := range states {
		RefreshSessionStatus(state)
		if !m.includeAll && state.Status == StatusExited {
			continue
		}
		// Apply status filter (show)
		if !m.matchesShowFilter(state.Status) {
			continue
		}
		// Apply hide filter
		if m.matchesHideFilter(state.Status) {
			continue
		}
		filtered = append(filtered, state)
	}

	// Apply sorting
	SortSessions(filtered, m.sort)
	m.allSessions = filtered

	// Apply search filter
	m = m.applySearchFilter()

	return m
}

// applySearchFilter filters sessions based on search input
func (m model) applySearchFilter() model {
	if m.searchInput == "" {
		m.sessions = m.allSessions
	} else {
		query := strings.ToLower(m.searchInput)
		var filtered []*SessionState
		for _, s := range m.allSessions {
			if sessionMatchesSearch(s, query) {
				filtered = append(filtered, s)
			}
		}
		m.sessions = filtered
	}

	// Keep cursor in bounds
	if m.cursor >= len(m.sessions) {
		m.cursor = len(m.sessions) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	return m
}

// sessionMatchesSearch checks if a session matches the search query
func sessionMatchesSearch(s *SessionState, query string) bool {
	return strings.Contains(strings.ToLower(s.ID), query) ||
		strings.Contains(strings.ToLower(s.Cwd), query) ||
		strings.Contains(strings.ToLower(s.Status), query) ||
		strings.Contains(strings.ToLower(s.StatusDetail), query) ||
		strings.Contains(strings.ToLower(s.ConvID), query)
}

// matchesShowFilter checks if a status matches the show filter
func (m model) matchesShowFilter(status string) bool {
	if len(m.statusFilter) == 0 {
		return true // no filter = show all
	}
	for _, f := range m.statusFilter {
		if f == "all" {
			return true
		}
		if f == status {
			return true
		}
		// Handle grouped filters
		if f == "attention" && (status == StatusAwaitingPermission || status == StatusAwaitingInput) {
			return true
		}
	}
	return false
}

// matchesHideFilter checks if a status should be hidden
func (m model) matchesHideFilter(status string) bool {
	if len(m.hideFilter) == 0 {
		return false // no filter = hide nothing
	}
	for _, f := range m.hideFilter {
		if f == status {
			return true
		}
		// Handle grouped filters
		if f == "attention" && (status == StatusAwaitingPermission || status == StatusAwaitingInput) {
			return true
		}
	}
	return false
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle confirmation dialogs first
		if m.confirmMode != confirmNone {
			switch msg.String() {
			case "y", "Y":
				if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
					if m.confirmMode == confirmKill {
						state := m.sessions[m.cursor]
						_ = killSession(state)
						m.confirmMode = confirmNone
						m = m.refreshSessions()
					} else if m.confirmMode == confirmAttachForce {
						m.shouldAttach = m.sessions[m.cursor].TmuxSession
						m.forceAttach = true
						m.confirmMode = confirmNone
						return m, tea.Quit
					}
				}
			case "n", "N", "esc", "q":
				m.confirmMode = confirmNone
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
				}
			case "down":
				// Exit search and navigate down
				m.searchFocused = false
				if m.cursor < len(m.sessions)-1 {
					m.cursor++
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

		// Handle filter menu
		if m.filterMenu {
			switch msg.String() {
			case "esc", "q":
				m.filterMenu = false
			case "up", "k":
				if m.filterCursor > 0 {
					m.filterCursor--
				}
			case "down", "j":
				if m.filterCursor < len(filterOptions)-1 {
					m.filterCursor++
				}
			case " ", "x":
				// Toggle checkbox
				key := filterOptions[m.filterCursor].key
				if key == "all" {
					// "All" clears all other selections
					m.filterChecked = make(map[string]bool)
					m.filterChecked["all"] = true
				} else {
					// Toggle this option, clear "all" if set
					delete(m.filterChecked, "all")
					m.filterChecked[key] = !m.filterChecked[key]
					if !m.filterChecked[key] {
						delete(m.filterChecked, key)
					}
					// If nothing selected, select "all"
					if len(m.filterChecked) == 0 {
						m.filterChecked["all"] = true
					}
				}
			case "enter", "f":
				// Apply filter and close
				m.statusFilter = nil
				if !m.filterChecked["all"] {
					for key, checked := range m.filterChecked {
						if checked && key != "all" {
							m.statusFilter = append(m.statusFilter, key)
						}
					}
				}
				// Enable includeAll if exited is selected
				if m.filterChecked[StatusExited] {
					m.includeAll = true
				}
				m.filterMenu = false
				m = m.refreshSessions()
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
			} else {
				return m, tea.Quit
			}
		case "/":
			m.searchFocused = true
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
				state := m.sessions[m.cursor]
				if state.Attached > 0 {
					// Session already has clients attached, ask for confirmation
					m.confirmMode = confirmAttachForce
				} else {
					m.shouldAttach = state.TmuxSession
					return m, tea.Quit
				}
			}
		case "delete", "backspace", "x":
			if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
				m.confirmMode = confirmKill
			}
		case "f":
			// Initialize filter checkboxes from current filter state
			m.filterChecked = make(map[string]bool)
			if len(m.statusFilter) == 0 {
				m.filterChecked["all"] = true
			} else {
				for _, f := range m.statusFilter {
					m.filterChecked[f] = true
				}
			}
			m.filterCursor = 0
			m.filterMenu = true
		case "h", "?":
			m.helpView = true
		case "r":
			// Force refresh
			m = m.refreshSessions()
		case "f1", "1":
			m.sort.Toggle(SortID)
			m = m.refreshSessions()
		case "f2", "2":
			m.sort.Toggle(SortDirectory)
			m = m.refreshSessions()
		case "f3", "3":
			m.sort.Toggle(SortStatus)
			m = m.refreshSessions()
		case "f4", "4":
			m.sort.Toggle(SortAge)
			m = m.refreshSessions()
		case "f5", "5":
			m.sort.Toggle(SortUpdated)
			m = m.refreshSessions()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m = m.refreshSessions()
		return m, tickCmd()
	}

	return m, nil
}

var (
	confirmStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	menuStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	filterBadge  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

func (m model) View() string {
	// Help view overlay
	if m.helpView {
		return m.renderHelpView()
	}

	// Filter menu overlay
	if m.filterMenu {
		return m.renderFilterMenu()
	}

	var b strings.Builder

	// Search box
	b.WriteString("\n  ")
	if m.searchFocused {
		b.WriteString(searchStyle.Render("Search: "))
		b.WriteString(searchStyle.Render("[" + m.searchInput + "_]"))
	} else if m.searchInput != "" {
		b.WriteString(searchStyle.Render("Search: [" + m.searchInput + "]"))
	} else {
		b.WriteString(helpStyle.Render("/ to search"))
	}

	// Show count
	if len(m.sessions) != len(m.allSessions) {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  [showing %d of %d]", len(m.sessions), len(m.allSessions))))
	} else if len(m.allSessions) > 0 {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  [%d sessions]", len(m.allSessions))))
	}
	b.WriteString("\n\n")

	// Show empty state
	if len(m.sessions) == 0 {
		if len(m.allSessions) == 0 {
			b.WriteString("  No active sessions")
			if len(m.statusFilter) > 0 {
				b.WriteString(fmt.Sprintf(" (filter: %s)", strings.Join(m.statusFilter, ", ")))
			}
			b.WriteString("\n")
		} else {
			b.WriteString("  No matches for \"" + m.searchInput + "\"\n")
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  / search • f filter • r refresh • q quit"))
		b.WriteString("\n")
		return b.String()
	}

	// Header with sort indicators and filter badge
	idHdr := "ID" + m.sort.Indicator(SortID)
	dirHdr := "DIRECTORY" + m.sort.Indicator(SortDirectory)
	statusHdr := "STATUS" + m.sort.Indicator(SortStatus)
	ageHdr := "AGE" + m.sort.Indicator(SortAge)
	updatedHdr := "UPDATED" + m.sort.Indicator(SortUpdated)
	header := fmt.Sprintf("  %-12s %-35s %-22s %-10s %s", idHdr, dirHdr, statusHdr, ageHdr, updatedHdr)
	b.WriteString(headerStyle.Render(header))

	// Show filter badge
	if len(m.statusFilter) > 0 {
		b.WriteString(filterBadge.Render(fmt.Sprintf("  [%s]", strings.Join(m.statusFilter, ", "))))
	}
	b.WriteString("\n")

	width := m.width
	if width < 10 {
		width = 90 // default width
	}
	b.WriteString(headerStyle.Render(strings.Repeat("─", min(width-2, 90))))
	b.WriteString("\n")

	// Rows
	for i, state := range m.sessions {
		status := state.Status
		if state.StatusDetail != "" {
			status = status + ": " + state.StatusDetail
		}

		// Add attached indicator
		attachedMark := "  "
		if state.Attached > 0 {
			attachedMark = "⚡"
		}

		dir := shortenPathForTable(state.Cwd, 33)
		age := FormatDuration(time.Since(state.Created))
		updated := FormatDuration(time.Since(state.Updated))

		row := fmt.Sprintf("%s %-10s %-35s %-22s %-10s %s", attachedMark, state.ID, dir, status, age, updated)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render(row))
		} else {
			style := getRowStyle(state.Status)
			b.WriteString(style.Render(row))
		}
		b.WriteString("\n")
	}

	// Confirmation dialog or help hint
	b.WriteString("\n")
	if m.cursor < len(m.sessions) {
		switch m.confirmMode {
		case confirmKill:
			b.WriteString(confirmStyle.Render(fmt.Sprintf("  Kill session %s? [y/n]", m.sessions[m.cursor].ID)))
		case confirmAttachForce:
			b.WriteString(confirmStyle.Render(fmt.Sprintf("  Session %s already attached. Detach other clients? [y/n]", m.sessions[m.cursor].ID)))
		default:
			b.WriteString(helpStyle.Render("  h help • / search • ↑/↓ navigate • enter attach • q quit"))
		}
	} else {
		b.WriteString(helpStyle.Render("  h help • / search • ↑/↓ navigate • enter attach • q quit"))
	}
	b.WriteString("\n")

	return b.String()
}

func (m model) renderFilterMenu() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(menuStyle.Render("  Filter by status:"))
	b.WriteString("\n\n")

	for i, opt := range filterOptions {
		// Cursor indicator
		if i == m.filterCursor {
			b.WriteString(selectedStyle.Render(" >"))
		} else {
			b.WriteString("  ")
		}

		// Checkbox
		if m.filterChecked[opt.key] {
			b.WriteString(" [x] ")
		} else {
			b.WriteString(" [ ] ")
		}

		// Label
		if i == m.filterCursor {
			b.WriteString(selectedStyle.Render(opt.label))
		} else {
			b.WriteString(opt.label)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  ↑/↓ navigate • space toggle • enter apply • esc cancel"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderHelpView() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(menuStyle.Render("  Session Watch - Keyboard Shortcuts"))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render("  Navigation"))
	b.WriteString("\n")
	b.WriteString("    ↑/k       Move cursor up\n")
	b.WriteString("    ↓/j       Move cursor down\n")
	b.WriteString("    enter     Attach to selected session\n")
	b.WriteString("    q/esc     Quit watch mode\n")
	b.WriteString("\n")

	b.WriteString(headerStyle.Render("  Search"))
	b.WriteString("\n")
	b.WriteString("    /         Start search\n")
	b.WriteString("    esc       Clear search / exit search mode\n")
	b.WriteString("    ^U        Clear search input\n")
	b.WriteString("\n")

	b.WriteString(headerStyle.Render("  Actions"))
	b.WriteString("\n")
	b.WriteString("    del/x     Kill selected session (with confirmation)\n")
	b.WriteString("    r         Refresh session list\n")
	b.WriteString("\n")

	b.WriteString(headerStyle.Render("  Filtering"))
	b.WriteString("\n")
	b.WriteString("    f         Open filter menu\n")
	b.WriteString("\n")

	b.WriteString(headerStyle.Render("  Sorting"))
	b.WriteString("\n")
	b.WriteString("    1/F1      Sort by ID\n")
	b.WriteString("    2/F2      Sort by Directory\n")
	b.WriteString("    3/F3      Sort by Status\n")
	b.WriteString("    4/F4      Sort by Age\n")
	b.WriteString("    5/F5      Sort by Updated\n")
	b.WriteString("              (press again to toggle asc/desc/off)\n")
	b.WriteString("\n")

	b.WriteString(helpStyle.Render("  Press any key to close"))
	b.WriteString("\n")
	return b.String()
}

func getRowStyle(status string) lipgloss.Style {
	switch status {
	case StatusIdle:
		return idleStyle
	case StatusWorking:
		return workingStyle
	case StatusAwaitingPermission, StatusAwaitingInput:
		return needsInput
	case StatusExited:
		return exitedStyle
	default:
		return needsInput
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// WatchState holds state that persists between attach cycles
type WatchState struct {
	Sort         SortState
	StatusFilter []string
	HideFilter   []string
	SearchInput  string
	Cursor       int
}

// AttachResult holds the result of selecting a session to attach
type AttachResult struct {
	TmuxSession string
	ForceAttach bool // true if we should detach other clients
}

// RunInteractive starts the interactive session viewer
// Returns the attach result (if any) and the final watch state
func RunInteractive(includeAll bool, state WatchState) (AttachResult, WatchState, error) {
	m := initialModel(includeAll, state.StatusFilter, state.HideFilter)
	m.sort = state.Sort
	m.searchInput = state.SearchInput
	m.cursor = state.Cursor
	m = m.refreshSessions()

	// Ensure cursor is still valid after loading
	if m.cursor >= len(m.sessions) {
		m.cursor = max(0, len(m.sessions)-1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return AttachResult{}, state, err
	}

	fm := finalModel.(model)
	result := AttachResult{
		TmuxSession: fm.shouldAttach,
		ForceAttach: fm.forceAttach,
	}
	newState := WatchState{
		Sort:         fm.sort,
		StatusFilter: fm.statusFilter,
		HideFilter:   fm.hideFilter,
		SearchInput:  fm.searchInput,
		Cursor:       fm.cursor,
	}
	return result, newState, nil
}

// RunWatchMode runs the interactive watch mode with attach support
func RunWatchMode(includeAll bool, initialSort SortState, initialFilter, initialHide []string) error {
	state := WatchState{Sort: initialSort, StatusFilter: initialFilter, HideFilter: initialHide}
	for {
		result, newState, err := RunInteractive(includeAll, state)
		state = newState // Preserve state between attach cycles
		if err != nil {
			return err
		}

		if result.TmuxSession == "" {
			// User quit without selecting
			return nil
		}

		// Attach to the session
		if result.ForceAttach {
			fmt.Printf("Attaching to %s (detaching others)... (Ctrl+B D to detach)\n", result.TmuxSession)
		} else {
			fmt.Printf("Attaching to %s... (Ctrl+B D to detach)\n", result.TmuxSession)
		}

		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return fmt.Errorf("tmux not found: %w", err)
		}

		// Run tmux attach as a subprocess (not exec, so we return here after detach)
		// Use -d flag to detach other clients if force attach was requested
		args := []string{"attach-session", "-t", result.TmuxSession}
		if result.ForceAttach {
			args = []string{"attach-session", "-d", "-t", result.TmuxSession}
		}
		cmd := exec.Command(tmuxPath, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			// Check if it's just a normal detach (exit code 0 or specific tmux codes)
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					// tmux returns various codes, but we want to continue the loop
					// unless it's a real error
					if status.ExitStatus() != 0 {
						// Session might have ended, continue to interactive view
						continue
					}
				}
			}
		}

		// After detach or session end, loop back to interactive view
	}
}
