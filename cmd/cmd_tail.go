package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

type TailParams struct {
	Files   []string `pos:"true" optional:"true" help:"Files to tail. If none specified, read from standard input."`
	Lines   int      `short:"n" help:"Output the last N lines, instead of the last 10" default:"10"`
	Follow  bool     `short:"f" help:"Output appended data as the file grows"`
	Quiet   bool     `short:"q" help:"Never output headers giving file names"`
	Verbose bool     `short:"v" help:"Always output headers giving file names"`
}

func TailCmd() *cobra.Command {
	return boa.CmdT[TailParams]{
		Use:         "tail",
		Short:       "Output the last part of files",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *TailParams, cmd *cobra.Command, args []string) {
			if len(params.Files) == 0 {
				params.Files = []string{"-"}
			}

			// if we have both stdin and regular files, error out
			if len(params.Files) > 1 {
				for _, f := range params.Files {
					if f == "-" {
						fmt.Fprintf(os.Stderr, "tail: cannot mix standard input with files\n")
						return
					}
				}
			}

			if params.Lines < 0 {
				params.Lines = 0
			}

			// Header logic:
			// If > 1 file, print header unless Quiet.
			// If Verbose, always print header.
			printHeaders := (len(params.Files) > 1 && !params.Quiet) || params.Verbose

			if params.Follow && !slices.Contains(params.Files, "-") {
				runTailFollow(params, os.Stdout, os.Stderr, printHeaders)
			} else {
				runTailStatic(params, os.Stdout, os.Stderr, printHeaders)
			}
		},
	}.ToCobra()
}

func runTailStatic(params *TailParams, stdout, stderr io.Writer, printHeaders bool) {
	for i, file := range params.Files {
		if printHeaders {
			if i > 0 {
				fmt.Fprintln(stdout)
			}
			fmt.Fprintf(stdout, "==> %s <==\n", file)
		}

		if file == "-" {
			tailReader(os.Stdin, stdout, stderr, params.Lines)
		} else {
			f, err := os.Open(file)
			if err != nil {
				fmt.Fprintf(stderr, "tail: cannot open '%s' for reading: %v\n", file, err)
				continue
			}
			tailReader(f, stdout, stderr, params.Lines)
			f.Close()
		}
	}
}

func tailReader(r io.Reader, stdout, stderr io.Writer, n int) {
	if n == 0 {
		return
	}

	lines := make([]string, 0, n)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(stderr, "tail: error reading: %v\n", err)
	}

	for _, line := range lines {
		fmt.Fprintln(stdout, line)
	}
}

func runTailFollow(params *TailParams, stdout, stderr io.Writer, printHeaders bool) {
	type fileState struct {
		name string
		f    *os.File
	}

	states := []*fileState{}
	defer func() {
		for _, s := range states {
			if s.f != nil {
				s.f.Close()
			}
		}
	}()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(stderr, "tail: failed to initialize watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	for i, filename := range params.Files {
		if printHeaders {
			if i > 0 {
				fmt.Fprintln(stdout)
			}
			fmt.Fprintf(stdout, "==> %s <==\n", filename)
		}

		if filename == "-" {
			fmt.Fprintf(stderr, "tail: following stdin is not fully supported in this version when mixed with other files\n")
			continue
		}

		f, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(stderr, "tail: cannot open '%s' for reading: %v\n", filename, err)
			continue
		}

		// Read last N lines
		tailReader(f, stdout, stderr, params.Lines)

		// Add to watcher
		err = watcher.Add(filename)
		if err != nil {
			fmt.Fprintf(stderr, "tail: error watching '%s': %v\n", filename, err)
			f.Close()
			continue
		}

		states = append(states, &fileState{name: filename, f: f})
	}

	lastPrintedFile := ""

	// Watch loop
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				for _, s := range states {
					if s.name == event.Name {
						if printHeaders && lastPrintedFile != s.name {
							fmt.Fprintf(stdout, "\n==> %s <==\n", s.name)
							lastPrintedFile = s.name
						}

						// Copy from current offset to EOF
						_, err := io.Copy(stdout, s.f)
						if err != nil {
							fmt.Fprintf(stderr, "tail: error reading from '%s': %v\n", s.name, err)
						}
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(stderr, "tail: watcher error: %v\n", err)
		}
	}
}
