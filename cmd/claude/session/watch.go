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
}

func initialModel(includeAll bool) model {
	return model{
		sessions:   []*SessionState{},
		cursor:     0,
		includeAll: includeAll,
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

	// Header
	b.WriteString("\n")
	header := fmt.Sprintf("  %-10s %-40s %-25s %s", "ID", "DIRECTORY", "STATUS", "UPDATED")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(headerStyle.Render(strings.Repeat("─", min(m.width-2, 90))))
	b.WriteString("\n")

	// Rows
	for i, state := range m.sessions {
		status := state.Status
		if state.StatusDetail != "" {
			status = status + ": " + state.StatusDetail
		}

		dir := shortenPathForTable(state.Cwd, 38)
		updated := FormatDuration(time.Since(state.Updated))

		row := fmt.Sprintf("  %-10s %-40s %-25s %s", state.ID, dir, status, updated)

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
	b.WriteString(helpStyle.Render("  ↑/↓ navigate • enter attach • r refresh • q quit"))
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
// Returns the tmux session name to attach to (if any)
func RunInteractive(includeAll bool) (string, error) {
	m := initialModel(includeAll)
	m = m.refreshSessions()

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	fm := finalModel.(model)
	return fm.shouldAttach, nil
}

// RunWatchMode runs the interactive watch mode with attach support
func RunWatchMode(includeAll bool) error {
	for {
		tmuxSession, err := RunInteractive(includeAll)
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
