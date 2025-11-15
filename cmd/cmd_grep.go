package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
)

type PatternType string

const (
	PatternTypeBasic    PatternType = "basic"
	PatternTypeExtended PatternType = "extended"
	PatternTypeFixed    PatternType = "fixed"
	PatternTypePerl     PatternType = "perl"
)

type GrepParams struct {
	Pattern     string      `pos:"true" help:"Pattern to search for in files."`
	Files       []string    `pos:"true" optional:"true" help:"Files or directories to search. If none specified, reads from standard input." default:"-"`
	PatternType PatternType `short:"t" help:"Type of pattern matching (basic,extended,fixed,perl)." default:"extended" alts:"basic,extended,fixed,perl"`
	IgnoreCase  bool        `short:"i" help:"Perform case-insensitive matching." default:"false"`
	InvertMatch bool        `short:"v" help:"Select non-matching lines." default:"false"`
	WordRegexp  bool        `short:"w" help:"Match only whole words." default:"false"`
	LineRegexp  bool        `short:"x" help:"Match only whole lines." default:"false"`

	// Output control
	LineNumber        bool `short:"n" help:"Print line number with output lines." default:"false"`
	WithFilename      bool `short:"H" help:"Print filename with output lines." default:"false"`
	NoFilename        bool `help:"Suppress filename prefix on output." default:"false"`
	Count             bool `short:"c" help:"Print only a count of matching lines per file." default:"false"`
	FilesWithMatch    bool `short:"l" help:"Print only names of files with matches." default:"false"`
	FilesWithoutMatch bool `short:"L" help:"Print only names of files without matches." default:"false"`
	OnlyMatching      bool `short:"o" help:"Show only the matched parts of lines." default:"false"`
	Quiet             bool `short:"q" help:"Suppress all normal output." default:"false"`
	IgnoreBinary      bool `help:"Suppress output for binary files." default:"false"`
	MaxCount          int  `short:"m" help:"Stop after NUM matches per file." default:"0"`

	// Context control
	BeforeContext int `short:"B" help:"Print NUM lines of leading context." default:"0"`
	AfterContext  int `short:"A" help:"Print NUM lines of trailing context." default:"0"`
	Context       int `short:"C" help:"Print NUM lines of output context." default:"0"`

	// File/directory handling
	Recursive  bool     `short:"r" help:"Search directories recursively." default:"false"`
	Include    []string `short:"I" optional:"true" help:"Search only files matching pattern (glob)."`
	Exclude    []string `short:"e" optional:"true" help:"Skip files matching pattern (glob)."`
	ExcludeDir []string `optional:"true" help:"Skip directories matching pattern (glob)."`

	// Misc
	NoMessages bool `short:"s" help:"Suppress error messages." default:"false"`
}

func GrepCmd() *cobra.Command {
	return boa.CmdT[GrepParams]{
		Use:         "grep",
		Short:       "Search for patterns in files",
		ParamEnrich: defaultParamEnricher(),
		PreExecuteFunc: func(params *GrepParams, cmd *cobra.Command, args []string) error {
			if params.Pattern == "" {
				return fmt.Errorf("pattern cannot be empty")
			}

			// Set context values if -C is specified
			if params.Context > 0 {
				if params.BeforeContext == 0 {
					params.BeforeContext = params.Context
				}
				if params.AfterContext == 0 {
					params.AfterContext = params.Context
				}
			}

			// Validate mutually exclusive flags
			exclusiveCount := 0
			if params.Count {
				exclusiveCount++
			}
			if params.FilesWithMatch {
				exclusiveCount++
			}
			if params.FilesWithoutMatch {
				exclusiveCount++
			}
			if params.OnlyMatching {
				exclusiveCount++
			}
			if exclusiveCount > 1 {
				return fmt.Errorf("flags -c, -l, -L, and -o are mutually exclusive")
			}

			return nil
		},
		RunFunc: func(params *GrepParams, cmd *cobra.Command, args []string) {
			exitCode := runGrep(params)
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}.ToCobra()
}

func runGrep(params *GrepParams) int {
	// Compile the pattern
	pattern, err := compilePattern(params)
	if err != nil {
		if !params.NoMessages {
			_, _ = fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		}
		return 2
	}

	found := false
	hadError := false

	// If recursive, search directory tree
	if params.Recursive {
		startDir := "."
		if len(params.Files) > 0 && params.Files[0] != "-" {
			startDir = params.Files[0]
		}

		err := filepath.WalkDir(startDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				if !params.NoMessages {
					_, _ = fmt.Fprintf(os.Stderr, "grep: %s: %v\n", path, err)
				}
				hadError = true
				return nil
			}

			// Skip directories
			if d.IsDir() {
				// Check if directory should be excluded
				if shouldExcludeDir(d.Name(), params.ExcludeDir) {
					return filepath.SkipDir
				}
				return nil
			}

			// Check include/exclude patterns
			if !shouldSearchFile(path, params.Include, params.Exclude) {
				return nil
			}

			matched, err := grepFile(path, pattern, params, len(params.Files) > 1 || params.Recursive)
			if err != nil {
				if !params.NoMessages {
					_, _ = fmt.Fprintf(os.Stderr, "grep: %s: %v\n", path, err)
				}
				hadError = true
				return nil
			}
			if matched {
				found = true
			}
			return nil
		})

		if err != nil && !params.NoMessages {
			_, _ = fmt.Fprintf(os.Stderr, "grep: %v\n", err)
			hadError = true
		}
	} else {
		// Process specified params.Files
		for _, file := range params.Files {
			var matched bool
			var err error

			if file == "-" {
				matched, err = grepReader(os.Stdin, "<stdin>", pattern, params, len(params.Files) > 1)
			} else {
				matched, err = grepFile(file, pattern, params, len(params.Files) > 1)
			}

			if err != nil {
				if !params.NoMessages {
					_, _ = fmt.Fprintf(os.Stderr, "grep: %s: %v\n", file, err)
				}
				hadError = true
				continue
			}
			if matched {
				found = true
			}
		}
	}

	if hadError && !params.Quiet {
		return 2
	}
	if found {
		return 0
	}
	return 1
}

func compilePattern(params *GrepParams) (*regexp.Regexp, error) {
	pattern := params.Pattern

	// Handle different pattern types
	switch params.PatternType {
	case PatternTypeFixed:
		pattern = regexp.QuoteMeta(pattern)
	case PatternTypeBasic:
		// Convert basic regex to extended (simplified conversion)
		pattern = convertBasicToExtended(pattern)
	case PatternTypePerl, PatternTypeExtended:
		// Use as-is
	}

	// Word boundary matching
	if params.WordRegexp {
		pattern = `\b` + pattern + `\b`
	}

	// Line matching
	if params.LineRegexp {
		pattern = `^` + pattern + `$`
	}

	// Case insensitive
	if params.IgnoreCase {
		pattern = `(?i)` + pattern
	}

	return regexp.Compile(pattern)
}

func convertBasicToExtended(pattern string) string {
	// Simplified basic to extended regex conversion
	// In basic regex, +?{}() need escaping to be special
	// In extended regex (Go default), they are special by default
	result := strings.ReplaceAll(pattern, `\+`, "+")
	result = strings.ReplaceAll(result, `\?`, "?")
	result = strings.ReplaceAll(result, `\{`, "{")
	result = strings.ReplaceAll(result, `\}`, "}")
	result = strings.ReplaceAll(result, `\(`, "(")
	result = strings.ReplaceAll(result, `\)`, ")")
	return result
}

func grepFile(filename string, pattern *regexp.Regexp, params *GrepParams, showFilename bool) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "grep: error closing file %s: %v\n", filename, err)
		}
	}(file)

	// TODO: Check if file is binary and handle accordingly (skip or process)
	if isBinary, err := isFileBinary(file); err == nil && isBinary {
		// Skip binary files, log if needed
		if !params.IgnoreBinary {
			_, _ = fmt.Fprintf(os.Stderr, "grep: %s: binary file skipped\n", filename)
		}
		return false, nil
	}

	res, err := grepReader(file, filename, pattern, params, showFilename)
	if err != nil {
		return false, fmt.Errorf("error reading file %s: %v", filename, err)
	}

	return res, nil
}

func isFileBinary(file *os.File) (bool, error) {
	const sampleSize = 8000
	buf := make([]byte, sampleSize)
	// Check if contains a NUL byte that is not at the end, indicating binary content
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}
	// Reset file pointer
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return false, err
	}
	return false, nil
}

func grepReader(reader io.Reader, filename string, pattern *regexp.Regexp, params *GrepParams, showFilename bool) (bool, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // 10MB max line size
	lineNum := 0
	matchCount := 0
	found := false

	// Context tracking
	var contextBefore []string
	var contextAfter int
	printedContext := false

	// Override filename display based on flags
	if params.NoFilename {
		showFilename = false
	}
	if params.WithFilename {
		showFilename = true
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matches := pattern.MatchString(line)

		// Invert match if requested
		if params.InvertMatch {
			matches = !matches
		}

		if matches {
			found = true
			matchCount++

			// Handle different output modes
			if params.Quiet {
				return true, nil
			}

			if params.FilesWithMatch {
				fmt.Println(filename)
				return true, nil
			}

			if params.Count {
				// Will be printed after loop
				if params.MaxCount > 0 && matchCount >= params.MaxCount {
					break
				}
				continue
			}

			// Print context before
			if params.BeforeContext > 0 && !printedContext {
				for _, ctxLine := range contextBefore {
					printLine(filename, 0, ctxLine, showFilename, false, false, nil, params)
				}
			}

			// Print matching line
			printLine(filename, lineNum, line, showFilename, params.LineNumber, params.OnlyMatching, pattern, params)
			printedContext = true

			// Set up context after
			contextAfter = params.AfterContext

			if params.MaxCount > 0 && matchCount >= params.MaxCount {
				break
			}
		} else {
			// Handle context after previous match
			if contextAfter > 0 {
				printLine(filename, lineNum, line, showFilename, false, false, nil, params)
				contextAfter--
				if contextAfter == 0 {
					printedContext = false
				}
			}

			// Track context before for next potential match
			if params.BeforeContext > 0 {
				contextBefore = append(contextBefore, line)
				if len(contextBefore) > params.BeforeContext {
					contextBefore = contextBefore[1:]
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return found, err
	}

	// Print count if requested
	if params.Count && !params.Quiet {
		if showFilename {
			fmt.Printf("%s:%d\n", filename, matchCount)
		} else {
			fmt.Println(matchCount)
		}
	}

	// Handle params.Files without match
	if !found && params.FilesWithoutMatch && !params.Quiet {
		fmt.Println(filename)
	}

	return found, nil
}

func printLine(filename string, lineNum int, line string, showFilename, showLineNum, onlyMatching bool, pattern *regexp.Regexp, params *GrepParams) {
	if params.Quiet {
		return
	}

	var output strings.Builder

	if showFilename {
		output.WriteString(filename)
		output.WriteString(":")
	}

	if showLineNum && lineNum > 0 {
		output.WriteString(fmt.Sprintf("%d:", lineNum))
	}

	if onlyMatching && pattern != nil {
		match := pattern.FindString(line)
		output.WriteString(colorRed)
		output.WriteString(match)
		output.WriteString(colorReset)
	} else {
		// Highlight matches in the line
		if pattern != nil && !params.InvertMatch {
			highlightedLine := highlightMatches(line, pattern)
			output.WriteString(highlightedLine)
		} else {
			output.WriteString(line)
		}
	}

	fmt.Println(output.String())
}

func highlightMatches(line string, pattern *regexp.Regexp) string {
	// Find all matches
	matches := pattern.FindAllStringIndex(line, -1)
	if len(matches) == 0 {
		return line
	}

	var result strings.Builder
	lastIndex := 0

	for _, match := range matches {
		start, end := match[0], match[1]
		// Add text before match
		result.WriteString(line[lastIndex:start])
		// Add highlighted match
		result.WriteString(colorRed)
		result.WriteString(line[start:end])
		result.WriteString(colorReset)
		lastIndex = end
	}

	// Add remaining text after last match
	result.WriteString(line[lastIndex:])

	return result.String()
}

func shouldSearchFile(filename string, include, exclude []string) bool {
	basename := filepath.Base(filename)

	// Check include patterns
	if len(include) > 0 {
		matched := false
		for _, pattern := range include {
			if match, _ := filepath.Match(pattern, basename); match {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check exclude patterns
	for _, pattern := range exclude {
		if match, _ := filepath.Match(pattern, basename); match {
			return false
		}
	}

	return true
}

func shouldExcludeDir(dirname string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if match, _ := filepath.Match(pattern, dirname); match {
			return true
		}
	}
	return false
}
