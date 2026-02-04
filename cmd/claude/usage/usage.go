package usage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/creack/pty"
	"github.com/gigurra/tofu/cmd/claude/common/convops"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type Params struct {
	Interactive bool `short:"i" help:"Show Claude Code UI interactively (for debugging)"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "usage",
		Short:       "Show Claude Code subscription usage limits",
		Long:        "Launches Claude Code and runs /usage to show current subscription usage limits.\nShows session and weekly token usage percentages.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := runUsage(params.Interactive); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runUsage(interactive bool) error {
	// Create a temp directory to run claude from (isolates session data)
	tmpDir, err := os.MkdirTemp("", "tofu-usage-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Resolve symlinks (macOS: /var -> /private/var) so path matches what claude sees
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to resolve temp directory path: %w", err)
	}

	// Pre-trust the temp directory in ~/.claude.json so claude doesn't prompt
	// This backs up the original file and restores it on cleanup
	originalClaudeJSON, err := addTrustedProject(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to add trusted project: %w", err)
	}
	defer restoreClaudeJSON(originalClaudeJSON)

	// The claude project path for this temp dir (for cleanup)
	projectPath := convops.GetClaudeProjectPath(tmpDir)
	defer os.RemoveAll(projectPath)

	// Find claude binary
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH: %w", err)
	}

	// Create command - run from temp dir
	c := exec.Command(claudePath)
	c.Dir = tmpDir

	// Start with PTY
	ptmx, err := pty.Start(c)
	if err != nil {
		return fmt.Errorf("failed to start claude with PTY: %w", err)
	}
	defer ptmx.Close()

	// Handle window size changes (platform-specific)
	setupWindowResize(ptmx)

	// Set stdin to raw mode (needed for PTY interaction)
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Channels for coordination
	outputActivity := make(chan struct{}, 100)
	claudeReady := make(chan struct{}, 1)
	usageDataLoaded := make(chan struct{}, 1)
	usageCapture := make(chan string, 1)

	// Buffer to capture output for non-interactive mode
	var capturedOutput bytes.Buffer
	capturingUsage := false
	waitingForUsageData := false

	// Copy PTY output, signaling activity and detecting ready state
	go func() {
		buf := make([]byte, 4096)
		var accumulated bytes.Buffer
		ready := false
		usageLoaded := false

		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				return
			}

			// In interactive mode, show output live
			if interactive {
				os.Stdout.Write(buf[:n])
			}

			// Capture output after /usage is sent
			if capturingUsage {
				capturedOutput.Write(buf[:n])
			}

			// Signal output activity
			select {
			case outputActivity <- struct{}{}:
			default:
			}

			// Accumulate for pattern detection
			accumulated.Write(buf[:n])
			if accumulated.Len() > 500 {
				data := accumulated.Bytes()
				accumulated.Reset()
				accumulated.Write(data[len(data)-500:])
			}

			// Detect when Claude is ready (shows "shortcuts" hint)
			if !ready {
				if bytes.Contains(accumulated.Bytes(), []byte("shortcuts")) {
					ready = true
					select {
					case claudeReady <- struct{}{}:
					default:
					}
				}
			}

			// Detect when usage data has loaded (contains "%" which appears in usage percentages)
			if waitingForUsageData && !usageLoaded {
				accBytes := accumulated.Bytes()
				if bytes.Contains(accBytes, []byte("%")) {
					usageLoaded = true
					select {
					case usageDataLoaded <- struct{}{}:
					default:
					}
				}
			}
		}
	}()

	// Helper to wait for output to settle
	waitForSettle := func() {
		for {
			select {
			case <-outputActivity:
				// More output, keep waiting
				continue
			case <-time.After(100 * time.Millisecond):
				// Output settled
				return
			}
		}
	}

	// Type /usage after Claude is ready, then exit
	go func() {
		// Wait for Claude to show shortcuts hint
		<-claudeReady

		// Wait for UI to finish rendering
		waitForSettle()

		// Type each character, waiting for echo after each
		for _, ch := range []byte("/usage") {
			ptmx.Write([]byte{ch})
			waitForSettle()
		}

		// Start capturing output and waiting for usage data BEFORE sending Enter
		capturedOutput.Reset()
		capturingUsage = true
		waitingForUsageData = true

		// Send Enter
		ptmx.Write([]byte{13})

		// Wait for actual usage data to load (contains "resets" or "%")
		// Use a timeout to avoid hanging forever
		select {
		case <-usageDataLoaded:
			// Data loaded, now wait for settle
		case <-time.After(10 * time.Second):
			// Timeout - proceed anyway
		}

		// Wait for output to settle
		waitForSettle()

		// Stop capturing and save the screenshot
		capturingUsage = false
		snapshot := capturedOutput.String()

		// Send the captured content
		select {
		case usageCapture <- snapshot:
		default:
		}

		if interactive {
			// In interactive mode, wait a bit then kill (graceful /exit unreliable)
			time.Sleep(3 * time.Second)
		}
		// Kill the process - we have what we need
		c.Process.Kill()
	}()

	// Allow user to interrupt with Ctrl+C
	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	// In non-interactive mode, wait for capture and print before process exits
	var snapshot string
	if !interactive {
		snapshot = <-usageCapture
	}

	// Wait for process to exit
	err = c.Wait()

	// In non-interactive mode, print the captured usage info
	// (cleanup of temp dir, project path, and trusted project is handled by defers)
	if !interactive {
		printUsageInfo(snapshot)
		// We killed the process intentionally, so ignore that error
		return nil
	}

	return err
}

// printUsageInfo prints the captured usage output with all formatting
func printUsageInfo(raw string) {
	// Clear screen and move cursor to top-left
	fmt.Print("\033[2J\033[H")
	fmt.Print(raw)
}

// claudeJSONPath returns the path to ~/.claude.json
func claudeJSONPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude.json")
}

// addTrustedProject backs up ~/.claude.json, then adds a directory with hasTrustDialogAccepted: true
// Returns the original file contents for restoration
func addTrustedProject(dir string) ([]byte, error) {
	path := claudeJSONPath()

	// Read and backup original
	original, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse as untyped map
	var data map[string]any
	if err := json.Unmarshal(original, &data); err != nil {
		return nil, err
	}

	// Get or create projects map
	projects, ok := data["projects"].(map[string]any)
	if !ok {
		projects = make(map[string]any)
		data["projects"] = projects
	}

	// Add minimal project entry with trust accepted
	projects[dir] = map[string]any{
		"hasTrustDialogAccepted": true,
	}

	// Write modified file
	modified, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, modified, 0600); err != nil {
		return nil, err
	}

	return original, nil
}

// restoreClaudeJSON restores ~/.claude.json from the backup
func restoreClaudeJSON(original []byte) {
	if original != nil {
		os.WriteFile(claudeJSONPath(), original, 0600)
	}
}
