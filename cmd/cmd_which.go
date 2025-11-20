package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type WhichParams struct {
	Programs []string `pos:"true" help:"Program names to locate."`
}

func WhichCmd() *cobra.Command {
	return boa.CmdT[WhichParams]{
		Use:         "which",
		Short:       "Locate a program in the user's PATH",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *WhichParams, cmd *cobra.Command, args []string) {
			if len(params.Programs) == 0 {
				cmd.Help()
				os.Exit(1)
			}
			os.Exit(runWhich(params, os.Stdout, os.Stderr))
		},
	}.ToCobra()
}

func runWhich(params *WhichParams, stdout, stderr io.Writer) int {
	exitCode := 0
	for _, program := range params.Programs {
		path, err := exec.LookPath(program)
		if err != nil {
			fmt.Fprintf(stderr, "%s not found\n", program)
			exitCode = 1
		} else {
			fmt.Fprintln(stdout, path)
		}
	}
	return exitCode
}
