package mkdir

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Dirs    []string `pos:"true" required:"true" help:"Directories to create."`
	Parents bool     `short:"p" optional:"true" help:"Make parent directories as needed, no error if existing."`
	Mode    string   `short:"m" optional:"true" help:"Set file mode (as in chmod), not a=rwx - umask." default:"0755"`
	Verbose bool     `short:"v" optional:"true" help:"Print a message for each created directory."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "mkdir",
		Short:       "Create directories",
		Long:        "Create the DIRECTORY(ies), if they do not already exist.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdout, stderr io.Writer) int {
	mode, err := parseMode(params.Mode)
	if err != nil {
		fmt.Fprintf(stderr, "mkdir: invalid mode '%s': %v\n", params.Mode, err)
		return 1
	}

	hadError := false
	for _, dir := range params.Dirs {
		if err := makeDir(dir, mode, params, stdout); err != nil {
			fmt.Fprintf(stderr, "mkdir: %v\n", err)
			hadError = true
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func parseMode(s string) (os.FileMode, error) {
	val, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(val), nil
}

func makeDir(path string, mode os.FileMode, params *Params, stdout io.Writer) error {
	var err error
	if params.Parents {
		err = os.MkdirAll(path, mode)
	} else {
		err = os.Mkdir(path, mode)
	}

	if err != nil {
		return fmt.Errorf("cannot create directory '%s': %v", path, err)
	}

	if params.Verbose {
		fmt.Fprintf(stdout, "mkdir: created directory '%s'\n", path)
	}

	return nil
}
