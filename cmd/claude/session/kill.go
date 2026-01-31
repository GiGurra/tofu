package session

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/GiGurra/boa/pkg/boa"
	clcommon "github.com/gigurra/tofu/cmd/claude/common"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type KillParams struct {
	ID    string `pos:"true" optional:"true" help:"Session ID to kill"`
	All   bool   `short:"a" long:"all" help:"Kill all sessions"`
	Idle  bool   `long:"idle" help:"Kill idle sessions (currently same as --all)"`
	Force bool   `short:"f" long:"force" help:"Force kill without confirmation"`
}

func KillCmd() *cobra.Command {
	return boa.CmdT[KillParams]{
		Use:         "kill [id]",
		Short:       "Kill Claude Code session(s)",
		Long:        "Kill one or more Claude Code sessions. Use --all to kill all sessions, or --idle to kill idle sessions (activity monitoring not yet implemented, currently same as --all).",
		ParamEnrich: common.DefaultParamEnricher(),
		ValidArgsFunc: func(p *KillParams, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return getSessionCompletions(false), cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveNoFileComp
		},
		RunFunc: func(params *KillParams, cmd *cobra.Command, args []string) {
			if err := runKill(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runKill(params *KillParams) error {
	// Handle --all or --idle (currently same behavior)
	if params.All || params.Idle {
		return killMultiple(params)
	}

	// Single session kill
	if params.ID == "" {
		return fmt.Errorf("session ID required (or use --all/--idle)")
	}

	// Extract just the ID from completion format
	sessionID := clcommon.ExtractIDFromCompletion(params.ID)

	// Find matching session
	state, err := findSession(sessionID)
	if err != nil {
		return err
	}

	return killSession(state)
}

func killMultiple(params *KillParams) error {
	states, err := ListSessionStates()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(states) == 0 {
		fmt.Println("No sessions to kill")
		return nil
	}

	// Refresh status and filter running sessions
	var running []*SessionState
	for _, state := range states {
		RefreshSessionStatus(state)
		if state.Status == StatusRunning {
			running = append(running, state)
		}
	}

	if len(running) == 0 {
		fmt.Println("No running sessions to kill")
		return nil
	}

	if !params.Force {
		fmt.Printf("This will kill %d session(s):\n", len(running))
		for _, state := range running {
			fmt.Printf("  %s (%s)\n", state.ID, state.Cwd)
		}
		fmt.Print("\nContinue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted")
			return nil
		}
	}

	killed := 0
	for _, state := range running {
		if err := killSession(state); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to kill %s: %v\n", state.ID, err)
			continue
		}
		killed++
	}

	fmt.Printf("Killed %d session(s)\n", killed)
	return nil
}

func killSession(state *SessionState) error {
	// Kill tmux session if alive
	if IsTmuxSessionAlive(state.TmuxSession) {
		cmd := exec.Command("tmux", "kill-session", "-t", state.TmuxSession)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to kill tmux session: %w", err)
		}
	}

	// Remove state file
	if err := DeleteSessionState(state.ID); err != nil {
		return fmt.Errorf("failed to delete session state: %w", err)
	}

	if state.ID != "" {
		fmt.Printf("Killed session %s\n", state.ID)
	}
	return nil
}
