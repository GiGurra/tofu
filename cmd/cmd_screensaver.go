package cmd

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type ScreensaverParams struct {
	FPS int `short:"f" optional:"true" help:"Frames per second" default:"10"`
}

func ScreensaverCmd() *cobra.Command {
	return boa.CmdT[ScreensaverParams]{
		Use:         "screensaver",
		Short:       "Display an animated tofu bowl screensaver",
		Long:        "Display an animated tofu bowl with chopsticks in your terminal. Press Ctrl+C to exit.",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *ScreensaverParams, cmd *cobra.Command, args []string) {
			if err := runScreensaver(params); err != nil {
				fmt.Fprintf(os.Stderr, "screensaver: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runScreensaver(params *ScreensaverParams) error {
	// Handle signals for clean exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Hide cursor
	fmt.Print("\033[?25l")
	// Clear screen
	fmt.Print("\033[2J")

	// Restore cursor and clear screen on exit
	defer func() {
		fmt.Print("\033[?25h") // Show cursor
		fmt.Print("\033[2J")   // Clear screen
		fmt.Print("\033[H")    // Move to home
	}()

	fps := params.FPS
	if fps < 1 {
		fps = 10
	}
	if fps > 60 {
		fps = 60
	}

	frameDuration := time.Second / time.Duration(fps)
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	frame := 0
	for {
		select {
		case <-sigChan:
			return nil
		case <-ticker.C:
			renderFrame(frame)
			frame++
		}
	}
}

func renderFrame(frame int) {
	// Get terminal size
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	// Clear screen and move to top
	fmt.Print("\033[H")

	// Animation phases
	chopstickPhase := float64(frame) * 0.15
	steamPhase := float64(frame) * 0.2

	// Calculate center position
	bowlWidth := 30
	bowlHeight := 12
	startX := (width - bowlWidth) / 2
	startY := (height - bowlHeight) / 2

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// Build the frame
	lines := make([]string, height)
	for i := range lines {
		lines[i] = ""
	}

	// Steam animation (3 steam columns)
	steamChars := []string{"~", ".", "'", "`", "^", "~"}
	steamOffsets := []int{5, 12, 19}

	for row := 0; row < 3; row++ {
		y := startY - 3 + row
		if y >= 0 && y < height {
			line := spaces(startX)
			for col := 0; col < bowlWidth; col++ {
				isSteam := false
				for _, offset := range steamOffsets {
					if col == offset || col == offset+1 {
						// Animate steam rising
						steamIdx := int(steamPhase+float64(row)+float64(col)*0.5) % len(steamChars)
						if (frame/5+row+col)%3 != 0 { // Make steam flicker
							line += steamChars[steamIdx]
							isSteam = true
							break
						}
					}
				}
				if !isSteam {
					line += " "
				}
			}
			lines[y] = line
		}
	}

	// Chopstick animation - they move up and down like picking tofu
	chopstickOffset := int(math.Sin(chopstickPhase) * 2)
	chopstickSpread := int(math.Abs(math.Sin(chopstickPhase*0.5)) * 2)

	// Draw chopsticks (above and going into the bowl)
	chopstickLeft := startX + 8 - chopstickSpread
	chopstickRight := startX + 10 + chopstickSpread

	for row := 0; row < 4; row++ {
		y := startY + row + chopstickOffset
		if y >= 0 && y < height && row < 3 {
			// Build chopstick line
			line := ""
			for col := 0; col < width; col++ {
				if col == chopstickLeft+row {
					line += "/"
				} else if col == chopstickRight-row {
					line += "\\"
				} else {
					line += " "
				}
			}
			if lines[y] == "" {
				lines[y] = line
			} else {
				// Merge with existing content
				lines[y] = mergeLines(lines[y], line)
			}
		}
	}

	// Bowl top edge
	y := startY + 3
	if y >= 0 && y < height {
		lines[y] = mergeLines(lines[y], spaces(startX)+"   ___________________   ")
	}

	// Bowl opening with tofu inside
	tofuChars := []string{"#", "@", "#", "@", "#"}
	y = startY + 4
	if y >= 0 && y < height {
		tofu := ""
		for i, tc := range tofuChars {
			if (frame/8+i)%2 == 0 {
				tofu += tc + " "
			} else {
				tofu += "  "
			}
		}
		lines[y] = mergeLines(lines[y], spaces(startX)+"  /  "+tofu+"        \\  ")
	}

	// Bowl body with tofu pieces
	y = startY + 5
	if y >= 0 && y < height {
		// Tofu pieces that move slightly
		tofuLine := ""
		pieces := []string{"[##]", " @@ ", "[##]", " @@ "}
		for i, p := range pieces {
			if (frame/10+i)%3 == 0 {
				tofuLine += " " + p
			} else {
				tofuLine += p + " "
			}
		}
		lines[y] = mergeLines(lines[y], spaces(startX)+" |   "+tofuLine+"  | ")
	}

	// More bowl body
	y = startY + 6
	if y >= 0 && y < height {
		greenOnion := "~~~"
		if frame%6 < 3 {
			greenOnion = "~~~"
		} else {
			greenOnion = "```"
		}
		lines[y] = mergeLines(lines[y], spaces(startX)+" |  "+greenOnion+"  [##]  "+greenOnion+"  [##]  | ")
	}

	y = startY + 7
	if y >= 0 && y < height {
		lines[y] = mergeLines(lines[y], spaces(startX)+" |    @@   [##]   @@     | ")
	}

	// Bowl curves inward
	y = startY + 8
	if y >= 0 && y < height {
		lines[y] = mergeLines(lines[y], spaces(startX)+"  \\                    /  ")
	}

	y = startY + 9
	if y >= 0 && y < height {
		lines[y] = mergeLines(lines[y], spaces(startX)+"   \\__________________/   ")
	}

	// Bowl base
	y = startY + 10
	if y >= 0 && y < height {
		lines[y] = mergeLines(lines[y], spaces(startX)+"      |__________|        ")
	}

	// "TOFU" text with gentle animation
	y = startY + 12
	if y >= 0 && y < height {
		tofuText := "~ T O F U ~"
		if frame%20 < 10 {
			tofuText = "* T O F U *"
		}
		textX := startX + (bowlWidth-len(tofuText))/2
		lines[y] = spaces(textX) + tofuText
	}

	// Print all lines
	for _, line := range lines {
		fmt.Println(line)
	}
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	s := ""
	for i := 0; i < n; i++ {
		s += " "
	}
	return s
}

func mergeLines(base, overlay string) string {
	// Merge two lines, with overlay taking precedence for non-space chars
	result := []rune(base)

	// Extend result if overlay is longer
	for len(result) < len(overlay) {
		result = append(result, ' ')
	}

	for i, c := range overlay {
		if c != ' ' {
			if i < len(result) {
				result[i] = c
			}
		}
	}

	return string(result)
}
