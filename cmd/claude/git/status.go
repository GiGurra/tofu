package git

import (
	"fmt"
	"os"
	"os/exec"

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

	// Show git status
	statusCmd := exec.Command("git", "status", "--short")
	statusCmd.Dir = syncDir
	statusCmd.Stdout = os.Stdout
	statusCmd.Stderr = os.Stderr
	statusCmd.Run()

	// Show last sync time (last commit)
	logCmd := exec.Command("git", "log", "-1", "--format=%cr (%ci)", "--date=local")
	logCmd.Dir = syncDir
	output, err := logCmd.Output()
	if err == nil && len(output) > 0 {
		fmt.Printf("\nLast sync: %s", output)
	}

	return nil
}
