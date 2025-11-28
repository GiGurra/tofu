package time

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Timestamp []string `pos:"true" optional:"true" help:"Timestamp to parse (Unix or formatted string)."`
	Format    string   `short:"f" help:"Explicit input format (e.g. '2006-01-02' or 'unix', 'unixmilli')." optional:"true"`
	UTC       bool     `short:"u" help:"Show output in UTC only (suppress Local)" default:"false"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:   "time [timestamp]",
		Short: "Show current time or parse a timestamp",
		Long: `Display the current time in various formats, or parse a provided timestamp.

If no argument is provided, the current system time is displayed.
If an argument is provided, it attempts to parse it as:
1. Unix timestamp (seconds, milliseconds, or nanoseconds).
2. Standard date/time formats (RFC3339, RFC1123, DateOnly, DateTime, etc.).

You can force a specific input format using the --format/-f flag.
Supported special format names: unix, unixmilli, unixmicro, unixnano.
Otherwise, provide a Go reference time layout (e.g. "2006-01-02 15:04").

Examples:
  tofu time
  tofu time 1698393600
  tofu time "2023-10-27T10:00:00Z"
  tofu time 2023-10-27
  tofu time "27/10/2023" -f "02/01/2006"`,
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			t := time.Now()
			if len(params.Timestamp) > 0 {
				input := strings.Join(params.Timestamp, " ")
				var parsed time.Time
				var err error

				if params.Format != "" {
					parsed, err = parseTimeWithFormat(input, params.Format)
				} else {
					parsed, err = parseTime(input)
				}

				if err != nil {
					fmt.Fprintf(os.Stderr, "time: could not parse '%s': %v\n", input, err)
					os.Exit(1)
				}
				t = parsed
			}

			printTime(t, params.UTC)
		},
	}.ToCobra()
}

func parseTimeWithFormat(input, format string) (time.Time, error) {
	// Handle special numeric formats
	if strings.HasPrefix(strings.ToLower(format), "unix") {
		val, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid numeric value for %s: %w", format, err)
		}
		switch strings.ToLower(format) {
		case "unix":
			return time.Unix(val, 0), nil
		case "unixmilli":
			return time.UnixMilli(val), nil
		case "unixmicro":
			return time.UnixMicro(val), nil
		case "unixnano":
			return time.Unix(0, val), nil
		default:
			return time.Time{}, fmt.Errorf("unknown unix format variant: %s", format)
		}
	}

	// Try standard Go layout parsing
	// 1. Try strict Parse (useful if timezone is included)
	if t, err := time.Parse(format, input); err == nil {
		return t, nil
	}

	// 2. Try ParseInLocation (assume Local if no timezone info)
	return time.ParseInLocation(format, input, time.Local)
}
func parseTime(input string) (time.Time, error) {
	// 1. Try numeric (Unix timestamp)
	if num, err := strconv.ParseInt(input, 10, 64); err == nil {
		// Guess precision based on magnitude
		// Unix Seconds ~ 1.7e9 (current)
		// Unix Millis ~ 1.7e12
		// Unix Micros ~ 1.7e15
		// Unix Nanos ~ 1.7e18

		if num < 100000000000 { // Assume Seconds (up to year 5138)
			return time.Unix(num, 0), nil
		} else if num < 100000000000000 { // Assume Millis
			return time.UnixMilli(num), nil
		} else if num < 100000000000000000 { // Assume Micros
			return time.UnixMicro(num), nil
		} else { // Assume Nanos
			return time.Unix(0, num), nil
		}
	}

	// 2. Try standard layouts
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.DateTime,
		time.DateOnly,
		time.TimeOnly,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, input); err == nil {
			return t, nil
		}
		// Try parsing in Local location if default Parse (UTC for some layouts) fails logic?
		// time.Parse parses as UTC if no timezone info is in string, usually.
		// time.ParseInLocation might be better for things like DateTime.
		if t, err := time.ParseInLocation(layout, input, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unknown format")
}

func printTime(t time.Time, utcOnly bool) {
	if !utcOnly {
		fmt.Printf("Local:      %s\n", t.Local().Format("2006-01-02 15:04:05.000 -0700 MST"))
	}
	fmt.Printf("UTC:        %s\n", t.UTC().Format("2006-01-02 15:04:05.000 -0700 MST"))
	fmt.Printf("Unix:       %d\n", t.Unix())
	fmt.Printf("UnixMilli:  %d\n", t.UnixMilli())
	fmt.Printf("RFC3339:    %s\n", t.Format(time.RFC3339))
	fmt.Printf("ISO8601:    %s\n", t.Format("2006-01-02T15:04:05.000Z07:00"))
}
