package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

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
				_ = cmd.Help()
				os.Exit(1)
			}
			os.Exit(runWhich(params, os.Stdout, os.Stderr))
		},
	}.ToCobra()
}

func splitOsList(pathEnv string) []string {
	if runtime.GOOS == "windows" {
		return strings.Split(pathEnv, ";")
	}
	return strings.Split(pathEnv, ":")
}

func runWhich(params *WhichParams, stdout, stderr io.Writer) int {
	exitCode := 0
	for _, program := range params.Programs {
		path, err := exec.LookPath(program)
		if err != nil {
			winPath := checkWindowsLocalExe(program)
			if winPath != "" {
				_, _ = fmt.Fprintln(stdout, winPath)
			} else {
				_, _ = fmt.Fprintf(stderr, "%s not found\n", program)
				exitCode = 1
			}
		} else {
			_, _ = fmt.Fprintln(stdout, path)
		}
	}
	return exitCode
}

func checkWindowsLocalExe(program string) string {
	// if on windows, scan local workdir for the cmd, cmd.bat, cmd.cmd, cmd.exe or similar, case insensitively
	if runtime.GOOS == "windows" {
		path := ""
		extensions := os.Getenv("PATHEXT")
		exts := []string{}
		for _, ext := range splitOsList(extensions) {
			exts = append(exts, ext)
		}
		files, err := os.ReadDir(".")
		if err == nil {
			for _, file := range files {
				name := file.Name()
				lowerName := strings.ToLower(name)
				lowerProgram := strings.ToLower(program)
				if strings.HasPrefix(lowerName, lowerProgram) {
					for _, ext := range exts {
						if strings.HasSuffix(lowerName, strings.ToLower(ext)) {
							absPath, err := os.Getwd()
							if err == nil {
								path = absPath + string(os.PathSeparator) + name
								break
							}
						}
					}
				}
				if path != "" {
					break
				}
			}
		}
		return path
	}
	return ""
}
