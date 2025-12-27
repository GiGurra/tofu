package touch

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Files      []string `pos:"true" required:"true" help:"Files to touch."`
	NoCreate   bool     `short:"c" optional:"true" help:"Do not create any files."`
	AccessOnly bool     `short:"a" optional:"true" help:"Change only the access time."`
	ModifyOnly bool     `short:"m" optional:"true" help:"Change only the modification time."`
	Reference  string   `short:"r" optional:"true" help:"Use this file's times instead of current time."`
	Date       string   `short:"d" optional:"true" help:"Parse date string and use it instead of current time."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "touch",
		Short:       "Change file access and modification times",
		Long:        "Update the access and modification times of each FILE to the current time. A FILE argument that does not exist is created empty, unless -c is supplied.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdout, stderr io.Writer) int {
	// Determine the time to use
	touchTime := time.Now()

	if params.Reference != "" {
		info, err := os.Stat(params.Reference)
		if err != nil {
			fmt.Fprintf(stderr, "touch: cannot stat '%s': %v\n", params.Reference, err)
			return 1
		}
		touchTime = info.ModTime()
	} else if params.Date != "" {
		parsed, err := parseDate(params.Date)
		if err != nil {
			fmt.Fprintf(stderr, "touch: invalid date format '%s': %v\n", params.Date, err)
			return 1
		}
		touchTime = parsed
	}

	hadError := false
	for _, file := range params.Files {
		if err := touchFile(file, touchTime, params); err != nil {
			fmt.Fprintf(stderr, "touch: %v\n", err)
			hadError = true
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func touchFile(path string, t time.Time, params *Params) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		if params.NoCreate {
			return nil // silently skip
		}
		// Create the file
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("cannot touch '%s': %v", path, err)
		}
		f.Close()
		// Now set the times
		return os.Chtimes(path, t, t)
	} else if err != nil {
		return fmt.Errorf("cannot stat '%s': %v", path, err)
	}

	// File exists, update times
	atime := t
	mtime := t

	if params.AccessOnly && !params.ModifyOnly {
		mtime = info.ModTime()
	} else if params.ModifyOnly && !params.AccessOnly {
		atime = getAtime(info)
	}

	return os.Chtimes(path, atime, mtime)
}

func parseDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
		"Jan 2, 2006",
		"Jan 2 2006",
		"02 Jan 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date")
}
