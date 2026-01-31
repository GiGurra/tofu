package session

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	clcommon "github.com/gigurra/tofu/cmd/claude/common"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type NewParams struct {
	Dir      string `pos:"true" optional:"true" help:"Directory to start session in (defaults to current directory)"`
	Resume   string `long:"resume" short:"r" optional:"true" help:"Resume an existing conversation by ID"`
	Label    string `long:"label" optional:"true" help:"Custom label for the session"`
	Detached bool   `long:"detached" short:"d" help:"Start detached (don't attach to session)"`
}

func NewCmd() *cobra.Command {
	cmd := boa.CmdT[NewParams]{
		Use:         "new",
		Short:       "Start a new Claude Code session",
		Long:        "Start a new Claude Code session in a tmux session. Attaches by default (Ctrl+B D to detach).",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *NewParams, cmd *cobra.Command, args []string) {
			if err := runNew(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()

	// Register completion for --resume flag
	cmd.RegisterFlagCompletionFunc("resume", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return clcommon.GetConversationCompletions(true), cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runNew(params *NewParams) error {
	// Check tmux is installed
	if err := CheckTmuxInstalled(); err != nil {
		return err
	}

	// Determine working directory
	cwd := params.Dir
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Make path absolute
	if cwd[0] != '/' {
		wd, _ := os.Getwd()
		cwd = wd + "/" + cwd
	}

	// Extract just the ID from autocomplete format (e.g., "0459cd73_[title]_prompt..." -> "0459cd73")
	shortID := clcommon.ExtractIDFromCompletion(params.Resume)
	// Resolve to full UUID for claude --resume
	fullConvID := clcommon.ResolveConvID(shortID)

	// Generate session ID (use short prefix for our tracking)
	// Priority: explicit label > conv ID prefix (when resuming) > random
	sessionID := GenerateSessionID()
	if shortID != "" {
		// Use conv ID prefix as session ID for easy association
		sessionID = shortID
		if len(sessionID) > 8 {
			sessionID = sessionID[:8]
		}
	}
	if params.Label != "" {
		sessionID = params.Label
	}
	tmuxSession := "tofu-claude-" + sessionID

	// Build claude command
	claudeArgs := []string{}
	if fullConvID != "" {
		claudeArgs = append(claudeArgs, "--resume", fullConvID)
	}

	// Create tmux session with claude
	// Use tmux new-session -d to create detached
	tmuxArgs := []string{
		"new-session",
		"-d",                  // detached
		"-s", tmuxSession,     // session name
		"-c", cwd,             // working directory
		"claude",              // command
	}
	tmuxArgs = append(tmuxArgs, claudeArgs...)

	tmuxCmd := exec.Command("tmux", tmuxArgs...)
	tmuxCmd.Stdout = os.Stdout
	tmuxCmd.Stderr = os.Stderr

	if err := tmuxCmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	// Get the PID of claude in the tmux session
	pid := ParsePIDFromTmux(tmuxSession)

	// Create session state
	state := &SessionState{
		ID:          sessionID,
		TmuxSession: tmuxSession,
		PID:         pid,
		Cwd:         cwd,
		ConvID:      fullConvID,
		Status:      StatusRunning,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	if err := SaveSessionState(state); err != nil {
		return fmt.Errorf("failed to save session state: %w", err)
	}

	fmt.Printf("Created session %s\n", sessionID)
	fmt.Printf("  Directory: %s\n", cwd)

	if params.Detached {
		fmt.Printf("\nAttach with: tofu claude session attach %s\n", sessionID)
		return nil
	}

	fmt.Println("\nAttaching... (Ctrl+B D to detach)")
	return attachToSession(tmuxSession)
}
