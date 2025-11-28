package head

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Files   []string `pos:"true" optional:"true" help:"Files to head. If none specified, read from standard input."`
	Lines   int      `short:"n" help:"Output the first N lines, instead of the first 10" default:"10"`
	Quiet   bool     `short:"q" help:"Never output headers giving file names"`
	Verbose bool     `short:"v" help:"Always output headers giving file names"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "head",
		Short:       "Output the first part of files",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if len(params.Files) == 0 {
				params.Files = []string{"-"}
			}

			if params.Lines < 0 {
				params.Lines = 0
			}

			// Header logic:
			// If > 1 file, print header unless Quiet.
			// If Verbose, always print header.
			printHeaders := (len(params.Files) > 1 && !params.Quiet) || params.Verbose

			runHead(params, os.Stdout, os.Stderr, printHeaders)
		},
	}.ToCobra()
}

func runHead(params *Params, stdout, stderr io.Writer, printHeaders bool) {
	for i, file := range params.Files {
		if printHeaders {
			if i > 0 {
				fmt.Fprintln(stdout)
			}
			fmt.Fprintf(stdout, "==> %s <==\n", file)
		}

		if file == "-" {
			headReader(os.Stdin, stdout, stderr, params.Lines)
		} else {
			f, err := os.Open(file)
			if err != nil {
				fmt.Fprintf(stderr, "head: cannot open '%s' for reading: %v\n", file, err)
				continue
			}
			headReader(f, stdout, stderr, params.Lines)
			f.Close()
		}
	}
}

func headReader(r io.Reader, stdout, stderr io.Writer, n int) {
	if n == 0 {
		return
	}

	scanner := bufio.NewScanner(r)
	count := 0
	for scanner.Scan() {
		fmt.Fprintln(stdout, scanner.Text())
		count++
		if count >= n {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(stderr, "head: error reading: %v\n", err)
	}
}
