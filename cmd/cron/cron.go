package cron

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
	Expression string `pos:"true" help:"Cron expression to parse/explain (5 or 6 fields)."`
	Next       int    `short:"n" help:"Show next N execution times." default:"5" optional:"true"`
	Validate   bool   `short:"v" help:"Only validate the expression, don't explain." optional:"true"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "cron",
		Short:       "Explain and validate cron expressions",
		Long:        "Parse cron expressions and show human-readable explanations with upcoming execution times.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := runCron(params); err != nil {
				fmt.Fprintf(os.Stderr, "cron: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

// CronField represents a parsed cron field
type CronField struct {
	Name     string
	Min      int
	Max      int
	Values   []int // nil means wildcard (*)
	Original string
}

// CronExpr represents a parsed cron expression
type CronExpr struct {
	Second     *CronField // optional, for 6-field cron
	Minute     *CronField
	Hour       *CronField
	DayOfMonth *CronField
	Month      *CronField
	DayOfWeek  *CronField
	HasSeconds bool
}

func runCron(params *Params) error {
	expr, err := parseCronExpression(params.Expression)
	if err != nil {
		return err
	}

	if params.Validate {
		fmt.Println("Valid cron expression")
		return nil
	}

	// Print explanation
	fmt.Println("Expression:", params.Expression)
	fmt.Println()
	fmt.Println("Schedule:")
	fmt.Println(explainCron(expr))
	fmt.Println()

	// Show next execution times
	if params.Next > 0 {
		fmt.Printf("Next %d execution times:\n", params.Next)
		times := getNextExecutions(expr, time.Now(), params.Next)
		for i, t := range times {
			fmt.Printf("  %d. %s\n", i+1, t.Format("Mon, 02 Jan 2006 15:04:05 MST"))
		}
	}

	return nil
}

func parseCronExpression(expr string) (*CronExpr, error) {
	fields := strings.Fields(expr)

	if len(fields) < 5 || len(fields) > 6 {
		return nil, fmt.Errorf("invalid cron expression: expected 5 or 6 fields, got %d", len(fields))
	}

	cron := &CronExpr{}
	idx := 0

	// Handle optional seconds field (6-field cron)
	if len(fields) == 6 {
		cron.HasSeconds = true
		second, err := parseField(fields[idx], "second", 0, 59)
		if err != nil {
			return nil, err
		}
		cron.Second = second
		idx++
	}

	// Minute
	minute, err := parseField(fields[idx], "minute", 0, 59)
	if err != nil {
		return nil, err
	}
	cron.Minute = minute
	idx++

	// Hour
	hour, err := parseField(fields[idx], "hour", 0, 23)
	if err != nil {
		return nil, err
	}
	cron.Hour = hour
	idx++

	// Day of month
	dom, err := parseField(fields[idx], "day-of-month", 1, 31)
	if err != nil {
		return nil, err
	}
	cron.DayOfMonth = dom
	idx++

	// Month
	month, err := parseField(fields[idx], "month", 1, 12)
	if err != nil {
		return nil, err
	}
	cron.Month = month
	idx++

	// Day of week
	dow, err := parseField(fields[idx], "day-of-week", 0, 6)
	if err != nil {
		return nil, err
	}
	cron.DayOfWeek = dow

	return cron, nil
}

func parseField(field, name string, min, max int) (*CronField, error) {
	cf := &CronField{
		Name:     name,
		Min:      min,
		Max:      max,
		Original: field,
	}

	// Handle wildcard
	if field == "*" {
		return cf, nil
	}

	// Handle lists (e.g., "1,2,3")
	var values []int
	parts := strings.Split(field, ",")

	for _, part := range parts {
		// Handle step values (e.g., "*/5" or "1-10/2")
		if strings.Contains(part, "/") {
			stepParts := strings.Split(part, "/")
			if len(stepParts) != 2 {
				return nil, fmt.Errorf("invalid step in %s: %s", name, part)
			}

			step, err := strconv.Atoi(stepParts[1])
			if err != nil || step <= 0 {
				return nil, fmt.Errorf("invalid step value in %s: %s", name, stepParts[1])
			}

			var rangeStart, rangeEnd int
			if stepParts[0] == "*" {
				rangeStart, rangeEnd = min, max
			} else if strings.Contains(stepParts[0], "-") {
				rangeStart, rangeEnd, err = parseRange(stepParts[0], name, min, max)
				if err != nil {
					return nil, err
				}
			} else {
				rangeStart, err = strconv.Atoi(stepParts[0])
				if err != nil {
					return nil, fmt.Errorf("invalid value in %s: %s", name, stepParts[0])
				}
				rangeEnd = max
			}

			for i := rangeStart; i <= rangeEnd; i += step {
				if i >= min && i <= max {
					values = append(values, i)
				}
			}
		} else if strings.Contains(part, "-") {
			// Handle ranges (e.g., "1-5")
			rangeStart, rangeEnd, err := parseRange(part, name, min, max)
			if err != nil {
				return nil, err
			}
			for i := rangeStart; i <= rangeEnd; i++ {
				values = append(values, i)
			}
		} else {
			// Single value
			val, err := parseValue(part, name, min, max)
			if err != nil {
				return nil, err
			}
			values = append(values, val)
		}
	}

	cf.Values = values
	return cf, nil
}

func parseRange(s, name string, min, max int) (int, int, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range in %s: %s", name, s)
	}

	start, err := parseValue(parts[0], name, min, max)
	if err != nil {
		return 0, 0, err
	}

	end, err := parseValue(parts[1], name, min, max)
	if err != nil {
		return 0, 0, err
	}

	if start > end {
		return 0, 0, fmt.Errorf("invalid range in %s: start > end (%d > %d)", name, start, end)
	}

	return start, end, nil
}

func parseValue(s, name string, min, max int) (int, error) {
	// Handle month names
	if name == "month" {
		if v, ok := monthNames[strings.ToLower(s)]; ok {
			return v, nil
		}
	}

	// Handle day names
	if name == "day-of-week" {
		if v, ok := dayNames[strings.ToLower(s)]; ok {
			return v, nil
		}
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value in %s: %s", name, s)
	}

	if val < min || val > max {
		return 0, fmt.Errorf("value out of range in %s: %d (must be %d-%d)", name, val, min, max)
	}

	return val, nil
}

var monthNames = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4,
	"may": 5, "jun": 6, "jul": 7, "aug": 8,
	"sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

var dayNames = map[string]int{
	"sun": 0, "mon": 1, "tue": 2, "wed": 3,
	"thu": 4, "fri": 5, "sat": 6,
}

var dayNamesReverse = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var monthNamesReverse = []string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

func explainCron(expr *CronExpr) string {
	var parts []string

	// Seconds (if present)
	if expr.HasSeconds && expr.Second != nil {
		parts = append(parts, explainField(expr.Second, "second"))
	}

	// Minutes
	parts = append(parts, explainField(expr.Minute, "minute"))

	// Hours
	parts = append(parts, explainField(expr.Hour, "hour"))

	// Day of month
	parts = append(parts, explainField(expr.DayOfMonth, "day of month"))

	// Month
	parts = append(parts, explainField(expr.Month, "month"))

	// Day of week
	parts = append(parts, explainField(expr.DayOfWeek, "day of week"))

	return strings.Join(parts, "\n")
}

func explainField(f *CronField, displayName string) string {
	if f.Values == nil {
		return fmt.Sprintf("  %-15s every %s", displayName+":", f.Name)
	}

	if len(f.Values) == 1 {
		return fmt.Sprintf("  %-15s %s", displayName+":", formatValue(f.Values[0], f.Name))
	}

	// Check if it's a continuous range
	if isContinuousRange(f.Values) {
		return fmt.Sprintf("  %-15s %s through %s", displayName+":",
			formatValue(f.Values[0], f.Name),
			formatValue(f.Values[len(f.Values)-1], f.Name))
	}

	// Check for step pattern
	if step := detectStep(f.Values); step > 0 {
		return fmt.Sprintf("  %-15s every %d %ss starting at %s", displayName+":",
			step, f.Name, formatValue(f.Values[0], f.Name))
	}

	// List values
	var formatted []string
	for _, v := range f.Values {
		formatted = append(formatted, formatValue(v, f.Name))
	}
	return fmt.Sprintf("  %-15s %s", displayName+":", strings.Join(formatted, ", "))
}

func formatValue(v int, name string) string {
	switch name {
	case "month":
		if v >= 1 && v <= 12 {
			return monthNamesReverse[v]
		}
	case "day-of-week":
		if v >= 0 && v <= 6 {
			return dayNamesReverse[v]
		}
	case "hour":
		if v == 0 {
			return "12:00 AM (midnight)"
		} else if v < 12 {
			return fmt.Sprintf("%d:00 AM", v)
		} else if v == 12 {
			return "12:00 PM (noon)"
		} else {
			return fmt.Sprintf("%d:00 PM", v-12)
		}
	}
	return strconv.Itoa(v)
}

func isContinuousRange(values []int) bool {
	if len(values) < 2 {
		return false
	}
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			return false
		}
	}
	return true
}

func detectStep(values []int) int {
	if len(values) < 2 {
		return 0
	}
	step := values[1] - values[0]
	if step <= 1 {
		return 0
	}
	for i := 2; i < len(values); i++ {
		if values[i]-values[i-1] != step {
			return 0
		}
	}
	return step
}

func getNextExecutions(expr *CronExpr, from time.Time, count int) []time.Time {
	var results []time.Time
	current := from.Truncate(time.Minute).Add(time.Minute)

	// Limit iterations to prevent infinite loops
	maxIterations := 366 * 24 * 60 // One year of minutes
	iterations := 0

	for len(results) < count && iterations < maxIterations {
		iterations++

		if matchesCron(expr, current) {
			results = append(results, current)
		}
		current = current.Add(time.Minute)
	}

	return results
}

func matchesCron(expr *CronExpr, t time.Time) bool {
	// Check second (if applicable)
	if expr.HasSeconds && expr.Second != nil {
		if !matchesField(expr.Second, t.Second()) {
			return false
		}
	}

	// Check minute
	if !matchesField(expr.Minute, t.Minute()) {
		return false
	}

	// Check hour
	if !matchesField(expr.Hour, t.Hour()) {
		return false
	}

	// Check month
	if !matchesField(expr.Month, int(t.Month())) {
		return false
	}

	// Check day of month and day of week
	// Standard cron: if both are restricted, either can match (OR)
	// If only one is restricted, that one must match
	domMatch := matchesField(expr.DayOfMonth, t.Day())
	dowMatch := matchesField(expr.DayOfWeek, int(t.Weekday()))

	domRestricted := expr.DayOfMonth.Values != nil
	dowRestricted := expr.DayOfWeek.Values != nil

	if domRestricted && dowRestricted {
		// Both restricted: OR logic
		if !domMatch && !dowMatch {
			return false
		}
	} else {
		// One or neither restricted: AND logic
		if !domMatch || !dowMatch {
			return false
		}
	}

	return true
}

func matchesField(f *CronField, value int) bool {
	if f.Values == nil {
		return true // wildcard
	}
	for _, v := range f.Values {
		if v == value {
			return true
		}
	}
	return false
}
