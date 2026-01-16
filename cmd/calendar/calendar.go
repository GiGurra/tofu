package calendar

import (
	"fmt"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Month int `short:"m" help:"Month (1-12). Default is current month." default:"0"`
	Year  int `short:"y" help:"Year. Default is current year." default:"0"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "calendar",
		Short:       "Display a calendar",
		Long:        "Display a terminal calendar with today highlighted.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	now := time.Now()

	month := time.Month(params.Month)
	year := params.Year

	if params.Month == 0 {
		month = now.Month()
	}
	if params.Year == 0 {
		year = now.Year()
	}

	// Is this the current month?
	isCurrentMonth := month == now.Month() && year == now.Year()
	today := now.Day()

	// Print header
	title := fmt.Sprintf("%s %d", month.String(), year)
	padding := (20 - len(title)) / 2
	fmt.Printf("%*s%s\n", padding, "", title)
	fmt.Println("Su Mo Tu We Th Fr Sa")

	// Get first day of month
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(firstDay.Weekday())

	// Get number of days in month
	lastDay := firstDay.AddDate(0, 1, -1)
	daysInMonth := lastDay.Day()

	// Print leading spaces
	for i := 0; i < startWeekday; i++ {
		fmt.Print("   ")
	}

	// Print days
	for day := 1; day <= daysInMonth; day++ {
		if isCurrentMonth && day == today {
			// Highlight today with reverse video
			fmt.Printf("\033[7m%2d\033[0m ", day)
		} else {
			fmt.Printf("%2d ", day)
		}

		// New line after Saturday
		if (startWeekday+day)%7 == 0 {
			fmt.Println()
		}
	}

	// Final newline if needed
	if (startWeekday+daysInMonth)%7 != 0 {
		fmt.Println()
	}
}
