package session

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type InstallHooksParams struct {
	Check bool `short:"c" long:"check" help:"Only check if hooks are installed, don't install"`
}

func InstallHooksCmd() *cobra.Command {
	return boa.CmdT[InstallHooksParams]{
		Use:         "install-hooks",
		Short:       "Install Claude hooks for session status tracking",
		Long:        "Install the required hooks in ~/.claude/settings.json for tofu session status tracking.\nThis complements your existing configuration without overwriting other settings.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *InstallHooksParams, cmd *cobra.Command, args []string) {
			if err := runInstallHooks(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runInstallHooks(params *InstallHooksParams) error {
	installed, missing := CheckHooksInstalled()

	if params.Check {
		if installed {
			fmt.Println("All tofu session hooks are installed.")
			return nil
		}
		fmt.Printf("Missing hooks for: %v\n", missing)
		fmt.Println("\nRun 'tofu claude session install-hooks' to install them.")
		os.Exit(1)
	}

	if installed {
		fmt.Println("All tofu session hooks are already installed.")
		return nil
	}

	fmt.Printf("Installing hooks for: %v\n", missing)

	if err := InstallHooks(); err != nil {
		return err
	}

	fmt.Println("\nHooks installed successfully!")
	fmt.Printf("Configuration updated: %s\n", ClaudeSettingsPath())
	fmt.Println("\nThe following hooks were added:")
	fmt.Println("  - UserPromptSubmit: Tracks when you send a prompt")
	fmt.Println("  - Stop: Tracks when Claude finishes responding")
	fmt.Println("  - PermissionRequest: Tracks when Claude needs permission")
	fmt.Println("  - PostToolUse: Tracks when a tool completes")
	fmt.Println("  - PostToolUseFailure: Tracks when a tool fails")
	fmt.Println("  - Notification: Tracks other notifications and dialogs")
	fmt.Println("\nAll hooks use a unified callback that logs events and updates session status.")

	return nil
}
