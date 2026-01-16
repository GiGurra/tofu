package pomodoro

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Work       int  `short:"w" help:"Work duration in minutes." default:"25"`
	Break      int  `short:"b" help:"Break duration in minutes." default:"5"`
	LongBreak  int  `short:"l" help:"Long break duration in minutes." default:"15"`
	Sessions   int  `short:"n" help:"Number of sessions before long break." default:"4"`
	Continuous bool `short:"c" help:"Run continuously (multiple pomodoros)." default:"false"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "pomodoro",
		Short:       "Pomodoro timer for productivity",
		Long:        "A simple pomodoro timer. Default: 25min work, 5min break, 15min long break after 4 sessions.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	session := 1
	totalSessions := params.Sessions

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h\n")

	for {
		// Work session
		fmt.Printf("\n🍅 Pomodoro #%d - Work time! (%d minutes)\n", session, params.Work)
		if !countdown(params.Work*60, "🍅 Working", sigChan) {
			return
		}
		playBell()
		fmt.Printf("\n✅ Work session complete!\n")

		// Break
		var breakDuration int
		var breakType string
		if session%totalSessions == 0 {
			breakDuration = params.LongBreak
			breakType = "☕ Long break"
		} else {
			breakDuration = params.Break
			breakType = "☕ Short break"
		}

		fmt.Printf("\n%s time! (%d minutes)\n", breakType, breakDuration)
		if !countdown(breakDuration*60, breakType, sigChan) {
			return
		}
		playBell()
		fmt.Printf("\n✅ Break complete!\n")

		session++

		if !params.Continuous {
			fmt.Printf("\n🎉 Pomodoro session finished! Great work!\n")
			return
		}
	}
}

func countdown(seconds int, label string, sigChan chan os.Signal) bool {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	remaining := seconds

	for remaining >= 0 {
		select {
		case <-sigChan:
			return false
		case <-ticker.C:
			mins := remaining / 60
			secs := remaining % 60

			// Progress bar
			progress := float64(seconds-remaining) / float64(seconds)
			barWidth := 30
			filled := int(progress * float64(barWidth))
			bar := ""
			for i := 0; i < barWidth; i++ {
				if i < filled {
					bar += "█"
				} else {
					bar += "░"
				}
			}

			fmt.Printf("\r\033[K%s [%s] %02d:%02d ", label, bar, mins, secs)
			remaining--
		}
	}
	return true
}

func playBell() {
	// Terminal bell
	fmt.Print("\a")
}
