package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type SedPatternType string

const (
	SedPatternTypeLiteral SedPatternType = "literal"
	SedPatternTypeRegex   SedPatternType = "regex"
)

type Sed2Params struct {
	From       string         `pos:"true" help:"Pattern to search for."`
	To         string         `pos:"true" help:"Replacement string."`
	Files      []string       `pos:"true" optional:"true" help:"Files to process. If none specified or -, read from standard input." default:"-"`
	SearchType SedPatternType `short:"t" help:"Type of pattern to search for (literal, regex)." default:"regex" alts:"literal,regex"`
	InPlace    bool           `short:"i" help:"Edit files in place." default:"false"`
	IgnoreCase bool           `short:"I" help:"Perform a case-insensitive search." default:"false"`
	Global     bool           `short:"g" help:"Replace all occurrences on each line (not just first)." default:"false"`
}

func Sed2Cmd() *cobra.Command {
	return boa.CmdT[Sed2Params]{
		Use:         "sed2",
		Short:       "sed-like-but-different stream editor for filtering and transforming text",
		ParamEnrich: defaultParamEnricher(),
		PreExecuteFunc: func(params *Sed2Params, cmd *cobra.Command, args []string) error {
			if params.From == "" {
				return fmt.Errorf("search pattern cannot be empty")
			}

			// InPlace only makes sense with actual files
			if params.InPlace && (len(params.Files) == 0 || params.Files[0] == "-") {
				return fmt.Errorf("-i (in-place) flag requires file arguments")
			}

			return nil
		},
		RunFunc: func(params *Sed2Params, cmd *cobra.Command, args []string) {
			exitCode := runSed2(params)
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}.ToCobra()
}

func runSed2(params *Sed2Params) int {
	// Build the pattern
	var pattern *regexp.Regexp
	var err error

	if params.SearchType == SedPatternTypeLiteral {
		// For literal search, escape regex special characters
		escapedPattern := regexp.QuoteMeta(params.From)
		if params.IgnoreCase {
			escapedPattern = "(?i)" + escapedPattern
		}
		pattern, err = regexp.Compile(escapedPattern)
	} else {
		// Regex search
		patternStr := params.From
		if params.IgnoreCase {
			patternStr = "(?i)" + patternStr
		}
		pattern, err = regexp.Compile(patternStr)
	}

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "sed2: invalid pattern: %v\n", err)
		return 2
	}

	hadError := false

	// If no files specified, default to stdin
	if len(params.Files) == 0 {
		params.Files = []string{"-"}
	}

	for _, file := range params.Files {
		var err error
		if file == "-" {
			err = processSed2Reader(os.Stdin, os.Stdout, pattern, params)
		} else {
			err = processSed2File(file, pattern, params)
		}

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "sed2: %s: %v\n", file, err)
			hadError = true
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func processSed2File(filename string, pattern *regexp.Regexp, params *Sed2Params) error {
	if params.InPlace {
		// Read entire file into memory
		file, err := os.Open(filename)
		if err != nil {
			return err
		}

		var output strings.Builder
		err = processSed2Reader(file, &output, pattern, params)
		closeErr := file.Close()

		if err != nil {
			return err
		}
		if closeErr != nil {
			return fmt.Errorf("error closing file: %v", closeErr)
		}

		// Write back to file
		err = os.WriteFile(filename, []byte(output.String()), 0644)
		if err != nil {
			return fmt.Errorf("error writing file: %v", err)
		}

		return nil
	} else {
		// Just read and output
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "sed2: error closing file %s: %v\n", filename, err)
			}
		}(file)

		return processSed2Reader(file, os.Stdout, pattern, params)
	}
}

func processSed2Reader(reader io.Reader, writer io.Writer, pattern *regexp.Regexp, params *Sed2Params) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // 10MB max line size

	for scanner.Scan() {
		line := scanner.Text()

		// Perform replacement
		var newLine string
		if params.Global {
			// Replace all occurrences
			newLine = pattern.ReplaceAllString(line, params.To)
		} else {
			// Replace only first occurrence
			newLine = replaceFirst(line, pattern, params.To)
		}

		_, err := fmt.Fprintln(writer, newLine)
		if err != nil {
			return err
		}
	}

	return scanner.Err()
}

func replaceFirst(line string, pattern *regexp.Regexp, replacement string) string {
	loc := pattern.FindStringIndex(line)
	if loc == nil {
		return line
	}

	// Get the matched string to support $1, $2, etc. in replacement
	match := line[loc[0]:loc[1]]

	// Build replacement with support for capture groups
	expandedReplacement := pattern.ReplaceAllString(match, replacement)

	return line[:loc[0]] + expandedReplacement + line[loc[1]:]
}
