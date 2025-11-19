package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type CatParams struct {
	Files           []string `pos:"true" optional:"true" help:"Files to concatenate. If none specified or -, read from standard input." default:"-"`
	ShowAll         bool     `short:"A" help:"Equivalent to -vET (show all non-printing chars, ends, and tabs)."`
	NumberNonblank  bool     `short:"b" help:"Number non-empty output lines, overrides -n."`
	ShowEnds        bool     `short:"E" help:"Display $ at end of each line."`
	Number          bool     `short:"n" help:"Number all output lines."`
	SqueezeBlank    bool     `short:"s" help:"Suppress repeated empty output lines."`
	ShowTabs        bool     `short:"T" help:"Display TAB characters as ^I."`
	ShowNonPrinting bool     `short:"v" help:"Use ^ and M- notation for non-printing characters (except LFD and TAB)."`
}

func CatCmd() *cobra.Command {
	return boa.CmdT[CatParams]{
		Use:         "cat",
		Short:       "Concatenate files to standard output",
		ParamEnrich: defaultParamEnricher(),
		PreExecuteFunc: func(params *CatParams, cmd *cobra.Command, args []string) error {
			// Handle -A flag (equivalent to -vET)
			if params.ShowAll {
				params.ShowNonPrinting = true
				params.ShowEnds = true
				params.ShowTabs = true
			}

			return nil
		},
		RunFunc: func(params *CatParams, cmd *cobra.Command, args []string) {
			exitCode := runCat(params, os.Stdout, os.Stderr)
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}.ToCobra()
}

func runCat(params *CatParams, stdout, stderr io.Writer) int {
	hadError := false

	// If no files specified, default to stdin
	if len(params.Files) == 0 {
		params.Files = []string{"-"}
	}

	lineNum := 0

	for _, file := range params.Files {
		var reader io.Reader
		var closeFn func() error
		var filename string

		if file == "-" {
			reader = os.Stdin
			closeFn = func() error { return nil }
			filename = "<stdin>"
		} else {
			f, err := os.Open(file)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "cat: %s: %v\n", file, err)
				hadError = true
				continue
			}
			reader = f
			closeFn = f.Close
			filename = file
		}

		err := catReader(reader, stdout, params, &lineNum)
		closeErr := closeFn()

		if err != nil {
			_, _ = fmt.Fprintf(stderr, "cat: %s: %v\n", filename, err)
			hadError = true
		}
		if closeErr != nil {
			_, _ = fmt.Fprintf(stderr, "cat: error closing %s: %v\n", filename, closeErr)
			hadError = true
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func catReader(reader io.Reader, stdout io.Writer, params *CatParams, lineNum *int) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // 10MB max line size

	previousLineEmpty := false
	displayLineNum := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Handle squeeze blank lines
		isEmpty := len(strings.TrimSpace(line)) == 0
		if params.SqueezeBlank && isEmpty && previousLineEmpty {
			continue
		}
		previousLineEmpty = isEmpty

		// Build output line
		var output strings.Builder

		// Handle line numbering
		if params.NumberNonblank {
			if !isEmpty {
				displayLineNum++
				output.WriteString(fmt.Sprintf("%6d\t", displayLineNum))
			} else {
				// Don't show number for empty lines with -b
				output.WriteString("      \t")
			}
		} else if params.Number {
			displayLineNum++
			output.WriteString(fmt.Sprintf("%6d\t", displayLineNum))
		}

		// Process the line content
		processedLine := line
		if params.ShowTabs {
			processedLine = strings.ReplaceAll(processedLine, "\t", "^I")
		}
		if params.ShowNonPrinting {
			processedLine = showNonPrinting(processedLine, params.ShowTabs)
		}

		output.WriteString(processedLine)

		// Handle end-of-line marker
		if params.ShowEnds {
			output.WriteString("$")
		}

		fmt.Fprintln(stdout, output.String())
	}

	return scanner.Err()
}

func showNonPrinting(line string, tabsAlreadyHandled bool) string {
	var result strings.Builder

	for _, r := range line {
		// Skip LFD and TAB as per spec (and tab is already handled if ShowTabs is true)
		if r == '\n' {
			result.WriteRune(r)
			continue
		}
		if r == '\t' {
			if tabsAlreadyHandled {
				result.WriteString("^I")
			} else {
				result.WriteRune(r)
			}
			continue
		}

		// Handle control characters (0-31, excluding TAB and LF)
		if r < 32 {
			result.WriteString("^")
			result.WriteRune(r + 64)
			continue
		}

		// Handle DEL (127)
		if r == 127 {
			result.WriteString("^?")
			continue
		}

		// Handle high bit characters (128-255)
		if r >= 128 && r < 256 {
			result.WriteString("M-")
			if r >= 128+32 && r < 256 {
				result.WriteRune(r - 128)
			} else {
				result.WriteString("^")
				result.WriteRune(r - 128 + 64)
			}
			continue
		}

		// Normal printable character
		result.WriteRune(r)
	}

	return result.String()
}
