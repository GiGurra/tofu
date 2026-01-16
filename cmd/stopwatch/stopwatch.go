package stopwatch

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type Params struct{}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "stopwatch",
		Short:       "Simple stopwatch",
		Long:        "A terminal stopwatch. Press Space to lap, Enter to pause/resume, Q to quit.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Clear screen and hide cursor
	fmt.Print("\033[2J\033[H")
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h\n")

	// Set terminal to raw mode to capture keypresses
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set raw mode: %v\n", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	startTime := time.Now()
	running := true
	var pausedDuration time.Duration
	var pauseStart time.Time
	var laps []time.Duration

	// Channel for keyboard input
	keyChan := make(chan byte)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				continue
			}
			keyChan <- buf[0]
		}
	}()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	fmt.Println("⏱️  STOPWATCH")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Space: Lap | Enter: Pause/Resume | Q: Quit")
	fmt.Println()

	for {
		select {
		case <-sigChan:
			return
		case key := <-keyChan:
			switch key {
			case 'q', 'Q', 3: // q, Q, or Ctrl+C
				return
			case ' ': // Space - lap
				if running {
					elapsed := time.Since(startTime) - pausedDuration
					laps = append(laps, elapsed)
				}
			case 13, 10: // Enter - pause/resume
				if running {
					pauseStart = time.Now()
					running = false
				} else {
					pausedDuration += time.Since(pauseStart)
					running = true
				}
			}
		case <-ticker.C:
			// Move cursor to time position
			fmt.Print("\033[5;1H")

			var elapsed time.Duration
			if running {
				elapsed = time.Since(startTime) - pausedDuration
			} else {
				elapsed = pauseStart.Sub(startTime) - pausedDuration
			}

			// Format time
			hours := int(elapsed.Hours())
			minutes := int(elapsed.Minutes()) % 60
			seconds := int(elapsed.Seconds()) % 60
			millis := int(elapsed.Milliseconds()) % 1000

			status := "▶"
			if !running {
				status = "⏸"
			}

			fmt.Printf("%s  %02d:%02d:%02d.%03d\033[K\n", status, hours, minutes, seconds, millis)

			// Show laps
			if len(laps) > 0 {
				fmt.Println("\nLaps:")
				// Show last 5 laps
				start := 0
				if len(laps) > 5 {
					start = len(laps) - 5
				}
				for i := start; i < len(laps); i++ {
					lap := laps[i]
					h := int(lap.Hours())
					m := int(lap.Minutes()) % 60
					s := int(lap.Seconds()) % 60
					ms := int(lap.Milliseconds()) % 1000
					fmt.Printf("  #%d  %02d:%02d:%02d.%03d\033[K\n", i+1, h, m, s, ms)
				}
			}
		}
	}
}
