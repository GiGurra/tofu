package session

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type AttachParams struct {
	ID string `pos:"true" help:"Session ID to attach to"`
}

func AttachCmd() *cobra.Command {
	return boa.CmdT[AttachParams]{
		Use:         "attach <id>",
		Short:       "Attach to a Claude Code session",
		Long:        "Attach to an existing Claude Code session. Use Ctrl+B D to detach.",
		ParamEnrich: common.DefaultParamEnricher(),
		ValidArgsFunc: func(p *AttachParams, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return getSessionCompletions(false), cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveNoFileComp
		},
		RunFunc: func(params *AttachParams, cmd *cobra.Command, args []string) {
			if err := runAttach(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runAttach(params *AttachParams) error {
	if params.ID == "" {
		return fmt.Errorf("session ID required")
	}

	// Find matching session
	state, err := findSession(params.ID)
	if err != nil {
		return err
	}

	// Check if session is alive
	if !IsTmuxSessionAlive(state.TmuxSession) {
		state.Status = StatusExited
		SaveSessionState(state)
		return fmt.Errorf("session %s has exited", state.ID)
	}

	fmt.Printf("Attaching to session %s... (Ctrl+B D to detach)\n", state.ID)
	return attachToSession(state.TmuxSession)
}

// AttachToTmuxSession attaches to a tmux session, replacing the current process
// Returns exit code (0 = success) for use by other packages
func AttachToTmuxSession(tmuxSession string) int {
	if err := attachToSession(tmuxSession); err != nil {
		return 1
	}
	return 0
}

// attachToSession attaches to a tmux session, replacing the current process
func attachToSession(tmuxSession string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	// Use syscall.Exec to replace current process with tmux attach
	args := []string{"tmux", "attach-session", "-t", tmuxSession}
	env := os.Environ()

	return syscall.Exec(tmuxPath, args, env)
}

// findSession finds a session by ID or prefix
func findSession(id string) (*SessionState, error) {
	states, err := ListSessionStates()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Try exact match first
	for _, state := range states {
		if state.ID == id {
			return state, nil
		}
	}

	// Try prefix match
	var matches []*SessionState
	for _, state := range states {
		if strings.HasPrefix(state.ID, id) {
			matches = append(matches, state)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no session found with ID %s", id)
	}
	if len(matches) > 1 {
		ids := make([]string, len(matches))
		for i, m := range matches {
			ids[i] = m.ID
		}
		return nil, fmt.Errorf("ambiguous ID %s, matches: %v", id, ids)
	}

	return matches[0], nil
}
