package clock

import (
	"fmt"
	"math"
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
	Size int `short:"s" help:"Clock radius (default auto-fits terminal)." default:"12"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "clock",
		Short:       "Display an analog clock in the terminal",
		Long:        "Shows a beautiful analog clock with hour, minute, and second hands.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Clear screen and hide cursor
	fmt.Print("\033[2J\033[H")
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h\n")

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			return
		case <-ticker.C:
			fmt.Print("\033[H") // Move cursor to top-left
			drawClock(params.Size)
		}
	}
}

func drawClock(radius int) {
	now := time.Now()
	hour := now.Hour() % 12
	minute := now.Minute()
	second := now.Second()

	// Grid dimensions (wider to compensate for terminal character aspect ratio)
	width := radius*4 + 4
	height := radius*2 + 4

	// Create grid
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	centerX := width / 2
	centerY := height / 2

	// Draw clock face (circle)
	for angle := 0.0; angle < 360; angle += 2 {
		rad := angle * math.Pi / 180
		x := int(float64(centerX) + float64(radius)*2*math.Cos(rad))
		y := int(float64(centerY) + float64(radius)*math.Sin(rad))
		if y >= 0 && y < height && x >= 0 && x < width {
			grid[y][x] = '·'
		}
	}

	// Draw hour markers
	for h := 1; h <= 12; h++ {
		angle := float64(h)*30 - 90 // 30 degrees per hour, offset by -90 to start at 12
		rad := angle * math.Pi / 180
		x := int(float64(centerX) + float64(radius-1)*2*math.Cos(rad))
		y := int(float64(centerY) + float64(radius-1)*math.Sin(rad))
		if y >= 0 && y < height && x >= 0 && x < width {
			if h == 12 {
				grid[y][x] = '1'
				if x+1 < width {
					grid[y][x+1] = '2'
				}
			} else if h < 10 {
				grid[y][x] = rune('0' + h)
			} else {
				grid[y][x] = '1'
				if x+1 < width {
					grid[y][x+1] = rune('0' + h - 10)
				}
			}
		}
	}

	// Calculate hand angles (0 = 12 o'clock, clockwise)
	secondAngle := float64(second)*6 - 90                    // 6 degrees per second
	minuteAngle := float64(minute)*6 - 90                    // 6 degrees per minute
	hourAngle := float64(hour)*30 + float64(minute)*0.5 - 90 // 30 degrees per hour + minute offset

	// Draw hands (second hand longest, hour hand shortest)
	drawHand(grid, centerX, centerY, hourAngle, float64(radius)*0.5, '●', 2)
	drawHand(grid, centerX, centerY, minuteAngle, float64(radius)*0.75, '○', 2)
	drawHand(grid, centerX, centerY, secondAngle, float64(radius)*0.9, '∙', 2)

	// Draw center
	grid[centerY][centerX] = '◎'

	// Print grid
	var sb strings.Builder
	for _, row := range grid {
		sb.WriteString(string(row))
		sb.WriteString("\n")
	}

	// Add digital time below
	sb.WriteString(fmt.Sprintf("\n%s%s\n",
		strings.Repeat(" ", centerX-4),
		now.Format("15:04:05")))

	fmt.Print(sb.String())
}

func drawHand(grid [][]rune, centerX, centerY int, angle, length float64, char rune, widthMult int) {
	rad := angle * math.Pi / 180
	steps := int(length * 2)

	for i := 1; i <= steps; i++ {
		progress := float64(i) / float64(steps)
		x := int(float64(centerX) + length*float64(widthMult)*progress*math.Cos(rad))
		y := int(float64(centerY) + length*progress*math.Sin(rad))

		if y >= 0 && y < len(grid) && x >= 0 && x < len(grid[0]) {
			grid[y][x] = char
		}
	}
}
