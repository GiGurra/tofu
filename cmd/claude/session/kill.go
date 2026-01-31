package session

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type KillParams struct {
	ID    string `pos:"true" help:"Session ID to kill"`
	Force bool   `short:"f" long:"force" help:"Force kill without confirmation"`
}

func KillCmd() *cobra.Command {
	return boa.CmdT[KillParams]{
		Use:         "kill <id>",
		Short:       "Kill a Claude Code session",
		ParamEnrich: common.DefaultParamEnricher(),
		ValidArgsFunc: func(p *KillParams, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return getSessionCompletions(true), cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveNoFileComp
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
	if params.ID == "" {
		return fmt.Errorf("session ID required")
	}

	// Find matching session
	state, err := findSession(params.ID)
	if err != nil {
		return err
	}

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

	fmt.Printf("Killed session %s\n", state.ID)
	return nil
}
