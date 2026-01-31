package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type StatusParams struct{}

func StatusCmd() *cobra.Command {
	return boa.CmdT[StatusParams]{
		Use:         "status",
		Short:       "Show git sync status",
		Long:        "Show the status of the Claude conversation sync repository.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *StatusParams, cmd *cobra.Command, args []string) {
			if err := runStatus(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runStatus(_ *StatusParams) error {
	syncDir := SyncDir()
	projectsDir := ProjectsDir()

	if !IsInitialized() {
		fmt.Printf("Git sync not initialized.\n")
		fmt.Printf("Run 'tofu claude git init <repo-url>' to set up sync.\n")
		return nil
	}

	fmt.Printf("Sync directory: %s\n\n", syncDir)

	// Show git remote
	remoteCmd := exec.Command("git", "remote", "-v")
	remoteCmd.Dir = syncDir
	remoteCmd.Stdout = os.Stdout
	remoteCmd.Stderr = os.Stderr
	remoteCmd.Run()

	fmt.Println()

	// Copy local changes to sync dir so we can see pending changes
	if err := copyProjectsToSync(projectsDir, syncDir, false); err != nil {
		return fmt.Errorf("failed to copy local changes: %w", err)
	}

	// Show pending changes
	statusCmd := exec.Command("git", "status", "--short")
	statusCmd.Dir = syncDir
	output, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	if len(output) == 0 {
		fmt.Printf("No pending changes to sync.\n")
	} else {
		// Parse and summarize the changes
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		conversations := []string{}
		for _, line := range lines {
			if strings.HasSuffix(line, ".jsonl") {
				// Extract conversation path
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					conversations = append(conversations, parts[len(parts)-1])
				}
			}
		}

		if len(conversations) > 0 {
			fmt.Printf("Conversations with new messages (%d):\n", len(conversations))
			for _, c := range conversations {
				fmt.Printf("  %s\n", c)
			}
		}

		// Show full status
		otherChanges := len(lines) - len(conversations)
		if otherChanges > 0 {
			fmt.Printf("\nOther changes: %d files\n", otherChanges)
		}
	}

	// Show last sync time (last commit)
	logCmd := exec.Command("git", "log", "-1", "--format=%cr (%ci)", "--date=local")
	logCmd.Dir = syncDir
	logOutput, err := logCmd.Output()
	if err == nil && len(logOutput) > 0 {
		fmt.Printf("\nLast sync: %s", logOutput)
	}

	return nil
}
