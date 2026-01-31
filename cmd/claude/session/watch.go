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
)

type tickMsg time.Time

type model struct {
	sessions     []*SessionState
	cursor       int
	width        int
	height       int
	shouldAttach string // session ID to attach to after quitting
	includeAll   bool
	sort         SortState
}

func initialModel(includeAll bool) model {
	return model{
		sessions:   []*SessionState{},
		cursor:     0,
		includeAll: includeAll,
		sort:       SortState{Column: SortNone},
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
		filtered = append(filtered, state)
	}

	// Apply sorting
	SortSessions(filtered, m.sort)
	m.sessions = filtered

	// Keep cursor in bounds
	if m.cursor >= len(m.sessions) {
		m.cursor = len(m.sessions) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
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
				m.shouldAttach = m.sessions[m.cursor].TmuxSession
				return m, tea.Quit
			}
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

func (m model) View() string {
	if len(m.sessions) == 0 {
		return "\n  No active sessions\n\n  Press 'q' to quit\n"
	}

	var b strings.Builder

	// Header with sort indicators
	b.WriteString("\n")
	idHdr := "ID" + m.sort.Indicator(SortID)
	dirHdr := "DIRECTORY" + m.sort.Indicator(SortDirectory)
	statusHdr := "STATUS" + m.sort.Indicator(SortStatus)
	ageHdr := "AGE" + m.sort.Indicator(SortAge)
	updatedHdr := "UPDATED" + m.sort.Indicator(SortUpdated)
	header := fmt.Sprintf("  %-12s %-35s %-22s %-10s %s", idHdr, dirHdr, statusHdr, ageHdr, updatedHdr)
	b.WriteString(headerStyle.Render(header))
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

		dir := shortenPathForTable(state.Cwd, 33)
		age := FormatDuration(time.Since(state.Created))
		updated := FormatDuration(time.Since(state.Updated))

		row := fmt.Sprintf("  %-10s %-35s %-22s %-10s %s", state.ID, dir, status, age, updated)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render(row))
		} else {
			style := getRowStyle(state.Status)
			b.WriteString(style.Render(row))
		}
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  ↑/↓ navigate • enter attach • 1-5 sort • r refresh • q quit"))
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

// RunInteractive starts the interactive session viewer
// Returns the tmux session name to attach to (if any) and the final sort state
func RunInteractive(includeAll bool, initialSort SortState) (string, SortState, error) {
	m := initialModel(includeAll)
	m.sort = initialSort
	m = m.refreshSessions()

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", initialSort, err
	}

	fm := finalModel.(model)
	return fm.shouldAttach, fm.sort, nil
}

// RunWatchMode runs the interactive watch mode with attach support
func RunWatchMode(includeAll bool, initialSort SortState) error {
	currentSort := initialSort
	for {
		tmuxSession, newSort, err := RunInteractive(includeAll, currentSort)
		currentSort = newSort // Preserve sort state between attach cycles
		if err != nil {
			return err
		}

		if tmuxSession == "" {
			// User quit without selecting
			return nil
		}

		// Attach to the session
		fmt.Printf("Attaching to %s... (Ctrl+B D to detach)\n", tmuxSession)

		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return fmt.Errorf("tmux not found: %w", err)
		}

		// Run tmux attach as a subprocess (not exec, so we return here after detach)
		cmd := exec.Command(tmuxPath, "attach-session", "-t", tmuxSession)
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
