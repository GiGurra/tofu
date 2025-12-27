package reverse

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
	Files []string `pos:"true" optional:"true" help:"Files to reverse. If none specified, read from standard input."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "reverse",
		Short:       "Output lines in reverse order",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if len(params.Files) == 0 {
				params.Files = []string{"-"}
			}
			exitCode := Run(params, os.Stdin, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdin io.Reader, stdout, stderr io.Writer) int {
	for _, file := range params.Files {
		var reader io.Reader
		if file == "-" {
			reader = stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				fmt.Fprintf(stderr, "reverse: cannot open '%s' for reading: %v\n", file, err)
				return 1
			}
			defer f.Close()
			reader = f
		}

		if err := reverseLines(reader, stdout); err != nil {
			fmt.Fprintf(stderr, "reverse: error reading: %v\n", err)
			return 1
		}
	}
	return 0
}

func reverseLines(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	for i := len(lines) - 1; i >= 0; i-- {
		fmt.Fprintln(w, lines[i])
	}

	return nil
}
