// Package setup provides the tofu claude setup command for one-time configuration.
package setup

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/claude/common/config"
	"github.com/gigurra/tofu/cmd/claude/common/wsl"
	"github.com/gigurra/tofu/cmd/claude/session"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

// Protocol version - bump this when the handler needs to be re-registered
const protocolVersion = "3"

type Params struct {
	Check bool `short:"c" long:"check" help:"Only check setup status, don't install anything"`
	Force bool `short:"f" long:"force" help:"Force re-registration of protocol handler"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "setup",
		Short:       "Set up tofu claude integration (hooks, protocol handler)",
		Long:        "One-time setup for tofu claude integration.\nInstalls hooks in ~/.claude/settings.json and registers the tofu:// protocol handler for clickable notifications.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := runSetup(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runSetup(params *Params) error {
	if params.Check {
		return checkStatus()
	}

	fmt.Println("Setting up tofu claude integration...")
	fmt.Println()

	// 1. Install hooks
	fmt.Println("=== Hooks ===")
	installed, missing, hasOldHooks := session.CheckHooksInstalled()
	if installed && !hasOldHooks {
		fmt.Println("✓ All hooks already installed")
	} else {
		if hasOldHooks {
			fmt.Println("  Removing old-style hooks...")
		}
		if len(missing) > 0 {
			fmt.Printf("  Installing hooks for: %v\n", missing)
		}
		if err := session.InstallHooks(); err != nil {
			return fmt.Errorf("failed to install hooks: %w", err)
		}
		fmt.Println("✓ Hooks installed")
	}

	// 2. Register protocol handler (WSL/Windows only)
	fmt.Println("\n=== Protocol Handler ===")
	if runtime.GOOS == "linux" && wsl.IsWSL() {
		registered, err := isProtocolRegistered()
		if err != nil {
			fmt.Printf("  Warning: could not check protocol status: %v\n", err)
		}

		if registered && !params.Force {
			fmt.Println("✓ Protocol handler already registered")
		} else {
			if params.Force {
				fmt.Println("  Force re-registering protocol handler...")
			}
			if err := registerProtocol(); err != nil {
				fmt.Printf("  Warning: failed to register protocol handler: %v\n", err)
				fmt.Println("  Clickable notifications may not work")
			} else {
				fmt.Println("✓ Protocol handler registered (tofu://)")
			}
		}
	} else if runtime.GOOS == "windows" {
		// Native Windows - could register protocol directly
		fmt.Println("  Skipped (not implemented for native Windows yet)")
	} else {
		fmt.Println("  Skipped (not needed on this platform)")
	}

	// 3. Configure notifications
	fmt.Println("\n=== Notifications ===")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("  Warning: could not load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	if cfg.Notifications != nil && cfg.Notifications.Enabled {
		fmt.Println("✓ Notifications already enabled")
	} else {
		if askYesNo("Enable desktop notifications when Claude needs attention?", true) {
			if cfg.Notifications == nil {
				cfg.Notifications = config.DefaultConfig().Notifications
			}
			cfg.Notifications.Enabled = true
			if err := config.Save(cfg); err != nil {
				fmt.Printf("  Warning: could not save config: %v\n", err)
			} else {
				fmt.Println("✓ Notifications enabled")
				fmt.Printf("  Config saved to: %s\n", config.ConfigPath())
			}
		} else {
			fmt.Println("  Notifications not enabled (you can enable later in config)")
		}
	}

	fmt.Println("\n=== Setup Complete ===")
	fmt.Println("You can verify with: tofu claude setup --check")

	return nil
}

// askYesNo prompts the user for a yes/no answer.
func askYesNo(prompt string, defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)

	defaultStr := "Y/n"
	if !defaultYes {
		defaultStr = "y/N"
	}

	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

func checkStatus() error {
	fmt.Println("Tofu Claude Setup Status")
	fmt.Println()

	// Check hooks
	fmt.Println("=== Hooks ===")
	installed, missing, hasOldHooks := session.CheckHooksInstalled()
	if hasOldHooks {
		fmt.Println("⚠ Old-style hooks detected (need upgrade)")
	}
	if installed {
		fmt.Println("✓ All hooks installed")
	} else {
		fmt.Printf("✗ Missing hooks: %v\n", missing)
	}

	// Check protocol handler
	fmt.Println("\n=== Protocol Handler ===")
	if runtime.GOOS == "linux" && wsl.IsWSL() {
		registered, err := isProtocolRegistered()
		if err != nil {
			fmt.Printf("⚠ Could not check: %v\n", err)
		} else if registered {
			fmt.Println("✓ Protocol handler registered (tofu://)")
		} else {
			fmt.Println("✗ Protocol handler not registered")
		}
	} else if runtime.GOOS == "windows" {
		fmt.Println("  Not implemented for native Windows yet")
	} else {
		fmt.Println("  Not applicable on this platform")
	}

	// Check config and notifications
	fmt.Println("\n=== Notifications ===")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("⚠ Could not load config: %v\n", err)
	} else if cfg.Notifications != nil && cfg.Notifications.Enabled {
		fmt.Println("✓ Notifications enabled")
		fmt.Printf("  Config: %s\n", config.ConfigPath())
	} else {
		fmt.Println("✗ Notifications disabled")
		fmt.Printf("  Run 'tofu claude setup' to enable\n")
	}

	return nil
}

// isProtocolRegistered checks if the tofu:// protocol handler is registered with current version.
func isProtocolRegistered() (bool, error) {
	psPath := wsl.FindPowerShell()
	if psPath == "" {
		return false, fmt.Errorf("powershell not found")
	}

	checkScript := fmt.Sprintf(`
$key = Get-ItemProperty -Path 'HKCU:\Software\Classes\tofu' -ErrorAction SilentlyContinue
if ($key -and $key.Version -eq '%s') { Write-Output 'registered' }
`, protocolVersion)

	cmd := exec.Command(psPath, "-NoProfile", "-NonInteractive", "-Command", checkScript)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.Contains(string(output), "registered"), nil
}

// registerProtocol registers the tofu:// protocol handler on Windows via WSL.
func registerProtocol() error {
	psPath := wsl.FindPowerShell()
	if psPath == "" {
		return fmt.Errorf("powershell not found")
	}

	// Register the protocol handler
	// The handler extracts session ID from tofu://focus/SESSION_ID and calls wsl to run tofu focus
	registerScript := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'

# Create protocol key with all required values
New-Item -Path 'HKCU:\Software\Classes\tofu' -Force | Out-Null
Set-ItemProperty -Path 'HKCU:\Software\Classes\tofu' -Name '(Default)' -Value 'URL:Tofu Protocol'
Set-ItemProperty -Path 'HKCU:\Software\Classes\tofu' -Name 'URL Protocol' -Value ''
Set-ItemProperty -Path 'HKCU:\Software\Classes\tofu' -Name 'Version' -Value '%s'

# Add DefaultIcon (uses Windows Terminal icon if available)
New-Item -Path 'HKCU:\Software\Classes\tofu\DefaultIcon' -Force | Out-Null
$wtPath = (Get-Command wt.exe -ErrorAction SilentlyContinue).Source
if ($wtPath) {
    Set-ItemProperty -Path 'HKCU:\Software\Classes\tofu\DefaultIcon' -Name '(Default)' -Value "$wtPath,0"
}

# Create shell/open/command key
New-Item -Path 'HKCU:\Software\Classes\tofu\shell\open\command' -Force | Out-Null

# The command extracts session ID and calls wsl to run tofu focus
# %%1 will be like: tofu://focus/abc12345
$cmd = 'powershell.exe -NoProfile -WindowStyle Hidden -Command "$url = ''%%1''; $sessionId = $url -replace ''tofu://focus/'',''''; wsl -- tofu claude session focus $sessionId"'
Set-ItemProperty -Path 'HKCU:\Software\Classes\tofu\shell\open\command' -Name '(Default)' -Value $cmd

Write-Output 'OK'
`, protocolVersion)

	cmd := exec.Command(psPath, "-NoProfile", "-NonInteractive", "-Command", registerScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("registration failed: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "OK") {
		return fmt.Errorf("registration may have failed: %s", string(output))
	}

	return nil
}

// IsProtocolRegistered is exported for use by the notify package.
func IsProtocolRegistered() bool {
	if runtime.GOOS != "linux" || !wsl.IsWSL() {
		return false
	}
	registered, _ := isProtocolRegistered()
	return registered
}
