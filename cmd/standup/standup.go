package standup

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Interval int  `short:"i" help:"Interval between reminders in minutes." default:"30"`
	Quiet    bool `short:"q" help:"No bell sound." default:"false"`
}

var reminders = []string{
	"Time to stand up and stretch!",
	"Your spine called - it wants a break!",
	"Stand up! Touch your toes! Or at least try...",
	"Hydration check! Drink some water and stretch!",
	"Your future self will thank you for standing up now.",
	"Standing desk users: sit down for a bit. Sitting users: stand up!",
	"Quick stretch break! Roll those shoulders!",
	"Look away from the screen. Stand up. Breathe.",
	"Your legs are not decorative. Use them!",
	"Time for a micro-adventure to the window!",
	"Stand up! Pretend you're excited about something!",
	"Movement break! Do a little dance, make a little code...",
	"Your chair misses you already, but stand up anyway!",
	"RSI prevention time! Stretch those wrists!",
	"Stand up and do a mass victory pose for no reason!",
}

var exercises = []string{
	"🙆 Stretch your arms above your head",
	"🔄 Roll your shoulders 5 times",
	"👀 Look at something 20 feet away for 20 seconds",
	"🦵 Do 5 standing squats",
	"🚶 Walk to get a glass of water",
	"🧘 Touch your toes (or try to)",
	"💪 Do 5 desk push-ups",
	"🦴 Twist your torso left and right",
	"🖐️ Stretch your fingers and wrists",
	"🦒 Gently stretch your neck side to side",
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "standup",
		Short:       "Periodic reminders to stand up and stretch",
		Long:        "Reminds you to stand up and stretch at regular intervals. Your body will thank you.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	interval := time.Duration(params.Interval) * time.Minute

	// Clear screen and hide cursor
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top-left
	fmt.Print("\033[?25l")     // Hide cursor
	defer fmt.Print("\033[?25h\n")

	fmt.Printf("🧍 Standup reminder started! You'll be reminded every %d minutes.\n", params.Interval)
	fmt.Println("   Press Ctrl+C to stop.")
	fmt.Println()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Also show countdown
	secondTicker := time.NewTicker(1 * time.Second)
	defer secondTicker.Stop()

	nextReminder := time.Now().Add(interval)
	reminderCount := 0

	for {
		select {
		case <-sigChan:
			fmt.Printf("\n\n👋 Stay healthy! You received %d reminders today.\n", reminderCount)
			return
		case <-ticker.C:
			reminderCount++
			if !params.Quiet {
				fmt.Print("\a") // Terminal bell
			}
			fmt.Printf("\r\033[K\n")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Printf("🔔 #%d: %s\n", reminderCount, reminders[rand.Intn(len(reminders))])
			fmt.Printf("   Try: %s\n", exercises[rand.Intn(len(exercises))])
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println()
			nextReminder = time.Now().Add(interval)
		case <-secondTicker.C:
			remaining := time.Until(nextReminder)
			if remaining < 0 {
				remaining = 0
			}
			mins := int(remaining.Minutes())
			secs := int(remaining.Seconds()) % 60
			fmt.Printf("\r\033[K⏱️  Next reminder in %02d:%02d", mins, secs)
		}
	}
}
