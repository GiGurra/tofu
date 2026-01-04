package diff

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	File1       string `pos:"true" help:"First file to compare."`
	File2       string `pos:"true" help:"Second file to compare."`
	Unified     int    `short:"u" help:"Output NUM lines of unified context." default:"3" optional:"true"`
	Context     int    `short:"c" help:"Output NUM lines of context." default:"0" optional:"true"`
	SideBySide  bool   `short:"y" help:"Output in two columns side by side." optional:"true"`
	Width       int    `short:"W" help:"Output at most NUM columns (for side-by-side)." default:"130" optional:"true"`
	Color       string `help:"Color output (auto, always, never)." default:"auto" optional:"true" alts:"auto,always,never"`
	NoColor     bool   `help:"Disable color output." optional:"true"`
	Brief       bool   `short:"q" help:"Report only when files differ." optional:"true"`
	IgnoreCase  bool   `short:"i" help:"Ignore case differences." optional:"true"`
	IgnoreSpace bool   `short:"b" help:"Ignore changes in whitespace." optional:"true"`
	IgnoreBlank bool   `short:"B" help:"Ignore blank lines." optional:"true"`
	Stats       bool   `short:"s" help:"Show statistics summary." optional:"true"`
}

// ANSI color codes for diff
const (
	diffColorReset  = "\033[0m"
	diffColorRed    = "\033[31m"
	diffColorGreen  = "\033[32m"
	diffColorCyan   = "\033[36m"
	diffColorYellow = "\033[33m"
)

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "diff",
		Short:       "Compare files line by line",
		Long:        "Compare two files and show differences with optional color output.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := runDiff(params); err != nil {
				fmt.Fprintf(os.Stderr, "diff: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runDiff(params *Params) error {
	// Read both files
	lines1, err := readFileLines(params.File1)
	if err != nil {
		return err
	}

	lines2, err := readFileLines(params.File2)
	if err != nil {
		return err
	}

	// Preprocess lines if needed
	if params.IgnoreCase {
		lines1 = toLowerLines(lines1)
		lines2 = toLowerLines(lines2)
	}
	if params.IgnoreSpace {
		lines1 = normalizeWhitespace(lines1)
		lines2 = normalizeWhitespace(lines2)
	}
	if params.IgnoreBlank {
		lines1 = removeBlankLines(lines1)
		lines2 = removeBlankLines(lines2)
	}

	// Compute diff using Myers algorithm
	diff := computeDiff(lines1, lines2)

	// Check if files are identical
	if len(diff) == 0 {
		if params.Brief {
			// No output for identical files in brief mode
			return nil
		}
		return nil
	}

	// Brief mode - just report difference
	if params.Brief {
		fmt.Printf("Files %s and %s differ\n", params.File1, params.File2)
		return nil
	}

	// Determine color usage
	useColor := shouldUseColor(params)

	// Output diff
	if params.SideBySide {
		printSideBySide(lines1, lines2, diff, params, useColor)
	} else {
		printUnified(params.File1, params.File2, lines1, lines2, diff, params.Unified, useColor)
	}

	// Print stats if requested
	if params.Stats {
		printStats(diff)
	}

	return nil
}

func readFileLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", filename, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}

	return lines, nil
}

func toLowerLines(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = strings.ToLower(line)
	}
	return result
}

func normalizeWhitespace(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		// Collapse multiple spaces/tabs into single space, trim
		fields := strings.Fields(line)
		result[i] = strings.Join(fields, " ")
	}
	return result
}

func removeBlankLines(lines []string) []string {
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

func shouldUseColor(params *Params) bool {
	if params.NoColor {
		return false
	}
	switch params.Color {
	case "always":
		return true
	case "never":
		return false
	default: // "auto"
		// Check if stdout is a terminal
		fi, err := os.Stdout.Stat()
		if err != nil {
			return false
		}
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
}

// DiffOp represents a diff operation
type DiffOp int

const (
	DiffEqual DiffOp = iota
	DiffInsert
	DiffDelete
)

// DiffLine represents a single line in the diff
type DiffLine struct {
	Op     DiffOp
	Line   string
	Index1 int // Line number in file1 (0-based, -1 if not applicable)
	Index2 int // Line number in file2 (0-based, -1 if not applicable)
}

// computeDiff computes the diff between two sets of lines using a simple LCS-based algorithm
func computeDiff(lines1, lines2 []string) []DiffLine {
	// Compute LCS (Longest Common Subsequence)
	m, n := len(lines1), len(lines2)

	// DP table for LCS length
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if lines1[i-1] == lines2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// Backtrack to find the diff
	var diff []DiffLine
	i, j := m, n

	// Build diff in reverse, then reverse it
	var reverseDiff []DiffLine

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && lines1[i-1] == lines2[j-1] {
			reverseDiff = append(reverseDiff, DiffLine{
				Op:     DiffEqual,
				Line:   lines1[i-1],
				Index1: i - 1,
				Index2: j - 1,
			})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			reverseDiff = append(reverseDiff, DiffLine{
				Op:     DiffInsert,
				Line:   lines2[j-1],
				Index1: -1,
				Index2: j - 1,
			})
			j--
		} else {
			reverseDiff = append(reverseDiff, DiffLine{
				Op:     DiffDelete,
				Line:   lines1[i-1],
				Index1: i - 1,
				Index2: -1,
			})
			i--
		}
	}

	// Reverse to get correct order
	for i := len(reverseDiff) - 1; i >= 0; i-- {
		diff = append(diff, reverseDiff[i])
	}

	return diff
}

func printUnified(file1, file2 string, lines1, lines2 []string, diff []DiffLine, context int, useColor bool) {
	// Print header
	if useColor {
		fmt.Printf("%s--- %s%s\n", diffColorYellow, file1, diffColorReset)
		fmt.Printf("%s+++ %s%s\n", diffColorYellow, file2, diffColorReset)
	} else {
		fmt.Printf("--- %s\n", file1)
		fmt.Printf("+++ %s\n", file2)
	}

	// Group changes into hunks
	hunks := groupIntoHunks(diff, context)

	for _, hunk := range hunks {
		// Find line ranges for this hunk
		startLine1 := -1
		startLine2 := -1

		for _, d := range hunk {
			if d.Index1 >= 0 && startLine1 < 0 {
				startLine1 = d.Index1
			}
			if d.Index2 >= 0 && startLine2 < 0 {
				startLine2 = d.Index2
			}
		}

		// Print hunk header
		count1 := 0
		count2 := 0
		for _, d := range hunk {
			if d.Op == DiffEqual || d.Op == DiffDelete {
				count1++
			}
			if d.Op == DiffEqual || d.Op == DiffInsert {
				count2++
			}
		}

		if startLine1 < 0 {
			startLine1 = 0
		}
		if startLine2 < 0 {
			startLine2 = 0
		}

		if useColor {
			fmt.Printf("%s@@ -%d,%d +%d,%d @@%s\n", diffColorCyan, startLine1+1, count1, startLine2+1, count2, diffColorReset)
		} else {
			fmt.Printf("@@ -%d,%d +%d,%d @@\n", startLine1+1, count1, startLine2+1, count2)
		}

		// Print hunk content
		for _, d := range hunk {
			switch d.Op {
			case DiffEqual:
				fmt.Printf(" %s\n", d.Line)
			case DiffDelete:
				if useColor {
					fmt.Printf("%s-%s%s\n", diffColorRed, d.Line, diffColorReset)
				} else {
					fmt.Printf("-%s\n", d.Line)
				}
			case DiffInsert:
				if useColor {
					fmt.Printf("%s+%s%s\n", diffColorGreen, d.Line, diffColorReset)
				} else {
					fmt.Printf("+%s\n", d.Line)
				}
			}
		}
	}
}

func groupIntoHunks(diff []DiffLine, context int) [][]DiffLine {
	if len(diff) == 0 {
		return nil
	}

	var hunks [][]DiffLine
	var currentHunk []DiffLine
	lastChangeIdx := -1

	for i, d := range diff {
		isChange := d.Op != DiffEqual

		if isChange {
			// Include context before this change
			contextStart := i - context
			if contextStart < 0 {
				contextStart = 0
			}
			if lastChangeIdx >= 0 && contextStart <= lastChangeIdx+context {
				// Merge with previous hunk - add lines from lastChangeIdx+1 to i
				for j := lastChangeIdx + 1; j <= i; j++ {
					currentHunk = append(currentHunk, diff[j])
				}
			} else {
				// Start new hunk
				if len(currentHunk) > 0 {
					hunks = append(hunks, currentHunk)
				}
				currentHunk = nil
				for j := contextStart; j <= i; j++ {
					currentHunk = append(currentHunk, diff[j])
				}
			}
			lastChangeIdx = i
		}
	}

	// Add trailing context to last hunk
	if lastChangeIdx >= 0 {
		contextEnd := lastChangeIdx + context
		if contextEnd >= len(diff) {
			contextEnd = len(diff) - 1
		}
		for j := lastChangeIdx + 1; j <= contextEnd; j++ {
			currentHunk = append(currentHunk, diff[j])
		}
		hunks = append(hunks, currentHunk)
	}

	return hunks
}

func printSideBySide(lines1, lines2 []string, diff []DiffLine, params *Params, useColor bool) {
	colWidth := (params.Width - 3) / 2 // -3 for separator " | "

	for _, d := range diff {
		left := ""
		right := ""
		sep := " "

		switch d.Op {
		case DiffEqual:
			left = d.Line
			right = d.Line
			sep = " "
		case DiffDelete:
			left = d.Line
			right = ""
			sep = "<"
		case DiffInsert:
			left = ""
			right = d.Line
			sep = ">"
		}

		// Truncate if needed
		if len(left) > colWidth {
			left = left[:colWidth-1] + "…"
		}
		if len(right) > colWidth {
			right = right[:colWidth-1] + "…"
		}

		// Pad to column width
		leftPadded := fmt.Sprintf("%-*s", colWidth, left)
		rightPadded := fmt.Sprintf("%-*s", colWidth, right)

		if useColor {
			switch d.Op {
			case DiffDelete:
				fmt.Printf("%s%s%s %s %s\n", diffColorRed, leftPadded, diffColorReset, sep, rightPadded)
			case DiffInsert:
				fmt.Printf("%s %s %s%s%s\n", leftPadded, sep, diffColorGreen, rightPadded, diffColorReset)
			default:
				fmt.Printf("%s %s %s\n", leftPadded, sep, rightPadded)
			}
		} else {
			fmt.Printf("%s %s %s\n", leftPadded, sep, rightPadded)
		}
	}
}

func printStats(diff []DiffLine) {
	insertions := 0
	deletions := 0

	for _, d := range diff {
		switch d.Op {
		case DiffInsert:
			insertions++
		case DiffDelete:
			deletions++
		}
	}

	fmt.Printf("\n%d insertion(s), %d deletion(s)\n", insertions, deletions)
}
