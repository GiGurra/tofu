package busy

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Duration int    `short:"d" help:"Duration in seconds (0 = infinite)." default:"0"`
	Message  string `short:"m" help:"Custom status message." default:""`
}

var tasks = []string{
	"Compiling",
	"Linking",
	"Optimizing",
	"Resolving dependencies",
	"Downloading packages",
	"Building modules",
	"Running tests",
	"Generating code",
	"Analyzing",
	"Minifying",
	"Bundling",
	"Transpiling",
	"Linting",
	"Type checking",
	"Formatting",
	"Deploying",
	"Syncing",
	"Indexing",
	"Caching",
	"Validating",
	"Processing",
	"Encrypting",
	"Hashing",
	"Initializing",
	"Bootstrapping",
	"Reticulating splines",
	"Reversing the polarity",
	"Calibrating flux capacitor",
	"Consulting the oracle",
	"Summoning dependencies",
}

var spinners = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "busy",
		Short:       "Look productive with a fake progress display",
		Long:        "Display a fake compilation/build progress. Perfect for looking busy.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	done := make(chan bool)
	if params.Duration > 0 {
		go func() {
			time.Sleep(time.Duration(params.Duration) * time.Second)
			done <- true
		}()
	}

	spinnerIdx := 0
	taskIdx := 0
	progress := 0
	currentTask := params.Message
	if currentTask == "" {
		currentTask = tasks[rand.Intn(len(tasks))]
	}

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	// Clear screen and hide cursor
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top-left
	fmt.Print("\033[?25l")     // Hide cursor
	defer fmt.Print("\033[?25h\n")

	for {
		select {
		case <-sigChan:
			return
		case <-done:
			// Final message
			fmt.Printf("\r\033[K✓ Done!%s\n", strings.Repeat(" ", 50))
			return
		case <-ticker.C:
			// Update spinner
			spinnerIdx = (spinnerIdx + 1) % len(spinners)

			// Occasionally change task
			if params.Message == "" && rand.Float32() < 0.02 {
				taskIdx = (taskIdx + 1) % len(tasks)
				currentTask = tasks[taskIdx]
				progress = rand.Intn(30)
			}

			// Increment progress
			if rand.Float32() < 0.3 {
				progress += rand.Intn(3)
				if progress > 100 {
					progress = rand.Intn(30)
					if params.Message == "" {
						currentTask = tasks[rand.Intn(len(tasks))]
					}
				}
			}

			// Build progress bar
			barWidth := 20
			filled := progress * barWidth / 100
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

			// Print status
			fmt.Printf("\r\033[K%s %s [%s] %d%%", spinners[spinnerIdx], currentTask, bar, progress)
		}
	}
}
