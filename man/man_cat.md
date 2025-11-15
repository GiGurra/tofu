# `tofu find`

Finds items on the filesystem with a modern and user-friendly interface.

## Interface

```
> tofu cat --help
Concatenate files to standard output

Usage:
  tofu cat [files] [flags]

Flags:
  -A, --show-all            Equivalent to -vET (show all non-printing chars, ends, and tabs). (default false)
  -b, --number-nonblank     Number non-empty output lines, overrides -n. (default false)
  -E, --show-ends           Display $ at end of each line. (default false)
  -n, --number              Number all output lines. (default false)
  -s, --squeeze-blank       Suppress repeated empty output lines. (default false)
  -T, --show-tabs           Display TAB characters as ^I. (default false)
  -v, --show-non-printing   Use ^ and M- notation for non-printing characters (except LFD and TAB). (default false)
  -h, --help                help for cat
```

### Examples

```
> tofu cat main.go
package main

import (
        "runtime/debug"

        "github.com/GiGurra/boa/pkg/boa"
        "github.com/gigurra/tofu/cmd"
        "github.com/spf13/cobra"
)

func main() {
        boa.CmdT[boa.NoParams]{
                Use:     "tofu",
                Short:   "Tools Of Fantastic Utility",
                Version: appVersion(),
                SubCmds: []*cobra.Command{
                        cmd.CatCmd(),
                        cmd.FindCmd(),
                        cmd.GrepCmd(),
                },
        }.Run()
}

func appVersion() string {
        bi, hasBuilInfo := debug.ReadBuildInfo()
        if !hasBuilInfo {
                return "unknown-(no build info)"
        }

        versionString := bi.Main.Version
        if versionString == "" {
                versionString = "unknown-(no version)"
        }

        return versionString
}
```
