package session

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type NewParams struct {
	Dir    string `pos:"true" optional:"true" help:"Directory to start session in (defaults to current directory)"`
	Resume string `long:"resume" short:"r" optional:"true" help:"Resume an existing conversation by ID"`
	Label  string `long:"label" short:"l" optional:"true" help:"Custom label for the session"`
	Attach bool   `long:"attach" short:"a" help:"Attach to the session immediately after creation"`
}

func NewCmd() *cobra.Command {
	return boa.CmdT[NewParams]{
		Use:         "new [dir]",
		Short:       "Start a new Claude Code session",
		Long:        "Start a new Claude Code session in a detached tmux session.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *NewParams, cmd *cobra.Command, args []string) {
			if err := runNew(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
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

	// Generate session ID
	sessionID := GenerateSessionID()
	if params.Label != "" {
		sessionID = params.Label
	}
	tmuxSession := "tofu-claude-" + sessionID

	// Build claude command
	claudeArgs := []string{}
	if params.Resume != "" {
		claudeArgs = append(claudeArgs, "--resume", params.Resume)
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
		ConvID:      params.Resume,
		Status:      StatusRunning,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	if err := SaveSessionState(state); err != nil {
		return fmt.Errorf("failed to save session state: %w", err)
	}

	fmt.Printf("Created session %s\n", sessionID)
	fmt.Printf("  Directory: %s\n", cwd)
	fmt.Printf("  Tmux: %s\n", tmuxSession)

	if params.Attach {
		fmt.Println("\nAttaching... (Ctrl+B D to detach)")
		return attachToSession(tmuxSession)
	}

	fmt.Printf("\nAttach with: tofu claude session attach %s\n", sessionID)
	return nil
}
