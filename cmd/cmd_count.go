package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type CountParams struct {
	Files      []string `pos:"true" optional:"true" help:"Files to count. If none specified or -, read from standard input." default:"-"`
	Lines      bool     `short:"l" help:"Print the line count." optional:"true"`
	Words      bool     `short:"w" help:"Print the word count." optional:"true"`
	Chars      bool     `short:"c" help:"Print the character count." optional:"true"`
	Bytes      bool     `short:"b" help:"Print the byte count." optional:"true"`
	MaxLine    bool     `short:"L" help:"Print the length of the longest line." optional:"true"`
	TotalOnly  bool     `short:"t" help:"Print only the total (when multiple files)." optional:"true"`
	NoFilename bool     `short:"n" help:"Never print filenames." optional:"true"`
}

type CountResult struct {
	Lines    int64
	Words    int64
	Chars    int64
	Bytes    int64
	MaxLine  int
	Filename string
}

func CountCmd() *cobra.Command {
	return boa.CmdT[CountParams]{
		Use:         "count",
		Short:       "Count lines, words, and characters",
		Long:        "Count lines, words, characters, and bytes in files. Similar to wc but with clearer flags.",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *CountParams, cmd *cobra.Command, args []string) {
			if err := runCount(params); err != nil {
				fmt.Fprintf(os.Stderr, "count: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runCount(params *CountParams) error {
	// If no specific flags set, default to lines, words, and chars (like wc)
	showAll := !params.Lines && !params.Words && !params.Chars && !params.Bytes && !params.MaxLine
	if showAll {
		params.Lines = true
		params.Words = true
		params.Chars = true
	}

	var results []CountResult
	var total CountResult
	total.Filename = "total"

	files := params.Files
	if len(files) == 0 {
		files = []string{"-"}
	}

	for _, file := range files {
		var reader io.Reader
		var filename string

		if file == "-" {
			reader = os.Stdin
			filename = "-"
		} else {
			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("cannot open %s: %w", file, err)
			}
			defer f.Close()
			reader = f
			filename = file
		}

		result, err := countReader(reader, filename, params)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", filename, err)
		}

		results = append(results, result)
		total.Lines += result.Lines
		total.Words += result.Words
		total.Chars += result.Chars
		total.Bytes += result.Bytes
		if result.MaxLine > total.MaxLine {
			total.MaxLine = result.MaxLine
		}
	}

	// Determine if we should print filenames
	showFilename := len(results) > 1 && !params.NoFilename
	if params.NoFilename {
		showFilename = false
	}

	// Print results
	if !params.TotalOnly {
		for _, result := range results {
			printResult(result, params, showFilename)
		}
	}

	// Print total if multiple files
	if len(results) > 1 {
		printResult(total, params, showFilename)
	}

	return nil
}

func countReader(reader io.Reader, filename string, params *CountParams) (CountResult, error) {
	result := CountResult{Filename: filename}

	// Read entire content if we need chars (for UTF-8) or max line length
	// Otherwise use line-by-line scanning for efficiency
	if params.Chars || params.MaxLine {
		content, err := io.ReadAll(reader)
		if err != nil {
			return result, err
		}

		result.Bytes = int64(len(content))
		result.Chars = int64(utf8.RuneCount(content))

		// Count lines and words, track max line length
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := scanner.Text()
			result.Lines++
			result.Words += int64(countWords(line))
			lineLen := utf8.RuneCountInString(line)
			if lineLen > result.MaxLine {
				result.MaxLine = lineLen
			}
		}

		// Check if content ends without newline (still count as a line if non-empty)
		if len(content) > 0 && content[len(content)-1] != '\n' {
			// Scanner already counted it, no adjustment needed
		}
	} else {
		// Efficient streaming mode for just lines/words/bytes
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Bytes()
			result.Lines++
			result.Words += int64(countWordsBytes(line))
			result.Bytes += int64(len(line) + 1) // +1 for newline
		}
		if err := scanner.Err(); err != nil {
			return result, err
		}
	}

	return result, nil
}

func countWords(s string) int {
	return len(strings.Fields(s))
}

func countWordsBytes(b []byte) int {
	return len(strings.Fields(string(b)))
}

func printResult(result CountResult, params *CountParams, showFilename bool) {
	var parts []string

	if params.Lines {
		parts = append(parts, fmt.Sprintf("%8d", result.Lines))
	}
	if params.Words {
		parts = append(parts, fmt.Sprintf("%8d", result.Words))
	}
	if params.Chars {
		parts = append(parts, fmt.Sprintf("%8d", result.Chars))
	}
	if params.Bytes {
		parts = append(parts, fmt.Sprintf("%8d", result.Bytes))
	}
	if params.MaxLine {
		parts = append(parts, fmt.Sprintf("%8d", result.MaxLine))
	}

	output := strings.Join(parts, "")
	if showFilename {
		output += " " + result.Filename
	}

	fmt.Println(output)
}
